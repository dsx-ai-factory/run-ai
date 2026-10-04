/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

type InstallClusterChart struct {
	ClusterName     string
	Domain          string
	ControlPlaneURL string
	ClusterUID      string
	BootstrapToken  string
	ClientSecret    string
	Credentials     *step.ControlPlaneCredentials
	CACertPath      string
	Repo            string
	Version         string
	ReleaseName     string
	ChartName       string
	Namespace       string
	IsOCI           bool
	Auth            step.ChartAuth
	HostRouting     bool
	ExtraValues     []string
}

func (a InstallClusterChart) Kind() string { return "runai.install_cluster_chart" }

func (a InstallClusterChart) Release() (name, namespace string) { return a.ReleaseName, a.Namespace }

func (a InstallClusterChart) ClusterRef() (clusterName string, controlPlane ControlPlaneRef) {
	return a.ClusterName, ControlPlaneRef{URL: a.ControlPlaneURL, Domain: a.Domain, CACertPath: a.CACertPath, Credentials: a.Credentials}
}

func (a InstallClusterChart) UsesSuppliedCredentials() bool {
	return a.ClusterUID != ""
}

func (a InstallClusterChart) Args() map[string]any {
	return map[string]any{
		"cluster":   a.ClusterName,
		"repo":      a.Repo,
		"version":   a.Version,
		"release":   a.ReleaseName,
		"chart":     a.ChartName,
		"namespace": a.Namespace,
	}
}

func (a InstallClusterChart) Describe() string {
	base := fmt.Sprintf("helm upgrade --install %s %s --version %s --namespace %s", a.ReleaseName, a.ChartName, a.Version, a.Namespace)
	cmd := describeHelmCommand(base, a.ExtraValues, nil)
	if a.UsesSuppliedCredentials() {
		return cmd
	}
	return fmt.Sprintf("%s (resolves cluster %q uid + control-plane credential)", cmd, a.ClusterName)
}

func (a InstallClusterChart) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
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

	return orchestrator2.ActionResult{}, installChart(ctx, deps.HelmInstaller, a.Repo, a.Version, a.ReleaseName, a.ChartName, a.Namespace, a.IsOCI, values, a.Auth)
}
