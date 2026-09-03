/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemRunAiWorkspaceV1Mapping is the system seed for the fictive
// run.ai/Workspace/v1 GVK, mapping WorkspaceSpecSpec (native workspace policy spec).
//
// This is a flat spec structure with all fields at the top level or nested
// under compute/security/storage. Unlike K8s workloads, there's no pod template
// hierarchy - the spec selector "workspaceSpec" points directly at the spec.
func BuildSystemRunAiWorkspaceV1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("run.ai"), Kind: "Workspace", Version: "v1"},
		buildRunAiWorkspaceV1SpecSelectors()
}

func buildRunAiWorkspaceV1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		"workspaceSpec": {
			Description: convutils.Ref("the native workspace spec (WorkspaceSpecSpec)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemNativeSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},
	}
}
