package metadata

type PodMetadata struct {
	Status         string
	GPUs           float64
	GpuMemory      int64
	FractionsCount int64
	MigGPUs        float64
}
