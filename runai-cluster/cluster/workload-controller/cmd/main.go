/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	stdruntime "runtime"
	"time"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"

	"k8s.io/client-go/tools/record"

	"github.com/grafana/pyroscope-go"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions/distributed"
	rjv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/migration"
	webhookUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/webhook"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/managers"
	factory2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/policy"
	webhookutils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/webhook"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/mutators"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	workloadspolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads/policy"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	mpiv1alpha2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v1alpha2"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"

	distp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"

	optionsutils "github.com/run-ai/runai/runai-cluster/cluster/common/utils/options"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"

	//+kubebuilder:scaffold:imports

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	scheme  = runtime.NewScheme()
	log     = ctrl.Log.WithName("setup.controller")
	options = &Options{}
)

type restMappingResolver interface {
	RESTMapping(schema.GroupKind, ...string) (*meta.RESTMapping, error)
}

type apiAvailability struct {
	interactiveWorkload bool
	interactivePolicy   bool
	trainingWorkload    bool
	trainingPolicy      bool
	distributedWorkload bool
	distributedPolicy   bool
}

func resolveAPIAvailability(resolver restMappingResolver) (apiAvailability, error) {
	availability := apiAvailability{}
	candidates := []struct {
		gvk       schema.GroupVersionKind
		available *bool
	}{
		{iwv1.GroupVersion.WithKind("InteractiveWorkload"), &availability.interactiveWorkload},
		{ipv1.GroupVersion.WithKind("InteractivePolicy"), &availability.interactivePolicy},
		{twv1.GroupVersion.WithKind("TrainingWorkload"), &availability.trainingWorkload},
		{tpv1.GroupVersion.WithKind("TrainingPolicy"), &availability.trainingPolicy},
		{distw.GroupVersion.WithKind("DistributedWorkload"), &availability.distributedWorkload},
		{distp.GroupVersion.WithKind("DistributedPolicy"), &availability.distributedPolicy},
	}

	var resolutionErr error
	for _, candidate := range candidates {
		_, err := resolver.RESTMapping(candidate.gvk.GroupKind(), candidate.gvk.Version)
		if meta.IsNoMatchError(err) {
			log.Info("Legacy API is not served; dependent paths will be skipped", "gvk", candidate.gvk.String())
			continue
		}
		if err != nil {
			resolutionErr = errors.Join(resolutionErr, fmt.Errorf("resolve %s: %w", candidate.gvk, err))
			continue
		}
		*candidate.available = true
	}
	return availability, resolutionErr
}

const (
	interactiveType = "interactive"
	trainingType    = "training"
	distributedType = "distributed"

	interactivePolicyType          = "interactivepolicy"
	trainingPolicyType             = "trainingpolicy"
	distributedPolicyType          = "distributedpolicy"
	validationPolicyWebhooksPrefix = "validate-run-ai-v2alpha1"
	requestTimeoutMigrationWebhook = 10 * time.Second
	eventRecorderSource            = "workload-controller"
)

