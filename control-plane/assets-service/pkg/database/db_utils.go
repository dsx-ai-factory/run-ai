package database

import (
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

// typically, joined tables are assumed to hold their asset id in column asset_id but there are some exceptions
// that the function will have to support
var assetIdColumns = map[string]string{
	"policy_rules": "policy_id",
}

func JoinAssetMetaAndSpec(tableName string) string {
	assetIdColumn := assetIdColumns[tableName]
	if assetIdColumn == "" {
		assetIdColumn = "asset_id"
	}
	return fmt.Sprintf("left join %s on assets_metadata.id = %s.%s", tableName, tableName, assetIdColumn)
}

type DbOperation int

const (
	GET DbOperation = iota
	LIST
	COUNT
	CREATE
	UPDATE
	DELETE
	UPSERT
	COMMIT_CREATIONS
	COMMIT_UPDATES
	COMMIT_DELETIONS
)

func (op DbOperation) String() string {
	return [...]string{
		OperationGet,
		OperationList,
		OperationCount,
		OperationCreate,
		OperationUpdate,
		OperationDelete,
		OperationUpsert,
		OperationCommitCreations,
		OperationCommitUpdates,
		OperationCommitDeletions,
	}[op]
}

func Error(operation DbOperation, assetKind api.AssetKind) *api.Error {
	msg := fmt.Sprintf("failed to %s requested %s asset", operation, assetKind)
	return api.NewError(http.StatusServiceUnavailable, msg, nil)
}

func ErrorWithCode(operation DbOperation, assetKind api.AssetKind, errorCode int) *api.Error {
	msg := fmt.Sprintf("failed to %s requested %s asset", operation, assetKind)
	return api.NewError(errorCode, msg, nil)
}
