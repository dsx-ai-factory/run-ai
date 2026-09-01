//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestGetNotificationTypes(t *testing.T) {
	notificationKinds := []*notifyclient.NotificationKind{
		{
			Category:    "category1",
			Type:        "type1",
			Description: "category1 type1",
			Template:    "integration-test-template",
		},
	}

	err := notifyclient.RegisterNotification("integration-tests-notification-types", notificationKinds, ptr.To(true))
	require.NoError(t, err)

	ctx := context.Background()
	data, response, err := client.NotificationTypesAPI.GetNotificationTypes(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, data, []openapi.NotificationType{
		{
			Category:    "category1",
			Type:        "type1",
			Description: "category1 type1",
		},
	})
}
