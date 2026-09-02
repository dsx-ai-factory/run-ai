/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"testing"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/karta-migrate/migration"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// testRayKartaName is the canonical Karta name we use as the migration
// target across happy-path / unhealthy-RI / no-conditions / rename tests.
// Factored out per goconst lint.
const testRayKartaName = "krt-v1-ray-io-raycluster"

// riWithConditions builds a minimal RI fixture with the given conditions.
// Pass conditions as alternating type/status strings, e.g.
//
//	riWithConditions("ri-a", "RIValidated", "True", "CRDExists", "True")
func riWithConditions(name string, condPairs ...string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(migration.LegacyRIGVK)
	u.SetName(name)
	conditions := make([]interface{}, 0, len(condPairs)/2)
	for i := 0; i+1 < len(condPairs); i += 2 {
		conditions = append(conditions, map[string]interface{}{
			"type":   condPairs[i],
			"status": condPairs[i+1],
		})
	}
	_ = unstructured.SetNestedSlice(u.Object, conditions, "status", "conditions")
	return u
}

// validKartaFor builds a Karta with a valid spec for the given GVK -
// passes kartav1alpha1.NewKartaValidator. Used as the post-Create state
// before copyRIConditionsToKarta runs.
func validKartaFor(name, group, version, kind string) *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name:             "main",
					Kind:             &kartav1alpha1.GroupVersionKind{Group: group, Version: version, Kind: kind},
					StatusDefinition: &kartav1alpha1.StatusDefinition{},
				},
			},
		},
	}
}

func newClientsForConditions(t *testing.T, kartas []*kartav1alpha1.Karta) client.Client {
	t.Helper()
	scheme, err := migration.SchemeForTesting()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}
	objs := make([]client.Object, 0, len(kartas))
	for _, k := range kartas {
		objs = append(objs, k)
	}
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&kartav1alpha1.Karta{}).
		WithObjects(objs...).
		Build()
}

// TestCopyRIConditionsToKarta_HappyPath: an RI with all 4 base conditions
// True (RIValidated/CRDExists/RBACReady/SyncEnabled) plus Ready=True
// produces a Karta with all 5 conditions True - including a
// RunaiReady=True (renamed from the RI's Ready).
func TestCopyRIConditionsToKarta_HappyPath(t *testing.T) {
	name := testRayKartaName
	ri := riWithConditions("ri-a",
		"RIValidated", "True",
		"CRDExists", "True",
		"RBACReady", "True",
		"SyncEnabled", "True",
		"Ready", "True",
	)
	c := newClientsForConditions(t, []*kartav1alpha1.Karta{validKartaFor(name, "ray.io", "v1", "RayCluster")})

	if err := copyRIConditionsToKarta(context.Background(), c, name, ri); err != nil {
		t.Fatalf("copy: %v", err)
	}

	got, err := getKarta(context.Background(), c, name)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	for _, ct := range []runv1alpha1.KartaConditionType{
		runv1alpha1.CRDExists, runv1alpha1.RBACReady, runv1alpha1.SyncEnabled,
		runv1alpha1.KartaValidated, runv1alpha1.RunaiReady,
	} {
		expectCondition(t, got, ct, metav1.ConditionTrue)
	}
	expectCondition(t, got, runv1alpha1.Ready, metav1.ConditionTrue)
}

// TestCopyRIConditionsToKarta_RIValidatedRename: the RI condition name
// RIValidated must be renamed to KartaValidated on the Karta (one of the
// renames in riToKartaCondition; the other is Ready -> RunaiReady).
func TestCopyRIConditionsToKarta_RIValidatedRename(t *testing.T) {
	name := testRayKartaName
	ri := riWithConditions("ri-a", "RIValidated", "True")
	c := newClientsForConditions(t, []*kartav1alpha1.Karta{validKartaFor(name, "ray.io", "v1", "RayCluster")})

	if err := copyRIConditionsToKarta(context.Background(), c, name, ri); err != nil {
		t.Fatalf("copy: %v", err)
	}

	got, _ := getKarta(context.Background(), c, name)
	expectCondition(t, got, runv1alpha1.KartaValidated, metav1.ConditionTrue)
	// The Karta never has a literal "RIValidated" condition - it was renamed.
	for _, c := range got.Status.Conditions {
		if c.Type == "RIValidated" {
			t.Fatalf("expected RIValidated to be renamed to KartaValidated, but it appears literally on the Karta: %+v", c)
		}
	}
}

