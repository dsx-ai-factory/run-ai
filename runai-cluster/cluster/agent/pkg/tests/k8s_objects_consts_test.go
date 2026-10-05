/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

// Backend-side fixtures for the K8sObjectsHandler integration flow.
// Lives in its own file to keep const_helpers_test.go focused on the
// legacy asset families; the k8s-objects wire shape (K8sObjectSpec =
// objectMeta + untyped objectSpec) does not reuse any of those helpers.

import (
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assets_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/google/uuid"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	rctAssetIdCreateProj    = "rct-asset-id-create-proj"
	rctAssetIdCreateCluster = "rct-asset-id-create-cluster"
	rctAssetIdToUpdate      = "rct-asset-id-to-update"
	rctAssetIdToDelete      = "rct-asset-id-to-delete"
	rctAssetIdAlreadyInSync = "rct-asset-id-already-in-sync"

	rctNameNewProj         = "newly-synced-proj-rct"
	rctNameNewCluster      = "newly-synced-cluster-rct"
	rctNameToUpdate        = "existing-rct-to-update"
	rctNameToDelete        = "existing-rct-to-delete"
	rctNameAlreadyInSync   = "existing-rct-already-in-sync"
	rctHighestSyncedCounter = 5
)

// rctUndefaultedSpec is what the control plane ships: AllocationMode,
// Count, and Toleration.Operator are left at their zero values and get
// filled in by the API server on the apply path. The agent's compare
// path runs the same defaulter before diffing, so an undefaulted expected
// and a defaulted existing must still be treated as equivalent.
func rctUndefaultedSpec(deviceClassName string, deviceCount int64) resourcev1.ResourceClaimSpec {
	req := resourcev1.DeviceRequest{
		Name: "gpu",
		Exactly: &resourcev1.ExactDeviceRequest{
			DeviceClassName: deviceClassName,
			Tolerations: []resourcev1.DeviceToleration{{
				Key:   "nvidia.com/gpu",
				Value: "present",
			}},
		},
	}
	if deviceCount > 0 {
		req.Exactly.Count = deviceCount
	}
	return resourcev1.ResourceClaimSpec{
		Devices: resourcev1.DeviceClaim{Requests: []resourcev1.DeviceRequest{req}},
	}
}

// rctDefaultedSpec mirrors what the API server would store after it
// defaults the undefaulted payload above. Pre-seeding an "existing"
// RCT with this is what lets the end-to-end defaults-no-churn test
// prove the agent's ShouldReCreate correctly treats the two as equal.
func rctDefaultedSpec(deviceClassName string, deviceCount int64) resourcev1.ResourceClaimSpec {
	spec := rctUndefaultedSpec(deviceClassName, deviceCount)
	ex := spec.Devices.Requests[0].Exactly
	ex.AllocationMode = resourcev1.DeviceAllocationModeExactCount
	if ex.Count == 0 {
		ex.Count = 1
	}
	ex.Tolerations[0].Operator = resourcev1.DeviceTolerationOpEqual
	return spec
}

// rctObjectSpecMap round-trips a typed ResourceClaimSpec into the
// map[string]interface{} shape the generated K8sObjectSpec wraps.
// We go through handlers' marshal pair indirectly by hand-building the
// map so the test does not depend on handler internals: the control
// plane really does send a free-form JSON object here.
func rctObjectSpecMap(spec resourcev1.ResourceClaimSpec) map[string]interface{} {
	exactly := spec.Devices.Requests[0].Exactly
	exactlyMap := map[string]interface{}{
		"deviceClassName": exactly.DeviceClassName,
	}
	if exactly.AllocationMode != "" {
		exactlyMap["allocationMode"] = string(exactly.AllocationMode)
	}
	if exactly.Count != 0 {
		exactlyMap["count"] = exactly.Count
	}
	tolerations := make([]interface{}, 0, len(exactly.Tolerations))
	for _, t := range exactly.Tolerations {
		tol := map[string]interface{}{
			"key":   t.Key,
			"value": t.Value,
		}
		if t.Operator != "" {
			tol["operator"] = string(t.Operator)
		}
		tolerations = append(tolerations, tol)
	}
	exactlyMap["tolerations"] = tolerations
	return map[string]interface{}{
		"devices": map[string]interface{}{
			"requests": []interface{}{
				map[string]interface{}{
					"name":    spec.Devices.Requests[0].Name,
					"exactly": exactlyMap,
				},
			},
		},
	}
}

// buildK8sObjectSpec wraps the typed spec in the generated envelope and
// lets a caller optionally seed the inner objectMeta (labels /
// annotations that each spawned ResourceClaim inherits).
func buildK8sObjectSpec(spec resourcev1.ResourceClaimSpec, innerMeta *assets_openapi.K8sObjectMeta) *assets_openapi.K8sObjectSpec {
	k8sSpec := assets_openapi.NewK8sObjectSpec(rctObjectSpecMap(spec))
	if innerMeta != nil {
		k8sSpec.SetObjectMeta(*innerMeta)
	}
	return k8sSpec
}

