/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	runaiRepoIdentifier                 = "runai"
	minVersionForDisplayingIngressClass = "2.24.0"
	defaultIngressClass                 = "haproxy"
)

var installationTemplate, _ = template.New("InstallationStr").Parse(`helm repo add {{.RepositoryName}} {{.RepoURL}}{{if .HelmRegistryUsername}} \
--username='{{.HelmRegistryUsername}}' --password='{{.PasswordPlaceholder}}'{{end}} --force-update
helm repo update
helm upgrade -i runai-cluster {{.RepositoryName }}/runai-cluster -n runai \
--set controlPlane.url={{.Host}} \
--set controlPlane.bootstrapToken={{.BootstrapToken}} \
--set cluster.uid={{.Uuid}} \
--set cluster.url={{.RemoteClusterUrl}} --version="{{.Version}}" --create-namespace{{.IngressClassFlag}}{{.FipsModeFlag}}`)

type InstallationData struct {
	RepositoryName       string
	RepoURL              string
	HelmRegistryUsername string
	PasswordPlaceholder  string
	Host                 string
	BootstrapToken       string
	Uuid                 string
	RemoteClusterUrl     string
	Version              string
	IngressClass         *string
	IngressClassFlag     string
	FipsMode             string
	FipsModeFlag         string
	LastLiveness         *time.Time
}

type ClusterInstallService interface {
	GetClusterSupportedVersions(ctx context.Context, config *config.Installation) (map[string]string, error)
	GetInstallationString(data *InstallationData) string
}

type ClusterInstallServiceImpl struct {
	helmAction HelmActions
}

func NewClusterInstallService(helmAction HelmActions) ClusterInstallService {
	return &ClusterInstallServiceImpl{
		helmAction: helmAction,
	}
}

func (v *ClusterInstallServiceImpl) GetInstallationString(data *InstallationData) string {
	data.RepositoryName = runaiRepoIdentifier
	data.IngressClassFlag = buildIngressClassFlag(data.IngressClass, data.LastLiveness, data.Version)
	data.FipsModeFlag = buildFipsModeFlag(data.FipsMode)
	var buf bytes.Buffer
	_ = installationTemplate.Execute(&buf, data)
	return buf.String()
}

func buildFipsModeFlag(fipsMode string) string {
	if fipsMode == "" {
		return ""
	}
	return " \\\n--set global.fipsMode=" + fipsMode
}

func buildIngressClassFlag(reportedIngressClass *string, lastLiveness *time.Time, desiredVersion string) string {
	isClusterAlreadyInstalled := lastLiveness != nil

	if isClusterAlreadyInstalled {
		if reportedIngressClass != nil && *reportedIngressClass != "" {
			return " \\\n--set clusterConfig.global.ingress.ingressClass=" + *reportedIngressClass
		}
		return ""
	}

	if shouldAddIngressClassFlag(desiredVersion) {
		return " \\\n--set clusterConfig.global.ingress.ingressClass=" + defaultIngressClass
	}
	return ""
}

func shouldAddIngressClassFlag(desiredVersion string) bool {
	minVer, err := semver.NewVersion(minVersionForDisplayingIngressClass)
	if err != nil {
		logging.Errorf("Failed to parse min version for haproxy ingress class: %s", err.Error())
		return false
	}

	ver, err := semver.NewVersion(desiredVersion)
	if err != nil {
		logging.Errorf("Failed to parse desired version: %s", err.Error())
		return false
	}

	return ver.GreaterThan(minVer) || ver.Equal(minVer)
}

func (v *ClusterInstallServiceImpl) GetClusterSupportedVersions(ctx context.Context, installationConfig *config.Installation) (map[string]string, error) {
	curVersion := installationConfig.Version
	minVersion, maxVersion := adjustMinMaxVersion(ctx,
		installationConfig.MinVersion,
		installationConfig.MaxVersion,
		curVersion)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
		"CLUSTER_INSTALLATION_MIN_VERSION - CLUSTER_INSTALLATION_MAX_VERSION is set to %s - %s. Effective range is %s - %s",
		installationConfig.MinVersion, installationConfig.MaxVersion, minVersion, maxVersion)
	url := installationConfig.HelmRegistryUrl
	if url == "" {
		// In case of air gapped installation, the url is set empty, we return the current version as the only available version
		defaultResponse := getCurrentVersionAsVersionList(curVersion, minVersion, maxVersion)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("GetClusterVersions - CLUSTER_INSTALLATION_HELM_REGISTRY_URL is not set. returning current cluster version:%s", curVersion)
		return defaultResponse, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("GetClusterVersions - trying to get versions list from %s ", url)
	versionList, err := v.helmAction.GetVersionListFromHelm(ctx, url)
	if err != nil {
		// if failed to get versions list, return the current version
		defaultResponse := getCurrentVersionAsVersionList(curVersion, minVersion, maxVersion)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("GetClusterVersions - Failed to get versions list from %s - %s", url, err.Error())
		return defaultResponse, nil
	}

	allVersions := getLatestMinorVersions(versionList, minVersion, maxVersion)
	response := excludeVersions(allVersions, installationConfig.ExcludeVersions)
	return response, nil
}
