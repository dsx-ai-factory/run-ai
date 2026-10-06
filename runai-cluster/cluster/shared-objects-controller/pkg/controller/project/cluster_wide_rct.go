/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"reflect"

	"k8s.io/apimachinery/pkg/selection"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	resourcev1 "k8s.io/api/resource/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	rctType ObjectType = "resourcev1.ResourceClaimTemplate"

	// ResourceClaimTemplateAssetKind matches the value the agent stamps in
	// the `run.ai/asset` label on sync (see
	// runai-cluster/cluster/agent/pkg/handlers/assets_builders.go). We
	// filter cluster-wide / tenant-wide / department-wide RCT sources
	// in the runai namespace by this label so the replicator never
	// touches an unrelated RCT (e.g. one created by hand or by a
	// different controller).
	ResourceClaimTemplateAssetKind = "resource-claim-template"
)

// RctObjectInspector handles equality and spec-copy for
// resource.k8s.io/v1 ResourceClaimTemplate.
//
// The whole RCT.Spec is immutable on resource.k8s.io/v1, so AreEqual MUST
// return false whenever the spec diverges — even on a one-field diff —
// otherwise the handler's in-place Update would be rejected by the
// apiserver. The compare runs the API-server defaulters on both sides so a
// zero-valued field in `from` doesn't look like a diff against an
// apiserver-defaulted `to`.
//
// Because the Spec is immutable, DeepCopySpec is only ever meaningful on
// the recreate path (delete + create a fresh object with the new spec);
// handleExisting routes RCT through that path via IsSpecImmutable below
// instead of calling Client.Update, which would fail.
type RctObjectInspector struct{}

var _ ObjectInspector = RctObjectInspector{}

func (r RctObjectInspector) DeepCopySpec(from, to client.Object) {
	fromRct, ok := from.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return
	}
	toRct, ok := to.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return
	}
	fromRct.Spec.DeepCopyInto(&toRct.Spec)
}

func (r RctObjectInspector) AreEqual(left, right client.Object) bool {
	leftRct, ok := left.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return false
	}
	rightRct, ok := right.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return false
	}

	// Name/namespace/owner equality is a cheap pre-check; the spec+
	// metadata compare below is where the defaulter work actually
	// matters. We DO NOT include labels/annotations here on purpose:
	// those are reconciled by the shared handler via
	// copyRunaiLabels / copyRunaiAnnotations, and gating recreate on
	// a metadata-only diff would churn the object unnecessarily.
	if leftRct.Name != rightRct.Name ||
		leftRct.Namespace != rightRct.Namespace ||
		!reflect.DeepEqual(leftRct.OwnerReferences, rightRct.OwnerReferences) {
		return false
	}

	// Delegating to the shared defaulter keeps this check in lock-step
	// with the agent-side rctBuilder.ShouldReCreate — same compare
	// semantics for the two code paths that have to agree on "the specs
	// are already equivalent".
	return !dra.RCTSpecDiffersAfterDefaults(leftRct, rightRct)
}

// IsSpecImmutable tells the shared ClusterWideReplicatedResourceHandler
// to use the delete-and-recreate flow instead of a plain Update when the
// source and destination specs diverge — RCT.Spec is wholly immutable on
// resource.k8s.io/v1.
func (r RctObjectInspector) IsSpecImmutable() bool { return true }

func (handler *ClusterWideGVKCommon) buildRctConfig() error {
	rctGvk, err := getGvk(&resourcev1.ResourceClaimTemplate{}, rctType)
	if err != nil {
		return err
	}

	// Only pick up RCTs stamped by the asset-sync agent. A stricter filter
	// than the per-scope label alone, so manually-created RCTs in the
	// runai namespace that happen to carry a cluster-wide label (e.g. a
	// user-authored one) are not swept up by the replicator.
	rctAssetLabelSelector, err := CreateLabelSelector(commonconstants.RunaiAssetLabel, selection.Equals,
		[]string{ResourceClaimTemplateAssetKind})
	if err != nil {
		return err
	}

	handler.typeToObjectInspector[rctType] = RctObjectInspector{}
	handler.gvkObjectConfigs[rctGvk] = []ObjectConfig{
		{
			ScopeWideLabel:         commonconstants.ClusterWideResourceLabel,
			AdditionalCallback:     nil,
			AdditionalObjectFilter: rctAssetLabelSelector,
			Scope:                  clusterScope,
		},
		{
			ScopeWideLabel:         commonconstants.TenantScopeLabel,
			AdditionalCallback:     nil,
			AdditionalObjectFilter: rctAssetLabelSelector,
			Scope:                  clusterScope,
		},
		{
			ScopeWideLabel:         commonconstants.DepartmentScopeLabel,
			AdditionalCallback:     nil,
			AdditionalObjectFilter: rctAssetLabelSelector,
			Scope:                  departmentScope,
		},
	}

	return nil
}
