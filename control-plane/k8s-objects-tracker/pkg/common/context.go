/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// insufficientPermissionsError represent and error in case of insufficient permissions
var insufficientPermissionsError = api.NewError(http.StatusForbidden, "Insufficient permissions", nil)

// TenantIdFromContext extracts the tenant id from the context
func TenantIdFromContext(ctx context.Context) (int32, *api.Error) {
	tenantIdStr := ctx.Value(runaictx.CfTenantId)
	if tenantIdStr == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("missing tenant ID header from request, rejected")
		return 0, insufficientPermissionsError
	}
	tenantId, err := strconv.Atoi(tenantIdStr.(string))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("invalid tenant ID value: %s", tenantIdStr)
		return 0, insufficientPermissionsError
	}
	return int32(tenantId), nil
}
