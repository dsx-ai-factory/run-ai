/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	"helm.sh/helm/v3/pkg/chartutil"
)

// ChartRepoBaseEnv names the environment variable holding the base URL of the
// classic Helm repositories the Run:ai charts are served from when the
// registry is not NGC. Each environment appends its own repository name to it,
// so the base is the common prefix of the control plane and cluster chart
// repositories (e.g. "https://charts.example.com/api/helm"). It is deliberately
// not compiled in: the installer has no default chart host outside NGC.
const ChartRepoBaseEnv = "RUNAI_CHART_REPO_BASE"

// ChartRepoBase returns the configured chart repository base without a
// trailing slash, or "" when none is set.
func ChartRepoBase() string {
	return strings.TrimSuffix(strings.TrimSpace(os.Getenv(ChartRepoBaseEnv)), "/")
}

// MissingChartRepoError explains how to configure the chart repository when a
// chart has no repository to install from.
func MissingChartRepoError(chart string) error {
	return fmt.Errorf("no chart repository configured for the %s chart: set the %s environment variable", chart, ChartRepoBaseEnv)
}

const NGCNonProdChartRepoEnv = "RUNAI_NGC_NONPROD_CHART_REPO"

const ngcChartRepoProd = "https://helm.ngc.nvidia.com/nvidia/runai"

const (
	CPReleaseName      = "runai-backend"
	CPChartName        = "control-plane"
	CPNamespace        = "runai-backend"
	ClusterReleaseName = "runai-cluster"
	ClusterChartName   = "runai-cluster"
	ClusterNamespace   = "runai"

	ReservationNamespace = "runai-reservation"
	ScaleAdjustNamespace = "runai-scale-adjust"
)

// ChartAuth carries optional chart credentials.
type ChartAuth struct {
	Username string
	Password string
}

func (a ChartAuth) IsZero() bool { return a.Username == "" && a.Password == "" }

// AuthenticatedHelmInstaller adds authenticated chart pulls.
type AuthenticatedHelmInstaller interface {
	InstallOrUpgradeReleaseWithAuth(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string, auth ChartAuth) error
}

// NGCNonProdChartRepo returns the configured non-prod NGC chart repository
// without a trailing slash, or "" when none is set.
func NGCNonProdChartRepo() string {
	return strings.TrimSuffix(strings.TrimSpace(os.Getenv(NGCNonProdChartRepoEnv)), "/")
}

// NGCChartRepo returns the NGC ChartMuseum root for env: the public repository
// for prod, otherwise NGCNonProdChartRepo.
func NGCChartRepo(env config.RunaiEnvType) string {
	if env == config.EnvProd {
		return ngcChartRepoProd
	}
	return NGCNonProdChartRepo()
}

// CPChartRepo returns the control plane chart repository for env under the
// configured chart repository base, or "" when no base is set.
func CPChartRepo(env config.RunaiEnvType) string {
	base := ChartRepoBase()
	if base == "" {
		return ""
	}
	switch env {
	case config.EnvProd:
		return base + "/cp-charts-prod"
	case config.EnvLab:
		return base + "/cp-charts-lab"
	case config.EnvStaging:
		return base + "/cp-charts-staging"
	default:
		return base + "/cp-charts-test"
	}
}

// ClusterChartRepo returns the cluster chart repository for env under the
// configured chart repository base, or "" when no base is set.
func ClusterChartRepo(env config.RunaiEnvType) string {
	base := ChartRepoBase()
	if base == "" {
		return ""
	}
	switch env {
	case config.EnvProd:
		return base + "/run-ai-charts"
	case config.EnvLab:
		return base + "/op-charts-lab"
	case config.EnvStaging:
		return base + "/op-charts-staging"
	default:
		return base + "/run-ai-charts-test"
	}
}

// HelmInstaller abstracts the Helm install/upgrade operation for dependency injection.
type HelmInstaller interface {
	InstallOrUpgradeRelease(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string) error
	InstallOrUpgradeFromLocalChart(ctx context.Context, chartPath, releaseName, targetNamespace string, values []string, valueFiles []string) error
	ReleaseStatus(releaseName, targetNamespace string) (exists bool, version string, err error)
}

type HelmInstallerImpl struct{}

func NewHelmInstaller() HelmInstaller {
	return &HelmInstallerImpl{}
}

func (h *HelmInstallerImpl) InstallOrUpgradeRelease(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string) error {
	return helmclient.InstallOrUpgradeRelease(ctx, repoUrl, version, releaseName, chartName, targetNamespace, isOci, values)
}

func (h *HelmInstallerImpl) InstallOrUpgradeReleaseWithAuth(ctx context.Context, repoUrl, version, releaseName, chartName, targetNamespace string, isOci bool, values []string, auth ChartAuth) error {
	return helmclient.InstallOrUpgradeReleaseWithAuth(ctx, repoUrl, version, releaseName, chartName, targetNamespace, isOci, values, helmclient.ChartAuth(auth))
}

func (h *HelmInstallerImpl) InstallOrUpgradeFromLocalChart(ctx context.Context, chartPath, releaseName, targetNamespace string, values []string, valueFiles []string) error {
	return helmclient.InstallOrUpgradeFromLocalChart(ctx, chartPath, releaseName, targetNamespace, values, valueFiles)
}

// InstallOrUpgradeReleaseWithKubeVersion renders the chart against a pinned
// Kubernetes version (e.g. "1.28.15") instead of the live cluster version. The
// major/minor are derived from the version string.
func (h *HelmInstallerImpl) InstallOrUpgradeReleaseWithKubeVersion(ctx context.Context, repoURL, version, releaseName, chartName, targetNamespace string, isOCI bool, values []string, kubeVersion string) error {
	return helmclient.InstallOrUpgradeReleaseWithKubeVersion(ctx, repoURL, version, releaseName, chartName, targetNamespace, isOCI, values, parseKubeVersion(kubeVersion))
}

func parseKubeVersion(kubeVersion string) *chartutil.KubeVersion {
	if kubeVersion == "" {
		return nil
	}
	kv := &chartutil.KubeVersion{Version: kubeVersion}
	parts := strings.Split(strings.TrimPrefix(kubeVersion, "v"), ".")
	if len(parts) >= 2 && parts[0] != "" && parts[1] != "" {
		kv.Major, kv.Minor = parts[0], parts[1]
	}
	return kv
}

func (h *HelmInstallerImpl) ReleaseStatus(releaseName, targetNamespace string) (exists bool, version string, err error) {
	return helmclient.GetReleaseStatus(releaseName, targetNamespace)
}
