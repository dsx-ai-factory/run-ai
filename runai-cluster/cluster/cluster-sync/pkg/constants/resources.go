package constants

type ClusterResource string

const (
	PodResourceName                          ClusterResource = "pod"
	AccessRuleResourceName                   ClusterResource = "accessrule"
	ClusterInfoResourceName                  ClusterResource = "clusterinfo"
	NodePoolResourceName                     ClusterResource = "nodepool"
	DepartmentResourceName                   ClusterResource = "department"
	DeploymentResourceName                   ClusterResource = "deployment"
	DataVolumeResourceName                   ClusterResource = "datavolume"
	DistributedWorkloadResourceName          ClusterResource = "distributedworkload"
	EventsResourceName                       ClusterResource = "events"
	ProjectResourceName                      ClusterResource = "project"
	NodeResourceName                         ClusterResource = "node"
	NumaNodeResourceName                     ClusterResource = "numaNode"
	RevisionResourceName                     ClusterResource = "revision"
	PriorityClassResourceName                ClusterResource = "priorityclass"
	DistributedInferenceWorkloadResourceName ClusterResource = "distributedinferenceworkload"
	NetworkTopologyResourceName              ClusterResource = "networktopology"
	ApplicationResourceName                  ClusterResource = "application"
	KartaResourceName                        ClusterResource = "karta"
)
