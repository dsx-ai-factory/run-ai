package router

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/utils"
)

func addFullHostToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		value := r.Header.Get("X-Forwarded-Host")
		ctx = context.WithValue(ctx, utils.FullHostKey, value)

		newR := r.WithContext(ctx)
		next.ServeHTTP(w, newR)
	})
}
