/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package advanced

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi"
	api2 "github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	runaimetrics "github.com/run-ai/runai/runai-common-packages/metrics/http"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	"k8s.io/utils/ptr"
)

type RoutesRegister struct {
	dao                    ConfigProviderDao
	secretFieldsHandler    encryption.SecretFieldHandlerer
	listenerUpdatesChannel chan string
}

func NewRoutesRegister(dao ConfigProviderDao, secretFieldsHandler encryption.SecretFieldHandlerer, updatesChan chan string) *RoutesRegister {
	return &RoutesRegister{
		dao:                    dao,
		secretFieldsHandler:    secretFieldsHandler,
		listenerUpdatesChannel: updatesChan,
	}
}

func (rr *RoutesRegister) Register(aRouter *chi.Mux) {
	rns := http_listener.NewRegisterNotificationAPIService(rr.dao)
	trs := http_listener.NewTenantRouteAPIService(rr.dao, rr.secretFieldsHandler)
	urs := http_listener.NewUserRouteAPIService(rr.dao, rr.secretFieldsHandler)
	hs := http_listener.NewHealthAPIService(rr.dao)
	ss := http_listener.NewSinksAPIService(rr.dao, rr.secretFieldsHandler)
	ls := http_listener.NewListenerAPIService(rr.dao, rr.secretFieldsHandler, rr.listenerUpdatesChannel)
	openapi.AddSwaggerEndpoints(aRouter)
	aRouter.Group(func(router chi.Router) {
		router.Use(runaictx.ContextMiddleware)
		router.Use(runaimetrics.HttpMiddleware)
		addRoutes(router, api2.NewRegisterNotificationAPIController(rns, api2.WithRegisterNotificationAPIErrorHandler(errorHandler)))
		addRoutes(router, api2.NewTenantRouteAPIController(trs, api2.WithTenantRouteAPIErrorHandler(errorHandler)))
		addRoutes(router, api2.NewUserRouteAPIController(urs, api2.WithUserRouteAPIErrorHandler(errorHandler)))
		addRoutes(router, api2.NewHealthAPIController(hs, api2.WithHealthAPIErrorHandler(errorHandler)))
		addRoutes(router, api2.NewSinksAPIController(ss, api2.WithSinksAPIErrorHandler(errorHandler)))
		addRoutes(router, api2.NewListenerAPIController(ls, api2.WithListenerAPIErrorHandler(errorHandler)))
	})
}

func (rr *RoutesRegister) GetName() string {
	return "advanced_routes_register"
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error, result *api2.ImplResponse) {
	var requiredErr *api2.RequiredError
	if errors.As(err, &requiredErr) {
		api2.EncodeJSONResponse(api2.ApiError{Code: int32(http.StatusUnprocessableEntity), Message: err.Error()}, ptr.To(http.StatusUnprocessableEntity), w)
		return
	}
	var parseErr *api2.ParsingError
	if errors.As(err, &parseErr) {
		api2.EncodeJSONResponse(api2.ApiError{Code: int32(http.StatusBadRequest), Message: parseErr.Error()}, ptr.To(http.StatusBadRequest), w)
		return
	}

	var apiErr *http_listener.ApiError
	if errors.As(err, &apiErr) {
		api2.EncodeJSONResponse(apiErr.ApiError, &result.Code, w)
		return
	}

	api2.EncodeJSONResponse(api2.ApiError{Code: int32(result.Code), Message: err.Error()}, &result.Code, w)
}
