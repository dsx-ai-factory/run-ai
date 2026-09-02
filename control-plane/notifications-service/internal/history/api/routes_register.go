/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	history "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history/openapi/generated"
	"k8s.io/utils/ptr"
)

type RoutesRegister struct {
	service EventsHistoryServicer
}

func NewHistoryRoutesRegister(repository history.EventsRepository) *RoutesRegister {
	return &RoutesRegister{
		service: history.NewEventsHistoryService(repository),
	}
}

func (h *RoutesRegister) Register(router *chi.Mux) {
	router.Group(func(router chi.Router) {
		addRoutes(router, openapi.NewEventsAPIController(NewEventsAPIService(h.service), openapi.WithEventsAPIErrorHandler(errorHandler)))
	})
}

func (h *RoutesRegister) GetName() string {
	return "history_routes"
}

func addRoutes(router chi.Router, r openapi.Router) chi.Router {
	for _, route := range r.Routes() {
		handler := route.HandlerFunc
		router.Method(route.Method, route.Pattern, handler)
	}

	return router
}

func errorHandler(w http.ResponseWriter, r *http.Request, err error, result *openapi.ImplResponse) {
	var requiredErr *openapi.RequiredError
	if errors.As(err, &requiredErr) {
		openapi.EncodeJSONResponse(openapi.ApiError{Code: int32(http.StatusUnprocessableEntity), Message: err.Error()}, ptr.To(http.StatusUnprocessableEntity), w)
		return
	}
	var parseErr *openapi.ParsingError
	if errors.As(err, &parseErr) {
		openapi.EncodeJSONResponse(openapi.ApiError{Code: int32(http.StatusBadRequest), Message: parseErr.Error()}, ptr.To(http.StatusBadRequest), w)
		return
	}

	openapi.EncodeJSONResponse(openapi.ApiError{Code: int32(result.Code), Message: err.Error()}, &result.Code, w)
}
