package dao_crud

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models_crud"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
	"net/http"
)

/* WAIT_FOR_OFER
func getTemplatesPresetsOfTenant(ctx context.Context, tx *gorm.DB, tenantId uint, presetTemplateId string) (*models_crud.TemplatesPresets, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get templates presets of tenant %v preset %s", tenantId, presetTemplateId), nil)

	key := models_crud.TemplatesPresets{
		TenantId:         tenantId,
		PresetTemplateId: presetTemplateId,
	}

	var presetsInfo models_crud.TemplatesPresets

	result := tx.Model(&key).Where(key).Scan(&presetsInfo)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return nil, dbErr.WithDetailsFromDb(result)
	}
	if result.RowsAffected == 0 {
		return nil, api.NewError(http.StatusNotFound, "no forked template found", nil)
	}

	return &presetsInfo, nil
}
*/

func saveTemplatesPresetsOfTenant(ctx context.Context, tx *gorm.DB, presetsInfo *models_crud.TemplatesPresets) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to store templates presets of tenant %v preset %s",
		presetsInfo.TenantId, presetsInfo.PresetTemplateId), nil)

	result := tx.Model(presetsInfo).Save(&presetsInfo)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, result.Error)
		return dbErr.WithDetailsFromDb(result)
	}

	return nil
}
