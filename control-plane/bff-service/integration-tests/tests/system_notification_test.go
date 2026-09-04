/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/stretchr/testify/require"
)

func TestGetSystemNotification_SystemMessage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	setting, httpResp, err := OpenapiClient.SystemNotificationAPI.
		GetSystemNotification(ctx).
		Execute()

	require.NoError(t, err, "Should not return an error")
	require.Equal(t, http.StatusOK, httpResp.StatusCode, "Should return 200 OK status")
	require.NotNil(t, setting, "System setting response should not be nil")
}

func TestUpdateSystemNotification_SystemMessage(t *testing.T) {
	t.Parallel()
	ctx := newTestContextWithManagementTenant()

	updateBody := openapi.NewUpdateSystemNotification()
	updateBody.SetMsg("Integration test system message")

	httpResp, err := OpenapiClient.SystemNotificationAPI.
		UpdateSystemNotification(ctx).
		UpdateSystemNotification(*updateBody).
		Execute()

	logging.Infof("UpdateSystemNotification response %+v, error %+v", httpResp, err)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	setting, httpResp, err := OpenapiClient.SystemNotificationAPI.
		GetSystemNotification(ctx).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, setting)
	require.Equal(t, "Integration test system message", setting.GetMsg(), "Msg should be set after update")
}
