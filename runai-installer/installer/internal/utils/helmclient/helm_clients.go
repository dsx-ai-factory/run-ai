/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helmclient

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/storage/driver"
)

type HelmClientAPI interface {
	init(releaseName, chartName, targetNamespace string) error
	runChart(ctx context.Context, chartName string, chart *chart.Chart, vals map[string]interface{}) error
	handleValues() (map[string]interface{}, error)
	getChartPath() string
	handleDependencies(chartRequested *chart.Chart) error
	getChart() (*chart.Chart, error)
}

// ChartAuth holds optional chart credentials.
type ChartAuth struct {
	Username string
	Password string
}

func (a ChartAuth) IsZero() bool { return a.Username == "" && a.Password == "" }

type helmChartData struct {
	repoUrl        string
	chartVersion   string
	valuesOverride []string
	valueFiles     []string
	auth           ChartAuth
}

type helmClient struct {
	data          *helmChartData
	chartFullPath string
	actionConfig  *action.Configuration
	settings      *cli.EnvSettings
	providers     getter.Providers
	ociChart      bool
}

type InstallClient struct {
	*helmClient
	clientInstall *action.Install
}

type UpgradeClient struct {
	*helmClient
	clientUpgrade *action.Upgrade
	releaseName   string
}

func newClient(chartData *helmChartData, releaseName, chartName, targetNamespace string, kubeVersion *chartutil.KubeVersion, isOci bool) (HelmClientAPI, error) {
	actionConfig, settings, err := createActionConfig(targetNamespace, kubeVersion)
	if err != nil {
		return nil, err
	}
	var client HelmClientAPI
	helmClient := &helmClient{
		data:         chartData,
		actionConfig: actionConfig,
		settings:     settings,
		providers:    getter.All(settings),
		ociChart:     isOci,
	}
	if isOci {
		client = &InstallClient{
			helmClient: helmClient,
		}
	} else {
		histClient := action.NewHistory(actionConfig)
		histClient.Max = 1
		_, err = histClient.Run(releaseName)
		if errors.Is(err, driver.ErrReleaseNotFound) {
			client = &InstallClient{helmClient: helmClient}
		} else if err != nil {
			return nil, fmt.Errorf("failed to check release history: %w", err)
		} else {
			client = &UpgradeClient{helmClient: helmClient}
		}

	}
	err = client.init(releaseName, chartName, targetNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm client: %v", err)
	}
	return client, nil
}

func (installActions *InstallClient) init(releaseName, chartName, targetNamespace string) error {
	clientInstall := action.NewInstall(installActions.actionConfig)
	clientInstall.RepoURL = installActions.data.repoUrl
	clientInstall.ReleaseName = releaseName
	clientInstall.Namespace = targetNamespace
	clientInstall.Version = installActions.data.chartVersion
	clientInstall.CreateNamespace = true
	clientInstall.DependencyUpdate = true
	clientInstall.SkipCRDs = false
	clientInstall.Wait = true
	clientInstall.Atomic = true
	clientInstall.Timeout = 15 * time.Minute
	clientInstall.Username = installActions.data.auth.Username
	clientInstall.Password = installActions.data.auth.Password
	installActions.clientInstall = clientInstall
	if installActions.ociChart {
		return nil
	}
	cp, err := clientInstall.LocateChart(chartName, installActions.settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %v", err)
	}
	installActions.chartFullPath = cp
	return nil
}

func (installActions *InstallClient) runChart(ctx context.Context, chartName string, chart *chart.Chart, vals map[string]interface{}) error {
	_, err := installActions.clientInstall.RunWithContext(ctx, chart, vals)
	if err != nil {
		return fmt.Errorf("failed to run install / upgrade chart: %v", err)
	}
	return nil
}

func (installActions *InstallClient) getChart() (*chart.Chart, error) {
	return installActions.helmClient.getChart()
}

