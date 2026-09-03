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

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/mocks"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestCreateListener_failedToCreate(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("CreateListener", "http", "type", datatypes.JSON("{}")).Return(errors.New("failed"))
	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.CreateListener(context.Background(), api.ListenerData{
		ListenerId:     "http",
		ListenerConfig: map[string]interface{}{},
		ListenerType:   "type",
	})
	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
	assert.Equal(t, "failed to create listener", apiErr.Message)
}

func TestCreateListener_success(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	val := ""
	called := make(chan struct{})
	go func() {
		val = <-listenerUpdatesChannel
		called <- struct{}{}
	}()

	dao := mocks.NewListenerDao(t)
	dao.On("CreateListener", "http", "type", datatypes.JSON("{}")).Return(nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
	}

	res, err := lap.CreateListener(context.Background(), api.ListenerData{
		ListenerId:     "http",
		ListenerConfig: map[string]interface{}{},
		ListenerType:   "type",
	})

	<-called
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, "http", val)
}

func TestDeleteListener_success(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	val := ""
	called := make(chan struct{})
	go func() {
		val = <-listenerUpdatesChannel
		called <- struct{}{}
	}()

	dao := mocks.NewListenerDao(t)
	dao.On("DeleteListener", "http").Return(nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
	}

	res, err := lap.DeleteListener(context.Background(), "http")

	<-called
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, "http", val)
}

func TestDeleteListener_failed(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	dao := mocks.NewListenerDao(t)
	dao.On("DeleteListener", "http").Return(errors.New("some error"))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
	}

	res, err := lap.DeleteListener(context.Background(), "http")

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
	assert.Equal(t, "failed to delete", apiErr.Message)
}

func TestGetListener_success(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(&models.ListenersConfig{
		ListenerId:     "http",
		ListenerConfig: datatypes.JSON(`{"port":5000}`),
		ListenerType:   "http_listener",
	}, nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
		secretDataHandler:       getSecretHandler(),
	}

	res, err := lap.GetListener(context.Background(), "http")

	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
	body := res.Body.(api.ListenerData)
	assert.Equal(t, "http_listener", body.ListenerType)
	assert.Equal(t, "http", body.ListenerId)
	assert.Equal(t, map[string]interface{}{"port": float64(5000)}, body.ListenerConfig)
}

func TestGetListener_NotFound(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(nil, nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
		secretDataHandler:       getSecretHandler(),
	}

	res, err := lap.GetListener(context.Background(), "http")

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusNotFound)
	assert.Equal(t, "listener not found", apiErr.Message)
}

func TestGetListener_FetchReturnsNotFoundError(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "missing").Return(nil, configprovider.NewError("listener not found in db", configprovider.ErrCodeNotFound))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
		secretDataHandler:       getSecretHandler(),
	}

	res, err := lap.GetListener(context.Background(), "missing")

	apiErr := err.(*ApiError)
	assert.Equal(t, "listener not found", apiErr.Message)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetListener_Failed(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(nil, fmt.Errorf("failed to fetch listener"))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
		secretDataHandler:       getSecretHandler(),
	}

	res, err := lap.GetListener(context.Background(), "http")

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
	assert.Equal(t, "failed to fetch listener", apiErr.Message)
}