// TestCopyRIConditionsToKarta_UnhealthyRI: an RI whose CRD is missing has
// CRDExists=False / RBACReady=False / SyncEnabled=False / Ready=False -
// the migration faithfully mirrors that state onto the Karta (with the
// Ready -> RunaiReady rename).
func TestCopyRIConditionsToKarta_UnhealthyRI(t *testing.T) {
	name := "krt-v1-egx-nvidia-io-spotrequest"
	ri := riWithConditions("ri-spot",
		"RIValidated", "True",
		"CRDExists", "False",
		"RBACReady", "False",
		"SyncEnabled", "False",
		"Ready", "False",
	)
	c := newClientsForConditions(t, []*kartav1alpha1.Karta{validKartaFor(name, "egx.nvidia.io", "v1", "SPOTRequest")})

	if err := copyRIConditionsToKarta(context.Background(), c, name, ri); err != nil {
		t.Fatalf("copy: %v", err)
	}

	got, _ := getKarta(context.Background(), c, name)
	expectCondition(t, got, runv1alpha1.KartaValidated, metav1.ConditionTrue)
	expectCondition(t, got, runv1alpha1.CRDExists, metav1.ConditionFalse)
	expectCondition(t, got, runv1alpha1.RBACReady, metav1.ConditionFalse)
	expectCondition(t, got, runv1alpha1.SyncEnabled, metav1.ConditionFalse)
	expectCondition(t, got, runv1alpha1.RunaiReady, metav1.ConditionFalse)
	expectCondition(t, got, runv1alpha1.Ready, metav1.ConditionFalse)
}

// TestCopyRIConditionsToKarta_MissingConditions: an RI with no status
// conditions at all yields a Karta with the validator-computed
// KartaValidated condition plus the derived Ready condition. The migration
// is a literal copy for everything else but always sets KartaValidated using
// karta's own validator and derives Ready from it. Because CRDExists is
// absent, Ready is False. EWI's first reconcile post-upgrade fills in the
// canonical set.
func TestCopyRIConditionsToKarta_MissingConditions(t *testing.T) {
	name := testRayKartaName
	ri := riWithConditions("ri-uninit") // no conditions at all
	c := newClientsForConditions(t, []*kartav1alpha1.Karta{validKartaFor(name, "ray.io", "v1", "RayCluster")})

	if err := copyRIConditionsToKarta(context.Background(), c, name, ri); err != nil {
		t.Fatalf("copy: %v", err)
	}

	got, _ := getKarta(context.Background(), c, name)
	if len(got.Status.Conditions) != 2 {
		t.Fatalf("expected the KartaValidated and Ready conditions, got: %+v", got.Status.Conditions)
	}
	expectCondition(t, got, runv1alpha1.KartaValidated, metav1.ConditionTrue)
	expectCondition(t, got, runv1alpha1.Ready, metav1.ConditionFalse)
}

// TestCopyRIConditionsToKarta_KartaValidated_FromValidator: even if the RI
// claims RIValidated=True, the Karta's KartaValidated is recomputed via
// kartav1alpha1.NewKartaValidator(k).Validate(). A Karta with an invalid
// spec (empty rootComponent) must get KartaValidated=False regardless of
// what the RI said.
func TestCopyRIConditionsToKarta_KartaValidated_FromValidator(t *testing.T) {
	name := testRayKartaName
	ri := riWithConditions("ri-a", "RIValidated", "True")
	// Karta with EMPTY spec - validator will reject.
	invalidKarta := &kartav1alpha1.Karta{ObjectMeta: metav1.ObjectMeta{Name: name}}
	c := newClientsForConditions(t, []*kartav1alpha1.Karta{invalidKarta})

	if err := copyRIConditionsToKarta(context.Background(), c, name, ri); err != nil {
		t.Fatalf("copy: %v", err)
	}

	got, _ := getKarta(context.Background(), c, name)
	expectCondition(t, got, runv1alpha1.KartaValidated, metav1.ConditionFalse)
}

func expectCondition(t *testing.T, k *kartav1alpha1.Karta, ct runv1alpha1.KartaConditionType, want metav1.ConditionStatus) {
	t.Helper()
	for _, c := range k.Status.Conditions {
		if c.Type == string(ct) {
			if c.Status != want {
				t.Fatalf("condition %s: got %s want %s, message=%q", ct, c.Status, want, c.Message)
			}
			return
		}
	}
	t.Fatalf("condition %s not found, conditions: %+v", ct, k.Status.Conditions)
}
