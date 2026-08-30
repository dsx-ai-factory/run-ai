package clients

import (
	"context"
	"encoding/json"
	"fmt"

	temp_backend_settings_client "github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/client"
	settings_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	settings_pkg_manager "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// BackendClient wraps settings_pkg.BackendClient to provide a local implementation
// backed by the temp-backend-settings service for the settings package initialization.
type BackendClient interface {
	settings_pkg.BackendClient
}

// NewBackendClient creates a new BackendClient that connects to the given base URL.
func NewBackendClient(baseURL string) BackendClient {
	return &defaultBackendClient{
		client: temp_backend_settings_client.NewTempBackendSettingsClient(baseURL),
	}
}

type defaultBackendClient struct {
	client *temp_backend_settings_client.TempBackendSettingsClient
}

func (c *defaultBackendClient) GetSettingByKeyForAllTenants(ctx context.Context, key string) (*settings_pkg_manager.SettingValuesForAllTenantsBackendResponse, error) {
	result, err := c.client.GetAllTenantsSettingByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var response settings_pkg_manager.SettingValuesForAllTenantsBackendResponse
	if err := json.Unmarshal(jsonBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal backend response: %w", err)
	}
	return &response, nil
}
