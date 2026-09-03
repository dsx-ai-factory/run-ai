/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
)

func NoContentResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusNoContent,
	}
}

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	apiError.Details = ""
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}
