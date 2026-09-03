/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/types"
)

type ReleaseClientFactory func(types.NamespacedName, DownloaderFactory) (*ReleaseClient, error)

// ReleaseClient is responsible for running various helm actions for a given helm release.
type ReleaseClient struct {
	ReleaseId types.NamespacedName // Identifier of the helm release

	DownloaderFactory DownloaderFactory

	Settings     *cli.EnvSettings
	ActionConfig *action.Configuration
}

func NewReleaseClient(releaseId types.NamespacedName, downloaderFactory DownloaderFactory) (*ReleaseClient, error) {
	if downloaderFactory == nil {
		return nil, fmt.Errorf("downloaderFactory must not be nil")
	}

	settings := cli.New() // picks up HELM_CACHE_HOME / HELM_CONFIG_HOME set in main()
	settings.Debug = config.Config().UseDebugLogLevel
	settings.SetNamespace(releaseId.Namespace)

	// Ensure Helm cache directories exist
	if err := ensureWorkingDirectories(settings); err != nil {
		return nil, fmt.Errorf("failed to create Helm directories: %w", err)
	}

	actionConfig := &action.Configuration{}

	err := actionConfig.Init(
		settings.RESTClientGetter(),
		releaseId.Namespace,
		config.Config().HelmDriver,
		log.Printf,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init helm action config: %w", err)
	}

	return &ReleaseClient{
		ReleaseId:         releaseId,
		Settings:          settings,
		ActionConfig:      actionConfig,
		DownloaderFactory: downloaderFactory,
	}, nil
}

// GetRelease returns information about the helm release
func (hrc *ReleaseClient) GetRelease() (*release.Release, error) {
	statusClient := action.NewStatus(hrc.ActionConfig)

	rel, err := statusClient.Run(hrc.ReleaseId.Name)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, nil // Release doesn't exist
		}
		return nil, fmt.Errorf("failed to get release status: %w", err)
	}

	return rel, nil
}

// run preparations to install a release, including:
// 1) download of the chart
// 2) check dependencies
func (hrc *ReleaseClient) prepareToInstall(params *InstallParams) (*chart.Chart, error) {
	log.Info().Msgf("HelmTask --> preparing to install %s started", hrc.ReleaseId.String())

	chartDownloader := hrc.DownloaderFactory(hrc, params)

	chartPath, err := chartDownloader.Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download chart %v", err)
	}

	log.Info().Msgf("HelmTask --> chart of %s downloaded to %s", hrc.ReleaseId.String(), chartPath)

	// Load the chart
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %v", err)
	}

	if deps := chartRequested.Metadata.Dependencies; deps != nil {
		err := action.CheckDependencies(chartRequested, deps)
		if err != nil {
			return nil, fmt.Errorf("failed to check dependencies: %v", err)
		}

		// Dependency update only works on unpacked chart directories; packaged charts (.tgz) already include deps.
		chartPathInfo, err := os.Stat(chartPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat chart path: %w", err)
		}

		if chartPathInfo.IsDir() {
			man := &downloader.Manager{
				Out:              os.Stdout,
				ChartPath:        chartPath,
				Keyring:          "",
				SkipUpdate:       false,
				Getters:          getter.All(hrc.Settings),
				RepositoryConfig: hrc.Settings.RepositoryConfig,
				RepositoryCache:  hrc.Settings.RepositoryCache,
				Debug:            hrc.Settings.Debug,
			}

			err = man.Update()
			if err != nil {
				return chartRequested, fmt.Errorf("failed to update dependencies: %v", err)
			}
			log.Info().Msgf("HelmTask --> update dependencies of %s completed", hrc.ReleaseId.String())
			chartRequested, err = loader.Load(chartPath)
			if err != nil {
				return nil, fmt.Errorf("failed to reload chart after dependency update: %w", err)
			}
		} else {
			log.Debug().Msgf("HelmTask --> chart is packaged (file), skipping dependency update for %s", hrc.ReleaseId.String())
		}
	} else {
		log.Debug().Msgf("HelmTask --> no dependencies for %s", hrc.ReleaseId.String())
	}

	return chartRequested, nil
}

func ChartUrl(helmSpec *v1alpha1.HelmSpec) string {
	if helmSpec == nil {
		return ""
	}
	if helmSpec.ChartUrl != "" {
		return helmSpec.ChartUrl
	}
	if helmSpec.ChartRef != nil {
		s := helmSpec.ChartRef.Repository + "/" + helmSpec.ChartRef.Name
		if helmSpec.ChartRef.Version != "" {
			s += ":" + helmSpec.ChartRef.Version
		}
		return s
	}
	return ""
}

// ReleaseMatchesDesired returns true if the current release already matches the desired spec
// (chart URL and values). Uses status.DeployedChartUrl when set; otherwise falls back to release metadata.
func ReleaseMatchesDesired(rel *release.Release, spec *v1alpha1.HelmSpec, status *v1alpha1.HelmStatus) bool {
	desiredChartUrl := ChartUrl(spec)
	if desiredChartUrl == "" {
		return false
	}

	// the values were already checked in NewInstallParams so we assume there is no conversion issues
	desiredValues, _ := convertHelmValues(spec.Values)

	deployedChartUrl := ""
	if status != nil {
		deployedChartUrl = status.ChartUrl
	}

	// Prefer comparing stored DeployedChartUrl to desired (covers both ChartUrl and ChartRef).
	if deployedChartUrl != "" {
		if deployedChartUrl != desiredChartUrl {
			return false
		}
		return valuesEqual(rel.Config, desiredValues)
	}

	// Fallback when DeployedChartUrl not yet set (e.g. old releases): use release metadata.
	if rel == nil || rel.Chart == nil || rel.Chart.Metadata == nil {
		return false
	}

	chartName := rel.Chart.Metadata.Name
	chartVersion := rel.Chart.Metadata.Version

	if spec.ChartRef != nil {
		if chartName != spec.ChartRef.Name {
			return false
		}
		if spec.ChartRef.Version != "" && chartVersion != spec.ChartRef.Version {
			return false
		}
	}

	return valuesEqual(rel.Config, desiredValues)
}

// valuesEqual compares two value maps (nil and empty map treated as equal).
func valuesEqual(a, b map[string]interface{}) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}
