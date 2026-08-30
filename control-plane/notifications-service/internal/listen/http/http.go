package http

import (
	"github.com/go-chi/chi/v5"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/http/openapi/generated"
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
