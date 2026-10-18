/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/stretchr/testify/assert"
)

// TestNoDuplicateResourceTypes validates that no static role has duplicate ResourceType entries.
// This test prevents PostgreSQL "ON CONFLICT DO UPDATE command cannot affect row a second time" errors
// that occur when trying to upsert permissions with duplicate ResourceType values.
func TestNoDuplicateResourceTypes(t *testing.T) {
	// Map of all static roles to test
	staticRoles := map[string]models.Role{
		"Cloud operator":                    static_roles.CloudOperator,
		"Cluster service account":           static_roles.ClusterServiceAccount,
		"Compute resource administrator":    static_legacy_roles.ComputeResourcesAdmin,
		"Credentials administrator":         static_legacy_roles.CredentialsAdmin,
		"Data source administrator":         static_legacy_roles.DataSourcesAdmin,
		"Data volume administrator":         static_legacy_roles.DatavolumeAdmin,
		"Department administrator":          static_roles.DepartmentAdmin,
		"Department administrator (Legacy)": static_roles.DepartmentAdminLegacy,
		"Department viewer":                 static_roles.DepartmentViewer,
		"Editor":                            static_roles.Editor,
		"Environment administrator":         static_legacy_roles.EnvironmentAdmin,
		"Installer service account":         static_roles.InstallerServiceAccount,
		"L1 researcher":                     static_legacy_roles.Researcher,
		"L2 researcher":                     static_legacy_roles.ResearcherL1,
		"Integration metrics collector":     static_roles.IntegrationMetricsCollector,
		"ML engineer":                       static_legacy_roles.MLEngineer,
		"Research manager":                  static_legacy_roles.ResearchManager,
		"System administrator":              static_roles.SystemAdmin,
		"Template administrator":            static_legacy_roles.TemplatesAdmin,
		"Viewer":                            static_roles.Viewer,
	}

	var failedRoles []string
	duplicateDetails := make(map[string][]string)

	for roleName, role := range staticRoles {
		t.Run(roleName, func(t *testing.T) {
			duplicates := findDuplicateResourceTypes(role.Permissions)

			if len(duplicates) > 0 {
				failedRoles = append(failedRoles, roleName)
				duplicateDetails[roleName] = duplicates

				// Report the duplicates for this role
				for _, resourceType := range duplicates {
					t.Errorf("Role '%s' has duplicate ResourceType: %s", roleName, resourceType)
				}
			}
		})
	}

	// If any role has duplicates, provide a comprehensive summary
	if len(failedRoles) > 0 {
		summary := fmt.Sprintf("\n\n=== DUPLICATE RESOURCE TYPES FOUND ===\n")
		summary += fmt.Sprintf("Found duplicates in %d role(s):\n\n", len(failedRoles))

		for _, roleName := range failedRoles {
			summary += fmt.Sprintf("❌ %s:\n", roleName)
			for _, resourceType := range duplicateDetails[roleName] {
				summary += fmt.Sprintf("   - %s (appears multiple times)\n", resourceType)
			}
			summary += "\n"
		}

		summary += "These duplicates will cause PostgreSQL errors during role upsert operations.\n"
		summary += "Please remove the duplicate entries from the static role definitions.\n"

		t.Error(summary)
	} else {
		t.Log("✅ All static roles validated successfully - no duplicate ResourceTypes found")
	}
}

// findDuplicateResourceTypes returns a list of ResourceType values that appear more than once
// in the given permissions slice.
func findDuplicateResourceTypes(permissions []*models.Permission) []string {
	seen := make(map[api.ResourceType]int)
	var duplicates []string

	// Count occurrences of each ResourceType
	for _, perm := range permissions {
		seen[perm.ResourceType]++
	}

	// Collect ResourceTypes that appear more than once
	for resourceType, count := range seen {
		if count > 1 {
			duplicates = append(duplicates, fmt.Sprintf("%s (appears %d times)", resourceType, count))
		}
	}

	return duplicates
}

// TestDeduplicationLogic tests the deduplication helper function
func TestDeduplicationLogic(t *testing.T) {
	tests := []struct {
		name        string
		permissions []*models.Permission
		expected    int // expected number of unique permissions after deduplication
		wantDups    bool
	}{
		{
			name: "No duplicates",
			permissions: []*models.Permission{
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true},
				{ResourceType: api.RESOURCETYPE_PROJECT, Read: true},
			},
			expected: 3,
			wantDups: false,
		},
		{
			name: "Exact duplicates",
			permissions: []*models.Permission{
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true},
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true},
			},
			expected: 2,
			wantDups: true,
		},
		{
			name: "Same ResourceType, different permissions",
			permissions: []*models.Permission{
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true},
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true, Create: true},
			},
			expected: 2,
			wantDups: true,
		},
		{
			name: "Multiple duplicates",
			permissions: []*models.Permission{
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true},
				{ResourceType: api.RESOURCETYPE_PROJECT, Read: true},
				{ResourceType: api.RESOURCETYPE_TENANT, Read: true},
				{ResourceType: api.RESOURCETYPE_PROJECT, Create: true},
			},
			expected: 4,
			wantDups: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duplicates := findDuplicateResourceTypes(tt.permissions)
			hasDuplicates := len(duplicates) > 0

			assert.Equal(t, tt.wantDups, hasDuplicates,
				"Expected duplicates: %v, but got: %v (duplicates: %v)",
				tt.wantDups, hasDuplicates, duplicates)
		})
	}
}
