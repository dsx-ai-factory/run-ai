/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// k8sObjectsInspector is the ObjectInspector for the k8s-objects asset
// family. The up-to-date check reuses the shared asset
// `run.ai/update-count` annotation convention (same as credentials and
// datasources); the per-kind spec copy is routed through the builder
// registry so each kind owns its own typed DeepCopyInto path.
type k8sObjectsInspector struct {
	builders map[assetsApi.AssetKind]k8sObjectBuilder
}

func newK8sObjectsInspector(builders map[assetsApi.AssetKind]k8sObjectBuilder) k8sObjectsInspector {
	return k8sObjectsInspector{builders: builders}
}

// DeepCopySpec stamps the labels/annotations from the freshly-built object
// (from, the control-plane state) onto the latest cluster read (to) and
// then dispatches the kind-specific spec copy to the registered builder.
// Metadata copy comes first so a builder failure doesn't leave the object
// with stale Run:ai labels.
func (i k8sObjectsInspector) DeepCopySpec(from, to client.Object) {
	if annotations := from.GetAnnotations(); annotations != nil {
		merged := make(map[string]string, len(annotations))
		for k, v := range annotations {
			merged[k] = v
		}
		to.SetAnnotations(merged)
	}
	if labels := from.GetLabels(); labels != nil {
		merged := make(map[string]string, len(labels))
		for k, v := range labels {
			merged[k] = v
		}
		to.SetLabels(merged)
	}

	// Dispatch by the asset-kind label we stamped on from — matches on the
	// label string (not the AssetKind enum) because that's what's actually
	// on the k8s object. A missing/unknown label leaves the spec untouched;
	// the common handler will still have refreshed labels/annotations so
	// the next sync converges once the registry learns the kind.
	kindLabel := from.GetLabels()[constants.RunaiAssetLabel]
	if kindLabel == "" {
		return
	}
	for kind, builder := range i.builders {
		if getAssetKindLabel(kind) == kindLabel {
			builder.CopySpec(to, from)
			return
		}
	}
}

// IsUpToDate defers to the shared asset update-count comparison —
// AssetsObjectInspector carries the same `run.ai/update-count` convention
// and its implementation is the single source of truth for "newer than
// cluster" semantics across all asset handlers.
func (i k8sObjectsInspector) IsUpToDate(left, right client.Object) bool {
	return AssetsObjectInspector(false).IsUpToDate(left, right)
}
