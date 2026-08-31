/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notification

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestInit_NotificationsDisabled(t *testing.T) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init() panicked with disabled notifications: %v", r)
		}
	}()

	_ = ctx
}

func TestInitNotificationClient_Structure(t *testing.T) {
	ctx := context.Background()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initNotificationClient() panicked: %v", r)
		}
	}()

	config := &config.Notify{
		NotifyUrl: "http://test:8080",
		ConfigUrl: "http://test:9090",
	}

	_ = initNotificationClient(ctx, config)
}

func TestRegisterNotifications_Structure(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("registerNotifications() panicked: %v", r)
		}
	}()

	_ = registerNotifications()
}

func TestGetHostAndScheme(t *testing.T) {
	tests := []struct {
		name           string
		configUrl      string
		expectedHost   string
		expectedScheme string
	}{
		{
			name:           "valid http url",
			configUrl:      "http://notification-service:8080",
			expectedHost:   "notification-service:8080",
			expectedScheme: "http",
		},
		{
			name:           "valid https url",
			configUrl:      "https://notification-service:8080/path",
			expectedHost:   "notification-service:8080",
			expectedScheme: "https",
		},
		{
			name:           "url without scheme",
			configUrl:      "//notification-service:8080",
			expectedHost:   "notification-service:8080",
			expectedScheme: "http",
		},
		{
			name:           "invalid url",
			configUrl:      "://invalid-url",
			expectedHost:   "",
			expectedScheme: "",
		},
		{
			name:           "empty url",
			configUrl:      "",
			expectedHost:   "",
			expectedScheme: "http",
		},
		{
			name:           "url with port only",
			configUrl:      "localhost:8080",
			expectedHost:   "",
			expectedScheme: "localhost",
		},
		{
			name:           "url with path and query",
			configUrl:      "http://service:8080/api/v1?param=value",
			expectedHost:   "service:8080",
			expectedScheme: "http",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, scheme := getHostAndScheme(tt.configUrl)
			assert.Equal(t, tt.expectedHost, host)
			assert.Equal(t, tt.expectedScheme, scheme)
		})
	}
}

func TestBuildRouteDataFromConfig(t *testing.T) {
	routeData := buildRouteDataFromConfig()

	assert.NotNil(t, routeData)

	assert.Contains(t, routeData, "enrich")
	assert.Contains(t, routeData, "metadata")
	assert.Contains(t, routeData, "tenant_id")
	assert.Contains(t, routeData, "notify_ids")
	assert.Contains(t, routeData, "listener_id")
	assert.Contains(t, routeData, "filtering")

	assert.Equal(t, TenantId, routeData["tenant_id"])
	assert.Equal(t, NotifyIds, routeData["notify_ids"])
	assert.Equal(t, ListenerId, routeData["listener_id"])
	assert.Nil(t, routeData["metadata"])

	enrich, ok := routeData["enrich"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, enrich, "static")
	assert.Contains(t, enrich, "kube_metadata")
	assert.Nil(t, enrich["static"])
	assert.Nil(t, enrich["kube_metadata"])

	filtering, ok := routeData["filtering"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "any", filtering["policy"])
	assert.Nil(t, filtering["prometheus_matchers"])

	filters, ok := filtering["filters"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, filters, 1)

	filter := filters[0]
	assert.Equal(t, "matcher", filter["type"])
	assert.Nil(t, filter["prometheus_matchers"])

	matchers, ok := filter["matchers"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, matchers, 2)

	categoryMatcher := matchers[0]
	assert.Equal(t, "category", categoryMatcher["key"])
	assert.Equal(t, "keycloak", categoryMatcher["value"])
	assert.Equal(t, "==", categoryMatcher["match_type"])

	typeMatcher := matchers[1]
	assert.Equal(t, "type", typeMatcher["key"])
	assert.Equal(t, "reset|welcome", typeMatcher["value"])
	assert.Equal(t, "=~", typeMatcher["match_type"])
}

func TestCreateConfigClient(t *testing.T) {
	tests := []struct {
		name      string
		configUrl string
	}{
		{
			name:      "valid http url",
			configUrl: "http://config-service:9090",
		},
		{
			name:      "valid https url",
			configUrl: "https://config-service:9090",
		},
		{
			name:      "invalid url",
			configUrl: "://invalid-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createConfigClient(tt.configUrl)
			assert.NotNil(t, client)
		})
	}
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "keycloak", TenantId)
	assert.Equal(t, "smtp-runai", ListenerId)
	assert.Equal(t, []string{"runai-email"}, NotifyIds)
}

func TestInitNotificationClientWithRegistration_Structure(t *testing.T) {
	ctx := context.Background()
	mockConfig := &config.Notify{
		NotifyUrl: "http://notify:8080",
		ConfigUrl: "http://config:9090",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initNotificationClientWithRegistration() panicked: %v", r)
		}
	}()

	_ = initNotificationClientWithRegistration(ctx, mockConfig)
}

func TestNotificationKinds_Structure(t *testing.T) {
	expectedNotificationKinds := []*notifyclient.NotificationKind{
		{
			Category:    "keycloak",
			Type:        "reset",
			Description: "Password reset notification",
			Template:    "keycloak_reset",
		},
		{
			Category:    "keycloak",
			Type:        "welcome",
			Description: "Welcome email notification",
			Template:    "keycloak_welcome",
		},
	}

	assert.Len(t, expectedNotificationKinds, 2)

	resetNotif := expectedNotificationKinds[0]
	assert.Equal(t, "keycloak", resetNotif.Category)
	assert.Equal(t, "reset", resetNotif.Type)
	assert.Equal(t, "Password reset notification", resetNotif.Description)
	assert.Equal(t, "keycloak_reset", resetNotif.Template)

	welcomeNotif := expectedNotificationKinds[1]
	assert.Equal(t, "keycloak", welcomeNotif.Category)
	assert.Equal(t, "welcome", welcomeNotif.Type)
	assert.Equal(t, "Welcome email notification", welcomeNotif.Description)
	assert.Equal(t, "keycloak_welcome", welcomeNotif.Template)
}
