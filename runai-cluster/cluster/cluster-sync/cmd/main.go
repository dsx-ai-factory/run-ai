/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"os"
	"time"

	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	runaikaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	infpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	ipv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	tpv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"

	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	idmclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/cmd/profiling"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/logger"
	dinfwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	inferencewv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	mpi "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/code-generator"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logger.InitLogger()
	config.Init()
}

func main() {
	zap.L().Info("Cluster Sync service has started")
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		errorMsg := err.Error()
		kubeConfigPath := os.Getenv("KUBE_CONFIG_FILE_PATH")
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			zap.S().Error(err)
			zap.S().Fatal(errorMsg)
		} else {
			zap.S().Debugf("Got kube config from  path %s", kubeConfigPath)
		}
	}

	kubeConfig.QPS = float32(config.Config().K8sClientConfigQPS)
	kubeConfig.Burst = config.Config().K8sClientConfigBurst

	scheme := initSchema()

	if config.Config().EnableProfiler {
		go profiling.RegisterProfiler(config.Config().ProfilerApiPort)
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   config.Config().EnableLeaderElection,
		LeaderElectionID: "cluster-sync.run.ai",
	})
	if err != nil {
		zap.S().Error(err, "Failed to configure manager")
		os.Exit(1)
	}
	managerContext := ctrl.SetupSignalHandler()

	if err = setupNRTCRDWatchIfNeeded(managerContext, mgr); err != nil {
		zap.S().Errorf("Failed watching for NodeResourceTopology CRD installation: %v", err)
		os.Exit(1)
	}

	go func() {
		<-mgr.Elected()
		zap.S().Info("Leader elected")
		startWatchClients(kubeConfig, scheme, managerContext)
	}()

	if err := mgr.Start(managerContext); err != nil {
		zap.S().Errorf("Manager stopped: %v", err)
		os.Exit(1)
	}
}

func setupNRTCRDWatchIfNeeded(ctx context.Context, manager ctrl.Manager) error {
	if controllers.IsNRTCRDInstalled(ctx, manager.GetAPIReader()) {
		zap.S().Info("NodeResourceTopology CRD detected; NUMA node sync will be enabled")
		return nil
	}

	zap.S().Info("NodeResourceTopology CRD not installed; watching for its installation")
	return controllers.WatchForNRTCRD(ctx, manager)
}

func initSchema() *apiruntime.Scheme {
	scheme := scheme_init.Scheme()

	addToScheme := func(fn func(*apiruntime.Scheme) error, name string) {
		if err := fn(scheme); err != nil {
			zap.S().Errorf("failed to add %s to scheme: %v", name, err)
		}
	}

	addToScheme(kubeflowv1.AddToScheme, "kubeflowv1") // pytorchjob, tfjob, xgboostjob
	addToScheme(mpi.AddToScheme, "mpijob")
	addToScheme(iwv2a1.AddKnownTypes, "interactive workload types")
	addToScheme(iwv2a1.AddToScheme, "interactive workload")
	addToScheme(twv2a1.AddKnownTypes, "training workload types")
	addToScheme(twv2a1.AddToScheme, "training workload")
	addToScheme(inferencewv2.AddToScheme, "inference workload")
	addToScheme(diwv2.AddToScheme, "distributed workload")
	addToScheme(knv1.AddToScheme, "knative")
	addToScheme(ipv2.AddToScheme, "interactive policy")
	addToScheme(tpv2.AddToScheme, "training policy")
	addToScheme(infpv2.AddToScheme, "inference policy")
	addToScheme(dpv2.AddToScheme, "distributed policy")
	addToScheme(dinfwv1.AddToScheme, "distributed inference workload")
	addToScheme(kaiv1alpha1.AddToScheme, "kai topology")
	addToScheme(runaikaiv1alpha1.AddToScheme, "kai resources project/department")
	addToScheme(grovev1alpha1.AddToScheme, "grove cluster topology")
	addToScheme(kartav1alpha1.AddToScheme, "kai bolt resource interface")
	addToScheme(nrtv1alpha2.AddToScheme, "node resource topology")
	return scheme
}

func initBackendClient(ctx context.Context) (*backendclient.BackendClient, error) {
	basePath := config.Config().RunaiApiEndPoint
	idmClient := idmclient.NewIdmClient(basePath)
	authnManager, err := idmclient.NewAuthnManager(ctx, idmclient.AuthnManagerConfig{
		ClientId:     config.Config().ClientName,
		ClientSecret: config.Config().ClientSecret,
		IdmClient:    idmClient,
	})
	if err != nil {
		return nil, err
	}
	runaiBackendClient := backendclient.NewBackendClient(basePath, authnManager.HttpClient())
	return runaiBackendClient, nil
}

