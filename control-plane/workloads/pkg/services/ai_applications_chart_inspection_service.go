/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"errors"
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
				return "", api.NewError(http.StatusBadGateway, "saved NGC chart credential has no key material", nil)
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
	return s.chartInspection.inspect(ctx, request)
}

func (s *applicationChartInspector) inspect(ctx context.Context, request api.AiApplicationChartInspectionRequest) (api.ImplResponse, error) {
	tenant, e := services_utils.ParseTenantId(ctx)
	if e != nil {
		return NewErrorResponse(e), nil
	}
	if actor := runaictx.GetString(ctx, runaictx.CfSubjectID); actor == nil || *actor == "" {
		if client := runaictx.GetString(ctx, runaictx.CfClientId); client == nil || *client == "" {
			return NewErrorResponse(api.NewError(http.StatusUnauthorized, "chart inspection requires authenticated caller identity", nil)), nil
		}
	}
	selection, e := chartInspectionSelection(request.ChartRef)
	if e != nil {
		return NewErrorResponse(e), nil
	}
	if request.Auth.UserCredential != nil && *request.Auth.UserCredential && request.Auth.AssetId == "" {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "assetId is required for a personal chart credential", nil)), nil
	}
	if selection.ChartID != "" && request.Auth.AssetId != "" {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, "uploaded charts do not accept source credentials", nil)), nil
	}

	// Authorize the source before every cache lookup; metadata is not an access grant.
	key := ""
	if request.Auth.AssetId != "" {
		if id, err := uuid.Parse(request.Auth.AssetId); err != nil || id == uuid.Nil {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, "assetId must be a valid uuid", nil)), nil
		}
		if s.loadKey == nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError, "chart credential lookup is not configured", nil)), nil
		}
		key, e = s.loadKey(ctx, request.Auth.AssetId)
		if e != nil {
			return NewErrorResponse(e), nil
		}
	}
	scope, err := chartInspectionScope(ctx, request.Auth)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "failed to resolve chart inspection caller scope", nil)), nil
	}

	var result *chartinspection.Result
	if selection.ChartID != "" {
		if s.loadUpload == nil {
			return NewErrorResponse(api.NewError(http.StatusInternalServerError, "uploaded chart lookup is not configured", nil)), nil
		}
		id, parseErr := uuid.Parse(selection.ChartID)
		if parseErr != nil {
			return NewErrorResponse(api.NewError(http.StatusBadRequest, "chartId must be a valid uuid", nil)), nil
		}
		result, err = s.inspector.InspectWithFetch(ctx, selection, scope, "", func(fetchCtx context.Context, _ chartinspection.Selection, _ string) ([]byte, error) {
			data, loadErr := s.loadUpload(fetchCtx, id, fmt.Sprint(tenant))
			if loadErr != nil {
				return nil, loadErr
			}
			return data, nil
		})
	} else {
		result, err = s.inspector.Inspect(ctx, selection, scope, key)
	}
	if err != nil {
		var apiErr *api.Error
		if errors.As(err, &apiErr) {
			return NewErrorResponse(apiErr), nil
		}
		var sourceErr *chartinspection.Error
		if errors.As(err, &sourceErr) {
			return NewErrorResponse(api.NewError(sourceErr.StatusCode, sourceErr.Code+": "+sourceErr.Message, nil)), nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return NewErrorResponse(api.NewError(http.StatusRequestTimeout, "chart inspection request cancelled or expired", nil)), nil
		}
		return NewErrorResponse(api.NewError(http.StatusUnprocessableEntity, "invalid Helm chart archive", nil).WithDetails(err)), nil
	}
	return api.Response(http.StatusOK, result), nil
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
			return selection, api.NewError(http.StatusBadRequest, "repository inspection requires repo, chartName and exact chartVersion", nil)
		}
	}
	if sources != 1 {
		return selection, api.NewError(http.StatusBadRequest, "provide exactly one chart source: chartId, chartURL, or repo with chartName and chartVersion", nil)
	}
	if selection.ChartID != "" {
		if id, err := uuid.Parse(selection.ChartID); err != nil || id == uuid.Nil {
			return selection, api.NewError(http.StatusBadRequest, "chartId must be a valid uuid", nil)
		}
	}
	return selection, nil
}

func chartInspectionScope(ctx context.Context, auth api.AiApplicationAuth) (string, error) {
	identity := make(map[string]any)
	for _, field := range []runaictx.ContextField{
		runaictx.CfTenantId, runaictx.CfSubjectID, runaictx.CfSubjectType, runaictx.CfClientId, runaictx.CfClientType, runaictx.CfGroups, runaictx.CfSupplementalGroups,
	} {
		identity[string(field)] = runaictx.GetField(ctx, field)
	}
	identity["assetId"] = auth.AssetId
	identity["userCredential"] = auth.UserCredential != nil && *auth.UserCredential
	encoded, err := json.Marshal(identity)
	return string(encoded), err
}
