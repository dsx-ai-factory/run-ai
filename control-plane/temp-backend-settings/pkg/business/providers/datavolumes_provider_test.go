package providers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	dv_openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

type mockDatavolumesSettingsAPI struct {
	getSettingsResp *dv_openapi.DatavolumesSettingsGetResponse
	getSettingsHTTP *http.Response
	getSettingsErr  error
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettings(_ context.Context) dv_openapi.ApiGetDatavolumesSettingsRequest {
	return dv_openapi.ApiGetDatavolumesSettingsRequest{ApiService: m}
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettingsExecute(_ dv_openapi.ApiGetDatavolumesSettingsRequest) (*dv_openapi.DatavolumesSettingsGetResponse, *http.Response, error) {
	return m.getSettingsResp, m.getSettingsHTTP, m.getSettingsErr
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettingByKey(_ context.Context, _ dv_openapi.DatavolumesSettingsKeyEnum) dv_openapi.ApiGetDatavolumesSettingByKeyRequest {
	return dv_openapi.ApiGetDatavolumesSettingByKeyRequest{ApiService: m}
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettingByKeyExecute(_ dv_openapi.ApiGetDatavolumesSettingByKeyRequest) (*dv_openapi.GetEnableDataVolumes, *http.Response, error) {
	return nil, nil, nil
}

func (m *mockDatavolumesSettingsAPI) UpdateDatavolumesSettingByKey(_ context.Context, _ dv_openapi.DatavolumesSettingsKeyEnum) dv_openapi.ApiUpdateDatavolumesSettingByKeyRequest {
	return dv_openapi.ApiUpdateDatavolumesSettingByKeyRequest{ApiService: m}
}

func (m *mockDatavolumesSettingsAPI) UpdateDatavolumesSettingByKeyExecute(_ dv_openapi.ApiUpdateDatavolumesSettingByKeyRequest) (*http.Response, error) {
	return nil, nil
}

func newDataVolumesSettingWithEnabled(enabled bool) *dv_openapi.GetEnableDataVolumes {
	v := dv_openapi.NewGetEnableDataVolumesValue()
	v.SetEnabled(enabled)
	s := dv_openapi.NewGetEnableDataVolumes()
	s.SetValue(*v)
	return s
}

func TestDatavolumesProvider_FetchSettings_EnabledTrue(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{
		getSettingsResp: &dv_openapi.DatavolumesSettingsGetResponse{
			EnableDataVolumes: newDataVolumesSettingWithEnabled(true),
		},
	}
	client := &dv_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewDatavolumesProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, settings.SettingKeyEnableDataVolumes)
	assert.Equal(t, true, result[settings.SettingKeyEnableDataVolumes].Value)
}

func TestDatavolumesProvider_FetchSettings_EnabledFalse(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{
		getSettingsResp: &dv_openapi.DatavolumesSettingsGetResponse{
			EnableDataVolumes: newDataVolumesSettingWithEnabled(false),
		},
	}
	client := &dv_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewDatavolumesProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, settings.SettingKeyEnableDataVolumes)
	assert.Equal(t, false, result[settings.SettingKeyEnableDataVolumes].Value)
}

func TestDatavolumesProvider_FetchSettings_WithMetadata(t *testing.T) {
	dvSetting := newDataVolumesSettingWithEnabled(true)
	dvSetting.SetDisplayName("Data Volumes")
	dvSetting.SetDescription("Enable data volumes feature")
	dvSetting.SetStage("GA")

	mock := &mockDatavolumesSettingsAPI{
		getSettingsResp: &dv_openapi.DatavolumesSettingsGetResponse{
			EnableDataVolumes: dvSetting,
		},
	}
	client := &dv_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewDatavolumesProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	override := result[settings.SettingKeyEnableDataVolumes]
	assert.Equal(t, true, override.Value)
	assert.Equal(t, "Data Volumes", override.DisplayName)
	assert.Equal(t, "Enable data volumes feature", override.Description)
	assert.Equal(t, "GA", override.Stage)
}

func TestDatavolumesProvider_FetchSettings_NilSetting(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{
		getSettingsResp: &dv_openapi.DatavolumesSettingsGetResponse{
			EnableDataVolumes: nil,
		},
	}
	client := &dv_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewDatavolumesProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDatavolumesProvider_FetchSettings_UnsetEnabled(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{
		getSettingsResp: &dv_openapi.DatavolumesSettingsGetResponse{
			EnableDataVolumes: dv_openapi.NewGetEnableDataVolumes(),
		},
	}
	client := &dv_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewDatavolumesProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDatavolumesProvider_FetchSettings_APIError(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{
		getSettingsErr: fmt.Errorf("connection refused"),
	}
	client := &dv_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewDatavolumesProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err, "provider should swallow API errors")
	assert.Empty(t, result)
}
