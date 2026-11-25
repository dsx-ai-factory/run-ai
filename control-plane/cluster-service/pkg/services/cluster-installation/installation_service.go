/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation

import (
	"bytes"
	"context"
	"slices"
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

// repoLookupTimeout bounds the index lookups done to pick the install repo, so an unreachable
// repository cannot stall command generation.
const repoLookupTimeout = 10 * time.Second

type HelmRepo struct {
	URL                 string
	Username            string
	PasswordPlaceholder string
}

// publicRepoEnabled reports whether the public repo may be used at all. It is off when no public URL is
// configured, and when the private URL is blank, which is how air-gapped installs signal that nothing outside
// their mirror is reachable. Both keep the old single-URL behavior.
func publicRepoEnabled(installationConfig *config.Installation) bool {
	return installationConfig.HelmRegistryUrlPublic != "" && installationConfig.HelmRegistryUrl != ""
}

func privateHelmRepo(installationConfig *config.Installation) HelmRepo {
	return HelmRepo{
		URL:                 installationConfig.HelmRegistryUrl,
		Username:            installationConfig.HelmRegistryUsername,
		PasswordPlaceholder: installationConfig.HelmRegistryPasswordPlaceholder,
	}
}

type ClusterInstallService interface {
	GetClusterSupportedVersions(ctx context.Context, config *config.Installation) (map[string]string, error)
	GetInstallationString(data *InstallationData) string
	// GetHelmRepoForVersion returns the repository whose chart should be installed for the given cluster version.
	GetHelmRepoForVersion(ctx context.Context, config *config.Installation, version string) HelmRepo
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

	results := v.fetchVersionLists(ctx, urls)
	var versionList VersionsList
	failed := 0
	for _, res := range results {
		if res.err != nil {
			failed++
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetClusterVersions - Failed to get versions list from %s - %s", res.url, res.err.Error())
			continue
		}
		versionList = append(versionList, res.list...)
	}
	if failed == len(results) {
		// every repo failed (e.g. no credentials, or no outbound access): return the current version
		return getCurrentVersionAsVersionList(curVersion, minVersion, maxVersion), nil
	}

	allVersions := getLatestMinorVersions(versionList, minVersion, maxVersion)
	response := excludeVersions(allVersions, installationConfig.ExcludeVersions)
	return response, nil
}

// registryUrls lists the repositories to read versions from: the public one first, then the private one.
func registryUrls(installationConfig *config.Installation) []string {
	var urls []string
	if publicRepoEnabled(installationConfig) {
		urls = append(urls, installationConfig.HelmRegistryUrlPublic)
	}
	if installationConfig.HelmRegistryUrl != "" && !slices.Contains(urls, installationConfig.HelmRegistryUrl) {
		urls = append(urls, installationConfig.HelmRegistryUrl)
	}
	return urls
}

type repoVersions struct {
	url  string
	list VersionsList
	err  error
}

// fetchVersionLists reads the version list of every repo concurrently, so an unreachable one costs a single
// timeout rather than one per repo. Results keep the order of urls.
func (v *ClusterInstallServiceImpl) fetchVersionLists(ctx context.Context, urls []string) []repoVersions {
	results := make([]repoVersions, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("GetClusterVersions - trying to get versions list from %s ", url)
			list, err := v.helmAction.GetVersionListFromHelm(ctx, url)
			results[i] = repoVersions{url: url, list: list, err: err}
		}(i, url)
	}
	wg.Wait()
	return results
}

func listsVersion(list VersionsList, version *semver.Version) bool {
	for _, entry := range list {
		if candidate, err := semver.NewVersion(entry.Version); err == nil && candidate.Equal(version) {
			return true
		}
	}
	return false
}

// GetHelmRepoForVersion picks the repo to install the given cluster version from. The public repo wins when it
// lists the version. Otherwise the private repo is used if it lists the version, or if credentials for it are
// configured (its index may not be readable without them, so it cannot be checked). Anything else, including
// non-concrete versions such as "latest", falls back to the public repo, which needs no credentials.
func (v *ClusterInstallServiceImpl) GetHelmRepoForVersion(ctx context.Context, installationConfig *config.Installation, version string) HelmRepo {
	private := privateHelmRepo(installationConfig)
	if !publicRepoEnabled(installationConfig) {
		return private
	}
	public := HelmRepo{URL: installationConfig.HelmRegistryUrlPublic}
	concrete, err := semver.NewVersion(version)
	if err != nil {
		return public
	}

	ctx, cancel := context.WithTimeout(ctx, repoLookupTimeout)
	defer cancel()
	results := v.fetchVersionLists(ctx, registryUrls(installationConfig))
	listed := map[string]bool{}
	for _, res := range results {
		if res.err == nil && listsVersion(res.list, concrete) {
			listed[res.url] = true
		}
	}

	if listed[public.URL] {
		return public
	}
	if listed[private.URL] || private.Username != "" {
		return private
	}
	return public
}
