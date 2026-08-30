package manipulate

type DatavolumeColumn string

const (
	ColumnKeyName      DatavolumeColumn = "name"
	ColumnKeyStatus    DatavolumeColumn = "status"
	ColumnKeyScope     DatavolumeColumn = "scope"
	ColumnKeyOriginPvc DatavolumeColumn = "originpvc"
)

const (
	DefaultWidth = 20
)
