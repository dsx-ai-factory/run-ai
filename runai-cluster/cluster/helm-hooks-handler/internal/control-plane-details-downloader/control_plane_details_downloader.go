/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package control_plane_details_downloader

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"

	cpclusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	platform_identifier "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/platform-identifier"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	rwlc "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs"
	runaiv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	defaultControlPlaneClientID         = "runai-internal.installer"
	installationConfigurationSecretName = "installation-configuration"
	installationConfigurationKey        = installationConfigurationSecretName
)

func DownloadInstallationConfiguration(k8s client.Client, hookParams *hooks.ClusterHookParams, customCA *common.CustomCA) (*InstallationConfigurationData, error) {
	platform, err := platform_identifier.Identify(context.Background(), k8s)
	if err != nil {
		return nil, fmt.Errorf("failed to identify kubernetes flavor: %w", err)
	}

	// We have a default client-id configured in the control-plane, but we still provide the option to use
	// a different one if there is a need for it
	cpDetails := &hookParams.ControlPlaneDetails
	cpDetails.ClientId = getClientIDOrDefault(hookParams.ClientId)

	cpClient, err := clients.NewAuthenticatedClusterClient(cpDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to create control-plane client: %w", err)
	}

	cpConfig, err := cpClient.GetClusterControlPlaneConfig(context.Background(),
		hookParams.ClusterDetails.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch control-plane details: %w", err)
	}

	if cpConfig == nil {
		return nil, fmt.Errorf("control-plane details are empty")
	}

	// Filter out infrastructure metrics remote write endpoint if requested
	if hookParams.DisableInfrastructureMetricsRemoteWrite {
		cpConfig.MetricsRemoteWrites = filterOutApplicativeRunaiRemoteWrite(cpConfig.MetricsRemoteWrites)
	}
	additionalImagePullSecrets := appendControlPlaneImagePullSecretIfNeeded(
		hookParams.AdditionalImagePullSecrets,
		cpConfig.GetDockerConfigJson(),
	)

	isLocalCluster, err := isLocalCluster(hookParams.ClusterDetails.URL, cpDetails.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to determine if cluster is local to the control-plane: %w", err)
	}

	installationConfiguration := &InstallationConfigurationData{
		Internal: runaiv1_internal.Internal{
			Global: &runaiv1_internal.Global{
				CustomCA:                   customCA,
				ImageRegistry:              hookParams.ImageRegistry,
				AdditionalImagePullSecrets: additionalImagePullSecrets,
				ClusterURL:                 hookParams.ClusterDetails.URL,
				Platform:                   platform,
				ControlPlane: &runaiv1_internal.ControlPlane{
					Config: &runaiv1_internal.ClusterControlPlaneConfig{
						ClusterControlPlaneConfig: *cpConfig,
					},
					URL:         cpDetails.URL,
					ClusterUUID: hookParams.ClusterDetails.UID,
				},
				NMC: runaiv1_internal.NMC{
					RemoteAPIServerURL:     hookParams.NMC.RemoteServerAPIServerURL,
					ServiceAccountName:     hookParams.NMC.ServiceAccountName,
					AllowedServiceAccounts: hookParams.NMC.AllowedServiceAccounts,
					MgmtPrometheusIngress:  hookParams.NMC.MgmtPrometheusIngress,
				},
			},
			WorkloadController: &runaiv1_internal.WorkloadController{},
			ResearcherService: &runaiv1_internal.ResearcherService{
				Ingress: &runaiv1_internal.ResearcherServiceIngress{
					Enabled: ptr.To(!isLocalCluster),
				},
			},
			Development: &runaiv1_internal.Development{
				Headless:             hookParams.Development.Headless,
				SkipDependencyChecks: hookParams.Development.SkipDependencyChecks,
			},
		},
	}

	return installationConfiguration, nil
}

func CreateOrReplaceInstallationConfigurationSecret(k8s client.Client,
	installationConfiguration *InstallationConfigurationData) error {

	installationConfigurationYAML, err := yaml.Marshal(installationConfiguration)
	if err != nil {
		return fmt.Errorf("failed to marshal installation configuration: %w", err)
	}

	installationConfigurationSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      installationConfigurationSecretName,
			Namespace: runaiv1.DefaultNamespace,
		},
		Data: map[string][]byte{
			installationConfigurationKey: installationConfigurationYAML,
		},
	}

	log.Printf("attempting to replace secret [%s]\n", installationConfigurationSecretName)

	err = k8s.Delete(context.Background(), installationConfigurationSecret)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete secret [%s]: %w", installationConfigurationSecretName, err)
	}

	err = k8s.Create(context.Background(), installationConfigurationSecret)
	if err != nil {
		return fmt.Errorf("failed to create secret [%s]: %w", installationConfigurationSecretName, err)
	}

	log.Printf("successfully replaced secret [%s]\n", installationConfigurationSecretName)

	return nil
}

func GetInstallationConfiguration(k8s client.Client) (*InstallationConfigurationData, error) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      installationConfigurationSecretName,
			Namespace: runaiv1.DefaultNamespace,
		},
	}

	err := k8s.Get(context.Background(), client.ObjectKeyFromObject(secret), secret)
	if err != nil {
		return nil, err
	}

	data := &InstallationConfigurationData{}

	err = yaml.Unmarshal(secret.Data[installationConfigurationKey], data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func InstallationConfigurationSecretExists(k8s client.Client) (bool, error) {
	_, err := GetInstallationConfiguration(k8s)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func getClientIDOrDefault(clientID string) string {
	if clientID == "" {
		return defaultControlPlaneClientID
	}

	return clientID
}

func appendControlPlaneImagePullSecretIfNeeded(additionalImagePullSecrets []string, dockerConfigJSON string) []string {
	if strings.TrimSpace(dockerConfigJSON) == "" ||
		containsString(additionalImagePullSecrets, hooks.RunAIControlPlaneImagePullSecretName) {
		return additionalImagePullSecrets
	}

	return append(additionalImagePullSecrets, hooks.RunAIControlPlaneImagePullSecretName)
}

func containsString(values []string, value string) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}

	return false
}

func isLocalCluster(clusterUrlStr, controlPlaneUrlStr string) (bool, error) {
	clusterUrl, err := url.Parse(clusterUrlStr)
	if err != nil {
		return false, fmt.Errorf("failed parsing cluster URL: %w", err)
	}
	controlPlaneUrl, err := url.Parse(controlPlaneUrlStr)
	if err != nil {
		return false, fmt.Errorf("failed parsing control plane URL: %w", err)
	}

	return clusterUrl.Hostname() == controlPlaneUrl.Hostname(), nil
}

// filterOutApplicativeRunaiRemoteWrite removes the "applicative_runai" remote write endpoint from the configuration
func filterOutApplicativeRunaiRemoteWrite(remoteWrites []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite) []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite {
	filtered := make([]cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite, 0, len(remoteWrites))
	for _, rw := range remoteWrites {
		if rw.Label != nil && *rw.Label == rwlc.ApplicationMetricsRemoteWriteLabel {
			log.Printf("Filtering out infrastructure metrics remote write endpoint with label '%s'", rwlc.ApplicationMetricsRemoteWriteLabel)
			continue
		}
		filtered = append(filtered, rw)
	}
	return filtered
}
