package manipulate

const (
	DefaultWidthName       = 20
	DefaultWidthType       = 15
	DefaultWidthStatus     = 15
	DefaultWidthProject    = 15
	DefaultWidthGPUCompute = 18
	DefaultWidthGPUMemory  = 18
	DefaultWidthCPUCompute = 18
	DefaultWidthCPUMemory  = 18
)

const (
	noDataFormat = "-"
)

type AIApplicationColumn string

const (
	ColumnKeyName            AIApplicationColumn = "name"
	ColumnKeyType            AIApplicationColumn = "type"
	ColumnKeyStatus          AIApplicationColumn = "status"
	ColumnKeyProject         AIApplicationColumn = "project"
	ColumnKeyGPUComputeReq   AIApplicationColumn = "gpucomputereq"
	ColumnKeyGPUComputeAlloc AIApplicationColumn = "gpucomputealloc"
	ColumnKeyGPUMemoryReq    AIApplicationColumn = "gpumemoryreq"
	ColumnKeyGPUMemoryAlloc  AIApplicationColumn = "gpumemoryalloc"
	ColumnKeyCPUComputeReq   AIApplicationColumn = "cpucomputereq"
	ColumnKeyCPUComputeAlloc AIApplicationColumn = "cpucomputealloc"
	ColumnKeyCPUMemoryReq    AIApplicationColumn = "cpumemoryreq"
	ColumnKeyCPUMemoryAlloc  AIApplicationColumn = "cpumemoryalloc"
)