type Options struct {
	certDir                            string
	enableWorkloadOwnershipProtection  bool
	interactiveMaxConcurrentReconciles int
	trainingMaxConcurrentReconciles    int
	distributedMaxConcurrentReconciles int
	disableMetric                      bool
	metricExporterBindAddr             string
	profiler                           bool
	profilerServerAddress              string
	mutexProfilerRate                  int
	blockProfilerRate                  int
	healthProbeBindAddress             string
	degradedPhaseDelay                 int
	pollingIntervalSec                 int
	forceExitTimeSec                   int
	webhookPort                        int
	skipMigration                      bool
	migrationMaxWorkers                int
	webhookExternalAddress             string
	EnableLeaderElection               bool
	additionalPodLabels                string
	additionalPodAnnotations           string
	failureResourceCleanupPolicy       string
	mnnvlSupported                     bool
	topologyEnabled                    bool
	elasticCDEnabled                   bool
	enableDanglingResourceCleanup      bool
	allowedServiceAccounts             optionsutils.ArrayFlags
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(ipv1.AddToScheme(scheme))    // interactive policy
	utilruntime.Must(iwv1.AddToScheme(scheme))    // interactive workload
	utilruntime.Must(tpv1.AddToScheme(scheme))    // training policy
	utilruntime.Must(twv1.AddToScheme(scheme))    // training workload
	utilruntime.Must(distp.AddToScheme(scheme))   // distributed policy
	utilruntime.Must(distw.AddToScheme(scheme))   // distributed workload
	utilruntime.Must(rjv1.AddToScheme(scheme))    // runai job
	utilruntime.Must(runaiv2.AddToScheme(scheme)) // runai project
	utilruntime.Must(kaiorg.AddToScheme(scheme))  // kai.resources project/department
	utilruntime.Must(schedv2.AddToScheme(scheme)) // runai department
	utilruntime.Must(mpiv1alpha2.AddKnownTypes(scheme))
	utilruntime.Must(mpiv1alpha2.AddToScheme(scheme)) // mpi job
	utilruntime.Must(mpiv2beta1.AddToScheme(scheme))  // mpi job
	utilruntime.Must(kubeflowv1.AddToScheme(scheme))  // pytorchjob, tfjob, xgboostjob
	utilruntime.Must(v1alpha1.AddToScheme(scheme))    // nodepools and accessrules
	utilruntime.Must(servingv1.AddToScheme(scheme))   // knativeservice
	utilruntime.Must(openshiftV1.Install(scheme))
	utilruntime.Must(nvapi.AddToScheme(scheme)) // ComputeDomain
	utilruntime.Must(gatewayv1.Install(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	flag.StringVar(&options.certDir, "cert-dir", "/etc/webhook/certs", "Directory that contains the server key and certificate for the webhooks")
	flag.BoolVar(&options.enableWorkloadOwnershipProtection, "enable-workloads-ownership-protection", false, "Should enable workloads ownership protection")
	flag.IntVar(&options.interactiveMaxConcurrentReconciles, "interactive-max-concurrent-reconciles", 10, "Max concurrent reconciles for interactive workloads")
	flag.IntVar(&options.trainingMaxConcurrentReconciles, "training-max-concurrent-reconciles", 10, "Max concurrent reconciles for training workloads")
	flag.BoolVar(&options.disableMetric, "disable-metrics", false, "Disable metric exporter")
	flag.BoolVar(&options.profiler, "enable-profiler", false, "Enable profiler")
	flag.StringVar(&options.profilerServerAddress, "profiler-server-address", "localhost:4040", "Profiler server address")
	flag.StringVar(&options.metricExporterBindAddr, "metrics-bind-address", ":8383", "Metric bind address for metric exporter")
	flag.IntVar(&options.mutexProfilerRate, "mutex-profiler-rate", 5, "Mutex profiler rate")
	flag.IntVar(&options.blockProfilerRate, "block-profiler-rate", 5, "Block profiler rate")
	flag.StringVar(&options.healthProbeBindAddress, "health-probe-bind-address", ":8081",
		"Probe bind address for health checks")
	flag.IntVar(&options.degradedPhaseDelay, "degraded-phase-delay", 60,
		"Workload status delay from running to degraded in seconds")
	flag.IntVar(&options.pollingIntervalSec, "polling-interval-sec", 30, "Polling interval in seconds")
	flag.IntVar(&options.forceExitTimeSec, "force-exit-time-sec", 10, "Force exit time in seconds")
	flag.IntVar(&options.webhookPort, "webhook-port", 9443, "Webhook port")
	flag.IntVar(&options.distributedMaxConcurrentReconciles, "distributed-max-concurrent-reconciles", 10, "Max concurrent reconciles for distributed workloads")
	flag.BoolVar(&options.skipMigration, "skip-migration", false, "Skip the entre migration flow without running it")
	flag.IntVar(&options.migrationMaxWorkers, "migration-max-workers", 50, "Max concurrent migration workers")
	flag.StringVar(&options.webhookExternalAddress, "webhook-external-address", "workload-controller.runai.svc:443", "Webhook external address")
	flag.BoolVar(&options.EnableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&options.additionalPodLabels, "additional-pod-labels", "", "Additional pod labels added to pods of all workloads")
	flag.StringVar(&options.additionalPodAnnotations, "additional-pod-annotations", "", "Additional pod annotations added to pods of all workloads")
	flag.StringVar(&options.failureResourceCleanupPolicy, "failure-cleanup-policy", managers.GetDefaultFailureCleanupPolicy().String(), "Policy for handling sub-resources cleanup in workload failure ('All' - default, 'None', 'KeepFailing')")
	flag.BoolVar(&options.mnnvlSupported, "enable-mnnvl-support", false, "Enable support for the MNNVL feature")
	flag.Var(&options.allowedServiceAccounts, "allowed-service-accounts", "Comma-separated list of service account patterns to allow in the webhook validation")
	flag.BoolVar(&options.topologyEnabled, "enable-topology-support", false, "Enable support for the scheduling topology feature")
	flag.BoolVar(&options.elasticCDEnabled, "enable-elastic-compute-domain", false, "Enable support for the scheduling topology feature")
	flag.BoolVar(&options.enableDanglingResourceCleanup, "enable-dangling-resource-cleanup", false, "Enable automatic cleanup of dangling Kubernetes resources left over from workload networking mode transitions")

	//
	//   initialize logger parameters
	//
	opts := zap.Options{
		Development:     true,
		StacktraceLevel: zapcore.LevelEnabler(zapcore.FatalLevel),
		TimeEncoder:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	if options.profiler {
		err := enableProfiler(options.profilerServerAddress, options.mutexProfilerRate, options.blockProfilerRate, nil)
		if err != nil {
			log.Error(err, "Failed to start profiler")
			os.Exit(1)
		}
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	//
	// set throttling parameters of the k8s client, allowing to increase them if necessary
	//
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = utils.GetK8sClientQPS()
	cfg.Burst = utils.GetK8sClientBurst()

	featureGateUtil := featuregate.GetFeatureGateUtilInstance()
	featureGateUtil.Initialize(cfg)

	log.Info("client throttling applied", "QPS", cfg.QPS, "burst", cfg.Burst)
	log.Info("max concurrent reconciles",
		"interactive", options.interactiveMaxConcurrentReconciles,
		"training", options.trainingMaxConcurrentReconciles)
	log.Info("topology enabled", "state", options.topologyEnabled)
	log.Info("elastic compute domain", "state", options.elasticCDEnabled)

	clusterVersion := os.Getenv(workloadcommon.EnvVarClusterVersion)
	if len(clusterVersion) == 0 {
		err := fmt.Errorf("env var %s is not set", workloadcommon.EnvVarClusterVersion)
		log.Error(err, "could not start controller")
		os.Exit(1)
	}

	// checks if the migration should be skipped
	if !options.skipMigration {
		err := runMigration(cfg, clusterVersion, options.migrationMaxWorkers)
		if err != nil {
			log.Error(err, "unable to finish successfully  migration")
			os.Exit(1)
		}
	}
	log.Info("Migration is skipped, by user request")
	// enable metric exporter
	metricExporterBindAddress := options.metricExporterBindAddr
	if options.disableMetric {
		log.Info("metric exporter is disabled")
		metricExporterBindAddress = "0"
	}
	//
	//   initialize the controller service
	//
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricExporterBindAddress,
		},
		WebhookServer: &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    options.webhookPort,
				CertDir: options.certDir,
			}},
		HealthProbeBindAddress: options.healthProbeBindAddress,
		LeaderElection:         options.EnableLeaderElection,
		LeaderElectionID:       "workload-controller.run.ai",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}
	availability, err := resolveAPIAvailability(mgr.GetRESTMapper())
	if err != nil {
		log.Error(err, "unable to resolve legacy API availability")
		os.Exit(1)
	}

	// when checking load balancer connections, we need to search for "allocation failed" event.
	// to avoid scanning all the events, we do the search by oid, and for this purpose we need
	// to set up an indexer by oid.
	err = jobs.IndexInvolvedOid(mgr)
	if err != nil {
		log.Error(err, fmt.Sprintf("Failed indexing field %s", workloadcommon.InvolvedObjectUidField))
		os.Exit(1)
	}
	configClient, err := client.New(mgr.GetConfig(), client.Options{Scheme: scheme})
	if err != nil {
		log.Error(err, "unable to create config client")
		os.Exit(1)
	}

	jobCRDsFrameworksList := utils.GetMapKeys[extensions.Type](distributed.SupportedExtensions)
	extClient := distributed.NewClient(configClient)
	extensionGroupStatus, err := extClient.List(context.Background(), jobCRDsFrameworksList)
	podAnnotatios := utils.ArgToMap(options.additionalPodAnnotations)
	podLabels := utils.ArgToMap(options.additionalPodLabels)
	shouldUseControlPlanePolicies := policyUtils.SetupControlPlanePoliciesUsage(context.Background(), configClient)

	enabledDistributedJobFramework, err := controller.GetDistributedJobFrameworksStatus(extensionGroupStatus)
	if err != nil {
		log.Error(err, "unable to get enabled distributed job frameworks")
		os.Exit(1)
	}
	if options.enableWorkloadOwnershipProtection {
		enableWorkloadOwnershipProtection(mgr, configClient)
	}

	controllerOptions, err := parseControllerOptions(options)
	if err != nil {
		log.Error(err, "unable to parse controller options")
		os.Exit(1)
	}

	setupControllers(mgr, enabledDistributedJobFramework, podLabels, podAnnotatios, *controllerOptions, availability)

	setupWebhook(mgr, enabledDistributedJobFramework, shouldUseControlPlanePolicies, clusterVersion, options, availability)

	setupHealthCheck(mgr)

	DistributedJobPoll := distributed.NewStatusChangedPoll(extensionGroupStatus, extClient)
	withCancel, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	go func() {
		err = wait.PollUntilContextCancel(withCancel, time.Duration(options.pollingIntervalSec)*time.Second, false, DistributedJobPoll.Poll)
		if err != nil {
			log.Error(err, "got error while polling for distributed job availability")
			// ignore error and continue
		}
		log.Info("found distributed job availability change, exiting")
		cancel()
	}()

	// force exit after the context is done
	go func() {
		<-withCancel.Done()
		// wait for a few seconds to allow the controller to exit gracefully
		<-time.After(time.Duration(options.forceExitTimeSec) * time.Second)
		log.Info("force exit")
		os.Exit(1)
	}()
	//
	//    start the service event loop
	//
	log.Info("starting manager")
	if err = mgr.Start(withCancel); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}

}

