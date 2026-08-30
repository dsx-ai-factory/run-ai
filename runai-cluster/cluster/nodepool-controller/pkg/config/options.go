package config

import (
	"flag"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
)

// AddLabelFlags binds the nodepool-controller's label-vocabulary
// configurable values to the given FlagSet. Other runtime flags (debug,
// ports, qps, leader-election) are declared separately in cmd/app/options.go
// and do not go through this function. Must be called before flag.Parse().
func AddLabelFlags(fs *flag.FlagSet, cfg *NodePoolControllerConfig) {
	fs.StringVar(&cfg.NodePoolNameLabel, "nodepool-label-key",
		kaiconstants.DefaultNodePoolLabelKey,
		"Label key used to associate Kubernetes resources with a node pool")
	fs.StringVar(&cfg.DefaultNodepoolName, "default-nodepool-name",
		kaiconstants.DefaultNodePoolName,
		"Name identifying the implicit default node pool")
	fs.StringVar(&cfg.SchedulerName, "scheduler-name",
		kaiconstants.DefaultSchedulerName,
		"Value of pod.Spec.SchedulerName for pods managed by this scheduler; also the base name of per-shard ServiceMonitor resources")
	fs.StringVar(&cfg.SchedulerNamespace, "scheduler-namespace",
		kaiconstants.DefaultKAINamespace,
		"Namespace where per-shard scheduler resources (ServiceMonitor) live")
	fs.StringVar(&cfg.MetricsNamespace, "metrics-namespace",
		kaiconstants.DefaultMetricsNamespace,
		"Prefix attached to every Prometheus metric the controller emits or queries (Subsystem of its own gauge and prefix of the queue-allocation metric names written into each SchedulingShard)")
	fs.StringVar(&cfg.FinalizerDomain, "finalizer-domain",
		defaultFinalizerDomain,
		"DNS-style domain prefix used to compose the nodepool finalizer string")
	fs.StringVar(&cfg.SchedulingShardArgsStr, "scheduling-shard-args",
		"",
		"JSON-encoded map of cluster-wide KAI scheduler args the controller merges into every SchedulingShard CR it writes")
	fs.StringVar(&cfg.ManagedNodesConfigName, "managed-nodes-config-name",
		defaultManagedNodesConfigName,
		"Name of the singleton ManagedNodesConfig CR the controller reconciles")
	fs.StringVar(&cfg.ExcludedNodepoolName, "excluded-nodepool-name",
		defaultExcludedNodepoolName,
		"Reserved nodepool that unmanaged nodes are moved into")
	fs.StringVar(&cfg.ShouldBeExcludedLabelKey, "to-exclude-label",
		defaultToExcludeLabel,
		"Node label key marking a node pending graceful drain before exclusion")
	fs.StringVar(&cfg.UnschedulableLabelKey, "unschedulable-label",
		defaultUnschedulableLabel,
		"Node label key marking a node the controller made unschedulable")
	fs.StringVar(&cfg.GroveTopologyAnnotation, "grove-topology-annotation",
		defaultGroveTopologyAnnotation,
		"Annotation key linking a topology to its Grove topology")
	fs.StringVar(&cfg.GroveTopologyResourceVersionAnnotation, "grove-topology-resource-version-annotation",
		defaultGroveTopologyRVAnnotation,
		"Annotation key tracking the synced Grove topology resource version")
}
