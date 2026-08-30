package database

const (
	OperationList   = "list"
	OperationGet    = "get"
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationCount  = "count"
	OperationUpsert = "upsert"

	OperationSync = "sync"

	OperationCommitCreations = "commit/creations"
	OperationCommitUpdates   = "commit/updates"
	OperationCommitDeletions = "commit/deletions"
)
