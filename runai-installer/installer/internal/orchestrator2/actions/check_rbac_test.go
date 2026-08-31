/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/stretchr/testify/assert"
)

// TestCheckRBAC_RequiredPermissionsPinned pins the exact permission set the
// installer asserts at preflight. The set is security-relevant: adding or
// removing a permission should be a deliberate, reviewed change, so this test
// is expected to be updated in the same commit that changes the list.
func TestCheckRBAC_RequiredPermissionsPinned(t *testing.T) {
	assert.Equal(t, []orchestrator2.ResourceAttributes{
		{Verb: "create", Resource: "namespaces"},
		{Verb: "create", Resource: "secrets"},
		{Verb: "create", Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"},
		{Verb: "delete", Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"},
		{Verb: "create", Group: "rbac.authorization.k8s.io", Resource: "clusterroles"},
		{Verb: "list", Group: "storage.k8s.io", Resource: "storageclasses"},
	}, requiredPermissions)
}
