/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	"helm.sh/helm/v3/pkg/chartutil"
)

const jfrogBase = "https://runai.jfrog.io/artifactory/api/helm"

const ngcChartHost = "helm.ngc.nvidia.com"

const (
	ngcChartOrgTeamProd    = "nvidia/runai"
	ngcChartOrgTeamNonProd = "nvstaging/runai"
)

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

func ngcChartOrgTeam(env config.RunaiEnvType) string {
	if env == config.EnvProd {
		return ngcChartOrgTeamProd
	}
	return ngcChartOrgTeamNonProd
}

// NGCChartRepo returns the NGC ChartMuseum root.
func NGCChartRepo(env config.RunaiEnvType) string {
	return fmt.Sprintf("https://%s/%s", ngcChartHost, ngcChartOrgTeam(env))
}

func CPChartRepo(env config.RunaiEnvType) string {
	switch env {
	case config.EnvProd:
		return jfrogBase + "/cp-charts-prod"
	case config.EnvLab:
		return jfrogBase + "/cp-charts-lab"
	case config.EnvStaging:
		return jfrogBase + "/cp-charts-staging"
	default:
		return jfrogBase + "/cp-charts-test"
	}
}

func ClusterChartRepo(env config.RunaiEnvType) string {
	switch env {
	case config.EnvProd:
		return jfrogBase + "/run-ai-charts"
	case config.EnvLab:
		return jfrogBase + "/op-charts-lab"
	case config.EnvStaging:
		return jfrogBase + "/op-charts-staging"
	default:
		return jfrogBase + "/run-ai-charts-test"
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
