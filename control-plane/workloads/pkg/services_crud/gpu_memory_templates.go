/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
)

func normalizeTemplateGPUMemory(source *api.SupersetTemplate) (*api.SupersetTemplate, *api.Error) {
	values, err := gojsonq.LoadJson(source)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "invalid template", err)
	}
	values, err = gpucompat.NormalizeLegacyGPUFields(values)
	if err != nil {
		return nil, api.NewValidationError(err.Error())
	}
	var result api.SupersetTemplate
	if err := convutils.CopySpecs(&result, values); err != nil {
		return nil, api.NewError(http.StatusBadRequest, "invalid template", err)
	}
	return &result, nil
}

func legacyCompatibleTemplate(source *api.SupersetTemplate) (*api.SupersetTemplate, *api.Error) {
	values, err := gojsonq.LoadJson(source)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to read template", err)
	}
	var result api.SupersetTemplate
	if err := convutils.CopySpecs(&result, gpucompat.LegacyCompatibleGPUFields(values)); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to prepare template response", err)
	}
	return &result, nil
}

func legacyCompatibleTemplateResponse(source *api.SupersetTemplate) (api.ImplResponse, error) {
	result, err := legacyCompatibleTemplate(source)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}
	return api.Response(http.StatusOK, *result), nil
}
