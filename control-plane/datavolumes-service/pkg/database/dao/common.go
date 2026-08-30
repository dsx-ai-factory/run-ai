package dao

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/services_utils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

const deleteBatchSize = 100

// DeleteDataVolumesByScope delete data volumes of a project
func DeleteDataVolumesByScope(ctx context.Context, scopeType api.ScopeType, scopeId string) error {
	db := runaidb.DB().WithContext(ctx)
	tenantId, _ := services_utils.ParseTenantId(ctx)

	scopeIdColumn, err := dbtables.ScopeIdColumnName(scopeType)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx orm.DB) error {
		for { // Loop until no more assets to delete
			var idsToDelete []string
			if err := tx.Table((&models.Datavolume{}).TableName()).
				Where(fmt.Sprintf("%s = ?", dbtables.DataVolumeColumnTenantId), tenantId).
				Where(fmt.Sprintf("%s = ?", scopeIdColumn), scopeId).
				Order(fmt.Sprintf("%s ASC", dbtables.DataVolumeColumnId)).
				Limit(deleteBatchSize). // Limit the number of records per query to avoid overloading the DB
				Pluck("id", &idsToDelete); err != nil {
				return fmt.Errorf("failed to fetch datavolume IDs: %w", err)
			}

			// Nothing to delete
			if len(idsToDelete) == 0 {
				return nil
			}

			//delete from shared-scopes table
			if _, err := tx.Where(fmt.Sprintf("%s IN (?)", dbtables.SharedScopesColumnId), idsToDelete).
				Unscoped().Delete(&models.SharedScope{}); err != nil {
				return fmt.Errorf("failed to delete shared-scopes: %w", err)
			}

			// Delete from data-volumes table
			if _, err := tx.
				Where(fmt.Sprintf("%s IN (?)", dbtables.DataVolumeColumnId), idsToDelete).
				Delete(&models.Datavolume{}); err != nil {
				return fmt.Errorf("failed to delete data-volumes: %w", err)
			}

		}
	})
}
