package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	bff_openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
)

type BffWorkloadsAdapter struct {
	WorkloadsSettingsAPI bff_openapi.SettingsAPI
}

func (a *BffWorkloadsAdapter) UpdateDeletedWorkloadsView(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateBoolSetting(ctx, tenantID, bff_openapi.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW, value)
}

func (a *BffWorkloadsAdapter) UpdateAllowGPT2Model(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateBoolSetting(ctx, tenantID, bff_openapi.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL, value)
}

func (a *BffWorkloadsAdapter) UpdateEnableModelCatalog(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateBoolSetting(ctx, tenantID, bff_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG, value)
}

func (a *BffWorkloadsAdapter) UpdateEnableNIM(ctx context.Context, tenantID int, value interface{}) error {
	return a.updateBoolSetting(ctx, tenantID, bff_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM, value)
}

func (a *BffWorkloadsAdapter) updateBoolSetting(ctx context.Context, tenantID int, key bff_openapi.WorkloadsSettingsKeyEnum, value interface{}) error {
	boolVal, _ := value.(bool)
	body := bff_openapi.UpdateWorkloadsSettingByKeyRequest{}

	switch key {
	case bff_openapi.WORKLOADSSETTINGSKEYENUM_DELETED_WORKLOADS_VIEW:
		body.UpdateDeletedWorkloadsView = &bff_openapi.UpdateDeletedWorkloadsView{}
		body.UpdateDeletedWorkloadsView.SetEnabled(boolVal)
	case bff_openapi.WORKLOADSSETTINGSKEYENUM_ALLOW_GPT2_MODEL:
		body.UpdateAllowGPT2Model = &bff_openapi.UpdateAllowGPT2Model{}
		body.UpdateAllowGPT2Model.SetEnabled(boolVal)
	case bff_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_MODEL_CATALOG:
		body.UpdateEnableModelCatalog = &bff_openapi.UpdateEnableModelCatalog{}
		body.UpdateEnableModelCatalog.SetEnabled(boolVal)
	case bff_openapi.WORKLOADSSETTINGSKEYENUM_ENABLE_NIM:
		body.UpdateEnableNIM = &bff_openapi.UpdateEnableNIM{}
		body.UpdateEnableNIM.SetEnabled(boolVal)
	}

	req := a.WorkloadsSettingsAPI.UpdateWorkloadsSettingByKey(ctx, key).
		UpdateWorkloadsSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateWorkloadsSettingByKey(%s): %w", key, err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateWorkloadsSettingByKey(%s) returned status %d", key, resp.StatusCode)
	}
	return nil
}
