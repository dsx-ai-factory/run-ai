/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helmclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

func InstallOrUpgradeRelease(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string) error {
	return installOrUpgrade(ctx, repoUrl, version, releaseName, chartName, targetNamespace, isOci, values, nil, ChartAuth{})
}

func InstallOrUpgradeReleaseWithAuth(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string, auth ChartAuth) error {
	return installOrUpgrade(ctx, repoUrl, version, releaseName, chartName, targetNamespace, isOci, values, nil, auth)
}

func InstallOrUpgradeReleaseWithKubeVersion(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string, kubeVersion *chartutil.KubeVersion) error {
	return installOrUpgrade(ctx, repoUrl, version, releaseName, chartName, targetNamespace, isOci, values, kubeVersion, ChartAuth{})
}

func installOrUpgrade(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string, kubeVersion *chartutil.KubeVersion, auth ChartAuth) error {
	logging.Infof("[Helm] Starting install/upgrade for release '%s' (chart: %s, version: %s) in namespace '%s'", releaseName, chartName, version, targetNamespace)
	logging.Infof("[Helm] Chart repository: %s (OCI: %v, authenticated: %v)", repoUrl, isOci, !auth.IsZero())
	if len(values) > 0 {
		logging.Infof("[Helm] Custom values: %v", values)
	}

	chartData := helmChartData{
		repoUrl:        repoUrl,
		chartVersion:   version,
		valuesOverride: values,
		auth:           auth,
	}
	client, err := newClient(&chartData, releaseName, chartName, targetNamespace, kubeVersion, isOci)
	if err != nil {
		logging.Errorf("[Helm] Failed to create Helm client for release '%s': %v", releaseName, err)
		return err
	}

	logging.Infof("[Helm] Merging values for chart '%s'...", chartName)
	mergedValues, err := client.handleValues()
	if err != nil {
		logging.Errorf("[Helm] Failed to merge values for chart '%s': %v", chartName, err)
		return fmt.Errorf("failed to merge values for chart: %v", err)
	}

	logging.Infof("[Helm] Fetching chart '%s' from repository...", chartName)
	chartRequested, err := client.getChart()
	if err != nil {
		logging.Errorf("[Helm] Failed to get chart '%s': %v", chartName, err)
		return fmt.Errorf("failed to get chart: %v", err)
	}

	logging.Infof("[Helm] Handling dependencies for chart '%s'...", chartName)
	err = client.handleDependencies(chartRequested)
	if err != nil {
		logging.Errorf("[Helm] Failed to handle dependencies for chart '%s': %v", chartName, err)
		return fmt.Errorf("failed to handle chart dependencies: %v", err)
	}

	logging.Infof("[Helm] Installing/upgrading chart '%s'...", chartName)
	err = retryOnTransientAPIError(ctx, helmRetryAttempts, helmRetryBaseBackoff,
		fmt.Sprintf("install/upgrade %s", releaseName),
		func() error {
			return client.runChart(ctx, chartName, chartRequested, mergedValues)
		})
	if err != nil {
		logging.Errorf("[Helm] Failed to install/upgrade chart '%s': %v", chartName, err)
		return fmt.Errorf("failed running charts: %v", err)
	}

	logging.Infof("[Helm] Successfully installed/upgraded release '%s' in namespace '%s'", releaseName, targetNamespace)
	return nil
}

func createActionConfig(targetNamespace string, kubeVersion *chartutil.KubeVersion) (*action.Configuration, *cli.EnvSettings, error) {
	actionConfig := &action.Configuration{}
	settings := cli.New()
	settings.Debug = true
	settings.SetNamespace(targetNamespace)
	settings.Debug = false
	if kubeVersion != nil {
		actionConfig.Capabilities = &chartutil.Capabilities{
			APIVersions: []string{},
			KubeVersion: *kubeVersion,
		}
	}
	os.Setenv("HELM_EXPERIMENTAL_OCI", "1")
	err := actionConfig.Init(settings.RESTClientGetter(),
		settings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		func(msg string, args ...interface{}) {
			// Route Helm internal logs through our logger.
			if len(args) > 0 {
				logging.Infof(msg, args...)
			} else {
				logging.Infof("%s", msg)
			}
		})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init helm client: %v", err)
	}
	return actionConfig, settings, err
}