func (upgradeActions *UpgradeClient) init(releaseName, chartName, targetNamespace string) error {
	clientUpgrade := action.NewUpgrade(upgradeActions.actionConfig)
	clientUpgrade.Namespace = targetNamespace
	clientUpgrade.RepoURL = upgradeActions.data.repoUrl
	clientUpgrade.Version = upgradeActions.data.chartVersion
	clientUpgrade.DependencyUpdate = true
	clientUpgrade.Wait = true
	clientUpgrade.Atomic = true
	clientUpgrade.CleanupOnFail = true
	clientUpgrade.Timeout = 20 * time.Minute
	clientUpgrade.Username = upgradeActions.data.auth.Username
	clientUpgrade.Password = upgradeActions.data.auth.Password
	upgradeActions.clientUpgrade = clientUpgrade
	upgradeActions.releaseName = releaseName
	cp, err := clientUpgrade.LocateChart(chartName, upgradeActions.settings)
	if err != nil {
		return fmt.Errorf("failed to locate chart: %v", err)
	}
	upgradeActions.chartFullPath = cp
	return nil
}

func (upgradeActions *UpgradeClient) runChart(ctx context.Context, chartName string, chart *chart.Chart, vals map[string]interface{}) error {
	_, err := upgradeActions.clientUpgrade.RunWithContext(ctx, upgradeActions.releaseName, chart, vals)
	if err != nil {
		return fmt.Errorf("failed to run install chart: %v", err)
	}
	return nil
}

func (upgradeActions *UpgradeClient) isOci() bool {
	return upgradeActions.ociChart
}

func (upgradeActions *UpgradeClient) getChart() (*chart.Chart, error) {
	return upgradeActions.helmClient.getChart()
}

func (helmClient *helmClient) handleDependencies(chartRequested *chart.Chart) error {
	if deps := chartRequested.Metadata.Dependencies; deps != nil {
		if err := action.CheckDependencies(chartRequested, deps); err != nil {
			err = errors.Wrap(err, "An error occurred while checking for chart dependencies. You may need to run `helm dependency build` to fetch missing dependencies")
			man := &downloader.Manager{
				Out:              logging.InfoWriter(),
				ChartPath:        helmClient.chartFullPath,
				Keyring:          "",
				SkipUpdate:       false,
				Getters:          helmClient.providers,
				RepositoryConfig: helmClient.settings.RepositoryConfig,
				RepositoryCache:  helmClient.settings.RepositoryCache,
				Debug:            false,
			}
			if err := man.Update(); err != nil {
				return fmt.Errorf("failed to update dependency: %v", err)
			}
			// Reload the chart with the updated Chart.lock file.
			if chartRequested, err = loader.Load(helmClient.chartFullPath); err != nil {
				return fmt.Errorf("failed reloading chart after repo update")
			}
		}
	}
	return nil
}
func (helmClient *helmClient) handleValues() (map[string]interface{}, error) {
	valueOpts := &values.Options{
		Values:     helmClient.data.valuesOverride,
		ValueFiles: helmClient.data.valueFiles,
	}
	return valueOpts.MergeValues(helmClient.providers)
}

func (helmClient *helmClient) getChartPath() string {
	return helmClient.chartFullPath
}

// ociRegistryHost extracts the login host from an OCI reference.
func ociRegistryHost(ref string) string {
	ref = strings.TrimPrefix(ref, "oci://")
	if i := strings.IndexByte(ref, '/'); i >= 0 {
		ref = ref[:i]
	}
	return ref
}

func (helmClient *helmClient) getChart() (*chart.Chart, error) {
	var chart *chart.Chart
	var err error
	if helmClient.ociChart {
		regClient, err := registry.NewClient(registry.ClientOptEnableCache(true))
		if err != nil {
			return nil, fmt.Errorf("failed to create registry client: %v", err)
		}
		if !helmClient.data.auth.IsZero() {
			host := ociRegistryHost(helmClient.data.repoUrl)
			if err := regClient.Login(host, registry.LoginOptBasicAuth(helmClient.data.auth.Username, helmClient.data.auth.Password)); err != nil {
				return nil, fmt.Errorf("failed to authenticate to registry %q: %v", host, err)
			}
		}
		ref := helmClient.data.repoUrl
		if helmClient.data.chartVersion != "" {
			ref = fmt.Sprintf("%s:%s", helmClient.data.repoUrl, helmClient.data.chartVersion)
		}
		pullResult, err := regClient.Pull(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to pull OCI chart: %v", err)
		}
		chart, err = loader.LoadArchive(bytes.NewReader(pullResult.Chart.Data))
		if err != nil {
			return nil, fmt.Errorf("failed to load OCI chart: %v", err)
		}
	} else {
		chart, err = loader.Load(helmClient.chartFullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load chart: %v", err)
		}
	}
	return chart, nil
}
