/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/mocks"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

func TestUpdateUserRoute_Succeed(t *testing.T) {
	data := datatypes.JSON(`{"listener_id":"listener1","prometheus_alerts":{"port":9090}}`)
	dao := mocks.NewUserDao(t)
	dao.On("UpdateUserRoute", "tenant1", "user1", "routeTest", "listener1", data).Return(nil)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "routeTest", mock.Anything).Return(models.NewEmptyRouteGlobalConfig(), nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteDataForUpdate{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteConfig: anyMap{"prometheus_alerts": anyMap{"port": 9090}, "listener_id": "listener1"},
	}

	res, err := uras.UpdateUserRoute(context.Background(), "tenant1", "user1", "routeTest", requestData)
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestUpdateUserRouteurass_ListenerNotPassed(t *testing.T) {
	dao := mocks.NewUserDao(t)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	requestData := api.UserRouteDataForUpdate{
		TenantId: "tenant1",
		UserId:   "user1",
		RouteConfig: anyMap{
			"tenantId":    "tenant1",
			"RouteId":     "routeTest",
			"RouteConfig": anyMap{"prometheus_alerts": anyMap{"": "port"}, "listener_id": "listener1"},
		},
	}

	res, err := uras.UpdateUserRoute(context.Background(), "tenant1", "user1", "routeTest", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "listener id not passed in config", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusBadRequest)
}

func TestGetUserRoute_failed(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoutes", "", "tenant1", "user1").Return(nil, errors.New("failed to fetch routes"))

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := uras.GetUserRoutes(context.Background(), "tenant1", "user1")
	apiErr := err.(*ApiError)
	assert.Equal(t, "failed to fetch routes to tenant: tenant1 user: user1", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestGetUserRoute_succeed(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoutes", "", "tenant1", "user1").Return([]*models.RouteGlobalConfig{
		{
			TenantId:    "tenant1",
			ListenerId:  "listener1",
			RouteId:     "route1",
			RouteConfig: datatypes.JSON(`{}`),
		},
	}, nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := uras.GetUserRoutes(context.Background(), "tenant1", "user1")
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestCreateUserRoute_Succeed(t *testing.T) {
	routeConfig := anyMap{"listener_id": "listener1", "notify_ids": []string{"email1"}}
	data := datatypes.JSON(`{"listener_id":"listener1","notify_ids":["email1"]}`)
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "route1").Return(nil, nil)
	dao.On("CreateUserRoute", "tenant1", "user1", "route1", "listener1", data).Return(nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteData{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteId:     "route1",
		RouteConfig: routeConfig,
	}

	res, err := uras.CreateUserRoute(context.Background(), "tenant1", "user1", requestData)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestCreateUserRoute_ListenerNotPassed(t *testing.T) {
	dao := mocks.NewUserDao(t)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteData{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteId:     "route1",
		RouteConfig: anyMap{"notify_ids": []string{"email1"}},
	}

	res, err := uras.CreateUserRoute(context.Background(), "tenant1", "user1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "listener id not passed in config", apiErr.Message)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestCreateUserRoute_AlreadyExists(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "route1").Return(&models.RouteGlobalConfig{
		TenantId:    "tenant1",
		RouteId:     "route1",
		ListenerId:  "listener1",
		RouteConfig: datatypes.JSON(`{}`),
	}, nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteData{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteId:     "route1",
		RouteConfig: anyMap{"listener_id": "listener1"},
	}

	res, err := uras.CreateUserRoute(context.Background(), "tenant1", "user1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "route already exists: route1", apiErr.Message)
	assert.Equal(t, http.StatusConflict, res.Code)
}

func TestCreateUserRoute_DaoCreateFails(t *testing.T) {
	routeConfig := anyMap{"listener_id": "listener1"}
	data := datatypes.JSON(`{"listener_id":"listener1"}`)
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "route1").Return(nil, nil)
	dao.On("CreateUserRoute", "tenant1", "user1", "route1", "listener1", data).Return(fmt.Errorf("db error"))

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteData{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteId:     "route1",
		RouteConfig: routeConfig,
	}

	res, err := uras.CreateUserRoute(context.Background(), "tenant1", "user1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "internal server error", apiErr.Message)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateUserRoute_RouteNotFound(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "route1").Return(nil, nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteDataForUpdate{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteConfig: anyMap{"listener_id": "listener1"},
	}

	res, err := uras.UpdateUserRoute(context.Background(), "tenant1", "user1", "route1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "route not found: route1", apiErr.Message)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestUpdateUserRoute_FetchFails(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "route1").Return(nil, fmt.Errorf("db error"))

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteDataForUpdate{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteConfig: anyMap{"listener_id": "listener1"},
	}

	res, err := uras.UpdateUserRoute(context.Background(), "tenant1", "user1", "route1", requestData)
	assert.Error(t, err)
	apiErr := err.(*ApiError)
	assert.Equal(t, "failed to read route: route1", apiErr.Message)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateUserRoute_DaoUpdateFails(t *testing.T) {
	data := datatypes.JSON(`{"listener_id":"listener1"}`)
	dao := mocks.NewUserDao(t)
	dao.On("FetchUserRoute", "listener1", "tenant1", "user1", "route1", mock.Anything).Return(models.NewEmptyRouteGlobalConfig(), nil)
	dao.On("UpdateUserRoute", "tenant1", "user1", "route1", "listener1", data).Return(fmt.Errorf("db error"))

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}
	requestData := api.UserRouteDataForUpdate{
		TenantId:    "tenant1",
		UserId:      "user1",
		RouteConfig: anyMap{"listener_id": "listener1"},
	}

	res, err := uras.UpdateUserRoute(context.Background(), "tenant1", "user1", "route1", requestData)
	apiErr := err.(*ApiError)
	assert.Equal(t, "internal server error", apiErr.Message)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestDeleteAllTenantsUserRoutes_succeed(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("DeleteAllTenantsUserRoutes", "tenant1").Return(nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := uras.DeleteAllUserRoutes(context.Background(), "tenant1")
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestDeleteAllTenantsUserRoutes_failed(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("DeleteAllTenantsUserRoutes", "tenant1").Return(fmt.Errorf("failed"))

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := uras.DeleteAllUserRoutes(context.Background(), "tenant1")
	assert.Error(t, err)
	apiErr := err.(*ApiError)
	assert.Equal(t, "internal server error", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestDeleteNotifyIdFromAllUserRoutes(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("DeleteNotifyIdFromAllUserRoutes", "1", "email1").Return(nil)

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := uras.DeleteNotifyIdFromAllUserRoutes(context.Background(), "1", "email1")
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
}

func TestDeleteNotifyIdFromAllUserRoutes_Failed(t *testing.T) {
	dao := mocks.NewUserDao(t)
	dao.On("DeleteNotifyIdFromAllUserRoutes", "tenant1", "email1").Return(fmt.Errorf("failed"))

	uras := UserRouteAPIService{dao: dao, secretHandler: getSecretHandler()}

	res, err := uras.DeleteNotifyIdFromAllUserRoutes(context.Background(), "tenant1", "email1")
	assert.Error(t, err)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}
