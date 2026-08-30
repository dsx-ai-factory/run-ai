package api

type SubresourceComponent string

const (
	SpecSubResource       SubresourceComponent = "spec"
	WorkerSpecSubResource SubresourceComponent = "workerSpec"
	MasterSpecSubResource SubresourceComponent = "masterSpec"
	LeaderSpecSubResource SubresourceComponent = "leaderSpec"
)
