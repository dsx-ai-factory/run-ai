/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	sdkglobal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	unreadyThreshold = 1 * time.Minute
)

type ClusterInfoParser struct{}

func NewClusterInfoParser() *ClusterInfoParser {
	return &ClusterInfoParser{}
}

func (parser *ClusterInfoParser) ParseClusterInfoToServerScheme(clusterInfoCm *corev1.ConfigMap) (swagger.ClusterInfoSyncRequest, error) {
	var runaiPublic runv1.RunaiPublic
	err := yaml.Unmarshal([]byte(clusterInfoCm.Data[constants.ClusterInfoConfigMapName]), &runaiPublic)
	if err != nil {
		return swagger.ClusterInfoSyncRequest{}, fmt.Errorf("failed to unmarshal runai-public configmap: %w", err)
	}

	var clusterInfo swagger.ClusterInfoSyncRequest
	clusterInfo.Domain = config.Config().ClusterDomain
	clusterInfo.Version = runaiPublic.Version

	clusterInfoStatus, err := parseClusterInfoStatusToServerScheme(&runaiPublic)
	if err != nil {
		return swagger.ClusterInfoSyncRequest{}, fmt.Errorf("failed to parse cluster info status to server scheme: %w", err)
	}

	if clusterInfoStatus.Dependencies == nil &&
		clusterInfoStatus.Operands == nil &&
		len(clusterInfoStatus.Conditions) == 0 &&
		!clusterInfoStatus.Platform.IsSet() &&
		!clusterInfoStatus.Config.IsSet() {
		return swagger.ClusterInfoSyncRequest{}, fmt.Errorf("empty cluster info, dont send update to the CP")
	}

	clusterInfo.Status = *swagger.NewNullableClusterReportedStatus(clusterInfoStatus)

	return clusterInfo, nil
}

func parseClusterInfoStatusToServerScheme(runaiPublic *runv1.RunaiPublic) (*swagger.ClusterReportedStatus, error) {
	if runaiPublic == nil {
		return nil, fmt.Errorf("runai-public configmap is empty")
	}

	var status swagger.ClusterReportedStatus

	status.Dependencies = parseDependenciesToServerScheme(runaiPublic.Dependencies)
	status.Operands = parseOperandStatusesToServerScheme(runaiPublic.RunaiConfigStatus.OperandStatuses)
	status.Conditions = parseConditionsToServerScheme(runaiPublic.RunaiConfigStatus.Conditions)
	platformStatus := parseClusterInfoPlatformToServerScheme(runaiPublic.Platform)
	if platformStatus != nil {
		status.Platform = *swagger.NewNullableClusterReportedStatusPlatform(platformStatus)
	}

	if err := validateWorkloadInjectedContainerResources(runaiPublic.Configuration); err != nil {
		return nil, fmt.Errorf("invalid workload-injected container resources: %w", err)
	}
	configStatus := parseClusterInfoConfigToServerScheme(runaiPublic.Configuration)
	if configStatus != nil {
		status.Config = *swagger.NewNullableClusterReportedConfig(configStatus)
	}

	return &status, nil
}

func validateWorkloadInjectedContainerResources(configuration *runv1.PublicConfiguration) error {
	if configuration == nil {
		return nil
	}
	configurations := []struct {
		name      string
		resources *runv1.PublicResourceRequirements
	}{
		{name: "runaiOauth2ProxyResources", resources: configuration.RunaiOauth2ProxyResources},
		{name: "usernameLoaderResources", resources: configuration.UsernameLoaderResources},
		{name: "gitSyncResources", resources: configuration.GitSyncResources},
		{name: "s3FileSystemResources", resources: configuration.S3FileSystemResources},
	}
	for _, configured := range configurations {
		if configured.resources == nil {
			continue
		}
		if _, err := sdkglobal.ResourceRequirementsFromStringMaps(
			configured.resources.Limits,
			configured.resources.Requests,
		); err != nil {
			return fmt.Errorf("%s: %w", configured.name, err)
		}
	}
	return nil
}

func parseConditionsToServerScheme(conditions []metav1.Condition) []swagger.ClusterReportedStatusConditionsInner {
	conditionsServerScheme := []swagger.ClusterReportedStatusConditionsInner{}

	for _, condition := range conditions {
		conditionsServerScheme = append(conditionsServerScheme, swagger.ClusterReportedStatusConditionsInner{
			LastTransitionTime: condition.LastTransitionTime.Time,
			Message:            condition.Message,
			ObservedGeneration: &condition.ObservedGeneration,
			Reason:             condition.Reason,
			Status:             string(condition.Status),
			Type:               condition.Type,
		})
	}

	return conditionsServerScheme
}

