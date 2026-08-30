package manipulate

type ComputeColumn string

const (
	ColumnKeyName       ComputeColumn = "name"
	ColumnKeyScope      ComputeColumn = "scope"
	ColumnKeyGPUDevices ComputeColumn = "gpudevices"
	ColumnKeyGPUMemory  ComputeColumn = "gpumemory"
	ColumnKeyCPUMemory  ComputeColumn = "cpumemory"
)

var DefaultColumns = []ComputeColumn{
	ColumnKeyName,
	ColumnKeyScope,
	ColumnKeyGPUDevices,
	ColumnKeyGPUMemory,
	ColumnKeyCPUMemory,
}