func InstallOrUpgradeFromLocalChart(ctx context.Context, chartPath, releaseName, targetNamespace string, setValues []string, valueFiles []string) error {
	logging.Infof("[Helm] Starting local chart install/upgrade for release '%s' from %s in namespace '%s'", releaseName, chartPath, targetNamespace)

	chartData := helmChartData{
		valuesOverride: setValues,
		valueFiles:     valueFiles,
	}
	client, err := newClient(&chartData, releaseName, chartPath, targetNamespace, nil, false)
	if err != nil {
		logging.Errorf("[Helm] Failed to create Helm client for release '%s': %v", releaseName, err)
		return err
	}

	logging.Infof("[Helm] Merging values for local chart '%s'...", chartPath)
	mergedValues, err := client.handleValues()
	if err != nil {
		logging.Errorf("[Helm] Failed to merge values for local chart '%s': %v", chartPath, err)
		return fmt.Errorf("failed to merge values for chart: %v", err)
	}

	logging.Infof("[Helm] Loading local chart '%s'...", chartPath)
	chartRequested, err := client.getChart()
	if err != nil {
		logging.Errorf("[Helm] Failed to load local chart '%s': %v", chartPath, err)
		return fmt.Errorf("failed to get chart: %v", err)
	}

	logging.Infof("[Helm] Handling dependencies for local chart '%s'...", chartPath)
	err = client.handleDependencies(chartRequested)
	if err != nil {
		logging.Errorf("[Helm] Failed to handle dependencies for local chart '%s': %v", chartPath, err)
		return fmt.Errorf("failed to handle chart dependencies: %v", err)
	}

	logging.Infof("[Helm] Installing/upgrading local chart '%s'...", chartPath)
	err = retryOnTransientAPIError(ctx, helmRetryAttempts, helmRetryBaseBackoff,
		fmt.Sprintf("install/upgrade %s (local)", releaseName),
		func() error {
			return client.runChart(ctx, chartPath, chartRequested, mergedValues)
		})
	if err != nil {
		logging.Errorf("[Helm] Failed to install/upgrade local chart '%s': %v", chartPath, err)
		return fmt.Errorf("failed running charts: %v", err)
	}

	logging.Infof("[Helm] Successfully installed/upgraded release '%s' from local chart in namespace '%s'", releaseName, targetNamespace)
	return nil
}

func isHelmWaitTimeout(err error) bool {
	message := err.Error()
	return strings.Contains(message, "context deadline exceeded") ||
		strings.Contains(message, "timed out waiting for the condition")
}

// NormalizeUninstallError converts a completed Helm uninstall's sole wait
// timeout into context.DeadlineExceeded.
func NormalizeUninstallError(response *release.UninstallReleaseResponse, err error) error {
	if err == nil || response == nil || response.Release == nil || response.Release.Info == nil {
		return err
	}
	if response.Release.Info.Status != release.StatusUninstalled ||
		!strings.Contains(err.Error(), "uninstallation completed with 1 error(s):") ||
		!isHelmWaitTimeout(err) {
		return err
	}
	return fmt.Errorf("%v: %w", err, context.DeadlineExceeded)
}

