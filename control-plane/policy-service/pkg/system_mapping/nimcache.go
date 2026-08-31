/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping

import (
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// BuildSystemAppsNvidiaComNimCacheV1alpha1Mapping is the system seed for
// apps.nvidia.com/NIMCache/v1alpha1 (spec selectors without scope).
//
// NIMCache is a flat CRD (no .spec.template wrapping, no selector). Most
// configuration lives directly under .spec. Unlike pod-based workloads,
// NIMCache has no explicit container — the image that pulls the model lives
// at .spec.source.{ngc,dataStore,hf}.modelPuller. We still expose an
// "allContainers" spec selector (pointing at the same .spec anchor as
// workloadSpec) so policy writers can author rules that look consistent
// across workload types — it simply binds a different field group that
// focuses on the env + resources fields a caching Job ends up using.
func BuildSystemAppsNvidiaComNimCacheV1alpha1Mapping() (api.Gvk, map[string]api.SpecSelector) {
	return api.Gvk{Group: convutils.Ref("apps.nvidia.com"), Kind: "NIMCache", Version: "v1alpha1"}, buildAppsNvidiaComNimCacheV1alpha1SpecSelectors()
}

func buildAppsNvidiaComNimCacheV1alpha1SpecSelectors() map[string]api.SpecSelector {
	return map[string]api.SpecSelector{
		"workloadSpec": {
			Description: convutils.Ref("the NIMCache spec (source, storage, proxy, certConfig, runtime, tolerations, nodeSelector)"),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemNimCacheSpecFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "spec",
			},
		},

		"workloadMetadata": {
			Description: convutils.Ref("the NIMCache metadata (annotations and labels on the CR)"),
			Query: `(
			  .metadata
			)`,
			FieldGroups: []string{SystemMetadataFgName},
			Dependency: &api.SpecSelectorDependency{
				Path: "metadata",
			},
		},

		"allContainers": {
			Description: convutils.Ref("the NIMCache caching-Job container (env + resources). Anchored at .spec — NIMCache has no containers[] because the model-puller image comes from .spec.source.*.modelPuller."),
			Query: `(
			  .spec
			)`,
			FieldGroups: []string{SystemNimCacheContainerFgName},
		},
	}
}
