package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func validateNotInSync(ctx context.Context, tx orm.DB, existingMeta *dbmodel.AssetMetaRec) *api.Error {
	if existingMeta == nil {
		return nil
	}

	//
	//   system scoped assets no need to validate (they do not support rename)
	//
	if existingMeta.TenantId == nil {
		return nil
	}

	//
	//   currently we do not sync tenant scoped assets, but if this changes this function should consider
	//   the sync status in all the cluster rather than checking status in specific cluster
	//
	if existingMeta.ClusterId == nil {
		return nil
	}

	info, err := GetSyncStatusByClusterAndAssetId(ctx, tx, *existingMeta.TenantId, *existingMeta.ClusterId, existingMeta.Id)
	if err != nil {
		// if no sync info for this asset it means that even the create did not complete.
		if err.Code == http.StatusNotFound {
			return api.NewError(http.StatusPreconditionFailed, fmt.Sprintf("%s asset rename is not allowed while the asset is pending to be synchronized to the cluster. try again later", existingMeta.Kind), nil)
		}

		// failure to get the sync just log the error and allow the rename, we do not want to fail a rename just
		// because we could not get the info about sync
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get sync status for asset %s: %v", existingMeta.Id, err.Message)
		return nil
	}

	maxSent := max(info.SentToAgentDelete, info.SentToAgentApply)

	if existingMeta.UpdateCount > maxSent || maxSent != info.ProcessedByAgent {
		return api.NewError(http.StatusPreconditionFailed, fmt.Sprintf("%s asset rename is not allowed while the asset is being synchronized to the cluster. try again later", existingMeta.Kind), nil)
	}

	if info.Message != "" {
		return api.NewError(http.StatusPreconditionFailed, fmt.Sprintf(" %s asset rename is not allowed due to synchronization error. try again later", existingMeta.Kind), nil)
	}

	return nil
}
