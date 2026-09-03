/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	"testing"
	"time"

	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/resourcing"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	testProjUID = "old-proj-uid"
	testProjNS  = "runai-proj-a"
)

// oldProject builds a run.ai/v2 project carrying a fully-populated spec so migration
// of every retained, renamed, and moved-to-label field can be asserted.
func oldProject(name string) *runaiv2.Project {
	return &runaiv2.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			UID:         testProjUID,
			Labels:      map[string]string{runaiv2.ProjectIdLabel: "proj-id-1", "team": "research"},
			Annotations: map[string]string{"run.ai/note": "hello"},
			Finalizers:  []string{"run.ai/project-finalize"},
		},
		Spec: runaiv2.ProjectSpec{
			Department:            "dep-a",
			EnforceRunaiScheduler: true,
			DefaultNodePools:      []string{"default", "gpu-pool"},
			Namespace:             "requested-ns",
			Queues: []runaiv2.QueueProjectSpec{
				{
					Name:     "q-default",
					Nodepool: "default",
					Priority: ptr.To(int32(50)),
					Resources: &resourcing.QueueResources{
						GPU: resourcing.SystemResource{Deserved: 2, OverQuotaWeight: 1, Limit: 4},
					},
				},
			},
			InteractiveJobTimeLimitSecs:                  3600,
			TrainingJobMaxIdleDurationSecs:               600,
			InteractivePreemptibleJobMaxIdleDurationSecs: 120,
			NodeAffinityInteractive:                      []string{"a", "b"},
			DeletionRequest:                              &runaiv2.DeletionRequest{Force: true, Type: runaiv2.Cascade},
		},
		Status: runaiv2.ProjectStatus{
			Namespace:  testProjNS,
			SyncStatus: runaiv2.ProjectSyncStatus{LastSyncTimestamp: metav1.NewTime(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC))},
		},
	}
}

func projectOwnerRef(proj *runaiv2.Project, blockOwnerDeletion bool) metav1.OwnerReference {
	ref := metav1.OwnerReference{
		APIVersion: runaiv2.GroupVersion.String(),
		Kind:       projectKind,
		Name:       proj.Name,
		UID:        proj.UID,
		Controller: &migrationTrueRef,
	}
	if blockOwnerDeletion {
		ref.BlockOwnerDeletion = &migrationTrueRef
	}
	return ref
}

// projectOwnedQueue builds a cluster-scoped queue owned by the given project and
// labelled with the project name, as project-controller stamps it.
func projectOwnedQueue(name string, proj *runaiv2.Project) *schedv2.Queue {
	return &schedv2.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Labels:          map[string]string{projectNameLabel: proj.Name},
			OwnerReferences: []metav1.OwnerReference{projectOwnerRef(proj, true)},
		},
	}
}

func projectOwnedLimitRange(name string, proj *runaiv2.Project) *corev1.LimitRange {
	return &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       testProjNS,
			OwnerReferences: []metav1.OwnerReference{projectOwnerRef(proj, true)},
		},
	}
}

// projectOwnedRoleBinding builds a RoleBinding in the project namespace owned by the
// project. accessrule-controller omits BlockOwnerDeletion; project-controller sets it.
func projectOwnedRoleBinding(name string, proj *runaiv2.Project, blockOwnerDeletion bool) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       testProjNS,
			OwnerReferences: []metav1.OwnerReference{projectOwnerRef(proj, blockOwnerDeletion)},
		},
	}
}

// replicatedSecret builds a per-project replicated Secret, tracked by the replication
// labels (not an owner ref).
func replicatedSecret(name string, proj *runaiv2.Project) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testProjNS,
			Labels: map[string]string{
				constants.ReplicatedProjectNameLabel: proj.Name,
				constants.ReplicatedProjectUIDLabel:  string(proj.UID),
			},
		},
	}
}

func mustGetProject(t *testing.T, k8sClient client.Client, name string) kaiv1alpha1.Project {
	t.Helper()
	var got kaiv1alpha1.Project
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name}, &got); err != nil {
		t.Fatalf("get kai project %s: %v", name, err)
	}
	return got
}

