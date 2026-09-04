/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"

	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func departmentCRD(name string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func legacyDepartment(name string) *schedv2alpha2.Department {
	return &schedv2alpha2.Department{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func legacyProject(name string) *runaiv2.Project {
	return &runaiv2.Project{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

// kaiProjectScheme registers everything the project cleanup path touches plus the new
// kai project type (absent from the shared cluster scheme) for seeding assertions.
func kaiProjectScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		apiextensionsv1.AddToScheme,
		runaiv2.AddToScheme,
		kaiv1alpha1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatalf("add to scheme: %v", err)
		}
	}
	return s
}

func oldProjectCRDExists(t *testing.T, k8sClient client.Client) bool {
	t.Helper()
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: hooks.OldProjectCRDName}, &apiextensionsv1.CustomResourceDefinition{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

// kaiDepartmentScheme registers everything the cleanup path touches plus the new
// kai department type, so a migrated kai department instance can be seeded and
// asserted on. The shared cluster scheme does not register the run:ai kai
// department type, so a dedicated scheme is built here.
func kaiDepartmentScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		apiextensionsv1.AddToScheme,
		schedv2alpha2.AddToScheme,
		kaiv1alpha1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatalf("add to scheme: %v", err)
		}
	}
	return s
}

// cleanupHook builds a post-upgrade hook with the cleanup gate set. The cleanup
// requires both the migration and the cleanup flags, so both are toggled together.
func cleanupHook(enabled bool) *PostUpgradeHook {
	return NewPostUpgradeHook(&hooks.ClusterHookParams{
		ResourceManagementMigrationEnabled: enabled,
		ResourceManagementMigrationCleanup: enabled,
	})
}

func oldDepartmentCRDExists(t *testing.T, k8sClient client.Client) bool {
	t.Helper()
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: hooks.OldDepartmentCRDName}, &apiextensionsv1.CustomResourceDefinition{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func TestDepartmentCleanupIsDisabledByDefault(t *testing.T) {
	hook := cleanupHook(false)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
	).Build()

	// The gate lives in finalizeResourceManagementMigration: with the flags off it
	// returns before touching the cluster, so cleanup never runs.
	require.NoError(t, hook.finalizeResourceManagementMigration(context.Background(),
		&clients.Clients{LocalClusterClient: k8sClient}))
	// The legacy CRD is left untouched when cleanup is disabled.
	require.True(t, oldDepartmentCRDExists(t, k8sClient))
}

func TestDepartmentCleanupNoOpWhenLegacyCRDAbsent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))
}

func TestDepartmentCleanupErrorsWhenLegacyDepartmentHasNoKaiTwin(t *testing.T) {
	hook := cleanupHook(true)
	// dep-a has a migrated kai twin, dep-b does not - the cleanup must refuse.
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartment("dep-a"),
		legacyDepartment("dep-b"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-a"}},
	).Build()

	err := hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background()))
	require.ErrorContains(t, err, `legacy department "dep-b" has no migrated kai department`)
	// Nothing is deleted while the migration is incomplete.
	require.True(t, oldDepartmentCRDExists(t, k8sClient))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-a"}, &schedv2alpha2.Department{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-b"}, &schedv2alpha2.Department{}))
}

func TestDepartmentCleanupDeletesLegacyDepartmentsWhenAllMigrated(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartment("dep-a"),
		legacyDepartment("dep-b"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-a"}},
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-b"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))

	// Every legacy department is deleted...
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-a"}, &schedv2alpha2.Department{})))
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-b"}, &schedv2alpha2.Department{})))
	// ...and so is the legacy CRD.
	require.False(t, oldDepartmentCRDExists(t, k8sClient))
	// The migrated kai departments are left untouched.
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-a"}, &kaiv1alpha1.Department{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-b"}, &kaiv1alpha1.Department{}))
}

func TestDepartmentCleanupErrorsWhenNewCRDMissing(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
	).Build()

	err := hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background()))
	require.ErrorContains(t, err, "new department CRD")
	require.True(t, oldDepartmentCRDExists(t, k8sClient))
}

func TestDepartmentCleanupDeletesLegacyCRDAndIsIdempotent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
	).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))
	require.False(t, oldDepartmentCRDExists(t, k8sClient))

	// Running again is a no-op now that the legacy CRD is gone.
	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))
}

func TestDepartmentCleanupPreservesNewDepartmentInstances(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-a"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))

	// The legacy CRD is deleted...
	require.False(t, oldDepartmentCRDExists(t, k8sClient))

	// ...but the migrated kai department instance is left untouched.
	var got kaiv1alpha1.Department
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-a"}, &got))
}

func TestProjectCleanupIsDisabledByDefault(t *testing.T) {
	hook := cleanupHook(false)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
	).Build()

	// The gate lives in finalizeResourceManagementMigration: with the flags off it
	// returns before touching the cluster, so cleanup never runs.
	require.NoError(t, hook.finalizeResourceManagementMigration(context.Background(),
		&clients.Clients{LocalClusterClient: k8sClient}))
	// The legacy CRD is left untouched when cleanup is disabled.
	require.True(t, oldProjectCRDExists(t, k8sClient))
}

