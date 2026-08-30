package stages

// WorkloadStage represents a part of the workload lifecycle.
type WorkloadStage string

const (
	// Since the user initiated the workload creation API until it got a response
	WorkloadStageSubmitted WorkloadStage = "Submitted"
	// Since the workload was submitted until it was created on the cluster
	WorkloadStageCreatedOnCluster WorkloadStage = "CreatedOnCluster"
	// Since the workload was created on the cluster until the first pod was created
	WorkloadStageFirstPodCreated WorkloadStage = "FirstPodCreated"
	// Since the first pod was created until it was bound to a node
	WorkloadStageFirstPodBound WorkloadStage = "FirstPodBound"
	// Since the workload changed its phase in the cluster until its phase was synced to the control plane
	WorkloadStagePhaseChangeSynced WorkloadStage = "PhaseChangeSynced"
	// Since the workload was created until it reached a terminal phase on the control plane
	WorkloadStageWholeLifecycle WorkloadStage = "WholeLifecycle"
)
