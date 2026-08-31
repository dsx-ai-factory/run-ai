/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputs

import (
	"sync"

	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
)

type InstallInputType int

const (
	InstallInputJfrogCreds InstallInputType = iota
	InstallInputFQDN
	InstallInputClusterName
	// the port where the ingress controller nodesPort will listen on for https requests
	InstallInputIngressControllerHTTPSPort
	InstallInputSSLPath
	InstallInputSSLKeyPath
	InstallInputCAChainPath
	InstallInputWildcardFQDN
	InstallInputWildcardCertPath
	InstallInputWildcardCertKeyPath
	InstallInputKubernetesDistribution
	// User email for Run:AI assets creation (local user)
	InstallInputResearcherUserEmail
	// Project name for Run:AI assets creation
	InstallInputReasercherProjectName
	// Workload routing type: path (default) or host
	InstallInputWorkloadRouting
	InstallInputHostRoutingCertPath
	InstallInputHostRoutingKeyPath
	InstallInputAirgapPackagePath
	InstallInputRegistryURL
	InstallInputRegistryUsername
	InstallInputRegistryPassword

	FlagJfrogCreds                 = "jfrog_creds"
	FlagFQDN                       = "fqdn"
	FlagClusterName                = "cluster_name"
	FlagIngressControllerHTTPSPort = "ingress_controller_https_port"
	FlagSSLCertPath                = "ssl_cert_path"
	FlagSSLKeyPath                 = "ssl_key_path"
	FlagCAChainPath                = "ca_chain_path"
	FlagWildcardFQDN               = "wildcard_fqdn"
	FlagWildcardCertPath           = "wildcard_cert_path"
	FlagWildcardCertKeyPath        = "wildcard_cert_key_path"
	FlagKubernetesDistribution     = "kubernetes_distribution"
	FlagResearcherUserEmail        = "researcher_user_email"
	FlagResearcherProjectName      = "researcher_project_name"
	FlagWorkloadRouting            = "workload_routing"
	FlagHostRoutingCertPath        = "host_routing_cert_path"
	FlagHostRoutingKeyPath         = "host_routing_key_path"
	FlagAirgapPackagePath          = "airgap_package_path"
	FlagRegistryURL                = "registry_url"
	FlagRegistryUsername           = "registry_username"
	FlagRegistryPassword           = "registry_password"
)

// FlagToInstallInput provides a reverse lookup from flag string to InstallInputType.
var (
	FlagToInstallInput = map[string]InstallInputType{
		FlagJfrogCreds:                 InstallInputJfrogCreds,
		FlagFQDN:                       InstallInputFQDN,
		FlagClusterName:                InstallInputClusterName,
		FlagIngressControllerHTTPSPort: InstallInputIngressControllerHTTPSPort,
		FlagSSLCertPath:                InstallInputSSLPath,
		FlagSSLKeyPath:                 InstallInputSSLKeyPath,
		FlagCAChainPath:                InstallInputCAChainPath,
		FlagWildcardFQDN:               InstallInputWildcardFQDN,
		FlagWildcardCertPath:           InstallInputWildcardCertPath,
		FlagWildcardCertKeyPath:        InstallInputWildcardCertKeyPath,
		FlagKubernetesDistribution:     InstallInputKubernetesDistribution,
		FlagResearcherUserEmail:        InstallInputResearcherUserEmail,
		FlagResearcherProjectName:      InstallInputReasercherProjectName,
		FlagWorkloadRouting:            InstallInputWorkloadRouting,
		FlagHostRoutingCertPath:        InstallInputHostRoutingCertPath,
		FlagHostRoutingKeyPath:         InstallInputHostRoutingKeyPath,
		FlagAirgapPackagePath:          InstallInputAirgapPackagePath,
		FlagRegistryURL:                InstallInputRegistryURL,
		FlagRegistryUsername:           InstallInputRegistryUsername,
		FlagRegistryPassword:           InstallInputRegistryPassword,
	}

	configInputs     map[InstallInputType]*ConfigurationItem
	configInputsOnce sync.Once
)