func setupPoliciesValidationWebhooks(mgr manager.Manager, availability apiAvailability) {
	k8sClient := mgr.GetClient()

	webhookServer := mgr.GetWebhookServer()

	if availability.interactivePolicy {
		handler := workloadspolicy.NewPolicyValidator(k8sClient, workloadspolicy.InitInteractivePolicyFromReq, options.allowedServiceAccounts)
		webhookServer.Register(fmt.Sprintf("/%s-%s", validationPolicyWebhooksPrefix, interactivePolicyType), &webhook.Admission{Handler: handler})
	}
	if availability.trainingPolicy {
		handler := workloadspolicy.NewPolicyValidator(k8sClient, workloadspolicy.InitTrainingPolicyFromReq, options.allowedServiceAccounts)
		webhookServer.Register(fmt.Sprintf("/%s-%s", validationPolicyWebhooksPrefix, trainingPolicyType), &webhook.Admission{Handler: handler})
	}
	if availability.distributedPolicy {
		handler := workloadspolicy.NewPolicyValidator(k8sClient, workloadspolicy.InitDistributedPolicyFromReq, options.allowedServiceAccounts)
		webhookServer.Register(fmt.Sprintf("/%s-%s", validationPolicyWebhooksPrefix, distributedPolicyType), &webhook.Admission{Handler: handler})
	}
}

