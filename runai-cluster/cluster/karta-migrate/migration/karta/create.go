/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/karta-migrate/migration/ri"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// generateKartaName derives the canonical Karta name from an RI's
// rootComponent.kind GVK.
func generateKartaName(riObj *unstructured.Unstructured) (string, error) {
	rootKind, found, err := unstructured.NestedMap(riObj.Object, "spec", "structureDefinition", "rootComponent", "kind")
	if err != nil {
		return "", err
	}
	if !found || rootKind == nil {
		return "", fmt.Errorf("RI %s has no spec.structureDefinition.rootComponent.kind", riObj.GetName())
	}
	group, _ := rootKind["group"].(string)
	version, _ := rootKind["version"].(string)
	kind, _ := rootKind["kind"].(string)
	return runv1alpha1.GenerateKartaNameByGvk(group, version, kind)
}

// getKarta fetches the Karta as the typed kartav1alpha1.Karta - used for
// controller-runtime status-subresource patches and condition reads.
func getKarta(ctx context.Context, c client.Client, name string) (*kartav1alpha1.Karta, error) {
	var k kartav1alpha1.Karta
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &k); err != nil {
		return nil, err
	}
	return &k, nil
}

// enrichKartaSpec fills in the karta CRD's required fields that some pre-v2.26
// RI specs are missing (rootComponent.name, statusDefinition.statusMappings).
// Operates on the typed *kartav1alpha1.KartaSpec (per natasharomm PR review).
// The defaults are bootstrap-only - the agent's next sync overwrites them.
func enrichKartaSpec(spec *kartav1alpha1.KartaSpec) {
	root := &spec.StructureDefinition.RootComponent
	if root.Name == "" && root.Kind != nil && root.Kind.Kind != "" {
		root.Name = strings.ToLower(root.Kind.Kind)
	}
	if root.StatusDefinition == nil {
		root.StatusDefinition = &kartav1alpha1.StatusDefinition{}
	}
	// StatusMappings is a non-pointer struct, leave it as its zero value; the
	// karta CRD accepts an empty StatusMappings. The agent's next sync fills
	// it with the canonical mappings for the workload kind.
}

// createKartaFromRI builds a Karta CR from the given legacy RI and posts it
// to the apiserver. The RI spec is round-tripped through json.Marshal /
// Unmarshal into the typed kartav1alpha1.KartaSpec (per natasharomm PR
// review) - avoids the brittle unstructured.SetNestedMap dance. After Create
// succeeds it copies the RI's status conditions onto the new Karta so EWI's
// first reconcile sees the same truth the RI had.
func createKartaFromRI(ctx context.Context, c client.Client, name string, riObj *unstructured.Unstructured) error {
	rootKind, _, _ := unstructured.NestedMap(riObj.Object, "spec", "structureDefinition", "rootComponent", "kind")
	group, _ := rootKind["group"].(string)
	version, _ := rootKind["version"].(string)
	kind, _ := rootKind["kind"].(string)

	specMap, found, err := unstructured.NestedMap(riObj.Object, "spec")
	if err != nil || !found {
		return fmt.Errorf("RI %s has no spec", riObj.GetName())
	}

	// Round-trip the RI's spec map into the typed KartaSpec via JSON. RI
	// and Karta share the same v1alpha1 spec shape, so this is a safe
	// no-loss conversion. From here on the spec is typed.
	specJSON, err := json.Marshal(specMap)
	if err != nil {
		return fmt.Errorf("marshal RI %s spec: %w", riObj.GetName(), err)
	}
	karta := &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	if err := json.Unmarshal(specJSON, &karta.Spec); err != nil {
		return fmt.Errorf("unmarshal RI %s spec into typed Karta: %w", riObj.GetName(), err)
	}
	enrichKartaSpec(&karta.Spec)

	workloadTypeID, _, _ := unstructured.NestedString(riObj.Object, "metadata", "labels", runv1alpha1.WorkloadTypeIdLabel)
	labels := runv1alpha1.GetKartaLabels(workloadTypeID, group, version, kind, "")
	labels[ri.GroupLabel] = group
	labels[ri.VersionLabel] = version
	labels[ri.KindLabel] = kind
	if updated, ok, _ := unstructured.NestedString(riObj.Object, "metadata", "labels", ri.UpdatedAtLabel); ok {
		labels[ri.UpdatedAtLabel] = updated
	}
	karta.SetLabels(labels)
	karta.SetFinalizers([]string{finalizer})

	if err := c.Create(ctx, karta); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create karta: %w", err)
	}

	return copyRIConditionsToKarta(ctx, c, name, riObj)
}
