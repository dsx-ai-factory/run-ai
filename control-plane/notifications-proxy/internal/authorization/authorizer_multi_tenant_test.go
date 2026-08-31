/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/authorization"
	"github.com/stretchr/testify/require"
)

func TestIsTenantM_WithTenantPermissions_Allowed(t *testing.T) {
	enableMultiTenancy()
	require.True(t, authorization.NewAuthorizer().IsTenant(ctxWithTenantPermissions))
}

func TestIsTenantM_WithoutPermissions_Denied(t *testing.T) {
	enableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsTenant(ctxNoPermissions))
}

func TestIsTenantM_WithSystemPermissions_Denied(t *testing.T) {
	enableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsTenant(ctxWithSystemPermissions))
}

func TestIsTenantM_WithoutTenantId_Denied(t *testing.T) {
	enableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsTenant(context.Background()))
}

func TestIsSystemM_WithSystemPermissions_Allowed(t *testing.T) {
	enableMultiTenancy()
	require.True(t, authorization.NewAuthorizer().IsSystem(ctxWithSystemPermissions))
}

func TestIsSystemM_WithoutPermissions_Denied(t *testing.T) {
	enableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsSystem(ctxNoPermissions))
}

func TestIsSystemM_WithTenantPermissions_Denied(t *testing.T) {
	enableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsSystem(ctxWithTenantPermissions))
}

func TestIsSystemM_WithoutTenantId_Denied(t *testing.T) {
	enableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsSystem(context.Background()))
}