func GetConfigInputs() map[InstallInputType]*ConfigurationItem {
	configInputsOnce.Do(func() {
		configInputs = map[InstallInputType]*ConfigurationItem{
			InstallInputJfrogCreds:  NewConfigurationItem("JFROG credentials (base64)", "Enter your NGC credentials...", InstallInputJfrogCreds, validation.ValidateJfrogCredentials),
			InstallInputFQDN:        NewConfigurationItem("Fully Qualified Domain Name (FQDN)", "e.g., runai.example.com", InstallInputFQDN, validation.ValidateFQDN),
			InstallInputClusterName: NewConfigurationItem("Cluster Name", "e.g., my-cluster", InstallInputClusterName, validation.ValidateClusterName),
			InstallInputIngressControllerHTTPSPort: NewConfigurationItem(
				"Ingress Controller HTTPS port ",
				"Leaving empty will let k8s pick the port. should be between 30000 and 32767. e.g., 30443",
				InstallInputIngressControllerHTTPSPort,
				validation.ValidateIngressControllerHttpsPort,
			),
			InstallInputSSLPath:                NewConfigurationItem("SSL cert path", "/path/to/cert.pem", InstallInputSSLPath, validation.ValidateSSLPaths),
			InstallInputSSLKeyPath:             NewConfigurationItem("SSL key path", "/path/to/key.pem", InstallInputSSLKeyPath, validation.ValidateSSLPaths),
			InstallInputCAChainPath:            NewConfigurationItem("CA Chain cert path", "/path/to/ca-chain.pem", InstallInputCAChainPath, validation.ValidateSSLPaths),
			InstallInputWildcardFQDN:           NewConfigurationItem("Wildcard FQDN", "*.inference.runai.example.com", InstallInputWildcardFQDN, validation.ValidateWildcardFQDN),
			InstallInputWildcardCertPath:       NewConfigurationItem("Inference SSL cert path", "/path/to/cert.pem", InstallInputWildcardCertPath, validation.ValidateSSLPaths),
			InstallInputWildcardCertKeyPath:    NewConfigurationItem("Inference SSL key path", "/path/to/key.pem", InstallInputWildcardCertKeyPath, validation.ValidateSSLPaths),
			InstallInputKubernetesDistribution: NewConfigurationItem("Kubernetes Distribution", "e.g., openshift, rancher, eks", InstallInputKubernetesDistribution, validation.ValidateKubernetesDistribution),
			InstallInputResearcherUserEmail:    NewConfigurationItem("User Email", "user@example.com", InstallInputResearcherUserEmail, validation.ValidateEmail),
			InstallInputReasercherProjectName:  NewConfigurationItem("Project Name", "e.g., test-project", InstallInputReasercherProjectName, validation.ValidateProjectName),
			InstallInputWorkloadRouting:        NewConfigurationItem("Workload Routing", "path or host (default: path)", InstallInputWorkloadRouting, validation.ValidateWorkloadRouting),
			InstallInputHostRoutingCertPath:    NewConfigurationItem("Host routing SSL cert path", "/path/to/wildcard-cert.pem", InstallInputHostRoutingCertPath, validation.ValidateSSLPaths),
			InstallInputHostRoutingKeyPath:     NewConfigurationItem("Host routing SSL key path", "/path/to/wildcard-key.pem", InstallInputHostRoutingKeyPath, validation.ValidateSSLPaths),
			InstallInputAirgapPackagePath:      NewConfigurationItem("Airgap package path", "/path/to/runai-airgapped-package.tar.gz", InstallInputAirgapPackagePath, validation.ValidateNonEmpty),
			InstallInputRegistryURL:            NewConfigurationItem("Registry URL", "registry.example.com:5000", InstallInputRegistryURL, validation.ValidateNonEmpty),
			InstallInputRegistryUsername:        NewConfigurationItem("Registry username", "admin", InstallInputRegistryUsername, validation.ValidateNonEmpty),
			InstallInputRegistryPassword:        NewConfigurationItem("Registry password", "", InstallInputRegistryPassword, validation.ValidateNonEmpty),
		}
	})
	return configInputs
}