func setupHealthCheck(mgr manager.Manager) {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
}

func setupControllers(mgr manager.Manager,
	enabledDistributedJobFramework map[string]controller.DistributedJobFrameworksStatus,
	podLabels map[string]string, podAnnotations map[string]string, controllerOptions controller.Options, availability apiAvailability) {
	recorder := mgr.GetEventRecorderFor(eventRecorderSource)
	if availability.interactiveWorkload {
		interactive := controller.NewSimpleInteractiveReconciler(mgr.GetClient(), mgr.GetAPIReader(), mgr.GetConfig(), recorder, controllerOptions)
		if err := interactive.SetupWithManager(mgr, options.interactiveMaxConcurrentReconciles); err != nil {
			log.Error(err, "unable to create interactive workload reconciler")
			os.Exit(1)
		}
	}
	if availability.trainingWorkload {
		train := controller.NewSimpleTrainingReconciler(mgr.GetClient(), mgr.GetAPIReader(), mgr.GetConfig(), recorder,
			podLabels, podAnnotations, controllerOptions)
		if err := train.SetupWithManager(mgr, options.trainingMaxConcurrentReconciles); err != nil {
			log.Error(err, "unable to create training workload reconciler")
			os.Exit(1)
		}
	}

	if availability.distributedWorkload {
		setupDistributedWorkloadController(mgr, enabledDistributedJobFramework, recorder,
			podLabels, podAnnotations, controllerOptions)
	}
}

