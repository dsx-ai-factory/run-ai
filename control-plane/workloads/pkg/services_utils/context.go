/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"context"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
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
