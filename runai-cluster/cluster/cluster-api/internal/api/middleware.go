/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
)

const (
	workloadTypeKey      = "workloadType"
	workloadNameKey      = "workloadName"
	workloadFrameworkKey = "workloadFramework"
)

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) Middleware(next http.Handler) http.Handler {
	return m.middleware(next)
}

func (m *Middleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		rawType := chi.URLParam(r, workloadTypeKey)
		valid := ValidateWorkloadType(rawType)

		if !valid {
			setStatusBadRequest(w, fmt.Sprintf("invalid workload type: %s", rawType))
			return
		}

		workloadType := WorkloadType(rawType)

		workloadName := chi.URLParam(r, workloadNameKey)
		if workloadName == "" {
			setStatusBadRequest(w, "workload name is required")
			return
		}

		rawFramework := chi.URLParam(r, workloadFrameworkKey)

		valid = IsValidWorkloadFramework(rawFramework)
		if !valid {
			setStatusBadRequest(w, fmt.Sprintf("invalid workload framework: %s", rawFramework))
			return
		}

		workloadFramework := WorkloadFramework(rawFramework)

		endpoint := GetEndpointPath(r)
		action := FromStringAction(endpoint)

		if action == -1 {
			setStatusBadRequest(w, fmt.Sprintf("unsupported endpoint: %s", endpoint))
			return
		}

		WorkloadNameCtx := WithWorkloadName(ctx, workloadName)
		WorkloadTypeCtx := WithWorkloadType(WorkloadNameCtx, workloadType)
		WorkloadActionCtx := WithWorkloadAction(WorkloadTypeCtx, action)
		WorkloadFrameworkCtx := WithWorkloadFramework(WorkloadActionCtx, workloadFramework)

		next.ServeHTTP(w, r.WithContext(WorkloadFrameworkCtx))
	})
}

func setStatusBadRequest(w http.ResponseWriter, msg string) {
	log.Error().Msg(msg)
	http.Error(w, msg, http.StatusBadRequest)
}