func startWatchClients(kubeConfig *rest.Config, scheme *apiruntime.Scheme, ctx context.Context) {
	runaiBackendClient, err := initBackendClient(ctx)
	if err != nil {
		zap.S().Fatalf("Failed to create authn manager: %v", err)
	}

	watchClient, err := client.NewWithWatch(kubeConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		zap.S().Fatalf("failed creating watchClient watch client: %+v", err)
	}
	zap.S().Debug("Created runai watch Client")

	//A channel that will make main goroutine wait for controllers, and allows for a means to stop the program
	defer runtime.HandleCrash()

	cachedClient := cachedclient.NewCachedClient(ctx, watchClient, kubeConfig, scheme)
	defer cachedClient.Shutdown()
	//accessRuleController := controllers.NewAccessRuleController(watchClient, runaiBackendClient)

	eventsController := controllers.NewEventController(cachedClient, runaiBackendClient)
	nodeHandler := handlers.NewNodeHandler(runaiBackendClient, cachedClient)
	numaNodeHandler := handlers.NewNumaNodeHandler(runaiBackendClient, cachedClient)
	podHandler := handlers.NewPodHandler(runaiBackendClient, cachedClient)
	controllers.AddNodeHandlerToCache(runaiBackendClient, cachedClient, nodeHandler)
	if controllers.AddNumaNodeHandlerToCache(cachedClient, numaNodeHandler) {
		nodeHandler.SetNumaNodeHandler(numaNodeHandler)
	}
	controllers.AddPVCHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddConfigMapHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddClusterInfoHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddSecretHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddNodePoolHandlerToCache(runaiBackendClient, cachedClient, nodeHandler)
	controllers.AddDatavolumeHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddAccessRuleHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddProjectHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddDepartmentHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddStorageClassHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddPriorityClassHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddNetworkTopologyHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddApplicationHandlerToCache(runaiBackendClient, cachedClient)
	controllers.AddKartaHandlerToCache(runaiBackendClient, cachedClient)

	podRateLimiter := controllers.AddPodHandlerToCache(runaiBackendClient, cachedClient, podHandler, nodeHandler)
	workloadHandler := handlers.NewWorkloadHandler(runaiBackendClient, cachedClient, podHandler, podRateLimiter)
	trainingRLC := controllers.AddTrainingHandlerToCache(runaiBackendClient, cachedClient, workloadHandler)
	workspaceRLC := controllers.AddWorkspaceHandlerToCache(runaiBackendClient, cachedClient, workloadHandler)
	distributedRLC := controllers.AddDistributedHandlerToCache(runaiBackendClient, cachedClient, workloadHandler)
	inferenceRLC := controllers.AddInferenceHandlerToCache(runaiBackendClient, cachedClient, workloadHandler)
	externalRLC := controllers.AddExternalWorkloadHandlerToCache(runaiBackendClient, cachedClient, workloadHandler)
	distributedInferenceRLC := controllers.AddDistributedInferenceHandlerToCache(runaiBackendClient, cachedClient, workloadHandler)
	controllers.AddPodGroupHandlerToCache(cachedClient, trainingRLC, workspaceRLC, distributedRLC, inferenceRLC, externalRLC, distributedInferenceRLC)
	controllers.AddWorkloadRunnerHandlerToCache(cachedClient, externalRLC)
	controllers.AddRevisionHandlerToCache(runaiBackendClient, cachedClient)

	policyHandler := handlers.NewPolicyHandler(runaiBackendClient, cachedClient)
	controllers.AddPolicyHandlerToCache(cachedClient, policyHandler)

	cachedClient.Start(ctx, ctx.Done())

	//go accessRuleController.Start(stopper)

	go eventsController.StartSyncingPeriodically(ctx.Done())
	go runaiBackendClient.StartSyncingClusterInfo(ctx.Done())

	// liveness handles sync once the backend becomes available - so informers should already be initialized
	go runaiBackendClient.StartSendingLivenessPeriodically(ctx.Done())

	if config.Config().SyncIntervalMinutes > 0 {
		go runaiBackendClient.SyncPeriodically(time.Duration(config.Config().SyncIntervalMinutes), ctx.Done())
	}

	<-ctx.Done()
}
