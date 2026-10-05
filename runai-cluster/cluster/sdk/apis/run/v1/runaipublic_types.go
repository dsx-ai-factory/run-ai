/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

// InferenceEndpointIntegrationV1 declares support for the v1 resolved intent
// envelope across Agent and the Anyworkload inference endpoint plugin.
const InferenceEndpointIntegrationV1 = "v1"

type RunaiPublic struct {
	Version           string               `yaml:"version,omitempty"`
	RunaiConfigStatus *RunaiConfigStatus   `yaml:"runaiConfigStatus,omitempty"`
	Dependencies      *PublicDependencies  `yaml:"dependencies,omitempty"`
	Platform          *PublicPlatform      `yaml:"platform,omitempty"`
	Configuration     *PublicConfiguration `yaml:"configuration,omitempty"`
}

type PublicDependencies struct {
	Required map[string]PublicStatus `yaml:"required,omitempty"`
	Optional map[string]PublicStatus `yaml:"optional,omitempty"`
}

type PublicStatus struct {
	Available  *bool                   `yaml:"available,omitempty"`
	Reason     string                  `yaml:"reason,omitempty"`
	Components map[string]PublicStatus `yaml:"components,omitempty"`
}

type PublicPlatform struct {
	Type        string `yaml:"type,omitempty"`
	KubeVersion string `yaml:"kubeVersion,omitempty"`
}

type PublicConfiguration struct {
	// InferenceEndpointSchemaVersion declares the installed integration contract;
	// operand readiness and administrator Gateway settings are checked separately.
	InferenceEndpointSchemaVersion         *string                     `yaml:"inferenceEndpointSchemaVersion,omitempty"`
	WorkloadOwnershipProtection            *bool                       `yaml:"workloadOwnershipProtection,omitempty"`
	SubdomainEnabled                       *bool                       `yaml:"subdomainEnabled,omitempty"`
	IngressClass                           *string                     `yaml:"ingressClass,omitempty"`
	IngressEnabled                         *bool                       `yaml:"ingressEnabled,omitempty"`
	GatewayName                            *string                     `yaml:"gatewayName,omitempty"`
	GatewayNamespace                       *string                     `yaml:"gatewayNamespace,omitempty"`
	GatewayAPIEnabled                      *bool                       `yaml:"gatewayAPIEnabled,omitempty"`
	GitSyncImage                           *string                     `yaml:"gitSyncImage,omitempty"`
	GitSyncResources                       *PublicResourceRequirements `yaml:"gitSyncResources,omitempty"`
	S3FileSystemImage                      *string                     `yaml:"s3FileSystemImage,omitempty"`
	S3FileSystemResources                  *PublicResourceRequirements `yaml:"s3FileSystemResources,omitempty"`
	FipsMode                               *string                     `yaml:"fipsMode,omitempty"`
	RunaiOauth2ProxyImage                  *string                     `yaml:"runaiOauth2ProxyImage,omitempty"`
	RunaiOauth2ProxyResources              *PublicResourceRequirements `yaml:"runaiOauth2ProxyResources,omitempty"`
	RunaiOauth2ProxyUsernameLoaderImage    *string                     `yaml:"runaiOauth2ProxyUsernameLoaderImage,omitempty"`
	UsernameLoaderResources                *PublicResourceRequirements `yaml:"usernameLoaderResources,omitempty"`
	RunaiOauth2ProxyRedisFQDN              *string                     `yaml:"runaiOauth2ProxyRedisFQDN,omitempty"`
	RunaiOauth2ProxyWorkloadAuthSecretName *string                     `yaml:"runaiOauth2ProxyWorkloadAuthSecretName,omitempty"`
	RunaiOauth2ProxySessionStoreType       *string                     `yaml:"runaiOauth2ProxySessionStoreType,omitempty"`
	RunaiOauth2ProxyOidcIssuerUrl          *string                     `yaml:"runaiOauth2ProxyOidcIssuerUrl,omitempty"`
	CustomCAGit                            *CustomCAConfig             `yaml:"customCAGit,omitempty"`
	CustomCAS3                             *CustomCAConfig             `yaml:"customCAS3,omitempty"`
	CustomCAOauth2Proxy                    *CustomCAConfig             `yaml:"customCAOauth2Proxy,omitempty"`
	ExternalAuthUrlEnabled                 *bool                       `yaml:"externalAuthUrlEnabled,omitempty"`
	OpenshiftRouteTerminationPolicy        *string                     `yaml:"openshiftRouteTerminationPolicy,omitempty"`

	// DistGPUNetworkAcceleration mirrors AnyWorkload's GPUNetworkAccelerationEnabled for distributed training.
	DistGPUNetworkAcceleration *bool `yaml:"distGPUNetworkAcceleration,omitempty"`
	// DinfGPUNetworkAcceleration mirrors AnyWorkload's GPUNetworkAccelerationEnabled for distributed inference.
	DinfGPUNetworkAcceleration *bool `yaml:"dinfGPUNetworkAcceleration,omitempty"`
}

// PublicResourceRequirements is a Kubernetes-independent representation of
// container resource requirements used by the runai-public wire format.
type PublicResourceRequirements struct {
	Limits   map[string]string `yaml:"limits,omitempty"`
	Requests map[string]string `yaml:"requests,omitempty"`
}

// CustomCAConfig flattens RunaiConfig.Spec.*.CustomCA (whose Secret is a
// k8s.io/api/core/v1.SecretKeySelector) into primitives so this YAML schema,
// the cluster-service wire, and the JSONB column avoid taking a dependency
// on the k8s API package.
type CustomCAConfig struct {
	Enabled    *bool   `yaml:"enabled,omitempty"`
	SecretName *string `yaml:"secretName,omitempty"`
	SecretKey  *string `yaml:"secretKey,omitempty"`
	Optional   *bool   `yaml:"optional,omitempty"`
}

func AvailableStatus(available bool) *bool {
	return &available
}

const (
	RunaiPublicMPI              = "mpi"
	RunaiPublicTF               = "tensorflow"
	RunaiPublicPyTorch          = "pytorch"
	RunaiPublicXGBoost          = "xgboost"
	RunaiPublicJax              = "jax"
	RunaiPublicGatewayAPI       = "gateway-api"
	RunaiPublicAdvice           = "advice"
	DistributedFrameworkSuspend = "suspend"
)
