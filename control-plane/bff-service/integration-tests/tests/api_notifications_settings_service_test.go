/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"
	"net/http"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client/generated"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetNotificationsSettings(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	settings, httpResp, err := OpenapiClient.SettingsAPI.GetNotificationsSettings(ctx).Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, settings, "Settings response should not be nil")

	require.NotNil(t, settings.AdminMessage, "AdminMessage should be present in notifications settings")
}

func TestGetNotificationsSettingByKey(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetNotificationsSettingByKey(ctx, openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE).
		Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, setting, "Setting should not be nil")
	require.Contains(t, setting, "msg", "Setting should contain 'msg' field")
	require.Contains(t, setting, "enableDismiss", "Setting should contain 'enableDismiss' field")
}

func TestUpdateNotificationsSettingByKey_AdminMessage(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	updateBody := openapi.NewUpdateAdminMessage()
	updateBody.SetMsg("Integration test admin message")
	updateBody.SetEnableDismiss(true)

	updateResp, err := OpenapiClient.SettingsAPI.
		UpdateNotificationsSettingByKey(ctx, openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE).
		UpdateAdminMessage(*updateBody).
		Execute()

	require.NoError(t, err, "Should not return an error on UPDATE")
	require.Equal(t, http.StatusOK, updateResp.StatusCode, "Should return 200 OK status for UPDATE")

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetNotificationsSettingByKey(ctx, openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE).
		Execute()

	require.NoError(t, err, "Should not return an error on GET after UPDATE")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status for GET after UPDATE")
	require.NotNil(t, setting, "Setting should not be nil after update")

	require.Equal(t, "Integration test admin message", setting["msg"], "Msg should match after update")
	require.Equal(t, true, setting["enableDismiss"], "EnableDismiss should match after update")
}

func TestUpdateNotificationsSettingByKey_ClearMessage(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithUniqueTenant()

	// Set a message first
	setBody := openapi.NewUpdateAdminMessage()
	setBody.SetMsg("Message to clear")
	setResp, err := OpenapiClient.SettingsAPI.
		UpdateNotificationsSettingByKey(ctx, openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE).
		UpdateAdminMessage(*setBody).
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, setResp.StatusCode)

	// Clear the message
	clearBody := openapi.NewUpdateAdminMessage()
	clearBody.SetMsgNil()
	clearResp, err := OpenapiClient.SettingsAPI.
		UpdateNotificationsSettingByKey(ctx, openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE).
		UpdateAdminMessage(*clearBody).
		Execute()

	require.NoError(t, err, "Should not return an error when clearing message")
	require.Equal(t, http.StatusOK, clearResp.StatusCode, "Should return 200 OK when clearing message")

	setting, httpResp, err := OpenapiClient.SettingsAPI.
		GetNotificationsSettingByKey(ctx, openapi.NOTIFICATIONSSETTINGSKEYENUM_ADMIN_MESSAGE).
		Execute()

	require.NoError(t, err, "Should not return an error on GET after clear")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK on GET after clear")
	require.Equal(t, "", setting["msg"], "Msg should be empty string after clear")
}

func TestGetNotificationsSettings_NonManagementTenantParam_Returns403(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/notifications/settings?tenantId=%s", hostUrl, BaseTenantId), nil)
	require.NoError(t, err)
	req.Header.Set("X-Runai-Subject-Id", "test@example.com")
	// no X-Runai-Is-Management-Tenant header → not a management tenant

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusForbidden, resp.StatusCode, "Should return 403 when tenantId query param is set by a non-management tenant")
}

func TestGetNotificationsSettingByKey_UnknownKey_Returns400(t *testing.T) {
	t.Parallel()
	tenantID := uuid.NewString()

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/api/v1/notifications/settings/unknownKey", hostUrl), nil)
	require.NoError(t, err)
	req.Header.Set("X-Runai-Tenant-Id", tenantID)
	req.Header.Set("X-Runai-Subject-Id", "test@example.com")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for unknown notification setting key")
}
