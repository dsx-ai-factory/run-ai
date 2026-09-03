/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	tenantIdHeader           = "X-Runai-Tenant-Id"
	isManagementTenantHeader = "X-Runai-Is-Management-Tenant"
	subjectIdHeader          = "X-Runai-Subject-Id"
	subjectTypeHeader        = "X-Runai-Subject-Type"
	groupsHeader             = "X-Runai-Groups"
	supplementalGroupsHeader = "X-Runai-Supplemental-Groups"
	traceIdHeader            = "x-runai-trace-id"
	spanIdHeader             = "x-runai-span-id"
	clientIdHeader           = "X-Runai-Client-Id"
	gIdHeader                = "X-Runai-GId"
	uIdHeader                = "X-Runai-UId"
	sIdHeader                = "X-Runai-Session-Id"
	acceptHeader             = "Accept"
)

var isMultiValueHeader = map[string]bool{
	groupsHeader: true,
}

func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := extractHeadersContext(r.Context(), r)

		newR := r.WithContext(ctx)
		next.ServeHTTP(w, newR)

		addResponseHeaders(ctx, w)
	})
}

func addResponseHeaders(ctx context.Context, w http.ResponseWriter) {
	if value, ok := ctx.Value(runaictx.CfTraceId).(string); ok {
		w.Header().Set(traceIdHeader, value)
	}
}

func extractHeadersContext(ctx context.Context, r *http.Request) context.Context {
	ctx = setContextValue(ctx, r, runaictx.CfTenantId, tenantIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfIsManagementTenant, isManagementTenantHeader)
	ctx = setContextValue(ctx, r, runaictx.CfSubjectID, subjectIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfSubjectType, subjectTypeHeader)
	ctx = setContextValue(ctx, r, runaictx.CfGroups, groupsHeader)
	ctx = setContextValue(ctx, r, runaictx.CfSupplementalGroups, supplementalGroupsHeader)
	ctx = setContextValue(ctx, r, runaictx.CfTraceId, traceIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfSpanId, spanIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfClientId, clientIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfAccept, acceptHeader)
	ctx = setContextValue(ctx, r, runaictx.CfGId, gIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfUId, uIdHeader)
	ctx = setContextValue(ctx, r, runaictx.CfSId, sIdHeader)
	return ctx
}

func setContextValue(ctx context.Context, r *http.Request, cf runaictx.ContextField, headerName string) context.Context {
	if isMultiValue, found := isMultiValueHeader[headerName]; found && isMultiValue {
		return setContextMultiValue(ctx, r, cf, headerName)
	}

	value := r.Header.Get(headerName)
	if value == "" {
		return ctx
	}
	return context.WithValue(ctx, cf, value)
}

func setContextMultiValue(ctx context.Context, r *http.Request, cf runaictx.ContextField, headerName string) context.Context {
	values := r.Header.Values(headerName)
	if len(values) == 0 {
		return ctx
	}

	if len(values) == 1 {
		values = strings.Split(values[0], ",")
	}
	return context.WithValue(ctx, cf, values)
}
