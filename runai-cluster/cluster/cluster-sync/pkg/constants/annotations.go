package constants

const (
	WorkloadCalculatedStatus           = "runai-calculated-status"
	JobID                              = "runai-job-id"
	WorkloadRunningPods                = "runai-running-pods"
	WorkloadPendingPods                = "runai-pending-pods"
	WorkloadUsedNodes                  = "runai-used-nodes"
	PodGroupRequestedGPUs              = "runai-podgroup-requested-gpus"
	PodGroupRequestedGPUsMemory        = "runai-podgroup-requested-gpus-memory"
	WorkloadCurrentAllocatedGPUs       = "runai-current-allocated-gpus"
	WorkloadCurrentAllocatedGPUsMemory = "runai-current-allocated-gpus-memory"
	WorkloadCurrentRequestedGPUs       = "runai-current-requested-gpus"
	WorkloadCurrentRequestedGPUsMemory = "runai-current-requested-gpus-memory"
	WorkloadTotalRequestedGPUs         = "runai-total-requested-gpus"
	WorkloadCliV2Command               = "runai-submit-command"
	PodGpuMemoryRequest                = "gpu-memory"
	PodGpuFractionRequest              = "gpu-fraction"
	PodAllocatedGPUs                   = "runai-allocated-gpus"
	PodAllocatedGpuMemory              = "runai-allocated-gpu-memory"
	PodAllocatedMigGPUs                = "runai-allocated-mig-gpus"
	PodGroupNameAnnotation             = "pod-group-name"
)

const MigResourcePrefix = "nvidia.com/mig-"
