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

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s, err := migration.SchemeForTesting()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}
	if err := rbacv1.AddToScheme(s); err != nil {
		t.Fatalf("add rbac scheme: %v", err)
	}
	return s
}

// healthyRI builds an RI fixture with the conditions a real cluster RI has
// (RBACReady, SyncEnabled, CRDExists all True). This is the shape the
// migration job sees on every cluster.
func healthyRI(name, group, version, kind string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(migration.LegacyRIGVK)
	u.SetName(name)
	_ = unstructured.SetNestedMap(u.Object,
		map[string]interface{}{"group": group, "version": version, "kind": kind},
		"spec", "structureDefinition", "rootComponent", "kind",
	)
	conditions := []interface{}{
		conditionMap("RBACReady", string(metav1.ConditionTrue)),
		conditionMap("SyncEnabled", string(metav1.ConditionTrue)),
		conditionMap("CRDExists", string(metav1.ConditionTrue)),
	}
	_ = unstructured.SetNestedSlice(u.Object, conditions, "status", "conditions")
	return u
}

func conditionMap(ctype, cstatus string) map[string]interface{} {
	return map[string]interface{}{"type": ctype, "status": cstatus}
}

// migratedKartaFor builds a Karta fixture that mirrors a real post-migration
// Karta: same generated name as the RI, same GVK on the root component, all
// 5 conditions True (CRDExists, RBACReady, SyncEnabled, KartaValidated,
// RunaiReady) - the state the migration job patches.
func migratedKartaFor(t *testing.T, ri *unstructured.Unstructured) *kartav1alpha1.Karta {
	t.Helper()
	rootKind, _, _ := unstructured.NestedMap(ri.Object, "spec", "structureDefinition", "rootComponent", "kind")
	group, _ := rootKind["group"].(string)
	version, _ := rootKind["version"].(string)
	kind, _ := rootKind["kind"].(string)
	name, err := runv1alpha1.GenerateKartaNameByGvk(group, version, kind)
	if err != nil {
		t.Fatalf("name: %v", err)
	}

	k := &kartav1alpha1.Karta{}
	k.Name = name
	k.Spec.StructureDefinition.RootComponent = kartav1alpha1.ComponentDefinition{
		Name:             "main",
		Kind:             &kartav1alpha1.GroupVersionKind{Group: group, Version: version, Kind: kind},
		StatusDefinition: &kartav1alpha1.StatusDefinition{},
	}
	k.Status.Conditions = []metav1.Condition{
		{Type: string(runv1alpha1.CRDExists), Status: metav1.ConditionTrue, Reason: "Migrated"},
		{Type: string(runv1alpha1.RBACReady), Status: metav1.ConditionTrue, Reason: "Migrated"},
		{Type: string(runv1alpha1.SyncEnabled), Status: metav1.ConditionTrue, Reason: "Migrated"},
		{Type: string(runv1alpha1.KartaValidated), Status: metav1.ConditionTrue, Reason: "Migrated"},
		{Type: string(runv1alpha1.RunaiReady), Status: metav1.ConditionTrue, Reason: "Migrated"},
	}
	return k
}

func newClients(t *testing.T, ris []*unstructured.Unstructured, kartas []*kartav1alpha1.Karta) client.Client {
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
		WithObjects(objs...).
		Build()
}

// ---- MigrationComplete ----

func TestIsMigrationComplete_TrueWhenAllRIsHaveHealthyKartas(t *testing.T) {
	ri := healthyRI("ri-a", "ray.io", "v1", "RayCluster")
	k := migratedKartaFor(t, ri)
	c := newClients(t, []*unstructured.Unstructured{ri}, []*kartav1alpha1.Karta{k})

	done, err := IsMigrationComplete(context.Background(), c)
	if err != nil {
		t.Fatalf("MigrationComplete: %v", err)
	}
	if !done {
		t.Fatal("expected migration to be complete")
	}
}

