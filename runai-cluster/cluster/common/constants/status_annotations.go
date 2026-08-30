package constants

const (
	PodStatusAnnotation = "runai-calculated-status"
	PodGroupStatusAnnotation = PodStatusAnnotation

	PodAllocatedGpusAnnotation      = "runai-allocated-gpus"
	PodAllocatedGpuMemoryAnnotation = "runai-allocated-gpu-memory"
	PodAllocatedMigGpusAnnotation   = "runai-allocated-mig-gpus"
)
