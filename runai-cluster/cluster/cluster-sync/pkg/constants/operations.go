package constants

type OperationName string

const (
	UpdateOperation OperationName = "update"
	SyncOperation   OperationName = "sync"
	DeleteOperation OperationName = "delete"
	AddOperation    OperationName = "add"
	UpsertOperation OperationName = "upsert"
)
