/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
)

type InstallClusterChartLocal struct {
	ClusterName     string
	Domain          string
	ControlPlaneURL string
	ClusterUID      string
	BootstrapToken  string
	ClientSecret    string
	Credentials     *step.ControlPlaneCredentials
	CACertPath      string
	ChartPath       string
	WorkingDir      string
	ChartPrefix     string
	LockChart       string
	Version         string
	ReleaseName     string
	Namespace       string
	HostRouting     bool
	ValueFiles      []string
	ExtraValues     []string
}

func (a InstallClusterChartLocal) chartPath() string {
	if a.LockChart != "" {
		return artifacts.LockedChartPreviewPath(a.WorkingDir, a.LockChart)
	}
	if a.ChartPrefix != "" {
		return airgap.FindChartFile(a.WorkingDir, a.ChartPrefix)
	}
	return a.ChartPath
}

func (a InstallClusterChartLocal) resolvedChartPath() (string, error) {
	if a.LockChart != "" {
		return artifacts.LockedChartPath(a.WorkingDir, a.LockChart)
	}
	return a.chartPath(), nil
}

func (a InstallClusterChartLocal) Kind() string { return "runai.install_cluster_chart_local" }

func (a InstallClusterChartLocal) Release() (name, namespace string) {
	return a.ReleaseName, a.Namespace
}

func (a InstallClusterChartLocal) ClusterRef() (clusterName string, controlPlane ControlPlaneRef) {
	return a.ClusterName, ControlPlaneRef{URL: a.ControlPlaneURL, Domain: a.Domain, CACertPath: a.CACertPath, Credentials: a.Credentials}
}

func (a InstallClusterChartLocal) UsesSuppliedCredentials() bool {
	return a.ClusterUID != ""
}

func (a InstallClusterChartLocal) Args() map[string]any {
	return map[string]any{
		"cluster":    a.ClusterName,
		"chartPath":  a.chartPath(),
		"release":    a.ReleaseName,
		"namespace":  a.Namespace,
		"valueFiles": a.ValueFiles,
	}
}

func (a InstallClusterChartLocal) Describe() string {
	base := fmt.Sprintf("helm upgrade --install %s %s --namespace %s", a.ReleaseName, a.chartPath(), a.Namespace)
	cmd := describeHelmCommand(base, a.ExtraValues, a.ValueFiles)
	if a.UsesSuppliedCredentials() {
		return cmd
	}
	return fmt.Sprintf("%s (resolves cluster %q uid + control-plane credential)", cmd, a.ClusterName)
}

func (a InstallClusterChartLocal) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	chartPath, err := a.resolvedChartPath()
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	var values []string
	if a.UsesSuppliedCredentials() {
		values = []string{
			"controlPlane.url=" + a.ControlPlaneURL,
			step.ClusterAuthValueFromCredentials(a.BootstrapToken, a.ClientSecret),
			"cluster.uid=" + a.ClusterUID,
			"cluster.url=" + a.Domain,
		}
	} else {
		_, controlPlane := a.ClusterRef()
		apiClient, err := controlPlane.Client(deps)
		if err != nil {
			return orchestrator2.ActionResult{}, err
		}
		controlPlaneURL := a.Domain
		if a.ControlPlaneURL != "" {
			controlPlaneURL = a.ControlPlaneURL
		}
		values, err = clusterChartCredentialValues(ctx, apiClient, a.ClusterName, a.Version, controlPlaneURL, a.Domain)
		if err != nil {
			return orchestrator2.ActionResult{}, err
		}
	}
	if a.CACertPath != "" {
		values = append(values, "global.customCA.enabled=true")
	}
	values = append(values, spec.SubdomainSupportHelmKey+"="+strconv.FormatBool(a.HostRouting))
	values = mergeValueStrings(values, a.ExtraValues)

	return orchestrator2.ActionResult{}, deps.HelmInstaller.InstallOrUpgradeFromLocalChart(ctx, chartPath, a.ReleaseName, a.Namespace, values, a.ValueFiles)
}