func parseOperandStatusesToServerScheme(operandStatuses map[string]runv1.TimedOperandStatus) *map[string]swagger.ClusterReportedStatusOperandsValue {
	if operandStatuses == nil {
		return nil
	}

	operandStatusesServerScheme := map[string]swagger.ClusterReportedStatusOperandsValue{}

	for operandName, operandStatus := range operandStatuses {
		unreadyThresholdCrossed := !operandStatus.Ready && (time.Since(operandStatus.LastTransitionTime.Time) > unreadyThreshold)

		operandStatusesServerScheme[operandName] = swagger.ClusterReportedStatusOperandsValue{
			Ready:                   operandStatus.Ready,
			Reasons:                 operandStatus.Reasons,
			LastTransitionTime:      &operandStatus.LastTransitionTime.Time,
			UnreadyThresholdCrossed: &unreadyThresholdCrossed,
		}
	}

	return &operandStatusesServerScheme
}

func parseDependenciesToServerScheme(dependencies *runv1.PublicDependencies) *swagger.ClusterDependenciesStatus {
	if dependencies == nil || (len(dependencies.Required) == 0 && len(dependencies.Optional) == 0) {
		return nil
	}

	clusterStatusDependencies := swagger.ClusterDependenciesStatus{
		Required: parseDependenciesStatusToServerScheme(dependencies.Required),
		Optional: parseDependenciesStatusToServerScheme(dependencies.Optional),
	}

	return &clusterStatusDependencies
}

func parseDependenciesStatusToServerScheme(dependenciesStatus map[string]runv1.PublicStatus) map[string]swagger.ClusterDependencyStatus {
	dependenciesStatusServerScheme := map[string]swagger.ClusterDependencyStatus{}

	for dependencyName, dependencyStatus := range dependenciesStatus {
		dependenciesStatusServerScheme[dependencyName] = swagger.ClusterDependencyStatus{
			Available:  dependencyStatus.Available != nil && *dependencyStatus.Available,
			Reason:     &dependencyStatus.Reason,
			Components: ptr.To(parseDependenciesStatusToServerScheme(dependencyStatus.Components)),
		}
	}

	return dependenciesStatusServerScheme
}

func parseClusterInfoPlatformToServerScheme(platform *runv1.PublicPlatform) *swagger.ClusterReportedStatusPlatform {
	if platform == nil {
		return nil
	}

	platformServerSchemeType := strings.ToLower(platform.Type)
	platformServerScheme := swagger.ClusterReportedStatusPlatform{
		Type:        *swagger.NewNullableString(&platformServerSchemeType),
		KubeVersion: *swagger.NewNullableString(&platform.KubeVersion),
	}

	return &platformServerScheme
}

