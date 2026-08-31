/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"slices"

	authorization_helpers "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	metrics_pkg "github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type MetricsAPIService struct {
	registry metrics_pkg.MetricsRegistrar
}

func NewMetricsAPIService(registry metrics_pkg.MetricsRegistrar) *MetricsAPIService {
	return &MetricsAPIService{
		registry: registry,
	}
}

func (s *MetricsAPIService) GetWorkloadsScrapeableMetrics(ctx context.Context) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest, Body: "missing tenant id"}, err
	}
	if !isPermitted(ctx, tenantId) {
		return api.Response(http.StatusForbidden, "permission denied"), nil
	}

	metricsString, err := s.registry.BuildMetricsString()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error building metrics string: %v", err)
		return api.Response(http.StatusInternalServerError, "something went wrong"), nil
	}
	return api.Response(http.StatusOK, metricsString), nil
}

func isPermitted(ctx context.Context, tenantId string) bool {
	multiTenancyEnabled := runai_config.Get[*config.MetricsExporter]().MultiTenancyEnabled
	if multiTenancyEnabled == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("multi tenancy enabled is not set")
		return false
	}
	if *multiTenancyEnabled {
		return authorization_helpers.HasSystemPermissions(ctx)
	}

	permittedScopes, ok := authorization_helpers.GetPermittedScopesFromCtx(ctx)
	if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing permitted scopes in request context")
		return false
	}
	return slices.Contains(permittedScopes.GetTenants(), tenantId)
}
