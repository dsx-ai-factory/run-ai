/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"
	"testing"

	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

func TestValidateRoleAssignable_RejectsInternallyAssignedRoles(t *testing.T) {
	for _, role := range []models.Role{
		static_roles.InstallerServiceAccount,
		static_roles.ClusterServiceAccount,
		static_roles.CloudOperator,
	} {
		err := validateRoleAssignable(&role)
		if err == nil {
			t.Fatalf("role %q must not be assignable through the public API", role.Name)
		}
		if err.Code != http.StatusForbidden {
			t.Fatalf("role %q: expected status %d, got %d", role.Name, http.StatusForbidden, err.Code)
		}
	}
}

func TestValidateRoleAssignable_AllowsRolesAssignableThroughThePublicAPI(t *testing.T) {
	// Integration metrics collector is internal-only (hidden from listings) but is intentionally assigned
	// through the public API.
	for _, role := range []models.Role{
		static_roles.IntegrationMetricsCollector,
		static_roles.SystemAdmin,
		static_roles.Viewer,
		static_roles.Editor,
	} {
		if err := validateRoleAssignable(&role); err != nil {
			t.Fatalf("role %q should be assignable, got %v", role.Name, err)
		}
	}
}

func TestValidateRoleAssignable_DoesNotBlockCustomRoleWithSameName(t *testing.T) {
	custom := models.Role{Name: static_roles.CloudOperator.Name, Custom: true, InternalOnly: false}
	if err := validateRoleAssignable(&custom); err != nil {
		t.Fatalf("a non internal-only role should be assignable, got %v", err)
	}
}
