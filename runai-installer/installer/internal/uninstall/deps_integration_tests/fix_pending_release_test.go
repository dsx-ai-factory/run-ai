//go:build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package uninstall_integration_test

// Integration tests for HelmWrapper.FixPendingRelease against a real cluster.
// Requires a live cluster pointed to by KUBECONFIG (a local kind cluster is fine).
//
// Run directly:
//
//	KUBECONFIG=<path> go test -v -tags integration -timeout 2m \
//	    ./internal/installations/uninstall/deps_integration_tests/
//
// Or via the Makefile target (creates and tears down a kind cluster automatically):
//
//	make test-uninstall-integration

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestHelmWrapper_FixPendingRelease_Integration(t *testing.T) {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
	require.NoError(t, err, "load kubeconfig")

	k8sClient, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	namespace := "test-fix-pending-helm"
	createNamespace(t, k8sClient, namespace)

	store := storage.Init(driver.NewSecrets(k8sClient.CoreV1().Secrets(namespace)))

	deps, err := uninstall.ProductionDeps()
	require.NoError(t, err)

	pendingCases := []struct {
		name          string
		releaseName   string
		initialStatus release.Status
	}{
		{
			name:          "pending-install transitions to failed",
			releaseName:   "release-pending-install",
			initialStatus: release.StatusPendingInstall,
		},
		{
			name:          "pending-upgrade transitions to failed",
			releaseName:   "release-pending-upgrade",
			initialStatus: release.StatusPendingUpgrade,
		},
		{
			name:          "pending-rollback transitions to failed",
			releaseName:   "release-pending-rollback",
			initialStatus: release.StatusPendingRollback,
		},
	}

	for _, tc := range pendingCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rel := newTestHelmRelease(tc.releaseName, tc.initialStatus)
			require.NoError(t, store.Create(rel), "create pending release in cluster")

			got, err := store.Last(tc.releaseName)
			require.NoError(t, err)
			require.Equal(t, tc.initialStatus, got.Info.Status, "pre-condition: release should be in pending state")

			require.NoError(t, deps.HelmWrapper.FixPendingRelease(tc.releaseName, namespace))

			got, err = store.Last(tc.releaseName)
			require.NoError(t, err)
			require.Equal(t, release.StatusFailed, got.Info.Status,
				"FixPendingRelease should transition %s to failed", tc.initialStatus)
		})
	}

	t.Run("deployed release is not modified", func(t *testing.T) {
		releaseName := "release-deployed"
		rel := newTestHelmRelease(releaseName, release.StatusDeployed)
		require.NoError(t, store.Create(rel), "create deployed release in cluster")

		require.NoError(t, deps.HelmWrapper.FixPendingRelease(releaseName, namespace))

		got, err := store.Last(releaseName)
		require.NoError(t, err)
		require.Equal(t, release.StatusDeployed, got.Info.Status,
			"FixPendingRelease should not modify a deployed release")
	})

	t.Run("non-existent release is a no-op", func(t *testing.T) {
		require.NoError(t, deps.HelmWrapper.FixPendingRelease("does-not-exist", namespace))
	})
}

// TestHelmWrapper_UninstallFailedRelease_Integration verifies that a release in
// failed state (which is what FixPendingRelease transitions a pending release to)
// can be uninstalled without Helm returning an error.  This closes the loop: if
// Helm blocked uninstall of failed releases, the fix would be useless.
func TestHelmWrapper_UninstallFailedRelease_Integration(t *testing.T) {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
	require.NoError(t, err, "load kubeconfig")

	k8sClient, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	namespace := "test-uninstall-failed-helm"
	createNamespace(t, k8sClient, namespace)

	store := storage.Init(driver.NewSecrets(k8sClient.CoreV1().Secrets(namespace)))

	deps, err := uninstall.ProductionDeps()
	require.NoError(t, err)

	releaseName := "release-failed"
	rel := newTestHelmRelease(releaseName, release.StatusFailed)
	require.NoError(t, store.Create(rel), "create failed release in cluster")

	require.NoError(t, deps.HelmWrapper.UninstallRelease(releaseName, namespace, 0),
		"UninstallRelease should succeed for a release in failed state")

	// After a successful uninstall with KeepHistory=false (the default), Helm
	// purges all release records from the store.
	_, err = store.Last(releaseName)
	require.Error(t, err, "release records should be purged after successful uninstall")
}

// TestHelmWrapper_PendingToFixToUninstall_Integration covers the end-to-end
// scenario from the ticket: a previous install was interrupted and left the
// release in pending-install state; FixPendingRelease transitions it to failed
// so that UninstallRelease can proceed without being blocked.
func TestHelmWrapper_PendingToFixToUninstall_Integration(t *testing.T) {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
	require.NoError(t, err, "load kubeconfig")

	k8sClient, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	namespace := "test-pending-fix-uninstall"
	createNamespace(t, k8sClient, namespace)

	store := storage.Init(driver.NewSecrets(k8sClient.CoreV1().Secrets(namespace)))

	deps, err := uninstall.ProductionDeps()
	require.NoError(t, err)

	releaseName := "release-was-pending"
	rel := newTestHelmRelease(releaseName, release.StatusPendingInstall)
	require.NoError(t, store.Create(rel), "seed cluster with a pending-install release")

	require.NoError(t, deps.HelmWrapper.FixPendingRelease(releaseName, namespace))

	got, err := store.Last(releaseName)
	require.NoError(t, err)
	require.Equal(t, release.StatusFailed, got.Info.Status, "release should be failed after fix")

	require.NoError(t, deps.HelmWrapper.UninstallRelease(releaseName, namespace, 0),
		"UninstallRelease should succeed after FixPendingRelease clears the pending state")

	_, err = store.Last(releaseName)
	require.Error(t, err, "release records should be purged after successful uninstall")
}

// TestHelmStorage_NonExistentReleaseError_Integration verifies the exact error
// returned by storage.Last for a non-existent release. This test exists to
// validate the assumption in FixPendingRelease that errors.Is(err,
// driver.ErrReleaseNotFound) correctly identifies a missing release so that
// real storage failures are not silently swallowed.
func TestHelmStorage_NonExistentReleaseError_Integration(t *testing.T) {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		nil,
	).ClientConfig()
	require.NoError(t, err, "load kubeconfig")

	k8sClient, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	namespace := "test-helm-error-type"
	createNamespace(t, k8sClient, namespace)

	store := storage.Init(driver.NewSecrets(k8sClient.CoreV1().Secrets(namespace)))

	_, err = store.Last("non-existent-release")
	require.Error(t, err, "Last should return an error for a non-existent release")
	t.Logf("Actual error returned by store.Last: %v (type: %T)", err, err)
	require.True(t, errors.Is(err, driver.ErrReleaseNotFound),
		"expected driver.ErrReleaseNotFound, got: %v (type: %T)", err, err)
}

// newTestHelmRelease returns a minimal Helm release object suitable for
// inserting directly into the Helm storage driver. Only fields required
// by the driver (Name, Version, Chart metadata, Info) are populated.
func newTestHelmRelease(name string, status release.Status) *release.Release {
	return &release.Release{
		Name:    name,
		Version: 1,
		Info:    &release.Info{Status: status},
		Chart:   &chart.Chart{Metadata: &chart.Metadata{Name: "test-chart"}},
	}
}
