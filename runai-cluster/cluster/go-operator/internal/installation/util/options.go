package util

import (
	"flag"
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const defaultProfilerApiPort = "8182"

type Options struct {
	MetricsAddr                              string
	RemoteClusterManagerMetricsAddr          string
	EnableLeaderElection                     bool
	RemoteClusterManagerEnableLeaderElection bool
	ProbeAddr                                string
	RemoteClusterManagerProbeAddr            string
	PreDelete                                bool
	EnableProfiler                           bool
	ProfilerApiPort                          string
	ZapOptions                               zap.Options
	Verbosity                                int
	EnableNMC                                bool
	NMCRemoteClusterAPIServerURL             string
	DisableRunaiConfigFinalizer              bool
	LegacyWorkloadCleanup                    bool
}

func SetOptions() *Options {
	opts := &Options{}
	opts.parseCommandLineArgs(flag.CommandLine)
	return opts
}

func (opts *Options) parseCommandLineArgs(flagSet *flag.FlagSet) {
	flagSet.StringVar(&opts.MetricsAddr, "metrics-bind-address", ":8082", "The address the metric endpoint binds to.")

	flagSet.BoolVar(&opts.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flagSet.BoolVar(&opts.EnableNMC, "nmc-enabled", false, "Enable NMC")

	flagSet.StringVar(&opts.NMCRemoteClusterAPIServerURL, "nmc-remote-cluster-api-server-url", "", "The URL of the remote cluster API server")

	flagSet.StringVar(&opts.ProbeAddr, "health-probe-bind-address", ":8083", "The address the probe endpoint binds to.")

	flagSet.BoolVar(&opts.PreDelete, "pre-delete", false, "Run pre-delete tasks")

	flagSet.IntVar(&opts.Verbosity, "log-level", 0,
		"User specified logging level, set to 3 for all debug logs")

	flagSet.BoolVar(&opts.EnableProfiler, "enable-profiler", false, "Enable profiler")

	flagSet.StringVar(&opts.ProfilerApiPort, "profiler-port", defaultProfilerApiPort, "The port to listen for profiler api requests")

	flagSet.StringVar(&opts.RemoteClusterManagerMetricsAddr, "remote-cluster-manager-metrics-bind-address", ":9082", "The address the remote cluster metrics endpoint binds to")

	flagSet.StringVar(&opts.RemoteClusterManagerProbeAddr, "remote-cluster-manager-health-probe-bind-address", ":9083", "The address the remote cluster probe endpoint binds to")

	flagSet.BoolVar(&opts.RemoteClusterManagerEnableLeaderElection, "remote-cluster-manager-leader-elect", false, "Enable leader election for remote cluster controller manager")

	flagSet.BoolVar(&opts.DisableRunaiConfigFinalizer, "disable-runai-config-finalizer", false, "Disable runai config finalizer")

	flagSet.BoolVar(&opts.LegacyWorkloadCleanup, "legacy-workload-cleanup", false,
		"Run the one-time legacy workload controller cleanup")

	opts.ZapOptions = zap.Options{
		Development: true,
	}
	opts.ZapOptions.BindFlags(flagSet)
	flagSet.Parse(os.Args[1:])

	opts.ValidateOrDie()
}

func (opts *Options) ValidateOrDie() {
	if opts.EnableNMC && opts.NMCRemoteClusterAPIServerURL == "" {
		panic(fmt.Errorf("NMC is enabled but remote cluster API server URL is not set"))
	}
}