func setupDistributedWorkloadController(mgr manager.Manager,
	enabledDistributedJobFramework map[string]controller.DistributedJobFrameworksStatus, eventRecorder record.EventRecorder,
	podLabels map[string]string, podAnnotations map[string]string, controllerOptions controller.Options) {
	if len(enabledDistributedJobFramework) == 0 {
		log.Info("no distributed job frameworks are enabled, skipping distributed workload controller")
		return
	}
	distributedWorkloadController := controller.NewSimpleDistributedWorkloadReconciler(mgr.GetClient(), mgr.GetAPIReader(),
		enabledDistributedJobFramework, controller.DistributedJobCRDsSupportedMap, eventRecorder,
		podLabels, podAnnotations, controllerOptions)
	if err := distributedWorkloadController.SetupWithManager(mgr,
		options.distributedMaxConcurrentReconciles, controllerOptions.MnnvlSupportedInCluster); err != nil {
		log.Error(err, "unable to create distributed workload reconciler")
		os.Exit(1)
	}
}

func setupWebhook(mgr manager.Manager,
	enabledDistributedJobFrameworks map[string]controller.DistributedJobFrameworksStatus,
	shouldUseControlPlanePolicies bool, clusterVersion string, options *Options, availability apiAvailability) {
	//
	//    setup webhooks
	//

	// register policies validation webhooks

	// policy crd
	setupPoliciesValidationWebhooks(mgr, availability)

	// register workload mutator webhooks
	registerMutators(mgr, shouldUseControlPlanePolicies, clusterVersion, options, availability)

	// register workload validators webhooks
	registerWorkloadsValidatorWebhooks(mgr, clusterVersion, enabledDistributedJobFrameworks, options, availability)
}

func enableProfiler(serverAddress string, mutexProfilerRate int, blockProfilerRate int, tags map[string]string) error {
	log.Info("Starting controller profiler")
	stdruntime.SetMutexProfileFraction(mutexProfilerRate)
	stdruntime.SetBlockProfileRate(blockProfilerRate)
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "workload-controller",

		// replace this with the address of pyroscope server
		ServerAddress: serverAddress,

		// you can disable logging by setting this to nil
		Logger: pyroscope.StandardLogger,

		// you can provide static tags via a map:
		Tags: tags,

		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to start pyroscope go agnet: %v", err)
	}
	return nil
}

func runMigration(cfg *rest.Config, clusterVersion string, migrationMaxWorkers int) error {
	migrationClient, err := client.New(cfg, client.Options{Scheme: scheme})
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "",
			CertDir:     options.certDir,
		},
		WebhookServer: &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    options.webhookPort,
				CertDir: options.certDir,
			}},
	})
	if err != nil {
		return fmt.Errorf("unabale to start webhooks: %w", err)
	}

	err = registerMigrationWebhooks(mgr, migrationClient)
	if err != nil {
		return err
	}

	migrator, err := migration.NewWorkloadMigration(migrationClient, clusterVersion, migrationMaxWorkers)
	if err != nil {
		return err
	}
	webserver := webhookUtils.NewExternalAccessCheckerServer(mgr.GetWebhookServer(), options.webhookExternalAddress,
		requestTimeoutMigrationWebhook)
	err = migrator.Run(context.Background(), webserver)
	if err != nil {
		return err
	}

	return nil
}

