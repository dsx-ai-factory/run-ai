/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
)

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}