// REPRODUCES THE BUG THAT THE BYTE-EQUAL SPEC CHECK USED TO TRIP ON.
//
// Real cluster context: the kserve InferenceService Karta CRD declares
// spec.statusMappings.statusMatchEntries.matcher.expressionMatcher.{
// messageFieldName, reasonFieldName} with defaults of "message" and
// "reason". The RI CRD does not declare those defaults. When the migration
// POSTs the Karta with the RI spec (which lacks those keys), the API
// server fills them in. The RI never gets them. So a byte-equal spec
// comparison would always fail for that one workload type, even though
// the migration produced a fully healthy Karta.
//
// The new health check ignores spec entirely and only verifies the Karta
// is at least as healthy as the RI plus RunaiReady=True. This case must pass.
func TestIsMigrationComplete_TrueDespiteSchemaDefaultingDifference(t *testing.T) {
	ri := healthyRI("ri-kserve-inferenceservice", "serving.kserve.io", "v1beta1", "InferenceService")
	// simulate RI spec WITHOUT the kserve defaults
	_ = unstructured.SetNestedMap(ri.Object, map[string]interface{}{
		"someOtherKnob": "some-value",
	}, "spec", "statusMappings", "statusMatchEntries", "matcher", "expressionMatcher")

	k := migratedKartaFor(t, ri)
	c := newClients(t, []*unstructured.Unstructured{ri}, []*kartav1alpha1.Karta{k})

	// the actual Karta on the API server would have messageFieldName /
	// reasonFieldName populated by CRD defaulting; the fake client doesn't
	// default, but the new check ignores spec so the difference is irrelevant
	// either way. Test passes because Karta covers RI on conditions.
	done, err := IsMigrationComplete(context.Background(), c)
	if err != nil {
		t.Fatalf("MigrationComplete: %v", err)
	}
	if !done {
		t.Fatal("expected migration to be complete despite schema-defaulting difference")
	}
}

func TestIsMigrationComplete_FalseWhenKartaMissing(t *testing.T) {
	ri := healthyRI("ri-a", "ray.io", "v1", "RayCluster")
	c := newClients(t, []*unstructured.Unstructured{ri}, nil)

	done, err := IsMigrationComplete(context.Background(), c)
	if err != nil {
		t.Fatalf("MigrationComplete: %v", err)
	}
	if done {
		t.Fatal("expected migration to be incomplete (no kartas)")
	}
}

func TestIsMigrationComplete_FalseWhenRIConditionNotMirroredOnKarta(t *testing.T) {
	ri := healthyRI("ri-a", "ray.io", "v1", "RayCluster")
	k := migratedKartaFor(t, ri)
	// Karta is Ready, but RBACReady regressed to False
	for i := range k.Status.Conditions {
		if k.Status.Conditions[i].Type == string(runv1alpha1.RBACReady) {
			k.Status.Conditions[i].Status = metav1.ConditionFalse
		}
	}
	c := newClients(t, []*unstructured.Unstructured{ri}, []*kartav1alpha1.Karta{k})

	done, err := IsMigrationComplete(context.Background(), c)
	if err != nil {
		t.Fatalf("MigrationComplete: %v", err)
	}
	if done {
		t.Fatal("expected migration to be incomplete (RI RBACReady=True not mirrored on Karta)")
	}
}

func TestIsMigrationComplete_FalseWhenPartiallyMigrated(t *testing.T) {
	riA := healthyRI("ri-a", "ray.io", "v1", "RayCluster")
	riB := healthyRI("ri-b", "kubeflow.org", "v1", "TFJob")
	kA := migratedKartaFor(t, riA)
	c := newClients(t, []*unstructured.Unstructured{riA, riB}, []*kartav1alpha1.Karta{kA})

	done, err := IsMigrationComplete(context.Background(), c)
	if err != nil {
		t.Fatalf("MigrationComplete: %v", err)
	}
	if done {
		t.Fatal("expected migration to be incomplete (riB has no karta)")
	}
}

// ---- kartaCoversRI primitive ----

func TestKartaCoversRI_HappyPath(t *testing.T) {
	ri := healthyRI("ri-a", "ray.io", "v1", "RayCluster")
	k := migratedKartaFor(t, ri)
	if ok, why := kartaCoversRI(k, ri); !ok {
		t.Fatalf("expected covers, got reason: %s", why)
	}
}

func TestKartaCoversRI_FalseWhenKartaMissingMirror(t *testing.T) {
	ri := healthyRI("ri-a", "ray.io", "v1", "RayCluster")
	k := migratedKartaFor(t, ri)
	for i := range k.Status.Conditions {
		if k.Status.Conditions[i].Type == "SyncEnabled" {
			k.Status.Conditions[i].Status = metav1.ConditionFalse
		}
	}
	if ok, _ := kartaCoversRI(k, ri); ok {
		t.Fatal("expected not-covers when SyncEnabled regressed on Karta")
	}
}
