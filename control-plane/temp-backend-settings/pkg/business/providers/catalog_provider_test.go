/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	catalog_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
)

type mockCatalogSettingsAPI struct {
	getCatalogSettingsResp *catalog_openapi.CatalogSettingsGetResponse
	getCatalogSettingsHTTP *http.Response
	getCatalogSettingsErr  error
}

func (m *mockCatalogSettingsAPI) GetCatalogSettings(_ context.Context) catalog_openapi.ApiGetCatalogSettingsRequest {
	return catalog_openapi.ApiGetCatalogSettingsRequest{ApiService: m}
}

func (m *mockCatalogSettingsAPI) GetCatalogSettingsExecute(_ catalog_openapi.ApiGetCatalogSettingsRequest) (*catalog_openapi.CatalogSettingsGetResponse, *http.Response, error) {
	return m.getCatalogSettingsResp, m.getCatalogSettingsHTTP, m.getCatalogSettingsErr
}

func (m *mockCatalogSettingsAPI) GetCatalogSettingByKey(_ context.Context, _ catalog_openapi.CatalogSettingsKeyEnum) catalog_openapi.ApiGetCatalogSettingByKeyRequest {
	return catalog_openapi.ApiGetCatalogSettingByKeyRequest{ApiService: m}
}

func (m *mockCatalogSettingsAPI) GetCatalogSettingByKeyExecute(_ catalog_openapi.ApiGetCatalogSettingByKeyRequest) (*catalog_openapi.GetEnableNgcIntegration, *http.Response, error) {
	return nil, nil, nil
}

func (m *mockCatalogSettingsAPI) UpdateCatalogSettingByKey(_ context.Context, _ catalog_openapi.CatalogSettingsKeyEnum) catalog_openapi.ApiUpdateCatalogSettingByKeyRequest {
	return catalog_openapi.ApiUpdateCatalogSettingByKeyRequest{ApiService: m}
}

func (m *mockCatalogSettingsAPI) UpdateCatalogSettingByKeyExecute(_ catalog_openapi.ApiUpdateCatalogSettingByKeyRequest) (*http.Response, error) {
	return nil, nil
}

func newNgcSettingWithEnabled(enabled bool) *catalog_openapi.GetEnableNgcIntegration {
	v := catalog_openapi.NewGetEnableNgcIntegrationValue()
	v.SetEnabled(enabled)
	s := catalog_openapi.NewGetEnableNgcIntegration()
	s.SetValue(*v)
	return s
}

func TestCatalogProvider_FetchSettings_EnabledTrue(t *testing.T) {
	mock := &mockCatalogSettingsAPI{
		getCatalogSettingsResp: &catalog_openapi.CatalogSettingsGetResponse{
			EnableNgcIntegration: newNgcSettingWithEnabled(true),
		},
	}
	client := &catalog_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewCatalogProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, settings.SettingKeyEnableNgcIntegration)
	assert.Equal(t, true, result[settings.SettingKeyEnableNgcIntegration].Value)
}

func TestCatalogProvider_FetchSettings_EnabledFalse(t *testing.T) {
	mock := &mockCatalogSettingsAPI{
		getCatalogSettingsResp: &catalog_openapi.CatalogSettingsGetResponse{
			EnableNgcIntegration: newNgcSettingWithEnabled(false),
		},
	}
	client := &catalog_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewCatalogProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, settings.SettingKeyEnableNgcIntegration)
	assert.Equal(t, false, result[settings.SettingKeyEnableNgcIntegration].Value)
}

func TestCatalogProvider_FetchSettings_WithMetadata(t *testing.T) {
	ngcSetting := newNgcSettingWithEnabled(true)
	ngcSetting.SetDisplayName("NGC Integration")
	ngcSetting.SetDescription("Enable NGC catalog integration")
	ngcSetting.SetStage("Beta")

	mock := &mockCatalogSettingsAPI{
		getCatalogSettingsResp: &catalog_openapi.CatalogSettingsGetResponse{
			EnableNgcIntegration: ngcSetting,
		},
	}
	client := &catalog_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewCatalogProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	override := result[settings.SettingKeyEnableNgcIntegration]
	assert.Equal(t, true, override.Value)
	assert.Equal(t, "NGC Integration", override.DisplayName)
	assert.Equal(t, "Enable NGC catalog integration", override.Description)
	assert.Equal(t, "Beta", override.Stage)
}

func TestCatalogProvider_FetchSettings_NilSetting(t *testing.T) {
	mock := &mockCatalogSettingsAPI{
		getCatalogSettingsResp: &catalog_openapi.CatalogSettingsGetResponse{
			EnableNgcIntegration: nil,
		},
	}
	client := &catalog_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewCatalogProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCatalogProvider_FetchSettings_UnsetEnabled(t *testing.T) {
	mock := &mockCatalogSettingsAPI{
		getCatalogSettingsResp: &catalog_openapi.CatalogSettingsGetResponse{
			EnableNgcIntegration: catalog_openapi.NewGetEnableNgcIntegration(),
		},
	}
	client := &catalog_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewCatalogProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCatalogProvider_FetchSettings_APIError(t *testing.T) {
	mock := &mockCatalogSettingsAPI{
		getCatalogSettingsErr: fmt.Errorf("connection refused"),
	}
	client := &catalog_openapi.APIClient{}
	client.SettingsAPI = mock
	provider := NewCatalogProvider(client)

	result, err := provider.FetchSettings(context.Background(), 0)

	assert.NoError(t, err, "provider should swallow API errors")
	assert.Empty(t, result)
}
