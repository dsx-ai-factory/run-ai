package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"

	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/managed-nodes-config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/topology_controller"
	"go.uber.org/zap/zapcore"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/cmd/app"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/cmd/version"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/nodepool_controller/metrics"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/scheme"
	nodepoolwebhook "github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/webhook"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	scheme   = scheme_init.Scheme()
	setupLog = ctrl.Log.WithName("setup")
)

func initLogging(debugLogLevel bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if !debugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	opts := zap.Options{
		Development:     true,
		StacktraceLevel: zapcore.LevelEnabler(zapcore.FatalLevel),
		TimeEncoder:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func main() {
	ops := app.InitOptions()
	initLogging(ops.DebugLogLevel)

	log.Info().Msg("Run:AI NodePool Controller")

	nodepoolControllerParams, err := parseNodePoolControllerParams()
	if err != nil {
		setupLog.Error(err, "unable to parse args")
		os.Exit(1)
	}

	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(ops.K8sClientConfigQPS)
	clientConfig.Burst = ops.K8sClientConfigBurst

	managerOptions := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf(":%s", ops.MetricsPort),
		},
		LeaderElection:   ops.EnableLeaderElection,
		LeaderElectionID: fmt.Sprintf("node-pool-controller.%s", config.Get().FinalizerDomain),
	}

	if ops.EnableNodepoolValidationWebhook {
		managerOptions.WebhookServer = &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    ops.WebhookPort,
				CertDir: ops.WebhookCertDir,
			}}
	}

	mgr, err := ctrl.NewManager(clientConfig, managerOptions)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	stopper := make(chan struct{})
	defer func() {
		close(stopper)
	}()
	managerContext := ctrl.SetupSignalHandler()

	nodePoolController := nodepool_controller.NewNodePoolController(
		mgr.GetClient(),
		mgr.GetScheme(),
		nodepoolControllerParams)
	err = nodePoolController.SetupWithManager(managerContext, mgr)
	if err != nil {
		setupLog.Error(err, "Error starting NodePoolController")
		os.Exit(1)
	}

	managedNodesConfigController := managed_nodes_config.NewManagedNodesConfigController(
		mgr.GetClient(),
		mgr.GetScheme(),
		nodePoolController)
	err = managedNodesConfigController.SetupWithManager(managerContext, mgr)
	if err != nil {
		setupLog.Error(err, "Error starting ManagedNodesConfigController")
		os.Exit(1)
	}

	if ops.EnableNodepoolValidationWebhook {
		if err = nodepoolwebhook.SetupNodePoolWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "Error setting up NodePool validating webhook")
			os.Exit(1)
		}
		setupLog.Info("NodePool validation webhook enabled",
			"port", ops.WebhookPort, "certDir", ops.WebhookCertDir)
	}

	go func() {
		<-mgr.Elected()
		setupLog.Info("Leader elected")
		metrics.Init()
		cachedClient := cachedclient.NewCachedClient(managerContext, mgr.GetClient(), ctrl.GetConfigOrDie(), scheme)
		metricsWatch, err := nodepool_controller.InitMetricsWatch(cachedClient)
		if err != nil {
			setupLog.Error(err, "Error starting NodePoolController metrics handler")
			os.Exit(1)
		}
		metricsWatch.Start(managerContext, stopper)

		topology_controller.StartWhenCRDReady(managerContext, mgr.GetAPIReader(), func() {
			tc := topology_controller.NewKaiTopologyController(mgr.GetClient())
			if err := tc.SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "Failed to start KAI topology controller")
				os.Exit(1)
			}
			setupLog.Info("KAI topology controller started successfully")
		})
	}()

	printVersion()
	log.Info().Msgf("Starting NodePoolController")
	if err := mgr.Start(managerContext); err != nil {
		setupLog.Error(err, "Error running manager")
		os.Exit(1)
	}

	<-stopper
}

func printVersion() {
	versionInfo := version.GetVersion()
	setupLog.Info(fmt.Sprintf("Version: %s", versionInfo.Version))
	if strings.HasSuffix(versionInfo.Version, "-DEVELOPMENT") {
		return
	}
	if versionInfo.BuildDate != "" {
		setupLog.Info(fmt.Sprintf("BuildDate: %s", versionInfo.BuildDate))
	}
	if versionInfo.GitCommit != "" {
		setupLog.Info(fmt.Sprintf("Commit Hash: %s", versionInfo.GitCommit))
	}
	if versionInfo.GitTag != "" {
		setupLog.Info(fmt.Sprintf("Tag: %s", versionInfo.GitTag))
	}
	setupLog.Info(fmt.Sprintf("Go Version: %s", versionInfo.GoVersion))
	setupLog.Info(fmt.Sprintf("Platform: %s", versionInfo.Platform))
}

func parseNodePoolControllerParams() (*common.NodePoolControllerParams, error) {
	schedulingShardArgs, err := app.ParseSchedulingShardArgs(config.Get().SchedulingShardArgsStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scheduler args: %w", err)
	}

	nodePoolControllerParams := &common.NodePoolControllerParams{
		SchedulingShardArgs: schedulingShardArgs,
	}
	return nodePoolControllerParams, nil
}
