/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

type InstallClusterChart struct {
	ClusterName string
	Domain      string
	CACertPath  string
	Repo        string
	Version     string
	ReleaseName string
	ChartName   string
	Namespace   string
	IsOCI       bool
	Auth        step.ChartAuth
	HostRouting bool
	ExtraValues []string
}

func (a InstallClusterChart) Kind() string { return "runai.install_cluster_chart" }

func (a InstallClusterChart) Release() (name, namespace string) { return a.ReleaseName, a.Namespace }

func (a InstallClusterChart) ClusterRef() (clusterName, domain, caCertPath string) {
	return a.ClusterName, a.Domain, a.CACertPath
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
	return fmt.Sprintf("%s (resolves cluster %q uid + control-plane credential)", cmd, a.ClusterName)
}

func (a InstallClusterChart) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	url := ControlPlaneURL(a.Domain)
	apiClient, err := deps.RunAIClientProvider(url, a.CACertPath)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	uuid, err := FindClusterUUID(ctx, apiClient, a.ClusterName)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if uuid == "" {
		return orchestrator2.ActionResult{}, fmt.Errorf("cluster %q not found in control plane", a.ClusterName)
	}
	installInfo, statusCode, err := apiClient.GetClusterInstallInfoByUUID(ctx, uuid, a.Version, a.Domain)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if statusCode != http.StatusOK {
		return orchestrator2.ActionResult{}, fmt.Errorf("get cluster install info: unexpected status %d", statusCode)
	}

	values := []string{
		"controlPlane.url=" + a.Domain,
		step.ClusterAuthValue(installInfo),
		"cluster.uid=" + uuid,
		"cluster.url=" + a.Domain,
	}
	if a.CACertPath != "" {
		values = append(values, "global.customCA.enabled=true")
	}
	values = append(values, spec.SubdomainSupportHelmKey+"="+strconv.FormatBool(a.HostRouting))
	values = mergeValueStrings(values, a.ExtraValues)

	return orchestrator2.ActionResult{}, installChart(ctx, deps.HelmInstaller, a.Repo, a.Version, a.ReleaseName, a.ChartName, a.Namespace, a.IsOCI, values, a.Auth)
}
