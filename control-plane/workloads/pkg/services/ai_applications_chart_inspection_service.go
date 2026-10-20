/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	chartinspection "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/chart_inspection"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	assetsclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
	runaiconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

type applicationChartInspector struct {
	inspector  *chartinspection.Inspector
	loadKey    func(context.Context, string) (string, *api.Error)
	loadUpload func(context.Context, uuid.UUID, string) ([]byte, *api.Error)
}

func newApplicationChartInspector() *applicationChartInspector {
	conf := runaiconfig.Get[*config.ChartUploadConfig]()
	ttl := time.Duration(conf.InspectionCacheTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = chartinspection.DefaultCacheTTL
	}
	assets := assetsclient.NewAssetsClient(runaiconfig.Get[*config.ServicesConfig]().AssetsServiceUrl)
	return &applicationChartInspector{
		inspector: chartinspection.NewInspector(chartinspection.FetchRemote, ttl),
		loadKey: func(ctx context.Context, id string) (string, *api.Error) {
			key, e := assets.GetNgcApiKeyByIdInternal(ctx, id)
			if e != nil {
				return "", api.NewError(int(e.Code), "failed to access saved NGC chart credential", nil)
			}
			if key == nil || key.Key == "" {
				return "", api.NewError(http.StatusServiceUnavailable, "saved NGC chart credential has no key material", nil)
			}
			return key.Key, nil
		},
		loadUpload: func(ctx context.Context, id uuid.UUID, tenant string) ([]byte, *api.Error) {
			chart, e := dao.GetApplicationChartUploadByID(ctx, id, tenant)
			if e != nil {
				return nil, e
			}
			return chart.ChartBytes, nil
		},
	}
}

func (s *AIApplicationsService) InspectApplicationChart(ctx context.Context, request api.AiApplicationChartInspectionRequest) (api.ImplResponse, error) {
	if s.chartInspection == nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "chart inspection is not configured", nil)), nil
	}
	result, err := s.chartInspection.inspect(ctx, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	return api.Response(http.StatusOK, result), nil
}

func (s *applicationChartInspector) inspect(ctx context.Context, request api.AiApplicationChartInspectionRequest) (*api.AiApplicationChartInspectionResponse, *api.Error) {
	tenant, e := services_utils.ParseTenantId(ctx)
	if e != nil {
		return nil, e
	}
	if actor := runaictx.GetString(ctx, runaictx.CfSubjectID); actor == nil || *actor == "" {
		if client := runaictx.GetString(ctx, runaictx.CfClientId); client == nil || *client == "" {
			return nil, api.NewError(http.StatusUnauthorized, "chart inspection requires authenticated caller identity", nil)
		}
	}
	selection, e := chartInspectionSelection(request.ChartRef)
	if e != nil {
		return nil, e
	}
	if request.Auth.UserCredential != nil && *request.Auth.UserCredential && request.Auth.AssetId == "" {
		return nil, api.NewValidationError("assetId is required for a personal chart credential")
	}
	if selection.ChartID != "" && request.Auth.AssetId != "" {
		return nil, api.NewValidationError("uploaded charts do not accept source credentials")
	}

	// Assets checks tenant/RBAC access and personal ownership before every cache lookup
	// or in-flight join, independently of the caller-supplied userCredential flag.
	key := ""
	if request.Auth.AssetId != "" {
		if id, err := uuid.Parse(request.Auth.AssetId); err != nil || id == uuid.Nil {
			return nil, api.NewValidationError("assetId must be a valid uuid")
		}
		if s.loadKey == nil {
			return nil, api.NewError(http.StatusInternalServerError, "chart credential lookup is not configured", nil)
		}
		key, e = s.loadKey(ctx, request.Auth.AssetId)
		if e != nil {
			return nil, e
		}
	}
	scope, e := chartInspectionScope(ctx, request.Auth)
	if e != nil {
		return nil, e
	}

	var result *api.AiApplicationChartInspectionResponse
	if selection.ChartID != "" {
		if s.loadUpload == nil {
			return nil, api.NewError(http.StatusInternalServerError, "uploaded chart lookup is not configured", nil)
		}
		id, parseErr := uuid.Parse(selection.ChartID)
		if parseErr != nil {
			return nil, api.NewValidationError("chartId must be a valid uuid")
		}
		result, e = s.inspector.InspectWithFetch(ctx, selection, scope, "", func(fetchCtx context.Context, _ chartinspection.Selection, _ string) ([]byte, *api.Error) {
			data, loadErr := s.loadUpload(fetchCtx, id, fmt.Sprint(tenant))
			if loadErr != nil {
				return nil, loadErr
			}
			return data, nil
		})
	} else {
		result, e = s.inspector.Inspect(ctx, selection, scope, key)
	}
	if e != nil {
		return nil, e
	}
	return result, nil
}

func chartInspectionSelection(ref api.AiApplicationSpecChartRef) (chartinspection.Selection, *api.Error) {
	selection := chartinspection.Selection{
		ChartID: strings.TrimSpace(ref.ChartId), ChartURL: strings.TrimSpace(ref.ChartURL),
		Repo: strings.TrimSpace(ref.Repo), ChartName: strings.TrimSpace(ref.ChartName), ChartVersion: strings.TrimSpace(ref.ChartVersion),
	}
	sources := 0
	if selection.ChartID != "" {
		sources++
	}
	if selection.ChartURL != "" {
		sources++
	}
	if selection.Repo != "" || selection.ChartName != "" || selection.ChartVersion != "" {
		sources++
		if selection.Repo == "" || selection.ChartName == "" || selection.ChartVersion == "" {
			return selection, api.NewValidationError("repository inspection requires repo, chartName and exact chartVersion")
		}
	}
	if sources != 1 {
		return selection, api.NewValidationError("provide exactly one chart source: chartId, chartURL, or repo with chartName and chartVersion")
	}
	if selection.ChartID != "" {
		if id, err := uuid.Parse(selection.ChartID); err != nil || id == uuid.Nil {
			return selection, api.NewValidationError("chartId must be a valid uuid")
		}
	}
	return selection, nil
}

// Shared assets and credential-free downloads return the same metadata to authorized
// callers within a tenant. Personal credentials retain their existing caller isolation.
// Inspector adds the chart selection, actual credential fingerprint and parser version.
func chartInspectionScope(ctx context.Context, auth api.AiApplicationAuth) (string, *api.Error) {
	personal := auth.UserCredential != nil && *auth.UserCredential
	identity := map[string]any{
		string(runaictx.CfTenantId): runaictx.GetField(ctx, runaictx.CfTenantId),
		"assetId":                   auth.AssetId,
		"userCredential":            personal,
	}
	if personal {
		for _, field := range []runaictx.ContextField{
			runaictx.CfSubjectID, runaictx.CfSubjectType, runaictx.CfClientId, runaictx.CfClientType, runaictx.CfGroups, runaictx.CfSupplementalGroups,
		} {
			identity[string(field)] = runaictx.GetField(ctx, field)
		}
	}
	encoded, err := json.Marshal(identity)
	if err != nil {
		return "", api.NewError(http.StatusInternalServerError, "failed to resolve chart inspection caller scope", nil)
	}
	return string(encoded), nil
}
