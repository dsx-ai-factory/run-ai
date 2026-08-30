package manipulate

const (
	DefaultWidthName           = 20
	DefaultWidthCategory       = 15
	DefaultWidthPriority       = 15
	DefaultWidthPreemptibility = 15
	DefaultWidthGroup          = 20
	DefaultWidthSource         = 10
)

const (
	nameColumnIndex = 0
	noDataFormat    = "-"
)

type WorkloadTypeColumn string

const (
	ColumnKeyName           WorkloadTypeColumn = "name"
	ColumnKeyCategory       WorkloadTypeColumn = "category"
	ColumnKeyPriority       WorkloadTypeColumn = "priority"
	ColumnKeyPreemptibility WorkloadTypeColumn = "preemptibility"
	ColumnKeyGroup          WorkloadTypeColumn = "group"
	ColumnKeySource         WorkloadTypeColumn = "source"
)
