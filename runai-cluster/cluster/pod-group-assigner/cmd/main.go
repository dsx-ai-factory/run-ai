package main

import (
	"flag"
	"os"
	"strings"

	"go.uber.org/zap/zapcore"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	podgroupcontroller "github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup"
	podmutator "github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/mutators/pod"
	podgroupmutator "github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/mutators/podgroup"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/profiling"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/scheme"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/version"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	//+kubebuilder:scaffold:imports

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	scheme  = scheme_init.Scheme()
	options = &Options{}
)

// Options is the main context object for the controller manager.
type Options struct {
	useDebugLogLevel     bool
	certDir              string
	webhookPort          int
	enableProfiler       bool
	profilerApiPort      string
	k8sClientConfigQPS   int
	k8sClientConfigBurst int
	EnableLeaderElection bool
	enablePodWebhook     bool
}

func init() {
	flag.BoolVar(&options.useDebugLogLevel, "debug", false, "Should use debug log level")
	config.AddLabelFlags(flag.CommandLine)
	flag.StringVar(&options.certDir, "cert-dir", "/etc/webhook/certs", "Directory that contains the server key and certificate for the webhooks")
	flag.IntVar(&options.webhookPort, "webhook-port", 8443, "The port to serve the webhook from")
	flag.BoolVar(&options.enableProfiler, "enable-profiling", false, "Should enable profiler")
	flag.StringVar(&options.profilerApiPort, "profiler-api-port", "8182", "Profiler API port")
	flag.IntVar(&options.k8sClientConfigQPS, "qps", 50, "Queries per second to the K8s API server")
	flag.IntVar(&options.k8sClientConfigBurst, "burst", 300, "Burst to the K8s API server")
	flag.BoolVar(&options.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&options.enablePodWebhook, "enable-pod-webhook", false,
		"Enable the pod mutating webhook. For pods of a project namespace that are either scheduled by "+
			"--scheduler-name or whose project enforces it, it enforces the scheduler name, labels the pod "+
			"with its project, and applies the project's defaultNodePools as node affinity.")
	flag.Parse()
	initLogging(options.useDebugLogLevel)
}

func initLogging(useDebugLogLevel bool) {
	if !useDebugLogLevel {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// controller log needs to be set up for the underlying webhook code, for example
	opts := zap.Options{
		Development: useDebugLogLevel,
		TimeEncoder: zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func main() {
	log.Info().Msg("Run:AI Pod Group Assigner")

	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(options.k8sClientConfigQPS)
	clientConfig.Burst = options.k8sClientConfigBurst

	mgr, err := ctrl.NewManager(clientConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0", // disable metrics
		},
		LeaderElection:   options.EnableLeaderElection,
		LeaderElectionID: "pod-group-assigner.run.ai",
	})
	if err != nil {
		log.Error().Msgf("unable to start manager, error: %s", err.Error())
		os.Exit(1)
	}

	registerWebhooks(mgr)

	managerContext := ctrl.SetupSignalHandler()

	podGroupController := podgroupcontroller.NewPodGroupReconciler(mgr.GetClient())
	if err = podGroupController.SetupWithManager(managerContext, mgr); err != nil {
		log.Error().Msgf("unable to create PodGroup controller, error: %s", err.Error())
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error().Msgf("unable to set up health check, error: %s", err.Error())
		os.Exit(1)
	}

	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error().Msgf("unable to set up ready check, error: %s", err.Error())
		os.Exit(1)
	}

	if options.enableProfiler {
		go profiling.RegisterProfiler(options.profilerApiPort)
	}

	printVersion()
	log.Info().Msg("Starting PodGroupController")

	if err = mgr.Start(managerContext); err != nil {
		log.Error().Msgf("problem running manager, error: %s", err.Error())
		os.Exit(1)
	}
}

func registerWebhooks(mgr manager.Manager) {
	webhookServer := webhook.NewServer(webhook.Options{
		Port:    options.webhookPort,
		CertDir: options.certDir,
	})

	webhookServer.Register("/mutate-pod-group", &webhook.Admission{Handler: podgroupmutator.NewPodGroupMutator()})

	if options.enablePodWebhook {
		log.Info().Msgf("Registering pod mutating webhook at %s", podmutator.WebhookPath)
		webhookServer.Register(podmutator.WebhookPath,
			&webhook.Admission{Handler: podmutator.NewPodMutator(mgr.GetClient())})
	}

	if err := mgr.Add(webhookServer); err != nil {
		log.Error().Msgf("unable to add webhook server to manager, error: %s", err.Error())
		os.Exit(1)
	}
}

func printVersion() {
	versionInfo := version.Version()
	log.Info().Msgf("Version: %s", versionInfo.Version)

	if strings.HasSuffix(versionInfo.Version, "-DEVELOPMENT") {
		return
	}

	if versionInfo.BuildDate != "" {
		log.Info().Msgf("BuildDate: %s", versionInfo.BuildDate)
	}

	if versionInfo.GitCommit != "" {
		log.Info().Msgf("Commit Hash: %s", versionInfo.GitCommit)
	}

	if versionInfo.GitTag != "" {
		log.Info().Msgf("Tag: %s", versionInfo.GitTag)
	}

	log.Info().Msgf("Go Version: %s", versionInfo.GoVersion)
	log.Info().Msgf("Platform: %s", versionInfo.Platform)
}
