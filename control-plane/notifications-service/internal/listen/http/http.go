/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/http/openapi/generated"
	"github.com/go-chi/chi/v5"
)

func AddRoutes(router chi.Router, routers ...api.Router) chi.Router {
	for _, aRouter := range routers {
		for _, route := range aRouter.Routes() {
			handler := route.HandlerFunc
			router.Method(route.Method, route.Pattern, handler)
		}
	}

	return router
}
