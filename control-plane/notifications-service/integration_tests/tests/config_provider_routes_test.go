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

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/route"

	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/stretchr/testify/require"
)

func TestGetRoute_WithMetadata(t *testing.T) {
	routeData := config_client.RouteData{
		RouteId:  "TestGetRoute_WithMetadata",
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email1"},
			Metadata:   &route.Metadata{"email_enabled": "true"},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.
		CreateRoute(context.Background(), "tenant").
		RouteData(routeData).
		Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestGetRoute_WithMetadata").Execute()

	res, r, err := configClient.TenantRouteAPI.GetRoute(context.Background(), "tenant", "TestGetRoute_WithMetadata").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)
	require.Equal(t, res.RouteConfig["metadata"], map[string]interface{}{"email_enabled": "true"})
}

func TestGetRoutes(t *testing.T) {
	route1 := config_client.RouteData{
		RouteId:  "TestGetRoutes_1",
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email1"},
		}),
	}
	route2 := config_client.RouteData{
		RouteId:  "TestGetRoutes_2",
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"slack1"},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.
		CreateRoute(context.Background(), "tenant").
		RouteData(route1).
		Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestGetRoutes_1").Execute()

	_, _, err = configClient.TenantRouteAPI.
		CreateRoute(context.Background(), "tenant").
		RouteData(route2).
		Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestGetRoutes_2").Execute()

	routes, r, err := configClient.TenantRouteAPI.GetRoutes(context.Background(), "tenant").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)

	routeIds := make([]string, len(routes))
	for i, rt := range routes {
		routeIds[i] = rt.RouteId
	}
	require.Contains(t, routeIds, "TestGetRoutes_1")
	require.Contains(t, routeIds, "TestGetRoutes_2")
}

func TestUpdateRoute(t *testing.T) {
	routeData := config_client.RouteData{
		RouteId:  "TestUpdateRoute",
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email1"},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.
		CreateRoute(context.Background(), "tenant").
		RouteData(routeData).
		Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestUpdateRoute").Execute()

	updatedRouteData := config_client.RouteDataForUpdate{
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email1", "slack1"},
			Metadata:   &route.Metadata{"email_enabled": "false"},
		}),
	}
	_, r, err := configClient.TenantRouteAPI.
		UpdateRoute(context.Background(), "tenant", "TestUpdateRoute").
		RouteDataForUpdate(updatedRouteData).
		Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)

	res, r, err := configClient.TenantRouteAPI.GetRoute(context.Background(), "tenant", "TestUpdateRoute").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)
	require.Equal(t, []interface{}{"email1", "slack1"}, res.RouteConfig["notify_ids"])
	require.Equal(t, map[string]interface{}{"email_enabled": "false"}, res.RouteConfig["metadata"])
}

func TestDeleteRoute_Exist(t *testing.T) {
	routeData := config_client.RouteData{
		RouteId:  "TestDeleteRoute_Exist",
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email1"},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.
		CreateRoute(context.Background(), "tenant").
		RouteData(routeData).
		Execute()
	require.NoError(t, err)

	res, r, err := configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestDeleteRoute_Exist").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)
	require.Empty(t, res)
}

func TestDeleteRoute_NotExist(t *testing.T) {
	res, httpResp, err := configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestDeleteRoute_NotExist").Execute()
	require.EqualError(t, err, "404 Not Found")
	require.Equal(t, httpResp.StatusCode, http.StatusNotFound)
	require.Empty(t, res)
}

func TestDeleteAllRoutes_Exist(t *testing.T) {
	routeData := config_client.RouteData{
		RouteId:  "TestDeleteAllRoutes_Exist",
		TenantId: "tenant",
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email1"},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.
		CreateRoute(context.Background(), "tenant").
		RouteData(routeData).
		Execute()
	require.NoError(t, err)
	_, _, err = configClient.UserRouteAPI.
		CreateUserRoute(context.Background(), "tenant", "userXYZ").UserRouteData(config_client.UserRouteData{
		TenantId:    "tenant",
		RouteId:     "TestDeleteAllRoutes_Exist",
		UserId:      "userXYZ",
		RouteConfig: routeData.RouteConfig,
	}).Execute()
	require.NoError(t, err)

	_, r, err := configClient.TenantRouteAPI.DeleteRoute(context.Background(), "tenant", "TestDeleteAllRoutes_Exist").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)

	_, r, err = configClient.UserRouteAPI.DeleteAllUserRoutes(context.Background(), "tenant").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)

	routes, r, err := configClient.UserRouteAPI.GetUserRoutes(context.Background(), "tenant", "userXYZ").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)
	require.Empty(t, routes)
}

