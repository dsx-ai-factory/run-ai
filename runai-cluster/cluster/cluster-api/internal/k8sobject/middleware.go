/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"context"
	"fmt"
	"net/http"
	"time"

	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WorkloadIDParam is the chi URL param name carrying the workload UUID.
const WorkloadIDParam = "workloadId"

// authorizerTimeout bounds the upstream GetWorkload call so a hung control-plane
// can't pin a cluster-api request handler indefinitely.
const authorizerTimeout = 10 * time.Second

// WorkloadAuthorizerMiddleware authorizes the caller for the workload referenced
// by {workloadId} via the workloads-service GetWorkload call (403 relayed
// verbatim), and stores the resolved workload on the request context.
//
// There is no separate authorization check — authorization happens implicitly via the
// upstream GetWorkload response. Handlers downstream can assume the caller is
// allowed to access the workload in ctx.
func WorkloadAuthorizerMiddleware(authorizer WorkloadAuthorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			workloadID := chi.URLParam(r, WorkloadIDParam)
			if workloadID == "" {
				http.Error(w, "missing workloadId in path", http.StatusBadRequest)
				return
			}
			if _, err := uuid.Parse(workloadID); err != nil {
				http.Error(w, fmt.Sprintf("workloadId must be a valid UUID: %v", err), http.StatusBadRequest)
				return
			}

			rawToken := auth.RawTokenFromContext(r.Context())
			if rawToken == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			subject := auth.SubjectFromContext(r.Context())
			if subject == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Bind the token under the generated workloads-client key once here so
			// downstream handlers can pass ctx straight to workloads-client calls.
			handlerCtx := context.WithValue(r.Context(), workloadsapi.ContextAccessToken, rawToken)

			authorizerCtx, cancel := context.WithTimeout(handlerCtx, authorizerTimeout)
			defer cancel()

			authorizationResponse, err := authorizer.Authorize(authorizerCtx, subject, workloadID)
			if err != nil {
				logging.Errorf("k8sobject: authorize workload %s: %v", workloadID, err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if authorizationResponse.HTTPStatus != http.StatusOK {
				http.Error(w, fmt.Sprintf("resolving workload: '%s'", authorizationResponse.UpstreamMsg), authorizationResponse.HTTPStatus)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithWorkloadInfo(handlerCtx, authorizationResponse.Workload)))
		})
	}
}