func TestMigrateProjects(t *testing.T) {
	ctx := context.Background()

	t.Run("creates the new kai project and re-parents its owned objects", func(t *testing.T) {
		old := oldProject("proj-a")
		queue := projectOwnedQueue("q-default", old)
		limitRange := projectOwnedLimitRange("kai-limit-range", old)
		rbCtrl := projectOwnedRoleBinding("rb-ctrl", old, true)
		rbAccess := projectOwnedRoleBinding("rb-access", old, false)
		secret := replicatedSecret("replicated-secret", old)
		c := newMigrationFakeClient(t, old, queue, limitRange, rbCtrl, rbAccess, secret)

		if err := migrateProjects(ctx, c); err != nil {
			t.Fatalf("migrateProjects: %v", err)
		}

		got := mustGetProject(t, c, "proj-a")

		// ObjectMeta carried over.
		if got.Annotations["run.ai/note"] != "hello" {
			t.Errorf("annotation run.ai/note = %q, want hello", got.Annotations["run.ai/note"])
		}
		if got.Labels["team"] != "research" {
			t.Errorf("label team = %q, want research", got.Labels["team"])
		}
		if len(got.Finalizers) != 1 || got.Finalizers[0] != "run.ai/project-finalize" {
			t.Errorf("finalizers = %v", got.Finalizers)
		}

		// Spec retained / renamed.
		if got.Spec.Parent != "dep-a" {
			t.Errorf("parent = %q, want dep-a", got.Spec.Parent)
		}
		if !got.Spec.EnforceKaiScheduler {
			t.Error("enforceKaiScheduler = false, want true")
		}
		if got.Spec.Namespace != "requested-ns" {
			t.Errorf("namespace = %q, want requested-ns", got.Spec.Namespace)
		}
		if len(got.Spec.DefaultNodePools) != 2 {
			t.Errorf("defaultNodePools = %v", got.Spec.DefaultNodePools)
		}
		if got.Spec.DeletionType == nil || *got.Spec.DeletionType != kaiv1alpha1.Blocking {
			t.Errorf("deletionType = %v, want Blocking", got.Spec.DeletionType)
		}
		if len(got.Spec.Queues) != 1 || got.Spec.Queues[0].Name != "q-default" {
			t.Fatalf("queues = %+v", got.Spec.Queues)
		}
		if got.Spec.Queues[0].Resources == nil ||
			got.Spec.Queues[0].Resources.GPU != (kaiv1alpha1.SystemResource{Deserved: 2, OverQuotaWeight: 1, Limit: 4}) {
			t.Errorf("queue GPU = %+v", got.Spec.Queues[0].Resources)
		}

		// Dropped spec fields moved onto annotations.
		if got.Annotations[constants.InteractiveJobTimeLimitSecondsAnnotation] != "3600" {
			t.Errorf("annotation %s = %q, want 3600", constants.InteractiveJobTimeLimitSecondsAnnotation, got.Annotations[constants.InteractiveJobTimeLimitSecondsAnnotation])
		}
		if got.Annotations[constants.TrainingJobMaxIdleDurationSecondsAnnotation] != "600" {
			t.Errorf("annotation %s = %q, want 600", constants.TrainingJobMaxIdleDurationSecondsAnnotation, got.Annotations[constants.TrainingJobMaxIdleDurationSecondsAnnotation])
		}
		if got.Annotations[constants.ForceDeleteAnnotation] != "true" {
			t.Errorf("annotation %s = %q, want true", constants.ForceDeleteAnnotation, got.Annotations[constants.ForceDeleteAnnotation])
		}
		if got.Annotations[hooks.LastSyncTimestampAnnotation] != "2024-01-02T03-04-05Z" {
			t.Errorf("annotation %s = %q", hooks.LastSyncTimestampAnnotation, got.Annotations[hooks.LastSyncTimestampAnnotation])
		}
		var affinity []string
		if err := json.Unmarshal([]byte(got.Annotations[constants.NodeAffinityInteractiveAnnotation]), &affinity); err != nil {
			t.Fatalf("unmarshal node affinity annotation: %v", err)
		}
		if len(affinity) != 2 || affinity[0] != "a" || affinity[1] != "b" {
			t.Errorf("node affinity annotation = %v, want [a b]", affinity)
		}

		// Queue re-owned by the kai project.
		gotQueue := mustGetQueue(t, c, "q-default")
		assertKaiProjectOwner(t, gotQueue.OwnerReferences, "proj-a")

		// LimitRange re-owned.
		var gotLR corev1.LimitRange
		if err := c.Get(ctx, types.NamespacedName{Namespace: testProjNS, Name: "kai-limit-range"}, &gotLR); err != nil {
			t.Fatalf("get limitrange: %v", err)
		}
		assertKaiProjectOwner(t, gotLR.OwnerReferences, "proj-a")

		// Both RoleBindings re-owned; accessrule-style keeps BlockOwnerDeletion unset.
		var gotRBCtrl, gotRBAccess rbacv1.RoleBinding
		if err := c.Get(ctx, types.NamespacedName{Namespace: testProjNS, Name: "rb-ctrl"}, &gotRBCtrl); err != nil {
			t.Fatalf("get rb-ctrl: %v", err)
		}
		assertKaiProjectOwner(t, gotRBCtrl.OwnerReferences, "proj-a")
		if gotRBCtrl.OwnerReferences[0].BlockOwnerDeletion == nil || !*gotRBCtrl.OwnerReferences[0].BlockOwnerDeletion {
			t.Error("rb-ctrl BlockOwnerDeletion = nil/false, want true (preserved)")
		}
		if err := c.Get(ctx, types.NamespacedName{Namespace: testProjNS, Name: "rb-access"}, &gotRBAccess); err != nil {
			t.Fatalf("get rb-access: %v", err)
		}
		assertKaiProjectOwner(t, gotRBAccess.OwnerReferences, "proj-a")
		if gotRBAccess.OwnerReferences[0].BlockOwnerDeletion != nil {
			t.Error("rb-access BlockOwnerDeletion set, want nil (preserved)")
		}

		// Replicated secret re-pointed via labels, not owner ref.
		var gotSecret corev1.Secret
		if err := c.Get(ctx, types.NamespacedName{Namespace: testProjNS, Name: "replicated-secret"}, &gotSecret); err != nil {
			t.Fatalf("get secret: %v", err)
		}
		if gotSecret.Labels[constants.ReplicatedProjectNameLabel] != "proj-a" {
			t.Errorf("secret name label = %q, want proj-a", gotSecret.Labels[constants.ReplicatedProjectNameLabel])
		}
		if gotSecret.Labels[constants.ReplicatedProjectUIDLabel] == testProjUID {
			t.Error("secret uid label still points at the old project uid")
		}
		if len(gotSecret.OwnerReferences) != 0 {
			t.Errorf("secret owner refs = %v, want none", gotSecret.OwnerReferences)
		}
	})

	t.Run("skips a project that already has a kai twin", func(t *testing.T) {
		old := oldProject("proj-a")
		existing := &kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}}
		c := newMigrationFakeClient(t, old, existing)

		if err := migrateProjects(ctx, c); err != nil {
			t.Fatalf("migrateProjects: %v", err)
		}
		got := mustGetProject(t, c, "proj-a")
		if len(got.Spec.Queues) != 0 || got.Spec.Parent != "" {
			t.Errorf("pre-existing twin was mutated: %+v", got.Spec)
		}
	})

	t.Run("rolls back the whole chain when re-parenting an owned object fails", func(t *testing.T) {
		old := oldProject("proj-a")
		queue := projectOwnedQueue("q-default", old)
		limitRange := projectOwnedLimitRange("kai-limit-range", old)

		// Queues are re-owned before limitranges; fail the limitrange patch so the
		// already-re-owned queue must be rolled back and the new project deleted.
		c := fake.NewClientBuilder().WithScheme(migrationScheme(t)).WithObjects(old, queue, limitRange).
			WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					if _, ok := obj.(*corev1.LimitRange); ok {
						return errors.New("forced limitrange patch failure")
					}
					return cl.Patch(ctx, obj, patch, opts...)
				},
			}).Build()

		if err := migrateProjects(ctx, c); err == nil {
			t.Fatal("migrateProjects succeeded, want error")
		}

		// New kai project deleted.
		var got kaiv1alpha1.Project
		err := c.Get(ctx, types.NamespacedName{Name: "proj-a"}, &got)
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected new project rolled back (deleted), get err = %v", err)
		}

		// Queue owner restored to the legacy run.ai/v2 project.
		gotQueue := mustGetQueue(t, c, "q-default")
		if len(gotQueue.OwnerReferences) != 1 || gotQueue.OwnerReferences[0].APIVersion != runaiv2.GroupVersion.String() {
			t.Errorf("queue owner not restored: %+v", gotQueue.OwnerReferences)
		}
	})
}

// assertKaiProjectOwner asserts owners contains exactly one owner ref pointing at the
// kai project with the given name.
func assertKaiProjectOwner(t *testing.T, owners []metav1.OwnerReference, name string) {
	t.Helper()
	if len(owners) != 1 {
		t.Fatalf("owner refs = %d, want 1: %+v", len(owners), owners)
	}
	if owners[0].APIVersion != kaiv1alpha1.GroupVersion.String() {
		t.Errorf("owner APIVersion = %q, want %q", owners[0].APIVersion, kaiv1alpha1.GroupVersion.String())
	}
	if owners[0].Kind != projectKind {
		t.Errorf("owner Kind = %q, want %q", owners[0].Kind, projectKind)
	}
	if owners[0].Name != name {
		t.Errorf("owner Name = %q, want %q", owners[0].Name, name)
	}
}
