/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func ParseTenantId(ctx context.Context) (uint, *api.Error) {
	var tenantIdStr *string

	if tenantIdStr = runaictx.GetString(ctx, runaictx.CfTenantId); tenantIdStr == nil {
		return 0, api.NewError(http.StatusBadRequest, "Missing tenant id", nil)
	}

	tenantId, parseErr := strconv.Atoi(*tenantIdStr)
	if parseErr != nil {
		return 0, api.NewError(http.StatusBadRequest, "Invalid tenant id", nil)
	}

	return uint(tenantId), nil
}

func ParseAuthData(ctx context.Context) (uint, string, *api.Error) {
	var tenantIdStr *string
	var subjectId *string

	if tenantIdStr = runaictx.GetString(ctx, runaictx.CfTenantId); tenantIdStr == nil {
		return 0, "", api.NewError(http.StatusBadRequest, "Missing tenant id", nil)
	}
	tenantId, parseErr := strconv.Atoi(*tenantIdStr)
	if parseErr != nil {
		return 0, "", api.NewError(http.StatusBadRequest, "Invalid tenant id", nil)
	}

	if subjectId = runaictx.GetString(ctx, runaictx.CfSubjectID); subjectId == nil {
		return 0, "", api.NewError(http.StatusBadRequest, "Missing subject id", nil)
	}

	return uint(tenantId), *subjectId, nil
}
