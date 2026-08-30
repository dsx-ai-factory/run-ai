package config

import "fmt"

const (
	// defaultKAIPrefix is the single source of truth for the KAI-style
	// label/domain prefix used to construct several default-value consts
	// below.
	defaultKAIPrefix = "kai.scheduler"

	defaultFinalizerDomain = defaultKAIPrefix

	// nodepoolFinalizerPrefix and nodepoolFinalizerSuffix are appended around
	// the configured FinalizerDomain to compose the nodepool finalizer string.
	// e.g., domain "run.ai" -> "nodepool.run.ai/finalize".
	nodepoolFinalizerPrefix = "nodepool."
	nodepoolFinalizerSuffix = "/finalize"

	// Managed-nodes vocabulary defaults.
	defaultManagedNodesConfigName = "kai-managed-nodes-config"
	defaultExcludedNodepoolName   = "kai-excluded-nodes"
	defaultToExcludeLabel         = defaultKAIPrefix + "/to-exclude"

	// Node/topology vocabulary defaults.
	defaultUnschedulableLabel        = defaultKAIPrefix + "/unschedulable"
	defaultGroveTopologyAnnotation   = defaultKAIPrefix + "/grove-topology-name"
	defaultGroveTopologyRVAnnotation = defaultKAIPrefix + "/grove-topology-resource-version"

	// PodGroupAnnotationForPod is the annotation carrying a pod's pod-group
	// name. Vendor-neutral and shared with the scheduler, so it is a fixed
	// const rather than a configurable value.
	PodGroupAnnotationForPod = "pod-group-name"
)

// NodePoolControllerConfig holds the binary's flag-configurable values.
// Production code reads it via Get(); tests replace it via SetForTest.
type NodePoolControllerConfig struct {
	// NodePoolNameLabel is the label key used to associate Kubernetes
	// resources (Nodes, PodGroups, Queues, SchedulingShards) with a specific
	// node pool.
	NodePoolNameLabel string

	// DefaultNodepoolName is the name identifying the implicit default
	// node pool — the bucket for nodes/workloads that don't carry an explicit
	// NodePoolNameLabel label. The controller treats nodepools named this as
	// the catch-all.
	DefaultNodepoolName string

	// SchedulerName is the value of pod.Spec.SchedulerName that identifies
	// pods managed by the scheduler this controller serves. Used (a) to
	// filter pods when computing per-nodepool capacity / draining state and
	// (b) as the base name of the ServiceMonitor resources written by
	// `pkg/operands/runai-scheduler`.
	SchedulerName string

	// SchedulerNamespace is the namespace in which the ServiceMonitor and
	// related per-shard scheduler resources live.
	SchedulerNamespace string

	// MetricsNamespace is the prefix attached to every Prometheus metric
	// the controller emits or queries: it is used both as the Subsystem of
	// the controller's own `<MetricsNamespace>_node_nodepool` gauge and as
	// the prefix of the queue-allocation metric names (e.g.,
	// `<MetricsNamespace>_queue_allocated_gpus`) written into each
	// SchedulingShard's UsageDB ExtraParams. Defaults to KAI's metrics
	// prefix ("kai"); runai overrides to "runai" so existing dashboards
	// keep working.
	MetricsNamespace string

	// FinalizerDomain is the DNS-style domain prefix used to compose the
	// nodepool finalizer string. See FinalizerName() for the composition.
	FinalizerDomain string

	// SchedulingShardArgsStr is the JSON-encoded map of cluster-wide KAI scheduler
	// args supplied at startup via the --scheduling-shard-args flag. The controller
	// parses it once at boot and merges it into every SchedulingShard's Args,
	// one per NodePool. Empty means no overrides (KAI built-in defaults).
	SchedulingShardArgsStr string

	// ManagedNodesConfigName is the name of the singleton ManagedNodesConfig
	// CR the controller reconciles.
	ManagedNodesConfigName string

	// ExcludedNodepoolName is the reserved nodepool unmanaged nodes are moved
	// into so the scheduler skips them.
	ExcludedNodepoolName string

	// ShouldBeExcludedLabelKey marks a node pending drain before exclusion.
	ShouldBeExcludedLabelKey string

	// UnschedulableLabelKey marks a node the controller made unschedulable.
	UnschedulableLabelKey string

	// GroveTopologyAnnotation and GroveTopologyResourceVersionAnnotation are
	// the annotation keys the controller writes onto a topology to link it to
	// its Grove topology and track the synced resource version.
	GroveTopologyAnnotation                string
	GroveTopologyResourceVersionAnnotation string
}

// current holds the parsed configuration. SetCurrent assigns it during
// startup; SetForTest replaces it temporarily in tests. Get() returns it for
// read-only access from deep-stack code.
var current *NodePoolControllerConfig

// Get returns the parsed configuration. Must be called after SetCurrent has
// run (i.e., after flag.Parse() in main, or after SetForTest in tests).
func Get() *NodePoolControllerConfig { return current }

// SetCurrent assigns the package-level configuration singleton. Called from
// the binary's startup once flags have been parsed.
func SetCurrent(c *NodePoolControllerConfig) { current = c }

// SetForTest replaces the global config and returns a restore closure.
// Tests only — production code goes through SetCurrent.
func SetForTest(c *NodePoolControllerConfig) func() {
	prev := current
	current = c
	return func() { current = prev }
}

// FinalizerName composes the nodepool-finalizer string from the configured
// FinalizerDomain. runai input "run.ai" -> "nodepool.run.ai/finalize"
// (unchanged); OSS default "kai.scheduler" -> "nodepool.kai.scheduler/finalize".
func FinalizerName() string {
	return fmt.Sprintf("%s%s%s", nodepoolFinalizerPrefix, Get().FinalizerDomain, nodepoolFinalizerSuffix)
}