func TestProjectCleanupNoOpWhenLegacyCRDAbsent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))
}

func TestProjectCleanupErrorsWhenLegacyProjectHasNoKaiTwin(t *testing.T) {
	hook := cleanupHook(true)
	// proj-a has a migrated kai twin, proj-b does not - the cleanup must refuse.
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		legacyProject("proj-a"),
		legacyProject("proj-b"),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
	).Build()

	err := hook.cleanupLegacyProjects(context.Background(), k8sClient)
	require.ErrorContains(t, err, `legacy project "proj-b" has no migrated kai project`)
	// Nothing is deleted while the migration is incomplete.
	require.True(t, oldProjectCRDExists(t, k8sClient))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &runaiv2.Project{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-b"}, &runaiv2.Project{}))
}

func TestProjectCleanupDeletesLegacyProjectsWhenAllMigrated(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		legacyProject("proj-a"),
		legacyProject("proj-b"),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-b"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))

	// Every legacy project is deleted...
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &runaiv2.Project{})))
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-b"}, &runaiv2.Project{})))
	// ...and so is the legacy CRD.
	require.False(t, oldProjectCRDExists(t, k8sClient))
	// The migrated kai projects are left untouched.
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &kaiv1alpha1.Project{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-b"}, &kaiv1alpha1.Project{}))
}

func TestProjectCleanupErrorsWhenNewCRDMissing(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
	).Build()

	err := hook.cleanupLegacyProjects(context.Background(), k8sClient)
	require.ErrorContains(t, err, "new project CRD")
	require.True(t, oldProjectCRDExists(t, k8sClient))
}

func TestProjectCleanupDeletesLegacyCRDAndIsIdempotent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
	).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))
	require.False(t, oldProjectCRDExists(t, k8sClient))

	// Running again is a no-op now that the legacy CRD is gone.
	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))
}

func TestProjectCleanupPreservesNewProjectInstances(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))
	require.False(t, oldProjectCRDExists(t, k8sClient))

	// The migrated kai project instance is left untouched.
	var got kaiv1alpha1.Project
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &got))
}

// legacyDepartmentWithFinalizer / legacyProjectWithFinalizer build a legacy org
// unit carrying the finalizer its (now kai-only) controller used to remove. The
// fake client honours finalizers exactly as the API server does, so a delete
// leaves the object Terminating rather than removing it - which is how the
// cleanup used to wedge the legacy CRD.
func legacyDepartmentWithFinalizer(name string) *schedv2alpha2.Department {
	dep := legacyDepartment(name)
	dep.Finalizers = []string{"run.ai.finalizers.department"}
	return dep
}

func legacyProjectWithFinalizer(name string) *runaiv2.Project {
	proj := legacyProject(name)
	proj.Finalizers = []string{"run.ai.finalizers.project"}
	return proj
}

func TestDepartmentCleanupClearsFinalizersSoDeletesComplete(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartmentWithFinalizer("dep-a"),
		legacyDepartmentWithFinalizer("dep-b"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-a"}},
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "dep-b"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))

	// The legacy departments are gone rather than stuck Terminating on their
	// finalizers, so the legacy CRD can finish terminating too.
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-a"}, &schedv2alpha2.Department{})))
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-b"}, &schedv2alpha2.Department{})))
	require.False(t, oldDepartmentCRDExists(t, k8sClient))
	// The migrated kai departments are left untouched.
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-a"}, &kaiv1alpha1.Department{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "dep-b"}, &kaiv1alpha1.Department{}))
}

func TestProjectCleanupClearsFinalizersSoDeletesComplete(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		legacyProjectWithFinalizer("proj-a"),
		legacyProjectWithFinalizer("proj-b"),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-b"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))

	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &runaiv2.Project{})))
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-b"}, &runaiv2.Project{})))
	require.False(t, oldProjectCRDExists(t, k8sClient))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &kaiv1alpha1.Project{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-b"}, &kaiv1alpha1.Project{}))
}

// TestProjectCleanupClearsFinalizersOnAlreadyTerminatingProjects covers a cluster
// wedged by an earlier upgrade: the legacy projects were already deleted but sit
// Terminating because nothing removed their finalizers. Re-running the cleanup
// must clear them and let both the objects and the CRD finish going away.
func TestProjectCleanupClearsFinalizersOnAlreadyTerminatingProjects(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		legacyProjectWithFinalizer("proj-a"),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
	).Build()

	// Delete first, as the previous (broken) upgrade did: the finalizer keeps the
	// object alive with a deletionTimestamp set.
	require.NoError(t, k8sClient.Delete(context.Background(), legacyProject("proj-a")))
	var terminating runaiv2.Project
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &terminating))
	require.NotNil(t, terminating.DeletionTimestamp)

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))

	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &runaiv2.Project{})))
	require.False(t, oldProjectCRDExists(t, k8sClient))
}
