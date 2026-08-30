package adapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/autorization"
	dv_openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

type DatavolumesAdapter struct {
	SettingsAPI dv_openapi.SettingsAPI
}

func (a *DatavolumesAdapter) UpdateEnableDataVolumes(ctx context.Context, tenantID int, value interface{}) error {
	boolVal, ok := value.(bool)
	if !ok {
		return fmt.Errorf("UpdateEnableDataVolumes: expected bool, got %T", value)
	}

	body := dv_openapi.UpdateDatavolumesSettingByKeyRequest{}
	body.SetEnabled(boolVal)

	req := a.SettingsAPI.UpdateDatavolumesSettingByKey(ctx, dv_openapi.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES).
		UpdateDatavolumesSettingByKeyRequest(body)
	if autorization.IsManagementTenant(ctx) {
		req = req.TenantId(strconv.Itoa(tenantID))
	}
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("UpdateEnableDataVolumes: %w", err)
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("UpdateEnableDataVolumes returned status %d", resp.StatusCode)
	}
	return nil
}
