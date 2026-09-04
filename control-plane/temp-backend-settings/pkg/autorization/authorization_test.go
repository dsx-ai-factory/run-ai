/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package autorization

import (
	"context"
	"testing"

	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
)

// ==================== ResolveTenantID ====================

func TestResolveTenantID_NonManagement_UsesContext(t *testing.T) {
	ctx := ctxWithTenantID("42")
	tenantID, err := ResolveTenantID(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, 42, tenantID)
}

func TestResolveTenantID_NonManagement_IgnoresTenantIdParam(t *testing.T) {
	ctx := ctxWithTenantID("42")
	tenantID, err := ResolveTenantID(ctx, "99")
	assert.NoError(t, err)
	assert.Equal(t, 42, tenantID)
}

func TestResolveTenantID_NonManagement_MissingContext(t *testing.T) {
	_, err := ResolveTenantID(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant id")
}

func TestResolveTenantID_Management_UsesParam(t *testing.T) {
	ctx := ctxManagementTenant()
	tenantID, err := ResolveTenantID(ctx, "99")
	assert.NoError(t, err)
	assert.Equal(t, 99, tenantID)
}

func TestResolveTenantID_Management_InvalidParam(t *testing.T) {
	ctx := ctxManagementTenant()
	_, err := ResolveTenantID(ctx, "not-a-number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant Id")
}

func TestResolveTenantID_Management_MissingParam_FallsBackToContext(t *testing.T) {
	ctx := ctxManagementTenantWithTenantID("42")
	tenantID, err := ResolveTenantID(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, 42, tenantID)
}

func TestResolveTenantID_Management_MissingParam_NoContext(t *testing.T) {
	ctx := ctxManagementTenant()
	_, err := ResolveTenantID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant id")
}

// ==================== IsManagementTenant ====================

func TestIsManagementTenant_True(t *testing.T) {
	ctx := ctxManagementTenant()
	assert.True(t, IsManagementTenant(ctx))
}

func TestIsManagementTenant_False(t *testing.T) {
	assert.False(t, IsManagementTenant(context.Background()))
}

func TestIsManagementTenant_NotTrue(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "false")
	assert.False(t, IsManagementTenant(ctx))
}

// ==================== extractTenantID ====================

func TestExtractTenantID_Success(t *testing.T) {
	ctx := ctxWithTenantID("42")
	tenantID, err := extractTenantID(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 42, tenantID)
}

func TestExtractTenantID_Missing(t *testing.T) {
	_, err := extractTenantID(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant id not found in context")
}

func TestExtractTenantID_Empty(t *testing.T) {
	ctx := ctxWithTenantID("")
	_, err := extractTenantID(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tenant id not found in context")
}

func TestExtractTenantID_Invalid(t *testing.T) {
	ctx := ctxWithTenantID("abc")
	_, err := extractTenantID(ctx)
	assert.Error(t, err)
}

// ==================== Helper Functions ====================

func ctxWithTenantID(tenantID string) context.Context {
	return context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)
}

func ctxManagementTenant() context.Context {
	return context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
}

func ctxManagementTenantWithTenantID(tenantID string) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfIsManagementTenant, "true")
	return context.WithValue(ctx, runaictx.CfTenantId, tenantID)
}
