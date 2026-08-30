package config

import (
	"flag"
	"sync"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	kaipgconstants "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

const (
	defaultUnexistingNodepoolSentinel = "kai-unexisting-node-pool"

	defaultAnnotationNodepoolsKey = "kai.scheduler/node-pools"

	// defaultEnforceSchedulerAnnotationKey mirrors project-controller's own default for the
	// annotation it writes on a project's namespace (see project-controller's
	// --enforce-scheduler-annotation-key). The two MUST agree: project-controller writes the
	// annotation from Project.Spec.EnforceKaiScheduler, and the pod mutator reads it back.
	defaultEnforceSchedulerAnnotationKey = "kai.resources/enforce-scheduler-name"
)

// PodGroupAssignerConfig holds the binary's flag-configurable values.
// It is scoped to the label/sentinel/default-name vocabulary that the
// assigner and mutator read at runtime.
type PodGroupAssignerConfig struct {
	// NodePoolLabelKey is the label key used to associate a Kubernetes resource
	// (PodGroup, Queue, NodePool selector) with a specific node pool. The mutating
	// webhook writes this label on incoming PodGroups (initially with the
	// UnexistingNodepoolSentinel value); the assigner later overwrites it with
	// the assigned node pool name. The same key is also used to match Queues
	// to a node pool when computing the queue name for a podgroup.
	NodePoolLabelKey string

	// QueueLabelKey is the label key that the assigner WRITES on a PodGroup to
	// record the queue assigned to that workload (the "workload side" of the
	// queue assignment). It is set by the assigner — never read from existing
	// labels — and is distinct from the namespace-side project label below.
	QueueLabelKey string

	// NamespaceProjectLabelKey is the label key READ from a Namespace object
	// to determine the project name a namespace belongs to. The assigner uses
	// this when a PodGroup has no project label of its own and we need to
	// derive the project from its namespace.
	//
	// Historically this was the *same* constant as QueueLabelKey (runai/queue),
	// because the same label was overloaded for two unrelated purposes:
	// (a) the workload's queue name and (b) the namespace's project name.
	// Splitting this into its own flag is the central change —
	// in runai deployments both flags still point at runai/queue (preserving
	// behavior), but in the OSS package they default to different keys.
	NamespaceProjectLabelKey string

	// ProjectLabelKey is the label key READ from a PodGroup to determine the
	// workload's project name. This is the workload-side project label
	// (the primary source); if absent, the assigner falls back to
	// NamespaceProjectLabelKey on the workload's namespace. It is also used
	// when matching Queues by project label.
	ProjectLabelKey string

	// UnexistingNodepoolSentinel is the label VALUE (not a key) written by the
	// mutating webhook into the NodePoolLabelKey slot when a PodGroup is
	// created without an explicit node pool assignment. It serves as a
	// placeholder that signals "this PodGroup is not yet assigned to any
	// real node pool", letting downstream controllers distinguish "no label"
	// from "intentionally unassigned and pending placement".
	UnexistingNodepoolSentinel string

	// DefaultNodepoolName is the name that identifies the implicit "default"
	// node pool — the bucket for workloads that don't carry an explicit
	// NodePoolLabelKey label. When the assigner is computing the queue name
	// for the default node pool, it matches Queues that LACK the node-pool
	// label (rather than queues that match a literal node pool name), since
	// the default pool is represented by the absence of a label.
	DefaultNodepoolName string

	// AnnotationNodepoolsKey is the annotation key (on Pods) whose value is a
	// space-separated list of explicitly requested node pool names. When set,
	// it takes precedence over affinity-, label-, and project-based node pool
	// resolution in the requested-nodepools converter.
	AnnotationNodepoolsKey string

	// EnforceSchedulerAnnotationKey is the annotation key READ from a Namespace to
	// decide whether the scheduler is enforced for every pod in it. project-controller
	// writes it from Project.Spec.EnforceKaiScheduler; the pod mutating webhook reads it
	// and, when the value parses as true, forces SchedulerName on incoming pods.
	// Must match project-controller's --enforce-scheduler-annotation-key.
	EnforceSchedulerAnnotationKey string

	// SchedulerName is the scheduler the pod mutating webhook writes into
	// pod.Spec.SchedulerName when enforcement applies. It is also the value the mutator
	// compares against to recognize a pod that already opted in to KAI, so it must match
	// the name the KAI scheduler itself runs under — every KAI component (podgrouper,
	// binder, admission hooks, scheduler cache) ignores pods scheduled by any other name.
	SchedulerName string
}

var (
	cfg     PodGroupAssignerConfig
	addOnce sync.Once
)

// AddLabelFlags binds the pod-group-assigner's label-vocabulary configurable
// values (label keys, the unassigned-nodepool sentinel, and the default
// nodepool name) to the given FlagSet. Other runtime flags (debug, ports,
// QPS, leader-election) are declared separately in main.go and do not go
// through this function. Idempotent — safe to call multiple times. Must be
// called before flag.Parse().
func AddLabelFlags(fs *flag.FlagSet) {
	addOnce.Do(func() {
		fs.StringVar(&cfg.NodePoolLabelKey, "nodepool-label-key",
			kaiconstants.DefaultNodePoolLabelKey,
			"Label key for node pools")
		fs.StringVar(&cfg.QueueLabelKey, "queue-label-key",
			kaiconstants.DefaultQueueLabel,
			"Label key for queue (workload side)")
		fs.StringVar(&cfg.NamespaceProjectLabelKey, "namespace-project-label-key",
			kaiv1alpha1.NamespaceProjectLabelKey,
			"Label key for the project name on a namespace")
		fs.StringVar(&cfg.ProjectLabelKey, "project-label-key",
			kaipgconstants.ProjectLabelKey,
			"Label key for the project name on a workload/podgroup")
		fs.StringVar(&cfg.UnexistingNodepoolSentinel, "unexisting-nodepool-sentinel",
			defaultUnexistingNodepoolSentinel,
			"Label value used to mark a podgroup as not yet assigned to any node pool")
		fs.StringVar(&cfg.DefaultNodepoolName, "default-nodepool-name",
			kaiconstants.DefaultNodePoolName,
			"Name used to identify the default node pool")
		fs.StringVar(&cfg.AnnotationNodepoolsKey, "annotation-nodepools-key",
			defaultAnnotationNodepoolsKey,
			"Annotation key on Pods whose value is a space-separated list of explicitly requested node pool names")
		fs.StringVar(&cfg.EnforceSchedulerAnnotationKey, "enforce-scheduler-annotation-key",
			defaultEnforceSchedulerAnnotationKey,
			"Annotation key on a Namespace marking that the scheduler is enforced for all of its pods. "+
				"Must match project-controller's --enforce-scheduler-annotation-key")
		fs.StringVar(&cfg.SchedulerName, "scheduler-name",
			kaiconstants.DefaultSchedulerName,
			"Scheduler name written on pods when scheduler enforcement applies")
	})
}

// Config returns the parsed package-level configuration. flag.Parse() must
// have been called before this is used.
func Config() *PodGroupAssignerConfig {
	return &cfg
}

// SetForTest replaces the global config and returns a restore closure.
// Tests only — production code must go through AddLabelFlags + flag.Parse.
func SetForTest(c PodGroupAssignerConfig) func() {
	prev := cfg
	cfg = c
	return func() { cfg = prev }
}
