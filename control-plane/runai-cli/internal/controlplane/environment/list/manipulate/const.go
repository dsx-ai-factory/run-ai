package manipulate

const (
	DefaultWidth              = 15
	DefaultWidthImage         = 30
	DefaultWidthTools         = 30
	DefaultWidthWorkloadTypes = 35
)

const (
	nameColumnIndex = 0
	noDataFormat    = "-"
)

// Column key constants
type EnvironmentColumn string

const (
	ColumnKeyName          EnvironmentColumn = "name"
	ColumnKeyScope         EnvironmentColumn = "scope"
	ColumnKeyImage         EnvironmentColumn = "image"
	ColumnKeyTools         EnvironmentColumn = "tools"
	ColumnKeyWorkloadTypes EnvironmentColumn = "workloadtypes"
)