func TestGetListener_failedToRemoveSecrets(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(&models.ListenersConfig{
		ListenerId:     "http",
		ListenerConfig: datatypes.JSON(`{port:5000,}`),
	}, nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
		secretDataHandler:       getSecretHandler(),
	}

	res, err := lap.GetListener(context.Background(), "http")

	apiErr := err.(*ApiError)
	assert.Equal(t, "failed to remove secret data: invalid character 'p' looking for beginning of object key string", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
}

func TestCreateListener_successOtherRouteExistWithSamePort(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListenersByPort", 5000).Return([]*models.ListenersConfig{{ListenerId: "other"}}, nil)
	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.CreateListener(context.Background(), api.ListenerData{
		ListenerId:     "http",
		ListenerConfig: map[string]interface{}{"http_listener": map[string]interface{}{"port": 5000}},
	})

	apiErr := err.(*ApiError)
	assert.Equal(t, "listener on same port already exist", apiErr.Message)
	assert.Equal(t, res.Code, http.StatusBadRequest)
}

func TestUpdateListener_success(t *testing.T) {
	listenerUpdatesChannel := make(chan string)
	val := ""
	called := make(chan struct{})
	go func() {
		val = <-listenerUpdatesChannel
		called <- struct{}{}
	}()

	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(&models.ListenersConfig{
		ListenerId:     "http",
		ListenerConfig: datatypes.JSON(`{"http_listener":{"port":5000}}`),
		ListenerType:   "http_listener",
	}, nil)
	dao.On("UpdateListener", "http", "http_listener", datatypes.JSON(`{"http_listener":{"port":5000}}`)).Return(nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: listenerUpdatesChannel,
	}

	res, err := lap.UpdateListener(context.Background(), "http", api.ListenerDataForUpdate{
		ListenerType:   "http_listener",
		ListenerConfig: map[string]interface{}{"http_listener": map[string]interface{}{"port": 5000}},
	})

	<-called
	assert.Nil(t, err)
	assert.Equal(t, res.Code, http.StatusOK)
	assert.Equal(t, "http", val)
}

func TestUpdateListener_listenerNotFound(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(nil, configprovider.NewError("listener not found", configprovider.ErrCodeNotFound))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.UpdateListener(context.Background(), "http", api.ListenerDataForUpdate{
		ListenerType:   "http_listener",
		ListenerConfig: map[string]interface{}{},
	})

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusNotFound)
	assert.Equal(t, "listener not found to be updated", apiErr.Message)
}

func TestUpdateListener_fetchListenerError(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(nil, errors.New("db error"))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.UpdateListener(context.Background(), "http", api.ListenerDataForUpdate{
		ListenerType:   "http_listener",
		ListenerConfig: map[string]interface{}{},
	})

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
	assert.Equal(t, "failed to update listener", apiErr.Message)
}

func TestUpdateListener_samePortAlreadyExists(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(&models.ListenersConfig{
		ListenerId:     "http",
		ListenerConfig: datatypes.JSON(`{"http_listener":{"port":5000}}`),
		ListenerType:   "http_listener",
	}, nil)
	dao.On("FetchListenersByPort", 5001).Return([]*models.ListenersConfig{{ListenerId: "other"}}, nil)

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.UpdateListener(context.Background(), "http", api.ListenerDataForUpdate{
		ListenerType:   "http_listener",
		ListenerConfig: map[string]interface{}{"http_listener": map[string]interface{}{"port": 5001}},
	})

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusBadRequest)
	assert.Equal(t, "listener on same port already exist", apiErr.Message)
}

func TestUpdateListener_updateReturnsNotFound(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(&models.ListenersConfig{
		ListenerId:     "http",
		ListenerConfig: datatypes.JSON(`{}`),
		ListenerType:   "http_listener",
	}, nil)
	dao.On("UpdateListener", "http", "http_listener", datatypes.JSON("{}")).Return(configprovider.NewError("listener not found", configprovider.ErrCodeNotFound))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.UpdateListener(context.Background(), "http", api.ListenerDataForUpdate{
		ListenerType:   "http_listener",
		ListenerConfig: map[string]interface{}{},
	})

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusNotFound)
	assert.Equal(t, "failed to update listener", apiErr.Message)
}

func TestUpdateListener_updateReturnsError(t *testing.T) {
	dao := mocks.NewListenerDao(t)
	dao.On("FetchListener", "http").Return(&models.ListenersConfig{
		ListenerId:     "http",
		ListenerConfig: datatypes.JSON(`{}`),
		ListenerType:   "http_listener",
	}, nil)
	dao.On("UpdateListener", "http", "http_listener", datatypes.JSON("{}")).Return(errors.New("db error"))

	lap := ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: nil,
	}

	res, err := lap.UpdateListener(context.Background(), "http", api.ListenerDataForUpdate{
		ListenerType:   "http_listener",
		ListenerConfig: map[string]interface{}{},
	})

	apiErr := err.(*ApiError)
	assert.Equal(t, res.Code, http.StatusInternalServerError)
	assert.Equal(t, "failed to update listener", apiErr.Message)
}