func enableWorkloadOwnershipProtection(mgr manager.Manager, k8sClient client.Client) {
	err := jobs.IndexAccessRules(mgr)
	if err != nil {
		log.Error(err, "unable to index access rules")
		os.Exit(1)
	}

	workloadControllerUID, err := webhookutils.GetWorkloadControllerUID(context.Background(), k8sClient)
	if err != nil {
		log.Error(err, "Failed getting workload-controller service account")
		os.Exit(1)
	}

	// register authorization validating webhook
	workloadOwnershipProtectionValidator := jobs.NewWorkloadOwnershipProtectionValidator(mgr.GetClient(), jobs.NewJobWorkloadWrapper, workloadControllerUID)
	workloadOwnershipMutator := jobs.NewWorkloadOwnershipMutator(mgr.GetClient(), jobs.NewJobWorkloadWrapper, workloadControllerUID)

	webhookServer := mgr.GetWebhookServer()
	registerWorkloadOwnershipProtection(webhookServer, workloadOwnershipProtectionValidator, workloadOwnershipMutator)
}

func registerWorkloadOwnershipProtection(webhookServer webhook.Server, workloadOwnershipProtectionValidator, workloadOwnershipMutator admission.Handler) {
	webhookServer.Register("/validate-workload-ownership-protection", &webhook.Admission{Handler: workloadOwnershipProtectionValidator})
	webhookServer.Register("/mutate-workload-ownership", &webhook.Admission{Handler: workloadOwnershipMutator})
}

func registerMutators(mgr manager.Manager, useControlPlanePolicies bool, clusterVersion string, options *Options, availability apiAvailability) {
	webserver := mgr.GetWebhookServer()
	// register interactive workload merge mutator webhook
	if availability.interactiveWorkload {
		interactiveMergeStackCreator := factory2.NewInteractiveWorkloadMergeCreator(mgr.GetClient(), useControlPlanePolicies)
		interactiveWorkloadPolicy := policy.NewSimplePolicyApplier(interactiveMergeStackCreator)
		interactiveHandler := mutators.NewWorkloadMutatorHandler(interactiveWorkloadPolicy, iwv1.GetNewInteractiveWorkloadFn,
			clusterVersion, &mutators.WorkloadStatusUpdater{}, false, false, false)
		registerWorkloadMutationWebhook(webserver, interactiveHandler, interactiveType)
	}

	// register training workload merge mutator webhook
	if availability.trainingWorkload {
		trainingMergeStackCreator := factory2.NewTrainingWorkloadMergeCreator(mgr.GetClient(), useControlPlanePolicies)
		trainingWorkloadPolicy := policy.NewSimplePolicyApplier(trainingMergeStackCreator)
		trainingHandler := mutators.NewWorkloadMutatorHandler(trainingWorkloadPolicy, twv1.GetNewTrainingWorkloadFn,
			clusterVersion, &mutators.WorkloadStatusUpdater{}, false, false, false)
		registerWorkloadMutationWebhook(webserver, trainingHandler, trainingType)
	}

	// register distributed workload merge mutator webhook
	if availability.distributedWorkload {
		distributedMergeStackCreator := factory2.NewDistributedWorkloadMergeCreator(mgr.GetClient(), useControlPlanePolicies)
		distributedWorkloadPolicy := policy.NewSimplePolicyApplier(distributedMergeStackCreator)
		distributedHandler := mutators.NewWorkloadMutatorHandler(distributedWorkloadPolicy, distw.GetNewDistributedWorkloadFn,
			clusterVersion, &mutators.WorkloadStatusUpdater{}, options.mnnvlSupported, options.topologyEnabled, options.elasticCDEnabled)
		registerWorkloadMutationWebhook(webserver, distributedHandler, distributedType)
	}

}