var (
	toCreateProjRct = assets_openapi.K8sObjectsUpdate{
		UpdateCount: 1,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      rctNameNewProj,
			Id:        rctAssetIdCreateProj,
			Kind:      assets_openapi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		},
		Spec: buildK8sObjectSpec(
			rctUndefaultedSpec("nvidia.com/gpu", 0),
			&assets_openapi.K8sObjectMeta{
				Labels: map[string]string{"pod-selector": "gpu-worker"},
			},
		),
	}

	toCreateClusterRct = assets_openapi.K8sObjectsUpdate{
		UpdateCount: 2,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_CLUSTER,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			Name:      rctNameNewCluster,
			Id:        rctAssetIdCreateCluster,
			Kind:      assets_openapi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		},
		Spec: buildK8sObjectSpec(rctUndefaultedSpec("nvidia.com/gpu", 2), nil),
	}

	toDeleteExistingRct = assets_openapi.K8sObjectsUpdate{
		UpdateCount: 3,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      rctNameToDelete,
			Id:        rctAssetIdToDelete,
			Kind:      assets_openapi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		},
		IsDeleted: *assets_openapi.NewNullableBool(convutils.Ref(true)),
	}

	// toUpdateExistingRct bumps the device count from 2 to 4. Because
	// ExactCount is immutable on the apiserver side, the handler must
	// delete-and-recreate rather than issue an in-place update — and the
	// test proves that by watching the UID change.
	toUpdateExistingRct = assets_openapi.K8sObjectsUpdate{
		UpdateCount: 4,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      rctNameToUpdate,
			Id:        rctAssetIdToUpdate,
			Kind:      assets_openapi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		},
		Spec: buildK8sObjectSpec(rctUndefaultedSpec("nvidia.com/gpu", 4), nil),
	}

	// toReconcileAlreadyInSyncRct ships the undefaulted spec whose
	// defaulted twin is already on the cluster (seeded as
	// existingRctAlreadyInSync below). The handler must not churn —
	// same UID expected after the sync — because ShouldReCreate runs
	// the API-server defaulter on both sides before diffing.
	toReconcileAlreadyInSyncRct = assets_openapi.K8sObjectsUpdate{
		UpdateCount: rctHighestSyncedCounter,
		Meta: assets_openapi.AssetMetaForSync{
			Scope:     assets_openapi.SCOPE_PROJECT,
			ClusterId: *assets_openapi.NewNullableString(convutils.Ref(TestClusterId)),
			ProjectId: *assets_openapi.NewNullableInt32(convutils.Ref(int32(Proj1Id))),
			Name:      rctNameAlreadyInSync,
			Id:        rctAssetIdAlreadyInSync,
			Kind:      assets_openapi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		},
		Spec: buildK8sObjectSpec(rctUndefaultedSpec("nvidia.com/gpu", 0), nil),
	}

	allK8sObjectsUpdates = []assets_openapi.K8sObjectsUpdate{
		toCreateProjRct,
		toCreateClusterRct,
		toDeleteExistingRct,
		toUpdateExistingRct,
		toReconcileAlreadyInSyncRct,
	}
)

// seedBasicRct builds a pre-existing cluster-side ResourceClaimTemplate
// with the Run:ai scope/kind/id/update-count metadata that the agent
// stamps on anything it syncs. Mirrors getBasicPvcObject /
// getBasicSecretObject in k8s_utils_test.go.
func seedBasicRct(name, namespace, assetId string, updateCounter int64,
	scope assets_openapi.ScopeType, scopeId string, spec resourcev1.ResourceClaimSpec,
) *resourcev1.ResourceClaimTemplate {
	rct := &resourcev1.ResourceClaimTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ResourceClaimTemplate",
			APIVersion: "resource.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uuid.New().String()),
			Labels: map[string]string{
				constants.RunaiAssetLabel: "resource-claim-template",
				"run.ai/asset-id":         assetId,
			},
			Annotations: map[string]string{
				"run.ai/asset-name":              name,
				handlers.UpdateCountAnnotation:   strconv.FormatInt(updateCounter, 10),
			},
		},
		Spec: resourcev1.ResourceClaimTemplateSpec{Spec: spec},
	}

	switch scope {
	case assets_openapi.SCOPETYPE_PROJECT:
		rct.Labels["run.ai/project"] = scopeId
	case assets_openapi.SCOPETYPE_DEPARTMENT:
		rct.Labels["run.ai/department"] = scopeId
	case assets_openapi.SCOPETYPE_CLUSTER:
		rct.Labels["run.ai/cluster-wide"] = "true"
	case assets_openapi.SCOPETYPE_TENANT:
		rct.Labels["run.ai/tenant-wide"] = "true"
	}

	return rct
}

var (
	existingRctToUpdate = seedBasicRct(
		rctNameToUpdate, NamespaceForSpec1, rctAssetIdToUpdate,
		int64(0),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		rctDefaultedSpec("nvidia.com/gpu", 2),
	)

	existingRctToDelete = seedBasicRct(
		rctNameToDelete, NamespaceForSpec1, rctAssetIdToDelete,
		int64(0),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		rctDefaultedSpec("nvidia.com/gpu", 1),
	)

	existingRctAlreadyInSync = seedBasicRct(
		rctNameAlreadyInSync, NamespaceForSpec1, rctAssetIdAlreadyInSync,
		int64(0),
		assets_openapi.SCOPETYPE_PROJECT, Proj1IdStr,
		rctDefaultedSpec("nvidia.com/gpu", 0),
	)
)

// expectedUpdateCountAnnotation makes the counter-annotation lookup
// readable at the It() site.
func expectedUpdateCountAnnotation(update assets_openapi.K8sObjectsUpdate) string {
	return fmt.Sprintf("%d", update.UpdateCount)
}
