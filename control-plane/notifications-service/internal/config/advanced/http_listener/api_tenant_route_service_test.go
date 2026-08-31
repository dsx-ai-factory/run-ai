/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/mocks"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

var encrypt = encryption.NewEncrypterDecrypter([]byte("0123456789abcdefghwertol"))

type anyMap map[string]interface{}

func getSecretHandler() *encryption.SecretFieldHandler {
	return encryption.NewSecretFieldHandler(encrypt, route.BasicRouteConfiguration{}, listen.ListeningConfig{}, sinks.EmailSinkConfig{}, sinks.SlackSinkConfig{})
}

func TestUpdateRoutetrass_ListenerNotPassed(t *testing.T) {
	dao := mocks.NewTenantDao(t)

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.RouteDataForUpdate{
		TenantId: "tenant1",
		RouteConfig: anyMap{
			"tenantId":    "tenant1",
			"RouteId":     "routeTest",
			"RouteConfig": anyMap{"prometheus_alerts": anyMap{"": "port"}, "listener_id": "listener1"},
		},
	}

	res, err := tras.UpdateRoute(context.Background(), "tenant1", "route1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "listener id not passed in config", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusBadRequest)
}

func TestUpdateRoute_ListenerNotPassed(t *testing.T) {
	dao := mocks.NewTenantDao(t)

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.RouteDataForUpdate{
		TenantId: "tenant1",
		RouteConfig: anyMap{
			"tenantId":    "tenant1",
			"RouteId":     "routeTest",
			"RouteConfig": anyMap{"prometheus_alerts": anyMap{"": "port"}, "listener_id": "listener1"},
		},
	}

	res, err := tras.UpdateRoute(context.Background(), "tenant1", "route1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "listener id not passed in config", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusBadRequest)
}

func TestGetRoutes_failed(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	dao.On("FetchRoutes", "", "tenant1").Return(nil, errors.New("failed to fetch routes"))

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := tras.GetRoutes(context.Background(), "tenant1")
	apiErr := err.(*ApiError)
	assert.Equal(t, "failed to fetch routes to tenant: tenant1", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestGetRoutes_succeed(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	dao.On("FetchRoutes", "", "tenant1").Return([]*models.RouteGlobalConfig{
		{
			TenantId:   "tenant1",
			ListenerId: "listener1",
			RouteId:    "route1",
			RouteConfig: datatypes.JSON(`{
				"listener_id": "listener1",
				"notify_ids": ["email1"]
			}`),
		},
	}, nil)

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := tras.GetRoutes(context.Background(), "tenant1")
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
	require.NotNil(t, res.Body)

	body := res.Body.([]api.RouteData)
	sinkId := body[0].RouteConfig["notify_ids"].([]interface{})[0]
	assert.Equal(t, "email1", sinkId)
}

func TestGetRoute_succeed(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	dao.On("FetchRoute", "", "tenant1", "route1").Return(&models.RouteGlobalConfig{
		TenantId:   "tenant1",
		ListenerId: "listener1",
		RouteId:    "route1",
		RouteConfig: datatypes.JSON(`{
			"listener_id": "listener1",
			"notify_ids": ["email1"]
		}`),
	}, nil)

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := tras.GetRoute(context.Background(), "tenant1", "route1")
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
	require.NotNil(t, res.Body)
}

func TestGetRoute_notFound(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	dao.On("FetchRoute", "", "tenant1", "route1").Return(nil, nil)

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := tras.GetRoute(context.Background(), "tenant1", "route1")
	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusNotFound)
	assert.Equal(t, "route not found: route1", apiErr.Message)
}

func TestCreateRoute_ListenerNotPassed(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.RouteData{
		TenantId: "tenant1",
		RouteId:  "route1",
		RouteConfig: anyMap{
			"tenantId":    "tenant1",
			"RouteId":     "routeTest",
			"RouteConfig": anyMap{"prometheus_alerts": anyMap{"": "port"}, "listener_id": "listener1"},
		},
	}

	res, err := tras.CreateRoute(context.Background(), "tenant1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "listener id not passed in config", apiErr.Message)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestCreateRoute_Succeed(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	dao.On("FetchRoute", "listener1", "tenant1", "route1").Return(nil, nil)
	dao.On("CreateRoute", "tenant1", "route1", "listener1", mock.AnythingOfType("datatypes.JSON")).Return(nil)

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.RouteData{
		TenantId: "tenant1",
		RouteId:  "route1",
		RouteConfig: anyMap{
			"listener_id": "listener1",
			"notify_ids":  []string{"email1"},
		},
	}

	res, err := tras.CreateRoute(context.Background(), "tenant1", requestData)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCreateRoute_daoError(t *testing.T) {
	dao := mocks.NewTenantDao(t)
	dao.On("FetchRoute", "listener1", "tenant1", "route1").Return(nil, nil)
	dao.On("CreateRoute", "tenant1", "route1", "listener1", mock.AnythingOfType("datatypes.JSON")).Return(errors.New("db error"))

	tras := TenantRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.RouteData{
		TenantId: "tenant1",
		RouteId:  "route1",
		RouteConfig: anyMap{
			"listener_id": "listener1",
			"notify_ids":  []string{"email1"},
		},
	}

	res, err := tras.CreateRoute(context.Background(), "tenant1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "internal server error", apiErr.Message)
}
