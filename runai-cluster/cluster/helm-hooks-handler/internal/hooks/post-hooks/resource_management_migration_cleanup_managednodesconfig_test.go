/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	testu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func legacyManagedNodesConfig() *runv1alpha1.ManagedNodesConfig {
	return &runv1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{Name: constants.ManagedNodesConfigName},
	}
}

func kaiManagedNodesConfig() *kaiv1alpha1.ManagedNodesConfig {
	return &kaiv1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{Name: constants.ManagedNodesConfigName},
	}
}

func oldManagedNodesConfigCRDExists(t *testing.T, k8sClient client.Client) bool {
	t.Helper()
	err := k8sClient.Get(context.Background(),
		client.ObjectKey{Name: hooks.OldManagedNodesConfigCRDName}, &apiextensionsv1.CustomResourceDefinition{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func legacyManagedNodesConfigExists(t *testing.T, k8sClient client.Client) bool {
	t.Helper()
	err := k8sClient.Get(context.Background(),
		client.ObjectKey{Name: constants.ManagedNodesConfigName}, &runv1alpha1.ManagedNodesConfig{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

// The managed-nodes-config cleanup rides the nodepool cleanup flag, so it must stay
// untouched while that flag is off - and run through finalizeResourceManagementMigration
// once it is on. Both halves of the wiring are asserted here; the specs below then
// exercise the guard rails directly.
func TestManagedNodesConfigCleanupIsGatedByTheNodePoolsCleanupFlag(t *testing.T) {
	objects := func() []client.Object {
		return []client.Object{
			departmentCRD(hooks.OldManagedNodesConfigCRDName),
			departmentCRD(newManagedNodesConfigCRDName),
			legacyManagedNodesConfig(),
			kaiManagedNodesConfig(),
		}
	}

	disabled := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(objects()...).Build()
	require.NoError(t, cleanupHook(false).finalizeResourceManagementMigration(context.Background(),
		&clients.Clients{LocalClusterClient: disabled}))
	require.True(t, legacyManagedNodesConfigExists(t, disabled))
	require.True(t, oldManagedNodesConfigCRDExists(t, disabled))

	enabled := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(objects()...).Build()
	hook := NewPostUpgradeHook(&hooks.ClusterHookParams{
		ResourceManagementMigrationNodePoolsEnabled: true,
		ResourceManagementMigrationNodePoolsCleanup: true,
	})
	// The nodepool flow re-enables the operands it disabled, which goes through the
	// dynamic client rather than the typed one.
	require.NoError(t, hook.finalizeResourceManagementMigration(context.Background(), &clients.Clients{
		LocalClusterClient:        enabled,
		LocalClusterDynamicClient: testu.NewTestDynamicClient(true, false),
	}))
	require.False(t, legacyManagedNodesConfigExists(t, enabled))
	require.False(t, oldManagedNodesConfigCRDExists(t, enabled))
}

func TestManagedNodesConfigCleanupNoOpWhenLegacyCRDAbsent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).Build()

	require.NoError(t, hook.cleanupLegacyManagedNodesConfigs(context.Background(), k8sClient))
}

// Deleting the legacy objects before the migration target exists would lose the
// configuration outright, so the cleanup refuses rather than proceeding.
func TestManagedNodesConfigCleanupRefusesWhenNewCRDAbsent(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldManagedNodesConfigCRDName),
		legacyManagedNodesConfig(),
	).Build()

	err := hook.cleanupLegacyManagedNodesConfigs(context.Background(), k8sClient)

	require.ErrorContains(t, err, newManagedNodesConfigCRDName)
	require.True(t, legacyManagedNodesConfigExists(t, k8sClient))
	require.True(t, oldManagedNodesConfigCRDExists(t, k8sClient))
}

// A legacy singleton with no kai twin means the pre-upgrade copy never ran (or
// failed), so deleting it would lose the configuration.
func TestManagedNodesConfigCleanupRefusesWhenTwinMissing(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldManagedNodesConfigCRDName),
		departmentCRD(newManagedNodesConfigCRDName),
		legacyManagedNodesConfig(),
	).Build()

	err := hook.cleanupLegacyManagedNodesConfigs(context.Background(), k8sClient)

	require.ErrorContains(t, err, "migration is not complete")
	require.True(t, legacyManagedNodesConfigExists(t, k8sClient))
	require.True(t, oldManagedNodesConfigCRDExists(t, k8sClient))
}

func TestManagedNodesConfigCleanupDeletesLegacyObjectAndCRD(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldManagedNodesConfigCRDName),
		departmentCRD(newManagedNodesConfigCRDName),
		legacyManagedNodesConfig(),
		kaiManagedNodesConfig(),
	).Build()

	require.NoError(t, hook.cleanupLegacyManagedNodesConfigs(context.Background(), k8sClient))

	require.False(t, legacyManagedNodesConfigExists(t, k8sClient))
	require.False(t, oldManagedNodesConfigCRDExists(t, k8sClient))
}

// A cluster that never had a singleton written still carries the legacy CRD, which
// should be retired along with the rest of the legacy group.
func TestManagedNodesConfigCleanupDeletesCRDWhenNoLegacyObject(t *testing.T) {
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiNodePoolScheme(t)).WithObjects(
		departmentCRD(hooks.OldManagedNodesConfigCRDName),
		departmentCRD(newManagedNodesConfigCRDName),
	).Build()

	require.NoError(t, hook.cleanupLegacyManagedNodesConfigs(context.Background(), k8sClient))

	require.False(t, oldManagedNodesConfigCRDExists(t, k8sClient))
}
