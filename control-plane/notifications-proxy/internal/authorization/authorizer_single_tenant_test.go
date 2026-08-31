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

func TestIsTenantS_WithTenantPermissions_Allowed(t *testing.T) {
	disableMultiTenancy()
	require.True(t, authorization.NewAuthorizer().IsTenant(ctxWithTenantPermissions))
}

func TestIsTenantS_WithoutPermissions_Denied(t *testing.T) {
	disableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsTenant(ctxNoPermissions))
}

func TestIsTenantS_WithSystemPermissions_Allowed(t *testing.T) {
	disableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsTenant(ctxWithSystemPermissions))
}

func TestIsTenantS_WithoutTenantId_Denied(t *testing.T) {
	disableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsTenant(context.Background()))
}

func TestIsSystemS_WithSystemPermissions_Allowed(t *testing.T) {
	disableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsSystem(ctxWithSystemPermissions))
}

func TestIsSystemS_WithoutPermissions_Denied(t *testing.T) {
	disableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsSystem(ctxNoPermissions))
}

func TestIsSystemS_WithTenantPermissions_Denied(t *testing.T) {
	disableMultiTenancy()
	require.True(t, authorization.NewAuthorizer().IsSystem(ctxWithTenantPermissions))
}

func TestIsSystemS_WithoutTenantId_Denied(t *testing.T) {
	disableMultiTenancy()
	require.False(t, authorization.NewAuthorizer().IsSystem(context.Background()))
}
