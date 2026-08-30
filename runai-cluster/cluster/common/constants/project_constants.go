package constants

const (
	ProjectScopeLabel = "run.ai/project"
	// DepartmentScopeLabel - resource (asset sync or user owned)
	//          needs to be replicated to all the projects of the department whose id is given in the value
	DepartmentScopeLabel = "run.ai/department"
	// ClusterWideResourceLabel - customer-owned secret that
	//          needs to be replicated to all the projects in the cluster
	ClusterWideResourceLabel = "run.ai/cluster-wide"
	// TenantScopeLabel - resource associated with a tenant scope assets.
	//          to be replicated to all projects in this cluster
	TenantScopeLabel = "run.ai/tenant-wide"

	DefaultQueuePriority = 100
)
