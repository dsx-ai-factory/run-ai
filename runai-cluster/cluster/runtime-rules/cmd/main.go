/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"flag"
	"os"
	"sort"
	"time"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions/distributed"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	pkg "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/kubeConfig"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/logger"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/validation"
	workloadcontroller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/controller"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/maps"
	coreV1 "k8s.io/api/core/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

type Options struct {
	logLevel                    int
	startPromClient             bool
	promServiceUrl              string
	frameworkPollingIntervalSec int
	enableLeaderElection        bool
	disableWorkloadRunnerRules  bool
}

var (
	options = &Options{}
)

func init() {
	InitArgs(options)
	logger.InitLogger(options.logLevel)
}

const (
	RuntimeRules = "runtime-rules"
)

func main() {
	scheme := initSchema()

	// init the kube client
	watchClient, err := kubeConfig.InitWatchClient(scheme)
	if err != nil {
		zap.S().Errorf("Failed creating watchClient: %+v", err)
		os.Exit(1)
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   options.enableLeaderElection,
		LeaderElectionID: "runtime-rules.run.ai",
	})
	if err != nil {
		zap.S().Error(err, "Failed to configure manager")
		os.Exit(1)
	}
	managerContext, cancel := context.WithCancel(ctrl.SetupSignalHandler())

	go func() {
		<-mgr.Elected()
		zap.S().Info("Leader elected")
		runRuntimeRules(watchClient, scheme, managerContext, cancel)
	}()

	if err := mgr.Start(managerContext); err != nil {
		zap.S().Errorf("Manager stopped: %v", err)
		os.Exit(1)
	}
}

func runRuntimeRules(watchClient client.WithWatch, scheme *apiruntime.Scheme, managerContext context.Context, cancel context.CancelFunc) {
	// init the queues manager
	qm := &queuesManager.QueuesManager{}

	// init the prometheus client
	promClient := prometheus.GetQuerierClient(options.startPromClient, options.promServiceUrl)

	// init recorder for k8s events
	recorder, err := InitRecorder(scheme)
	if err != nil {
		zap.S().Errorf("Failed creating watchClient: %+v", err)
		os.Exit(2)
	}

	// get event list
	eventList := util.GetEnvOrDefaultIntArray(util.EventList, util.EventListDefault)
	if !CheckEventList(eventList) {
		zap.S().Errorf("Failed validating eventList: %+v", eventList)
		os.Exit(3)
	}
	eventList = append(eventList, queuesManager.UnknownEvent, queuesManager.DeactivateEvent)
	sort.Ints(eventList)
	zap.S().Infof("event list for runtime-rules: %v", eventList)

	jobCRDsFrameworksList := maps.Keys(distributed.SupportedExtensions)
	extClient := distributed.NewClient(watchClient)

	extensionGroupStatus, _ := extClient.List(context.Background(), jobCRDsFrameworksList)
	// the current status of the supported frameworks
	currentDistributedFrameworksStatus, err := workloadcontroller.GetDistributedJobFrameworksStatus(extensionGroupStatus)
	if err != nil {
		zap.S().Errorf("unable to get the status of distributed job frameworks. got error: %s. exiting", err.Error())
		os.Exit(4)
	}
	startDistributedFrameworksStatusPoll(managerContext, cancel, extensionGroupStatus, extClient)

	distributedValidator := validation.NewDistributedValidator(currentDistributedFrameworksStatus)
	workloadValidator := validation.NewWorkloadValidator(distributedValidator)

	// init & start the runtime-rules
	runtimeRules := pkg.NewRuntimeRules(watchClient, qm, promClient, recorder, eventList, workloadValidator)
	runtimeRules.Start(managerContext)
}

func CheckEventList(eventList []int) bool {
	for index := range eventList {
		if eventList[index] >= queuesManager.DeactivateEvent {
			zap.S().Errorf("event list has value higher then 100, event list: %v, value: %d", eventList, eventList[index])
			return false
		}
		if eventList[index] <= queuesManager.UnknownEvent {
			zap.S().Errorf("event list has value lower then 0, event list: %v, value: %d", eventList, eventList[index])
			return false
		}
	}
	return true
}

func InitRecorder(scheme *apiruntime.Scheme) (record.EventRecorder, error) {
	kubeClient, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		return nil, err
	}
	broadcastRecorder := record.NewBroadcaster()
	broadcastRecorder.StartLogging(zap.S().Infof)
	broadcastRecorder.StartRecordingToSink(&v1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	return broadcastRecorder.NewRecorder(scheme, coreV1.EventSource{Component: RuntimeRules}), nil
}

func InitArgs(options *Options) {
	flag.IntVar(&options.logLevel, "log-level", int(zapcore.DebugLevel), "Log level to use, -1 is Debug, 0 is Info and so forth")
	flag.BoolVar(&options.startPromClient, "start-prom-client", true,
		"Defines rather prom client should be started or not.")
	flag.StringVar(&options.promServiceUrl, "prom-service-url", prometheus.DefaultPromServiceUrl,
		"Prometheus service url to use")
	flag.IntVar(&options.frameworkPollingIntervalSec, "framework-polling-interval-sec", 60, "Polling interval in seconds for distributed framework config changes")
	flag.BoolVar(&options.enableLeaderElection, "leader-elect", false,
		"Enable leader election for runtime-rules. "+
			"Enabling this will ensure there is only one active runtime-rules.")
	flag.BoolVar(&options.disableWorkloadRunnerRules, "disable-workload-runner-rules", false,
		"Deprecated compatibility flag. WorkloadRunner observation is always enabled.")
	flag.Parse()
}

func initSchema() *apiruntime.Scheme {
	scheme := schemeinit.Scheme()
	err := dromosv1alpha1.AddToScheme(scheme)
	if err != nil {
		zap.S().Error("failed to add run.ai/v1alpha1 (WorkloadRunner) to scheme", err)
	}
	err = kaiorg.AddToScheme(scheme)
	if err != nil {
		zap.S().Error("failed to add kai.resources (Project) to scheme", err)
	}
	return scheme
}

func startDistributedFrameworksStatusPoll(withCancel context.Context, cancel context.CancelFunc, currentStateExtensions extensions.Group, extClient extensions.Client) {
	distributedStatusPoll := distributed.NewStatusChangedPoll(currentStateExtensions, extClient)
	go func() {
		defer cancel() // reset the runtime-rules if the poll exits

		err := wait.PollUntilContextCancel(withCancel, time.Duration(options.frameworkPollingIntervalSec)*time.Second, false, distributedStatusPoll.Poll)
		if err != nil {
			zap.S().Errorf("got error while polling for distributed job status: %s. exiting", err.Error())
		} else {
			zap.S().Info("found distributed job availability change, exiting")
		}
	}()
}
