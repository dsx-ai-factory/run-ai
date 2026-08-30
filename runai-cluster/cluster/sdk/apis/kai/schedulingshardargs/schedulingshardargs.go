// Package schedulingshardargs documents the KAI scheduler arguments that
// nodepool-controller writes into each SchedulingShard's spec.args
// (kai.scheduler/v1 SchedulingShard). These keys are KAI scheduler CLI flag
// names; KAI does not export them, so they are mirrored here as the accepted
// --scheduling-shard-args contract. It lives beside the KAI NodePool CRD types so the
// contract and the CRD migrate together. Producers (the enterprise go-operator,
// and the open-source Helm chart) must use these keys so the values land on the
// flags the KAI scheduler actually reads.
package schedulingshardargs

const (
	// Base args set by the controller on every shard. The key is the KAI flag
	// name; the value is the node label the scheduler matches on.
	CPUWorkerNodeLabelKey = "cpu-worker-node-label-key"
	GPUWorkerNodeLabelKey = "gpu-worker-node-label-key"
	MIGWorkerNodeLabelKey = "mig-worker-node-label-key"

	// Cluster-wide args supplied via --scheduling-shard-args.
	RestrictNodeScheduling      = "restrict-node-scheduling"
	DetailedFitErrors           = "detailed-fit-errors"
	ScheduleCSIStorage          = "schedule-csi-storage"
	EnableProfiler              = "enable-profiler"
	Verbosity                   = "v"
	SchedulerName               = "scheduler-name"
	SchedulePeriod              = "schedule-period"
	MaxConsolidationPreemptees  = "max-consolidation-preemptees"
	QPS                         = "qps"
	Burst                       = "burst"
	UseSchedulingSignatures     = "use-scheduling-signatures"
	FullHierarchyFairness       = "full-hierarchy-fairness"
	AllowConsolidatingReclaim   = "allow-consolidating-reclaim"
	DefaultStalenessGracePeriod = "default-staleness-grace-period"
	NumOfStatusRecordingWorkers = "num-of-status-recording-workers"
)
