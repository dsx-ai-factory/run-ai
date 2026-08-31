/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package adapters

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	dv_openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
)

const testTenantID = 42

type mockDatavolumesSettingsAPI struct {
	updateResp *http.Response
	updateErr  error
	updateKey  dv_openapi.DatavolumesSettingsKeyEnum
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettings(_ context.Context) dv_openapi.ApiGetDatavolumesSettingsRequest {
	return dv_openapi.ApiGetDatavolumesSettingsRequest{ApiService: m}
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettingsExecute(_ dv_openapi.ApiGetDatavolumesSettingsRequest) (*dv_openapi.DatavolumesSettingsGetResponse, *http.Response, error) {
	return nil, nil, nil
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettingByKey(_ context.Context, _ dv_openapi.DatavolumesSettingsKeyEnum) dv_openapi.ApiGetDatavolumesSettingByKeyRequest {
	return dv_openapi.ApiGetDatavolumesSettingByKeyRequest{ApiService: m}
}

func (m *mockDatavolumesSettingsAPI) GetDatavolumesSettingByKeyExecute(_ dv_openapi.ApiGetDatavolumesSettingByKeyRequest) (*dv_openapi.GetEnableDataVolumes, *http.Response, error) {
	return nil, nil, nil
}

func (m *mockDatavolumesSettingsAPI) UpdateDatavolumesSettingByKey(_ context.Context, key dv_openapi.DatavolumesSettingsKeyEnum) dv_openapi.ApiUpdateDatavolumesSettingByKeyRequest {
	m.updateKey = key
	return dv_openapi.ApiUpdateDatavolumesSettingByKeyRequest{ApiService: m}
}

func (m *mockDatavolumesSettingsAPI) UpdateDatavolumesSettingByKeyExecute(_ dv_openapi.ApiUpdateDatavolumesSettingByKeyRequest) (*http.Response, error) {
	return m.updateResp, m.updateErr
}

func TestDatavolumesAdapter_UpdateEnableDataVolumes_Enable(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{updateResp: &http.Response{StatusCode: http.StatusOK}}
	adapter := &DatavolumesAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableDataVolumes(context.Background(), testTenantID, true)

	assert.NoError(t, err)
	assert.Equal(t, dv_openapi.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, mock.updateKey)
}

func TestDatavolumesAdapter_UpdateEnableDataVolumes_Disable(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{updateResp: &http.Response{StatusCode: http.StatusOK}}
	adapter := &DatavolumesAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableDataVolumes(context.Background(), testTenantID, false)

	assert.NoError(t, err)
	assert.Equal(t, dv_openapi.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES, mock.updateKey)
}

func TestDatavolumesAdapter_UpdateEnableDataVolumes_APIError(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{updateErr: fmt.Errorf("connection refused")}
	adapter := &DatavolumesAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableDataVolumes(context.Background(), testTenantID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateEnableDataVolumes")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestDatavolumesAdapter_UpdateEnableDataVolumes_ServerError(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{updateResp: &http.Response{StatusCode: http.StatusInternalServerError}}
	adapter := &DatavolumesAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableDataVolumes(context.Background(), testTenantID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestDatavolumesAdapter_UpdateEnableDataVolumes_InvalidStringValue(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{}
	adapter := &DatavolumesAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableDataVolumes(context.Background(), testTenantID, "not-a-bool")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected bool")
}

func TestDatavolumesAdapter_UpdateEnableDataVolumes_NilValue(t *testing.T) {
	mock := &mockDatavolumesSettingsAPI{}
	adapter := &DatavolumesAdapter{SettingsAPI: mock}

	err := adapter.UpdateEnableDataVolumes(context.Background(), testTenantID, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected bool")
}
