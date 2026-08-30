package clients

import (
	"context"

	settings_pkg_manager "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type MockBackendClient struct{}

func (m *MockBackendClient) GetSettingByKeyForAllTenants(ctx context.Context,
	key string) (*settings_pkg_manager.SettingValuesForAllTenantsBackendResponse, error) {
	return &settings_pkg_manager.SettingValuesForAllTenantsBackendResponse{}, nil
}
