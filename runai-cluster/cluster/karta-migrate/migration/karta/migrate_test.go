/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/karta-migrate/migration"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// newClientsForSync wires a controller-runtime client (with the Karta
// status subresource) seeded with the given RIs and Kartas.
func newClientsForSync(
	t *testing.T,
	kartas []*kartav1alpha1.Karta,
	ris []*unstructured.Unstructured,
) client.Client {
	t.Helper()
	scheme := newScheme(t)

	objs := []client.Object{}
	for _, k := range kartas {
		objs = append(objs, k)
	}
	for _, r := range ris {
		objs = append(objs, r)
	}
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&kartav1alpha1.Karta{}).
		WithObjects(objs...).
		Build()
}

// fullyHealthyRI builds an RI with all 5 conditions True - the shape a real
// cluster RI has when the underlying workload type is wired up correctly.
// Used as the source of truth that the migration copies onto a Karta.
func fullyHealthyRI(name, group, version, kind string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(migration.LegacyRIGVK)
	u.SetName(name)
	_ = unstructured.SetNestedMap(u.Object,
		map[string]interface{}{"group": group, "version": version, "kind": kind},
		"spec", "structureDefinition", "rootComponent", "kind",
	)
	conds := []interface{}{
		conditionMap("RIValidated", string(metav1.ConditionTrue)),
		conditionMap("RBACReady", string(metav1.ConditionTrue)),
		conditionMap("SyncEnabled", string(metav1.ConditionTrue)),
		conditionMap("CRDExists", string(metav1.ConditionTrue)),
		conditionMap("Ready", string(metav1.ConditionTrue)),
	}
	_ = unstructured.SetNestedSlice(u.Object, conds, "status", "conditions")
	return u
}

// TestMigrate_ReconcilesStaleKarta: rerun bug. A previous migration left a
// Karta with one condition False; on rerun the migration must recopy the
// RI's truth onto it.
func TestMigrate_ReconcilesStaleKarta(t *testing.T) {
	ri := fullyHealthyRI("ri-ray", "ray.io", "v1", "RayCluster")
	stale := migratedKartaFor(t, ri)
	// Regress RBACReady on the Karta - it now does not cover the RI.
	for i := range stale.Status.Conditions {
		if stale.Status.Conditions[i].Type == string(runv1alpha1.RBACReady) {
			stale.Status.Conditions[i].Status = metav1.ConditionFalse
		}
	}
	c := newClientsForSync(t,
		[]*kartav1alpha1.Karta{stale},
		[]*unstructured.Unstructured{ri},
	)

	if err := Migrate(context.Background(), c); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	got, err := getKarta(context.Background(), c, stale.Name)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	expectCondition(t, got, runv1alpha1.RBACReady, metav1.ConditionTrue)
	expectCondition(t, got, runv1alpha1.RunaiReady, metav1.ConditionTrue)

	done, err := IsMigrationComplete(context.Background(), c)
	if err != nil {
		t.Fatalf("MigrationComplete: %v", err)
	}
	if !done {
		t.Fatal("expected MigrationComplete=true after migrate reconciled the stale karta")
	}
}

// TestMigrate_NoopOnConvergedSystem: every RI already has a Karta that
// covers it. Migrate should not patch anything.
func TestMigrate_NoopOnConvergedSystem(t *testing.T) {
	ri := fullyHealthyRI("ri-ray", "ray.io", "v1", "RayCluster")
	healthy := migratedKartaFor(t, ri)
	c := newClientsForSync(t,
		[]*kartav1alpha1.Karta{healthy},
		[]*unstructured.Unstructured{ri},
	)

	if err := Migrate(context.Background(), c); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	got, err := getKarta(context.Background(), c, healthy.Name)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	for _, ct := range []runv1alpha1.KartaConditionType{
		runv1alpha1.CRDExists, runv1alpha1.RBACReady, runv1alpha1.SyncEnabled,
		runv1alpha1.KartaValidated, runv1alpha1.RunaiReady,
	} {
		expectCondition(t, got, ct, metav1.ConditionTrue)
	}
}

// TestMigrate_CreatesMissingKarta: brand-new RI (e.g. agent created it
// during the upgrade window between pre and post hook). Karta is missing.
// Migrate creates it and copies conditions.
func TestMigrate_CreatesMissingKarta(t *testing.T) {
	ri := fullyHealthyRI("ri-tf", "kubeflow.org", "v1", "TFJob")
	c := newClientsForSync(t,
		nil,
		[]*unstructured.Unstructured{ri},
	)

	if err := Migrate(context.Background(), c); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	name, err := runv1alpha1.GenerateKartaNameByGvk("kubeflow.org", "v1", "TFJob")
	if err != nil {
		t.Fatalf("name: %v", err)
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
}
