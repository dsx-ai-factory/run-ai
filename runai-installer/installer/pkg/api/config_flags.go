/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"

// ============================================================================
// Configuration Flag Constants
// ============================================================================
// These constants define the configuration keys that clients can use
// to provide installation parameters

const (
	// FlagRegistryPullSecret is the key for the registry pull secret: a base64-encoded docker config (.dockerconfigjson)
	FlagRegistryPullSecret = inputs.FlagRegistryPullSecret
	// FlagFQDN is the key for Fully Qualified Domain Name
	// Example: "runai.example.com"
	FlagFQDN = inputs.FlagFQDN

	// FlagClusterName is the key for the cluster name
	// Example: "my-cluster"
	FlagClusterName = inputs.FlagClusterName

	// FlagIngressControllerHTTPSPort is the key for the ingress controller HTTPS port
	// Should be between 30000 and 32767 for NodePort services
	// Example: "30443"
	FlagIngressControllerHTTPSPort = inputs.FlagIngressControllerHTTPSPort

	// FlagSSLCertPath is the key for SSL certificate file path
	// Example: "/path/to/cert.pem"
	FlagSSLCertPath = inputs.FlagSSLCertPath

	// FlagSSLKeyPath is the key for SSL certificate key file path
	// Example: "/path/to/key.pem"
	FlagSSLKeyPath = inputs.FlagSSLKeyPath

	// FlagCAChainPath is the key for CA chain certificate file path
	// Example: "/path/to/ca-chain.pem"
	FlagCAChainPath = inputs.FlagCAChainPath

	// FlagWildcardFQDN is the key for wildcard FQDN for inference services
	// Example: "*.inference.runai.example.com"
	FlagWildcardFQDN = inputs.FlagWildcardFQDN

	// FlagWildcardCertPath is the key for wildcard SSL certificate file path
	// Example: "/path/to/wildcard-cert.pem"
	FlagWildcardCertPath = inputs.FlagWildcardCertPath

	// FlagWildcardCertKeyPath is the key for wildcard SSL certificate key file path
	// Example: "/path/to/wildcard-key.pem"
	FlagWildcardCertKeyPath = inputs.FlagWildcardCertKeyPath

	// FlagKubernetesDistribution is the key for Kubernetes distribution type
	// Example: "openshift", "rancher", "eks"
	FlagKubernetesDistribution = inputs.FlagKubernetesDistribution

	FlagResearcherUserEmail = inputs.FlagResearcherUserEmail

	FlagResearcherProjectName = inputs.FlagResearcherProjectName

	FlagWorkloadRouting     = inputs.FlagWorkloadRouting
	FlagHostRoutingCertPath = inputs.FlagHostRoutingCertPath
	FlagHostRoutingKeyPath  = inputs.FlagHostRoutingKeyPath

	FlagAirgapPackagePath = inputs.FlagAirgapPackagePath
	FlagRegistryURL       = inputs.FlagRegistryURL
	FlagRegistryUsername  = inputs.FlagRegistryUsername
	FlagRegistryPassword  = inputs.FlagRegistryPassword
)
