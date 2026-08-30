package manipulate

const (
	DefaultWidth                 = 10
	DefaultWidthPreemptible      = 15
	DefaultWidthTotalPendingTime = 22
	DefaultWidthMemoryAllocation = 18
	DefaultWidthGpuAllocation    = 15
)

const (
	nameColumnIndex        = 0
	timeFieldsLayout       = "2006-01-02T15:04:05Z07:00"
	RunaiStandardFramework = "Runai"
	noDataFormat           = "-"
)

// Column key constants

type WorkloadColumn string

const (
	ColumnKeyName                WorkloadColumn = "name"
	ColumnKeyType                WorkloadColumn = "type"
	ColumnKeyFramework           WorkloadColumn = "framework"
	ColumnKeyStatus              WorkloadColumn = "status"
	ColumnKeyProject             WorkloadColumn = "project"
	ColumnKeyCategory            WorkloadColumn = "category"
	ColumnKeyPriority            WorkloadColumn = "priority"
	ColumnKeyDepartment          WorkloadColumn = "department"
	ColumnKeySubmittedBy         WorkloadColumn = "submittedby"
	ColumnKeyCreatedAt           WorkloadColumn = "createdat"
	ColumnKeyTotalPendingTime    WorkloadColumn = "totalpendingtimeseconds"
	ColumnKeyAllocatedGPUMemory  WorkloadColumn = "allocatedgpumemory"
	ColumnKeyAllocatedCPUCompute WorkloadColumn = "allocatedcpucompute"
	ColumnKeyAllocatedCPUMemory  WorkloadColumn = "allocatedcpumemory"
	ColumnKeyCompletedAt         WorkloadColumn = "completedat"
	ColumnKeyUrls                WorkloadColumn = "urls"
	ColumnKeyPreemptible         WorkloadColumn = "preemptible"
	ColumnKeyPods                WorkloadColumn = "pods"
	ColumnKeyGPUAlloc            WorkloadColumn = "allocatedgpu"
	ColumnKeyImage               WorkloadColumn = "image"
	ColumnKeyCurrentNodepool     WorkloadColumn = "currentnodepool"
	ColumnKeyRequestedNodepool   WorkloadColumn = "requestednodepool"
)
