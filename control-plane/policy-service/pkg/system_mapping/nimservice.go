/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

// BuildSystemAppsNvidiaComNimServiceV1alpha1Mapping is the system seed for
// apps.nvidia.com/NIMService/v1alpha1 (spec selectors without scope).
//
// NIMService is a flat CRD (no .spec.template wrapping, no selector): .spec
// mixes workload-level fields (replicas, autoscaling, expose), pod-level
// fields (nodeSelector, tolerations, podAffinity, userID/groupID), and
// "container-level" fields (image, command, args, env, resources). We
// expose the whole spec at a single anchor (.spec) and distinguish intent
// via two spec selectors with different field groups:
//
//   - workloadSpec   — everything non-container (storage, expose, scale, ...)
//   - allContainers  — image, command, args, env, resources
//
// "podMetadata" exists too, pointing at .spec, because NIMService stores the
// pod labels / annotations inside .spec.labels and .spec.annotations (not
// under .spec.template.metadata like K8s controllers). workloadMetadata
// targets .metadata of the CR itself.
func BuildSystemAppsNvidiaComNimServiceV1alpha1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("apps.nvidia.com"), Kind: "NIMService", Version: "v1alpha1"}, buildAppsNvidiaComNimServiceV1alpha1SpecSelectors()
}

func buildAppsNvidiaComNimServiceV1alpha1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		"workloadSpec": {
			Description: convutils.Ref("the NIMService spec (storage, expose, scale, multiNode, metrics, proxy, probes, runtime, tolerations, nodeSelector, etc.)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemNimServiceSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the NIMService metadata (annotations and labels on the CR)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		"podMetadata": {
			Description: convutils.Ref("the pod metadata propagated by NIMService (annotations and labels copied onto pods) at .spec.labels / .spec.annotations"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemMetadataFgName},
		},

		"allContainers": {
			Description: convutils.Ref("the NIMService container (image, command, args, env, resources). Anchored at .spec — NIMService has no containers[]."),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemNimServiceContainerFgName},
		},
	}
}
