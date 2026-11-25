/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"sync"
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

const minVersionForPublicRepo = "3.0.0"

type HelmRepo struct {
	URL                 string
	Username            string
	PasswordPlaceholder string
}

const ngcHelmHost = "https://helm.ngc.nvidia.com"

func isNgcUrl(url string) bool {
	return url == ngcHelmHost || strings.HasPrefix(url, ngcHelmHost+"/")
}

// publicRepoEnabled reports whether the public repo may be used at all. It is off when no public URL is
// configured and for air-gapped installs (installation type "airgap", or the legacy URL blanked, which is how
// the air-gap values signal that nothing outside the mirror is reachable), so those keep the old behavior.
func publicRepoEnabled(installationConfig *config.Installation) bool {
	return installationConfig.HelmRegistryUrlPublic != "" &&
		installationConfig.HelmRegistryUrl != "" &&
		installationConfig.InstallationType != "airgap"
}

// legacyRepoUsable reports whether the legacy repo can serve requests. When the public repo is also in play,
// the legacy URL is the private NGC repo, which cannot be used without credentials, so it is skipped instead of
// producing a failing command. In single-URL mode the configured URL is always used, as before.
func legacyRepoUsable(installationConfig *config.Installation) bool {
	if installationConfig.HelmRegistryUrl == "" {
		return false
	}
	if !publicRepoEnabled(installationConfig) {
		return true
	}
	return !isNgcUrl(installationConfig.HelmRegistryUrl) || installationConfig.HelmRegistryUsername != ""
}

func SelectHelmRepoForVersion(installationConfig *config.Installation, version string) HelmRepo {
	legacy := HelmRepo{
		URL:                 installationConfig.HelmRegistryUrl,
		Username:            installationConfig.HelmRegistryUsername,
		PasswordPlaceholder: installationConfig.HelmRegistryPasswordPlaceholder,
	}
	if !publicRepoEnabled(installationConfig) {
		return legacy
	}
	minVer, err := semver.NewVersion(minVersionForPublicRepo)
	if err != nil {
		return legacy
	}
	public := HelmRepo{URL: installationConfig.HelmRegistryUrlPublic}
	ver, err := semver.NewVersion(version)
	if err != nil {
		return public
	}
	if ver.LessThan(minVer) && legacyRepoUsable(installationConfig) {
		return legacy
	}
	return public
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
	urls := registryUrls(installationConfig)
	if len(urls) == 0 {
		// In case of air gapped installation, the url is set empty, we return the current version as the only available version
		defaultResponse := getCurrentVersionAsVersionList(curVersion, minVersion, maxVersion)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("GetClusterVersions - CLUSTER_INSTALLATION_HELM_REGISTRY_URL is not set. returning current cluster version:%s", curVersion)
		return defaultResponse, nil
	}

	// Repos are fetched concurrently so an unreachable one (e.g. a blocked OSS endpoint) costs one timeout,
	// not one per repo. Results are merged in urls order to stay deterministic.
	lists := make([]VersionsList, len(urls))
	errs := make([]error, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("GetClusterVersions - trying to get versions list from %s ", url)
			lists[i], errs[i] = v.helmAction.GetVersionListFromHelm(ctx, url)
		}(i, url)
	}
	wg.Wait()

	var versionList VersionsList
	failed := 0
	for i, url := range urls {
		if errs[i] != nil {
			failed++
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetClusterVersions - Failed to get versions list from %s - %s", url, errs[i].Error())
			continue
		}
		versionList = append(versionList, lists[i]...)
	}
	if failed == len(urls) {
		// every repo failed (e.g. no credentials, or no outbound access): return the current version
		return getCurrentVersionAsVersionList(curVersion, minVersion, maxVersion), nil
	}

	allVersions := getLatestMinorVersions(versionList, minVersion, maxVersion)
	response := excludeVersions(allVersions, installationConfig.ExcludeVersions)
	return response, nil
}

func registryUrls(installationConfig *config.Installation) []string {
	var urls []string
	if publicRepoEnabled(installationConfig) {
		urls = append(urls, installationConfig.HelmRegistryUrlPublic)
	}
	if legacyRepoUsable(installationConfig) && !slices.Contains(urls, installationConfig.HelmRegistryUrl) {
		urls = append(urls, installationConfig.HelmRegistryUrl)
	}
	return urls
}
