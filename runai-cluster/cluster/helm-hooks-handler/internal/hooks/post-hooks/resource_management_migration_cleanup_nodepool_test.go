/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"testing"

	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func legacyNodePool(name string) *runv1alpha1.NodePool {
	return &runv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

// kaiNodePoolScheme registers everything the nodepool cleanup path touches plus the
// new kai nodepool type (absent from the shared cluster scheme) for seeding
// assertions.
func kaiNodePoolScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	for _, add := range []func(*runtime.Scheme) error{
		apiextensionsv1.AddToScheme,
		runv1alpha1.AddToScheme,
		kaiv1alpha1.AddToScheme,
	} {
		if err := add(s); err != nil {
			t.Fatalf("add to scheme: %v", err)
		}
	}
	return s
}

func oldNodePoolCRDExists(t *testing.T, k8sClient client.Client) bool {
	t.Helper()
	err := k8sClient.Get(context.Background(), client.ObjectKey{Name: hooks.OldNodePoolCRDName}, &apiextensionsv1.CustomResourceDefinition{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func TestNodePoolCleanupIsDisabledByDefault(t *testing.T) {
	hook := cleanupHook(false)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldNodePoolCRDName),
	).Build()

	// The gate lives in finalizeResourceManagementMigration: with the flags off it
	// returns before touching the cluster, so cleanup never runs.
	require.NoError(t, hook.finalizeResourceManagementMigration(context.Background(),
		&clients.Clients{LocalClusterClient: k8sClient}))
	// The legacy CRD is left untouched when cleanup is disabled.
	require.True(t, oldNodePoolCRDExists(t, k8sClient))
}

func TestNodePoolCleanupNoOpWhenLegacyCRDAbsent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()

	require.NoError(t, hook.cleanupLegacyNodePools(context.Background(), k8sClient))
}

func TestNodePoolCleanupErrorsWhenLegacyNodePoolHasNoKaiTwin(t *testing.T) {
	hook := cleanupHook(true)
	// np-a has a migrated kai twin, np-b does not - the cleanup must refuse.
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldNodePoolCRDName),
		departmentCRD(newNodePoolCRDName),
		legacyNodePool("np-a"),
		legacyNodePool("np-b"),
		&kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-a"}},
	).Build()

	err := hook.cleanupLegacyNodePools(context.Background(), k8sClient)
	require.ErrorContains(t, err, `legacy nodepool "np-b" has no migrated kai nodepool`)
	// Nothing is deleted while the migration is incomplete.
	require.True(t, oldNodePoolCRDExists(t, k8sClient))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-a"}, &runv1alpha1.NodePool{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-b"}, &runv1alpha1.NodePool{}))
}

func TestNodePoolCleanupDeletesLegacyNodePoolsWhenAllMigrated(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldNodePoolCRDName),
		departmentCRD(newNodePoolCRDName),
		legacyNodePool("np-a"),
		legacyNodePool("np-b"),
		&kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-a"}},
		&kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-b"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyNodePools(context.Background(), k8sClient))

	// Every legacy nodepool is deleted...
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-a"}, &runv1alpha1.NodePool{})))
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-b"}, &runv1alpha1.NodePool{})))
	// ...and so is the legacy CRD.
	require.False(t, oldNodePoolCRDExists(t, k8sClient))
	// The migrated kai nodepools are left untouched.
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-a"}, &kaiv1alpha1.NodePool{}))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-b"}, &kaiv1alpha1.NodePool{}))
}

func TestNodePoolCleanupErrorsWhenNewCRDMissing(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldNodePoolCRDName),
	).Build()

	err := hook.cleanupLegacyNodePools(context.Background(), k8sClient)
	require.ErrorContains(t, err, "new nodepool CRD")
	require.True(t, oldNodePoolCRDExists(t, k8sClient))
}

func TestNodePoolCleanupDeletesLegacyCRDAndIsIdempotent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		departmentCRD(hooks.OldNodePoolCRDName),
		departmentCRD(newNodePoolCRDName),
	).Build()

	require.NoError(t, hook.cleanupLegacyNodePools(context.Background(), k8sClient))
	require.False(t, oldNodePoolCRDExists(t, k8sClient))

	// Running again is a no-op now that the legacy CRD is gone.
	require.NoError(t, hook.cleanupLegacyNodePools(context.Background(), k8sClient))
}

func TestNodePoolCleanupPreservesNewNodePoolInstances(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldNodePoolCRDName),
		departmentCRD(newNodePoolCRDName),
		&kaiv1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-a"}},
	).Build()

	require.NoError(t, hook.cleanupLegacyNodePools(context.Background(), k8sClient))
	require.False(t, oldNodePoolCRDExists(t, k8sClient))

	// The migrated kai nodepool instance is left untouched.
	var got kaiv1alpha1.NodePool
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "np-a"}, &got))
}
