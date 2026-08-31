/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"

	"github.com/rs/zerolog/log"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
	"github.com/xyproto/jpath"
	"gorm.io/datatypes"
)

type ListenerDao interface {
	FetchListeners() ([]*models.ListenersConfig, error)
	FetchListenersByPort(port int) ([]*models.ListenersConfig, error)
	CreateListener(listenerId string, listenerType string, route datatypes.JSON) error
	UpdateListener(listenerId string, listenerType string, route datatypes.JSON) error
	FetchListener(id string) (*models.ListenersConfig, error)
	IsHealthyer
	DeleteListener(listenerId string) error
}

// ListenerAPIService is a service that implements the logic for the ListenerAPIServicer
// This service should implement the business logic for every endpoint for the ListenerAPI API.
// Include any external packages or services that will be required by this service.
type ListenerAPIService struct {
	dao                     ListenerDao
	listenersUpdatesChannel chan string
	secretDataHandler       encryption.SecretFieldHandlerer
}

// NewListenerAPIService creates a default api service
func NewListenerAPIService(dao ListenerDao, handler encryption.SecretFieldHandlerer, updatesChan chan string) api.ListenerAPIServicer {
	return &ListenerAPIService{
		dao:                     dao,
		listenersUpdatesChannel: updatesChan,
		secretDataHandler:       handler,
	}
}

func (las *ListenerAPIService) CreateListener(_ context.Context, listenerData api.ListenerData) (api.ImplResponse, error) {
	config, err := json.Marshal(listenerData.ListenerConfig)
	if err != nil {
		return returnError(err, http.StatusBadRequest, "failed to parse listener configuration", listenerData.ListenerId)
	}

	if !las.isPortAvailable(config) {
		return returnError(errors.New("listener on same port already exist"), http.StatusBadRequest, "listener on same port already exist", listenerData.ListenerId)
	}

	err = las.dao.CreateListener(listenerData.ListenerId, listenerData.ListenerType, config)
	if err == nil {
		log.Info().Msgf("listener created %s", listenerData.ListenerId)
		las.listenersUpdatesChannel <- listenerData.ListenerId
		return api.Response(http.StatusOK, "listener created"), nil
	}

	log.Err(err).Msgf("failed to create listener %s", listenerData.ListenerId)
	var daoError *configprovider.Error
	if errors.As(err, &daoError) && daoError.Code == configprovider.ErrCodeDuplicate {
		return returnError(err, http.StatusConflict, "listener on same port already exist", listenerData.ListenerId)
	}

	return returnError(err, http.StatusInternalServerError, "failed to create listener", listenerData.ListenerId)
}

func (las *ListenerAPIService) DeleteListener(_ context.Context, listenerId string) (api.ImplResponse, error) {
	err := las.dao.DeleteListener(listenerId)
	if err == nil {
		log.Info().Msgf("listener deleted %s", listenerId)
		las.listenersUpdatesChannel <- listenerId
		return api.Response(http.StatusOK, nil), nil
	}

	log.Err(err).Msgf("failed to delete listener %s", listenerId)
	var daoError *configprovider.Error
	if errors.As(err, &daoError) && daoError.Code == configprovider.ErrCodeNotFound {
		return newErrorResponse(http.StatusNotFound, fmt.Sprintf("listener not found: %s", listenerId), err)
	}
	return returnError(err, http.StatusInternalServerError, "failed to delete", listenerId)
}

func (las *ListenerAPIService) UpdateListener(_ context.Context, listenerId string, listenerData api.ListenerDataForUpdate) (api.ImplResponse, error) {
	listener, err := las.dao.FetchListener(listenerId)
	if err != nil {
		var daoError *configprovider.Error
		if errors.As(err, &daoError) && daoError.Code == configprovider.ErrCodeNotFound {
			return newErrorResponse(http.StatusNotFound, "listener not found to be updated", err)
		}

		log.Err(err).Msgf("failed to fetch listener %s", listenerId)
		return returnError(err, http.StatusInternalServerError, "failed to update listener", listenerId)
	}
	if listener == nil {
		return newErrorResponse(http.StatusNotFound, "listener not found to be updated", nil)
	}

	config, err := json.Marshal(listenerData.ListenerConfig)
	if err != nil {
		return returnError(err, http.StatusBadRequest, "failed to parse listener configuration", listenerId)
	}

	if las.isPortChanged(config, listener.ListenerConfig) && !las.isPortAvailable(config) {
		return returnError(errors.New("listener on same port already exist"), http.StatusBadRequest, "listener on same port already exist", listenerId)
	}

	err = las.dao.UpdateListener(listenerId, listenerData.ListenerType, config)
	if err == nil {
		log.Info().Msgf("listener updated %s", listenerId)
		las.listenersUpdatesChannel <- listenerId
		return api.Response(http.StatusOK, "listener updated"), nil
	}

	log.Err(err).Msgf("failed to update listener %s", listenerId)
	var daoError *configprovider.Error
	if errors.As(err, &daoError) && daoError.Code == configprovider.ErrCodeNotFound {
		return returnError(err, http.StatusNotFound, "failed to update listener", listenerId)
	}
	return returnError(err, http.StatusInternalServerError, "failed to update listener", listenerId)
}

func (las *ListenerAPIService) GetListener(_ context.Context, listenerId string) (api.ImplResponse, error) {
	log.Debug().Msgf("getting listener %s", listenerId)
	listener, err := las.dao.FetchListener(listenerId)
	if err != nil {
		log.Err(err).Msgf("failed to fetch listener %s", listenerId)
		var daoError *configprovider.Error
		if errors.As(err, &daoError) && daoError.Code == configprovider.ErrCodeNotFound {
			return newErrorResponse(http.StatusNotFound, "listener not found", err)
		}
		return returnError(err, http.StatusInternalServerError, "failed to fetch listener", listenerId)
	}
	if listener == nil {
		log.Warn().Msgf("listener not found %s", listenerId)
		return newErrorResponse(http.StatusNotFound, "listener not found", nil)
	}

	response, err := las.secretDataHandler.RemoveSecrets(listener.ListenerConfig)
	if err != nil {
		return returnError(err, http.StatusInternalServerError, fmt.Sprintf("failed to remove secret data: %s", err.Error()), listenerId)
	}

	log.Debug().Msgf("listener found %s", listenerId)
	return api.Response(http.StatusOK, api.ListenerData{
		ListenerId:     listenerId,
		ListenerConfig: response,
		ListenerType:   listener.ListenerType,
	}), nil
}

func (las *ListenerAPIService) isPortAvailable(config []byte) bool {
	jconf, _ := jpath.New(config)
	port := jconf.GetNode("http_listener.port").Int(-1)
	if port == -1 {
		return true
	}

	routes, err := las.dao.FetchListenersByPort(port)
	return err != nil || len(routes) == 0
}

func (las *ListenerAPIService) isPortChanged(oldConfig []byte, newConfig []byte) bool {
	oldConf, _ := jpath.New(oldConfig)
	newConf, _ := jpath.New(newConfig)
	return oldConf.GetNode("http_listener.port").Int(-1) != newConf.GetNode("http_listener.port").Int(-1)
}
