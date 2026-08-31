/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	controlplanedetailsdownloader "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	rwlc "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs"
	"golang.org/x/exp/slices"
	"k8s.io/utils/ptr"
)

type prometheusUpgrade struct {
	installationData *controlplanedetailsdownloader.InstallationConfigurationData
}

func newPrometheusUpgrader(installationData *controlplanedetailsdownloader.InstallationConfigurationData) *prometheusUpgrade {
	return &prometheusUpgrade{
		installationData: installationData,
	}
}

func (u *prometheusUpgrade) shouldPerformUpgrade(resource *runaiv1.RunaiConfig) bool {
	prometheus := resource.Spec.Prometheus
	if prometheus == nil {
		return false
	}
	remoteWrite := prometheus.Spec.RemoteWrite
	if len(remoteWrite) == 0 {
		return false
	}
	runaiRemoteWriteUrl := u.extractRunaiRemoteWriteUrl()
	if runaiRemoteWriteUrl == nil {
		return false
	}
	url := remoteWrite[0].URL
	name := remoteWrite[0].Name
	if url == *runaiRemoteWriteUrl && (name == nil || *name == "") {
		return true
	}
	return false
}

func (u *prometheusUpgrade) extractRunaiRemoteWriteUrl() *string {
	for _, remoteWrite := range u.installationData.Global.ControlPlane.Config.MetricsRemoteWrites {
		if remoteWrite.GetLabel() == rwlc.ClusterMetricsRemoteWriteLabel {
			return ptr.To(remoteWrite.GetUrl())
		}
	}
	return nil
}

func (u *prometheusUpgrade) updateRunaiConfig(resource *runaiv1.RunaiConfig) (*runaiv1.RunaiConfig, error) {
	remoteWrites := resource.Spec.Prometheus.Spec.RemoteWrite
	resource.Spec.Prometheus.Spec.RemoteWrite = slices.Delete(remoteWrites, 0, 1)
	return resource, nil
}

func (u *prometheusUpgrade) upgradeHandlerIdentifier() string {
	return "prometheus-remote-write"
}