func parseClusterInfoConfigToServerScheme(configuration *runv1.PublicConfiguration) *swagger.ClusterReportedConfig {
	if configuration == nil {
		return nil
	}

	workloadOwnershipProtection := configuration.WorkloadOwnershipProtection != nil && *configuration.WorkloadOwnershipProtection
	var openshiftRouteTerminationPolicy *swagger.OpenshiftRouteTerminationPolicy
	if configuration.OpenshiftRouteTerminationPolicy != nil {
		policy := swagger.OpenshiftRouteTerminationPolicy(*configuration.OpenshiftRouteTerminationPolicy)
		openshiftRouteTerminationPolicy = &policy
	}
	configServerScheme := swagger.ClusterReportedConfig{
		InferenceEndpointSchemaVersion:         *swagger.NewNullableString(configuration.InferenceEndpointSchemaVersion),
		WorkloadOwnershipProtection:            *swagger.NewNullableBool(&workloadOwnershipProtection),
		SubdomainEnabled:                       *swagger.NewNullableBool(configuration.SubdomainEnabled),
		IngressClass:                           *swagger.NewNullableString(configuration.IngressClass),
		IngressEnabled:                         *swagger.NewNullableBool(configuration.IngressEnabled),
		GatewayName:                            *swagger.NewNullableString(configuration.GatewayName),
		GatewayNamespace:                       *swagger.NewNullableString(configuration.GatewayNamespace),
		GatewayAPIEnabled:                      *swagger.NewNullableBool(configuration.GatewayAPIEnabled),
		GitSyncImage:                           *swagger.NewNullableString(configuration.GitSyncImage),
		S3FileSystemImage:                      *swagger.NewNullableString(configuration.S3FileSystemImage),
		RunaiOauth2ProxyImage:                  *swagger.NewNullableString(configuration.RunaiOauth2ProxyImage),
		RunaiOauth2ProxyResources:              *swagger.NewNullableClusterReportedResourceRequirements(parseResourceRequirementsToServerScheme(configuration.RunaiOauth2ProxyResources)),
		UsernameLoaderResources:                *swagger.NewNullableClusterReportedResourceRequirements(parseResourceRequirementsToServerScheme(configuration.UsernameLoaderResources)),
		GitSyncResources:                       *swagger.NewNullableClusterReportedResourceRequirements(parseResourceRequirementsToServerScheme(configuration.GitSyncResources)),
		S3FileSystemResources:                  *swagger.NewNullableClusterReportedResourceRequirements(parseResourceRequirementsToServerScheme(configuration.S3FileSystemResources)),
		RunaiOauth2ProxyUsernameLoaderImage:    *swagger.NewNullableString(configuration.RunaiOauth2ProxyUsernameLoaderImage),
		RunaiOauth2ProxyRedisFQDN:              *swagger.NewNullableString(configuration.RunaiOauth2ProxyRedisFQDN),
		RunaiOauth2ProxyWorkloadAuthSecretName: *swagger.NewNullableString(configuration.RunaiOauth2ProxyWorkloadAuthSecretName),
		RunaiOauth2ProxySessionStoreType:       *swagger.NewNullableString(configuration.RunaiOauth2ProxySessionStoreType),
		RunaiOauth2ProxyOidcIssuerUrl:          *swagger.NewNullableString(configuration.RunaiOauth2ProxyOidcIssuerUrl),
		ExternalAuthUrlEnabled:                 *swagger.NewNullableBool(configuration.ExternalAuthUrlEnabled),
		OpenshiftRouteTerminationPolicy:        *swagger.NewNullableOpenshiftRouteTerminationPolicy(openshiftRouteTerminationPolicy),
		DistGPUNetworkAcceleration:             *swagger.NewNullableBool(configuration.DistGPUNetworkAcceleration),
		DinfGPUNetworkAcceleration:             *swagger.NewNullableBool(configuration.DinfGPUNetworkAcceleration),
	}
	if configuration.FipsMode != nil {
		configServerScheme.SetFipsMode(swagger.FipsModeEnum(*configuration.FipsMode))
	}
	if ca := parseCustomCAConfigToServerScheme(configuration.CustomCAGit); ca != nil {
		configServerScheme.CustomCAGit = *swagger.NewNullableClusterReportedCustomCA(ca)
	}
	if ca := parseCustomCAConfigToServerScheme(configuration.CustomCAS3); ca != nil {
		configServerScheme.CustomCAS3 = *swagger.NewNullableClusterReportedCustomCA(ca)
	}
	if ca := parseCustomCAConfigToServerScheme(configuration.CustomCAOauth2Proxy); ca != nil {
		configServerScheme.CustomCAOauth2Proxy = *swagger.NewNullableClusterReportedCustomCA(ca)
	}

	return &configServerScheme
}

func parseResourceRequirementsToServerScheme(resources *runv1.PublicResourceRequirements) *swagger.ClusterReportedResourceRequirements {
	if resources == nil {
		return nil
	}
	result := &swagger.ClusterReportedResourceRequirements{}
	if resources.Limits != nil {
		result.SetLimits(resources.Limits)
	}
	if resources.Requests != nil {
		result.SetRequests(resources.Requests)
	}
	return result
}

func parseCustomCAConfigToServerScheme(ca *runv1.CustomCAConfig) *swagger.ClusterReportedCustomCA {
	if ca == nil {
		return nil
	}
	return &swagger.ClusterReportedCustomCA{
		Enabled:    *swagger.NewNullableBool(ca.Enabled),
		SecretName: *swagger.NewNullableString(ca.SecretName),
		SecretKey:  *swagger.NewNullableString(ca.SecretKey),
		Optional:   *swagger.NewNullableBool(ca.Optional),
	}
}