// FixPendingRelease transitions a Helm release stuck in a pending state
// (pending-install, pending-upgrade, or pending-rollback) to failed so that a
// subsequent UninstallRelease call can proceed. A failed first install leaves
// the release in pending-install state; Helm's pending protection would then
// prevent a clean uninstall without this fix. It is a no-op when the release
// does not exist or is not in a pending state.
func FixPendingRelease(releaseName, namespace string) error {
	actionConfig, _, err := createActionConfig(namespace, nil)
	if err != nil {
		return fmt.Errorf("failed to init helm client for pending release fix: %v", err)
	}
	rel, err := actionConfig.Releases.Last(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil
		}
		return fmt.Errorf("failed to read release '%s' for pending fix: %w", releaseName, err)
	}
	if !rel.Info.Status.IsPending() {
		return nil
	}
	logging.Infof("[Helm] Release '%s' is in %s state, forcing to failed to allow uninstall", releaseName, rel.Info.Status)
	rel.SetStatus(release.StatusFailed, "installer: clearing pending state to allow uninstall")
	if err := actionConfig.Releases.Update(rel); err != nil {
		return fmt.Errorf("failed to update pending release '%s': %w", releaseName, err)
	}
	return nil
}

// UninstallRelease removes a Helm release, waiting up to waitTimeout for all
// chart resources to be deleted. Pass 0 for fire-and-forget (no wait).
// If the wait times out, the returned error wraps context.DeadlineExceeded so
// callers can detect it with errors.Is.
func UninstallRelease(releaseName, namespace string, waitTimeout time.Duration) error {
	logging.Infof("[Helm] Uninstalling release '%s' from namespace '%s'", releaseName, namespace)
	actionConfig, _, err := createActionConfig(namespace, nil)
	if err != nil {
		return fmt.Errorf("failed to init helm client for uninstall: %v", err)
	}
	uninstall := action.NewUninstall(actionConfig)
	if waitTimeout > 0 {
		uninstall.Wait = true
		uninstall.Timeout = waitTimeout
	}
	response, err := uninstall.Run(releaseName)
	if err = NormalizeUninstallError(response, err); err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			logging.Infof("[Helm] Release '%s' not found in namespace '%s', skipping uninstall", releaseName, namespace)
			return nil
		}
		if isHelmWaitTimeout(err) {
			return fmt.Errorf("helm uninstall of '%s' timed out waiting for resources: %w", releaseName, context.DeadlineExceeded)
		}
		return fmt.Errorf("helm uninstall of '%s' failed: %w", releaseName, err)
	}
	logging.Infof("[Helm] Successfully uninstalled release '%s' from namespace '%s'", releaseName, namespace)
	return nil
}

// GetReleaseStatus reports whether a Helm release is installed in the namespace
// and, if so, its chart version. Only deployed/superseded releases count as
// installed. A missing release, or a reinstallable one (failed/uninstalled), is
// not an error: it returns exists=false with a nil error. A release stuck in a
// pending/uninstalling state from an interrupted operation is returned as an
// error so the caller can surface it instead of silently re-driving it.
func GetReleaseStatus(releaseName, namespace string) (exists bool, version string, err error) {
	actionConfig, _, err := createActionConfig(namespace, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to init helm client for release status: %v", err)
	}
	rel, err := actionConfig.Releases.Last(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("failed to read release '%s': %w", releaseName, err)
	}
	if rel.Chart != nil && rel.Chart.Metadata != nil {
		version = rel.Chart.Metadata.Version
	}
	if rel.Info == nil {
		return false, version, nil
	}
	switch rel.Info.Status {
	case release.StatusDeployed, release.StatusSuperseded:
		return true, version, nil
	case release.StatusPendingInstall, release.StatusPendingUpgrade, release.StatusPendingRollback, release.StatusUninstalling:
		return false, version, fmt.Errorf("helm release '%s' is in a broken state (%s) from an interrupted operation; resolve it (helm rollback or uninstall) before retrying", releaseName, rel.Info.Status)
	default:
		// failed and uninstalled are reinstallable: report not-installed so the
		// step proceeds with install/upgrade.
		return false, version, nil
	}
}

func GetChartVersion(devel bool, version string) string {
	if devel {
		return ">0.0.0-0" // latest development version
	}
	return version
}