func TestDeleteNotifyIdFromAllUserRoutes(t *testing.T) {
	tenantId := "tenant"
	userId := "user1"
	routeId := "DeleteNotifyIdFromAllUserRoutes"
	defer userRoutesCleanup(t, tenantId)

	createUserRoute(t, tenantId, userId, routeId, []string{"email-1", "slack-1", "slack-2"})

	routes, _, err := configClient.UserRouteAPI.GetUserRoutes(context.Background(), tenantId, userId).Execute()
	require.NoError(t, err)
	require.Len(t, routes, 1)
	require.Equal(t, routeId, routes[0].RouteId)

	beforeConfig := routes[0].RouteConfig
	notifyIds := beforeConfig["notify_ids"].([]interface{})
	require.Equal(t, notifyIds, []interface{}{"email-1", "slack-1", "slack-2"})

	_, _, err = configClient.UserRouteAPI.DeleteNotifyIdFromAllUserRoutes(context.Background(), tenantId, "email-1").Execute()
	require.NoError(t, err)

	routes, _, err = configClient.UserRouteAPI.GetUserRoutes(context.Background(), tenantId, userId).Execute()
	require.NoError(t, err)
	require.Len(t, routes, 1)

	verifyOnlyNotifyIdsChanged(t, beforeConfig, routes[0].RouteConfig, []interface{}{"slack-1", "slack-2"})
}

func TestUpdateUserRoute(t *testing.T) {
	tenantId := "tenant"
	userId := "userUpdate1"
	routeId := "TestUpdateUserRoute"
	defer userRoutesCleanup(t, tenantId)

	createUserRoute(t, tenantId, userId, routeId, []string{"email-1"})

	routes, r, err := configClient.UserRouteAPI.GetUserRoutes(context.Background(), tenantId, userId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, r.StatusCode)
	require.Len(t, routes, 1)
	require.Equal(t, routeId, routes[0].RouteId)
	notifyIds := routes[0].RouteConfig["notify_ids"].([]interface{})
	require.Equal(t, []interface{}{"email-1"}, notifyIds)

	updatedRouteData := config_client.UserRouteDataForUpdate{
		TenantId: tenantId,
		UserId:   userId,
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  []string{"email-1", "slack-1"},
		}),
	}
	_, r, err = configClient.UserRouteAPI.
		UpdateUserRoute(context.Background(), tenantId, userId, routeId).
		UserRouteDataForUpdate(updatedRouteData).
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, r.StatusCode)

	routes, r, err = configClient.UserRouteAPI.GetUserRoutes(context.Background(), tenantId, userId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, r.StatusCode)
	require.Len(t, routes, 1)
	require.Equal(t, routeId, routes[0].RouteId)
	updatedNotifyIds := routes[0].RouteConfig["notify_ids"].([]interface{})
	require.Equal(t, []interface{}{"email-1", "slack-1"}, updatedNotifyIds)
}

func createUserRoute(t *testing.T, tenantId string, userId string, routeId string, notifyIds []string) {
	routeData := config_client.RouteData{
		RouteId:  routeId,
		TenantId: tenantId,
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: "listener",
			NotifyIds:  notifyIds,
		}),
	}
	_, _, err := configClient.UserRouteAPI.
		CreateUserRoute(context.Background(), tenantId, userId).
		UserRouteData(config_client.UserRouteData{
			TenantId:    tenantId,
			UserId:      userId,
			RouteId:     routeId,
			RouteConfig: routeData.RouteConfig,
		}).Execute()
	require.NoError(t, err)
}

func userRoutesCleanup(t *testing.T, tenantId string) {
	_, _, err := configClient.UserRouteAPI.DeleteAllUserRoutes(context.Background(), tenantId).Execute()
	require.NoError(t, err)
}

func verifyOnlyNotifyIdsChanged(t *testing.T, beforeConfig, newConfig map[string]interface{}, expectedNotifyIds []interface{}) {
	require.ElementsMatch(t, getMapKeys(beforeConfig), getMapKeys(newConfig))

	for key, value := range beforeConfig {
		if key == "notify_ids" { // only this should be different
			continue
		}
		require.Equal(t, value, newConfig[key])
	}

	require.Equal(t, expectedNotifyIds, newConfig["notify_ids"].([]interface{}))
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
