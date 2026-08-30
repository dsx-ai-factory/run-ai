package metrics

type WorkloadSource string

const (
	PodNameLabel   = "pod_name"
	PodUUIDLabel   = "pod_uuid"
	NamespaceLabel = "pod_namespace"

	queueLabel      = "queue_name"
	nodepoolLabel   = "nodepool"
	projectLabel    = "project"
	departmentLabel = "department"

	WorkloadIdLabel       = "workload_id"
	workloadNameLabel     = "workload_name"
	workloadCategoryLabel = "workload_category"
	workloadTypeLabel     = "workload_type"

	workloadObjectNameLabel = "workload_object_name"
	workloadNamespaceLabel  = "workload_namespace"
	workloadSourceLabel     = "workload_source"

	ExternalSource             WorkloadSource = "external"
	TrainingSource             WorkloadSource = "training"
	DistributedSource          WorkloadSource = "distributed"
	InteractiveSource          WorkloadSource = "interactive"
	InferenceSource            WorkloadSource = "inference"
	DistributedInferenceSource WorkloadSource = "distributed_inference"

	elementIDLabel    = "element_id"
	elementLevelLabel = "element_level"

	nodeLabel          = "node"
	gpuIndexLabel      = "gpu_index"
	gpuLabel           = "gpu"
	gpuInstanceIdLabel = "gpu_instance_id"
	isGpuLabel         = "is_gpu"
	userLabel          = "user"

	// phase - will be either Pending, Running or Suspended
	phaseLabel = "phase"
	// detailed_status - will be the calculated status of the podgroup of the workload
	detailedStatusLabel = "detailed_status"
	// status - will be the workload.Status.Phase - which is either:
	// workloadcontroller_apicommon.WorkloadCommonStatus.Phase (Creating;Initializing;Running;Degraded;Pending;Stopped;Completed;Failed;Terminating;Unknown)
	// or:
	// v1alpha1.ExternalWorkloadPhase (Creating;Pending;Initializing;Running;Stopping;Stopped;Degraded;Failed;Completed;Unknown)
	statusLabel = "status"
)
