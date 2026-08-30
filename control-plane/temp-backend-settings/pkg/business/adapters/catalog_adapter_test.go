package adapters

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	catalog_openapi "github.com/run-ai/runai/runai-clients/catalog-service-client/generated"
)

type mockCatalogSettingsAPI struct {
	updateResp *http.Response
	updateErr  error
	updateKey  catalog_openapi.CatalogSettingsKeyEnum
}

func (m *mockCatalogSettingsAPI) GetCatalogSettings(_ context.Context) catalog_openapi.ApiGetCatalogSettingsRequest {
	return catalog_openapi.ApiGetCatalogSettingsRequest{ApiService: m}
}

func (m *mockCatalogSettingsAPI) GetCatalogSettingsExecute(_ catalog_openapi.ApiGetCatalogSettingsRequest) (*catalog_openapi.CatalogSettingsGetResponse, *http.Response, error) {
	return nil, nil, nil
}

func (m *mockCatalogSettingsAPI) GetCatalogSettingByKey(_ context.Context, _ catalog_openapi.CatalogSettingsKeyEnum) catalog_openapi.ApiGetCatalogSettingByKeyRequest {
	return catalog_openapi.ApiGetCatalogSettingByKeyRequest{ApiService: m}
}

func (m *mockCatalogSettingsAPI) GetCatalogSettingByKeyExecute(_ catalog_openapi.ApiGetCatalogSettingByKeyRequest) (*catalog_openapi.GetEnableNgcIntegration, *http.Response, error) {
	return nil, nil, nil
}

func (m *mockCatalogSettingsAPI) UpdateCatalogSettingByKey(_ context.Context, key catalog_openapi.CatalogSettingsKeyEnum) catalog_openapi.ApiUpdateCatalogSettingByKeyRequest {
	m.updateKey = key
	return catalog_openapi.ApiUpdateCatalogSettingByKeyRequest{ApiService: m}
}

func (m *mockCatalogSettingsAPI) UpdateCatalogSettingByKeyExecute(_ catalog_openapi.ApiUpdateCatalogSettingByKeyRequest) (*http.Response, error) {
	return m.updateResp, m.updateErr
}

func TestCatalogAdapter_UpdateEnableNgcIntegration_Enable(t *testing.T) {
	mock := &mockCatalogSettingsAPI{updateResp: &http.Response{StatusCode: http.StatusOK}}
	adapter := &CatalogAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableNgcIntegration(context.Background(), testTenantID, true)

	assert.NoError(t, err)
	assert.Equal(t, catalog_openapi.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, mock.updateKey)
}

func TestCatalogAdapter_UpdateEnableNgcIntegration_Disable(t *testing.T) {
	mock := &mockCatalogSettingsAPI{updateResp: &http.Response{StatusCode: http.StatusOK}}
	adapter := &CatalogAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableNgcIntegration(context.Background(), testTenantID, false)

	assert.NoError(t, err)
	assert.Equal(t, catalog_openapi.CATALOGSETTINGSKEYENUM_ENABLE_NGC_INTEGRATION, mock.updateKey)
}

func TestCatalogAdapter_UpdateEnableNgcIntegration_APIError(t *testing.T) {
	mock := &mockCatalogSettingsAPI{updateErr: fmt.Errorf("connection refused")}
	adapter := &CatalogAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableNgcIntegration(context.Background(), testTenantID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateEnableNgcIntegration")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestCatalogAdapter_UpdateEnableNgcIntegration_ServerError(t *testing.T) {
	mock := &mockCatalogSettingsAPI{updateResp: &http.Response{StatusCode: http.StatusInternalServerError}}
	adapter := &CatalogAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableNgcIntegration(context.Background(), testTenantID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestCatalogAdapter_UpdateEnableNgcIntegration_InvalidStringValue(t *testing.T) {
	mock := &mockCatalogSettingsAPI{}
	adapter := &CatalogAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableNgcIntegration(context.Background(), testTenantID, "not-a-bool")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected bool")
}

func TestCatalogAdapter_UpdateEnableNgcIntegration_NilValue(t *testing.T) {
	mock := &mockCatalogSettingsAPI{}
	adapter := &CatalogAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableNgcIntegration(context.Background(), testTenantID, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected bool")
}
