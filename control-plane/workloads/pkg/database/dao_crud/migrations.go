package dao_crud

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"
)

func GetMigrationsInfo(ctx context.Context) (*models_crud.Migrations, *api.Error) {
	tenantId := int64(auth.TenantIdOf(ctx))

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get migration info of tenant '%d'", tenantId), nil)

	db := dao_utils.DB().WithContext(ctx)

	rec := models_crud.Migrations{
		TenantId: tenantId,
	}

	result := db.Model(rec).Where(rec).Scan(&rec)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return nil, dbErr.WithDetailsFromDb(result)
	}
	if result.RowsAffected == 0 {
		return &models_crud.Migrations{
			TenantId:              tenantId,
			AssetsToFlatTemplates: 0,
		}, nil
	}

	return &rec, nil
}

func SetAssetsToFlatMigration(ctx context.Context, count int) *api.Error {
	tenantId := int64(auth.TenantIdOf(ctx))

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get migration info of tenant '%d'", tenantId), nil)

	db := dao_utils.DB().WithContext(ctx)

	rec := models_crud.Migrations{
		TenantId: tenantId,
	}

	result := db.Model(rec).Where(rec).Scan(&rec)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: load: %v", dbErr.Message, result.Error)
		return dbErr.WithDetailsFromDb(result)
	}

	rec.AssetsToFlatTemplates = int32(count)

	result = db.Save(&rec)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: save: %v", dbErr.Message, result.Error)
		return dbErr.WithDetailsFromDb(result)
	}

	return nil
}