func registerWorkloadsValidatorWebhooks(
	mgr manager.Manager,
	clusterVersion string,
	enabledDistributedJobFrameworks map[string]controller.DistributedJobFrameworksStatus,
	options *Options,
	availability apiAvailability,
) {
	workloadParamsValidator := workloads.NewGenericParamsValidator()
	workloadGpuValidator := jobs.NewWorkloadGpuValidator()

	frameworkValidator := workloads.NewWorkloadFrameworkValidator(mgr.GetClient(), enabledDistributedJobFrameworks)
	singleRunaiValidator := workloads.NewSingleRunaiJobValidator(mgr.GetClient())
	existingNodepoolsValidator := workloads.NewExistingNodepoolsValidator(mgr.GetClient())
	existingPvcsValidator := workloads.NewExistingPvcsValidator(mgr.GetClient())
	specValidator := workloads.NewSpecValidator()
	distSpecValidator := workloads.NewDistSpecValidator()
	mnnvlValidator := workloads.NewMnnvlValidator(options.mnnvlSupported, options.elasticCDEnabled)
	topologyValidator := workloads.NewTopologyValidator(options.topologyEnabled)
	workloadValidatorHandler := validator.NewWorkloadParamsValidator(mgr.GetClient(), options.allowedServiceAccounts, frameworkValidator, workloadParamsValidator, workloadGpuValidator, singleRunaiValidator, existingNodepoolsValidator, existingPvcsValidator, specValidator)
	distributedWorkloadValidatorHandler := validator.NewWorkloadParamsValidator(mgr.GetClient(), options.allowedServiceAccounts,
		frameworkValidator, workloadParamsValidator, workloadGpuValidator, existingNodepoolsValidator, existingPvcsValidator, specValidator,
		distSpecValidator, mnnvlValidator, topologyValidator)
	distributedWorkloadValidatorHandler.RegisterUpdateValidators(frameworkValidator)

	webhookServer := mgr.GetWebhookServer()

	if availability.interactiveWorkload {
		registerWorkloadValidatorsWebhook(webhookServer, workloadValidatorHandler, interactiveType)
	}
	if availability.trainingWorkload {
		registerWorkloadValidatorsWebhook(webhookServer, workloadValidatorHandler, trainingType)
	}
	if availability.distributedWorkload {
		registerWorkloadValidatorsWebhook(webhookServer, distributedWorkloadValidatorHandler, distributedType)
	}
}

func registerWorkloadValidatorsWebhook(webhookServer webhook.Server, workloadParamsValidatorHandler webhook.AdmissionHandler, workloadType string) {

	// register workload validator webhook
	webhookServer.Register(fmt.Sprintf("/validate-%s-workload", workloadType),
		&webhook.Admission{Handler: workloadParamsValidatorHandler})
}

func registerWorkloadMutationWebhook(webhookServer webhook.Server, workloadMutationHandler webhook.AdmissionHandler, workloadType string) {
	webhookServer.Register(fmt.Sprintf("/mutate-%s-workload", workloadType), &webhook.Admission{Handler: workloadMutationHandler})
}

func registerMigrationWebhooks(mgr manager.Manager, k8sClient client.Client) error {
	migrationWebookHandler, err := migration.NewWebhookHandler(k8sClient, "")
	if err != nil {
		return fmt.Errorf("failed to create migration webhook handler: %w", err)
	}
	webhookServer := mgr.GetWebhookServer()
	registerWorkloadOwnershipProtection(webhookServer, migrationWebookHandler, migrationWebookHandler)
	registerWorkloadValidatorsWebhook(webhookServer, migrationWebookHandler, interactiveType)
	registerWorkloadValidatorsWebhook(webhookServer, migrationWebookHandler, trainingType)
	registerWorkloadValidatorsWebhook(webhookServer, migrationWebookHandler, distributedType)
	registerWorkloadMutationWebhook(webhookServer, migrationWebookHandler, trainingType)
	registerWorkloadMutationWebhook(webhookServer, migrationWebookHandler, interactiveType)
	registerWorkloadMutationWebhook(webhookServer, migrationWebookHandler, distributedType)
	webhookServer.Register("/validate-restartrequest", &webhook.Admission{Handler: migrationWebookHandler})
	return nil
}

func parseControllerOptions(options *Options) (*controller.Options, error) {
	failureResourceCleanupPolicy := managers.ParseFailureCleanupPolicy(options.failureResourceCleanupPolicy)
	if failureResourceCleanupPolicy == managers.Undefined {
		return nil, fmt.Errorf("%s is undefined failure cleanup policy", options.failureResourceCleanupPolicy)
	}

	return &controller.Options{
		FailureResourceCleanupPolicy:  failureResourceCleanupPolicy,
		MnnvlSupportedInCluster:       options.mnnvlSupported,
		ElasticCDSupportedInCluster:   options.elasticCDEnabled,
		EnableDanglingResourceCleanup: options.enableDanglingResourceCleanup,
	}, nil
}
