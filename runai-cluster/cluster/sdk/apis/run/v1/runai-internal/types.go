/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package internal

import (
	"os"

	cpclusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"k8s.io/utils/ptr"
)

// +kubebuilder:object:generate=false
type ClusterFullConfig struct {
	cpclusterclient.ClusterFullConfigData `json:",inline"`
}

func (c *ClusterFullConfig) SetDefaultsWhereNeeded() {
	if c.UserConfig == nil {
		c.UserConfig = &cpclusterclient.ClusterUserConfigData{}
	}
	if c.SystemConfig == nil {
		c.SystemConfig = &cpclusterclient.ClusterSystemConfigData{}
	}
}

// Internal defines internal configuration for the Run:AI deployment
type Internal struct {
	// Global specifies global configuration for the Run:AI components
	// +kubebuilder:validation:Optional
	Global *Global `json:"global,omitempty"`

	// WorkloadController specifies the workload controller configuration
	// +kubebuilder:validation:Optional
	WorkloadController *WorkloadController `json:"workloadController,omitempty"`

	// ResearcherService specifies the researcher service configuration
	// +kubebuilder:validation:Optional
	ResearcherService *ResearcherService `json:"researcherService,omitempty"`

	// RemoteConfig specifies the desired state of the remote configuration
	RemoteConfig *ClusterFullConfig `json:"remoteConfig,omitempty"`

	// Development defines development deployments configurations
	// +kubebuilder:validation:Optional
	Development *Development `json:"development,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (g *Internal) SetDefaultsWhereNeeded() {
	if g.Global == nil {
		g.Global = &Global{}
	}
	g.Global.SetDefaultsWhereNeeded()

	if g.WorkloadController == nil {
		g.WorkloadController = &WorkloadController{}
	}
	g.WorkloadController.SetDefaultsWhereNeeded()

	if g.RemoteConfig == nil {
		g.RemoteConfig = &ClusterFullConfig{}
	}
	g.RemoteConfig.SetDefaultsWhereNeeded()

	if g.Development == nil {
		g.Development = &Development{}
	}
}

type Global struct {
	// CustomCA specifies custom CA configuration
	// +kubebuilder:validation:Optional
	CustomCA *common.CustomCA `json:"customCA,omitempty"`

	// ImageRegistry specifies custom image registry (i.e for Air-gaped usage)
	// +kubebuilder:validation:Optional
	ImageRegistry string `json:"image.registry,omitempty"`

	// ClusterURL specifies the cluster URL
	// +kubebuilder:validation:Optional
	ClusterURL string `json:"clusterURL,omitempty"`

	// Platform specifies the platform details
	// +kubebuilder:validation:Optional
	Platform *Platform `json:"platform,omitempty"`

	// ControlPlane specifies the control plane details
	// +kubebuilder:validation:Optional
	ControlPlane *ControlPlane `json:"controlPlane,omitempty"`

	// Version specifies the version of the installed Run:AI package
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`

	// AdditionalImagePullSecrets defines the container registry additional secret credentials
	// +kubebuilder:validation:Optional
	AdditionalImagePullSecrets []string `json:"additionalImagePullSecrets,omitempty"`

	// NMC specifies the configurations for NMC deployments
	// +kubebuilder:validation:Optional
	NMC NMC `json:"nmc"`

	// FipsMode specifies the Go FIPS 140-3 runtime mode for cluster components (off, on, or only).
	// Set from chart values at install/upgrade; independent of control-plane fipsMode.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=off;on;only
	FipsMode string `json:"fipsMode,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (g *Global) SetDefaultsWhereNeeded() {
	if g.CustomCA == nil {
		g.CustomCA = &common.CustomCA{}
	}
	g.CustomCA.SetDefaultsWhereNeeded()

	if g.Platform == nil {
		g.Platform = &Platform{}
	}
	g.Platform.SetDefaultsWhereNeeded()

	if g.ControlPlane == nil {
		g.ControlPlane = &ControlPlane{}
	}
	g.ControlPlane.SetDefaultsWhereNeeded()

	if g.Version == "" {
		g.Version = os.Getenv(common.DefaultVersionEnvVarName)
	}

	//only happens in tests
	if g.ImageRegistry == "" {
		g.ImageRegistry = os.Getenv(common.DefaultRegistryEnvVarName)
	}
}

type Platform struct {
	// KubernetesFlavor specifies the platform type
	// +kubebuilder:validation:Optional
	KubernetesFlavor KubernetesFlavor `json:"type,omitempty"`

	// Version specifies the platform version
	// Only supported for OpenShift at the moment
	// +kubebuilder:validation:Optional
	Version string `json:"version,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (p *Platform) SetDefaultsWhereNeeded() {}

type KubernetesFlavor string

// NOTE: Upon changing this list, make sure to update the it on the Control Plane in:
// control-plane/cluster-service/swagger/source/schemas/Clusters/ClusterReportedStatus.yaml
const (
	KubernetesFlavorVanilla   KubernetesFlavor = "Vanilla"
	KubernetesFlavorOpenShift KubernetesFlavor = "OpenShift"
	KubernetesFlavorRKE       KubernetesFlavor = "RKE"
	KubernetesFlavorGKE       KubernetesFlavor = "GKE"
	KubernetesFlavorOKE       KubernetesFlavor = "OKE"
	KubernetesFlavorAKS       KubernetesFlavor = "AKS"
	KubernetesFlavorEKS       KubernetesFlavor = "EKS"
	KubernetesFlavorUnknown   KubernetesFlavor = ""
)

func (k KubernetesFlavor) String() string {
	return string(k)
}

func ParseKubernetesFlavor(s string) KubernetesFlavor {
	return KubernetesFlavor(s)
}

type ControlPlane struct {
	// Config specifies the configuration fetched from the control plane
	// (we can't embed it directly since it implements the Marshaler interface)
	// +kubebuilder:validation:Optional
	Config *ClusterControlPlaneConfig `json:"config,omitempty"`

	// URL specifies the control plane URL
	// +kubebuilder:validation:Optional
	URL string `json:"url,omitempty"`

	// ClusterUUID specifies the cluster UUID
	// +kubebuilder:validation:Optional
	ClusterUUID string `json:"clusterUUID,omitempty"`
}

func (c *ControlPlane) SetDefaultsWhereNeeded() {
	if c.Config == nil {
		c.Config = &ClusterControlPlaneConfig{}
	}
	c.Config.SetDefaultsWhereNeeded()
}

// +kubebuilder:object:generate=false
type ClusterControlPlaneConfig struct {
	cpclusterclient.ClusterControlPlaneConfig `json:",inline"`
}

func (c *ClusterControlPlaneConfig) SetDefaultsWhereNeeded() {}

func (c *ClusterControlPlaneConfig) Sanitize() {
	c.DockerConfigJson = nil
	if c.Agent != nil && c.Agent.Client != nil {
		c.Agent.Client.Secret = nil
	}
	if c.Workloads != nil && c.Workloads.Client != nil {
		c.Workloads.Client.Secret = nil
	}
	for i := range c.MetricsRemoteWrites {
		if c.MetricsRemoteWrites[i].Credentials != nil {
			c.MetricsRemoteWrites[i].Credentials.Username = nil
			c.MetricsRemoteWrites[i].Credentials.Password = nil
		}
	}
	if c.MetricsRemoteWrite != nil && c.MetricsRemoteWrite.Credentials != nil {
		c.MetricsRemoteWrite.Credentials.Username = nil
		c.MetricsRemoteWrite.Credentials.Password = nil
	}
}

type WorkloadController struct {
}

// SetDefaultsWhereNeeded sets default for unset fields
func (w *WorkloadController) SetDefaultsWhereNeeded() {}

type ResearcherService struct {
	// Ingrees specifies the ingress configuration of the researcher-service
	// +kubebuilder:validation:Optional
	Ingress *ResearcherServiceIngress `json:"ingress,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *ResearcherService) SetDefaultsWhereNeeded() {
	if r.Ingress == nil {
		r.Ingress = &ResearcherServiceIngress{}
	}
	r.Ingress.SetDefaultsWhereNeeded()
}

type ResearcherServiceIngress struct {
	// Enabled specifies whether the ingress is enabled
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (r *ResearcherServiceIngress) SetDefaultsWhereNeeded() {
	if r.Enabled == nil {
		r.Enabled = ptr.To[bool](true)
	}
}

// Development defines development deployments configurations
type Development struct {
	// Headless defines whether this deployment should be headless - meaning no control-plane communication
	Headless bool `json:"headless,omitempty"`

	// Skip dependency checks and assume they are fulfilled
	SkipDependencyChecks bool `json:"skipDependencyChecks,omitempty"`
}

type NMC struct {
	// RemoteAPIServerURL specifies the API server URL of the remote cluster in an NMC deployment
	// +kubebuilder:validation:Optional
	RemoteAPIServerURL string `json:"remoteAPIServerURL,omitempty"`

	// ServiceAccountName specifies the service account name for the NMC deployment
	// +kubebuilder:validation:Optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// AllowedServiceAccounts specifies additional service accounts allowed to perform privileged operations.
	// CR webhooks match full usernames exactly; bare names refer to service accounts in the installation namespace.
	// Use system:serviceaccount:<namespace>:<name> for accounts in other namespaces.
	// Secret/PVC admission policies retain their legacy substring matching for these entries.
	// +kubebuilder:validation:Optional
	AllowedServiceAccounts []string `json:"allowedServiceAccounts,omitempty"`

	MgmtPrometheusIngress *common.Ingress `json:"mgmtPrometheusIngress,omitempty"`

	// MgmtPrometheusGatewayAPI specifies a per-operand Gateway reference override for NMC management prometheus.
	// When set, this takes precedence over the global gatewayAPI name/namespace.
	// +kubebuilder:validation:Optional
	MgmtPrometheusGatewayAPI *common.GatewayRef `json:"mgmtPrometheusGatewayAPI,omitempty"`
}

func (n *NMC) IsEnabled() bool {
	return n.RemoteAPIServerURL != ""
}

func (n *NMC) SetDefaultsWhereNeeded() {
	if !n.IsEnabled() {
		return
	}
	if n.MgmtPrometheusIngress == nil {
		n.MgmtPrometheusIngress = &common.Ingress{}
	}
	n.MgmtPrometheusIngress.SetDefaultsWhereNeeded()
}
