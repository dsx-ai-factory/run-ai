/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package remote_write_relabel_configs

import (
	"fmt"

	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/utils/ptr"
)

const (
	ApplicationMetricsRemoteWriteLabel = "applicative_runai"
)

type ApplicationRemoteWriteRelabelConfigs struct {
	internalSpec *runv1_internal.Internal
}

func newDefaultApplicationRemoteWriteRelabelConfigs(internalSpec *runv1_internal.Internal) RemoteWriteRelabelConfigs {
	return &ApplicationRemoteWriteRelabelConfigs{
		internalSpec: internalSpec,
	}
}

func (d *ApplicationRemoteWriteRelabelConfigs) Get() []monitoringv1.RelabelConfig {
	retval := []monitoringv1.RelabelConfig{
		{
			TargetLabel: "env",
			Replacement: ptr.To(d.getEnvironment()),
			Action:      "replace",
		},
		{
			TargetLabel: "releaseName",
			Replacement: ptr.To(d.getReleaseName()),
			Action:      "replace",
		},
		{
			TargetLabel: "tenantId",
			Replacement: ptr.To(d.getTenantId()),
			Action:      "replace",
		},
		{
			TargetLabel: "clusterId",
			Replacement: ptr.To(d.getClusterId()),
			Action:      "replace",
		},
		{
			TargetLabel: "clusterVersion",
			Replacement: ptr.To(d.getClusterVersion()),
			Action:      "replace",
		},
		{
			SourceLabels: []monitoringv1.LabelName{"type"},
			Regex:        d.getRelabelConfigRegex(),
			Action:       "keep",
		},
	}
	if d.internalSpec.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift {
		// In openshift we use deduplication in the prometheus remote write configuration,
		// for this, we add the runai_replica label to the remote write
		retval = append(retval, monitoringv1.RelabelConfig{
			TargetLabel: "runai_replica",
			Replacement: ptr.To("A"),
			Action:      "replace",
		})
	}
	return retval
}

func (d *ApplicationRemoteWriteRelabelConfigs) GetRemoteWriteMatchingLabel() string {
	return ApplicationMetricsRemoteWriteLabel
}

func (d *ApplicationRemoteWriteRelabelConfigs) getRelabelConfigRegex() string {
	return "stats"
}

func (d *ApplicationRemoteWriteRelabelConfigs) getEnvironment() string {
	env := ""
	if d.internalSpec.Global.ControlPlane.Config.Environment != nil {
		env = *d.internalSpec.Global.ControlPlane.Config.Environment
	}
	return env
}

func (d *ApplicationRemoteWriteRelabelConfigs) getReleaseName() string {
	releaseName := ""
	if d.internalSpec.Global.ControlPlane.Config.ReleaseName != nil {
		releaseName = *d.internalSpec.Global.ControlPlane.Config.ReleaseName
	}
	return releaseName
}

func (d *ApplicationRemoteWriteRelabelConfigs) getTenantId() string {
	tenantId := int32(0)
	if d.internalSpec.Global.ControlPlane.Config.TenantId != nil {
		tenantId = *d.internalSpec.Global.ControlPlane.Config.TenantId
	}
	return fmt.Sprintf("%d", tenantId)
}

func (d *ApplicationRemoteWriteRelabelConfigs) getClusterId() string {
	return d.internalSpec.Global.ControlPlane.ClusterUUID

}

func (d *ApplicationRemoteWriteRelabelConfigs) getClusterVersion() string {
	return d.internalSpec.Global.Version
}
