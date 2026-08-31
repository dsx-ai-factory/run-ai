/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func SetHeaders(ctx context.Context, r *http.Request) {
	if r.Header == nil {
		r.Header = map[string][]string{}
	}

	setHeadersFromCtx(ctx, &r.Header)
}

func GetHeaders(ctx context.Context) *http.Header {
	headers := http.Header{}
	setHeadersFromCtx(ctx, &headers)
	return &headers
}

func setHeadersFromCtx(ctx context.Context, h *http.Header) {
	if value, ok := ctx.Value(runaictx.CfTenantId).(string); ok {
		h.Set(tenantIdHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfIsManagementTenant).(string); ok {
		h.Set(isManagementTenantHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfSubjectID).(string); ok {
		h.Set(subjectIdHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfSubjectType).(string); ok {
		h.Set(subjectTypeHeader, value)
	}

	if values, ok := ctx.Value(runaictx.CfGroups).([]string); ok {
		for _, value := range values {
			h.Add(groupsHeader, value)
		}
	}

	if value, ok := ctx.Value(runaictx.CfSupplementalGroups).(string); ok {
		h.Set(supplementalGroupsHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfClientId).(string); ok {
		h.Set(clientIdHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfGId).(string); ok {
		h.Set(gIdHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfUId).(string); ok {
		h.Set(uIdHeader, value)
	}

	if value, ok := ctx.Value(runaictx.CfSId).(string); ok {
		h.Set(sIdHeader, value)
	}
}
