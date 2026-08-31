/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8s_objects_tracker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
	k8sObjectsTracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	k8sObjectsMock "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/mocks"
)

// SimpleMockClient supports fixed set of pvcs and possible error
type SimpleMockClient struct {
	k8sObjectsMock.K8SObjectTrackerMock

	PvcResponse            *k8sObjectsTracker.PVCListResponse
	SecretsResponse        *k8sObjectsTracker.SecretListResponse
	ConfigMapsResponse     *k8sObjectsTracker.ConfigMapsListResponse
	ClusterObjectsResponse *k8sObjectsTracker.ClusterObjectsStatus
	Error                  *k8sObjectsTracker.Error
}

func NewSimpleMockClient() SimpleMockClient {
	return SimpleMockClient{
		PvcResponse: &k8sObjectsTracker.PVCListResponse{
			Pvcs: make([]k8sObjectsTracker.PVCListInfo, 0),
		},
		SecretsResponse: &k8sObjectsTracker.SecretListResponse{
			Secrets: make([]k8sObjectsTracker.SecretListInfo, 0),
		},
		ConfigMapsResponse: &k8sObjectsTracker.ConfigMapsListResponse{
			ConfigMaps: make([]k8sObjectsTracker.ConfigMapListInfo, 0),
		},
		ClusterObjectsResponse: &k8sObjectsTracker.ClusterObjectsStatus{
			Entries: make([]k8sObjectsTracker.ClusterObjectStatus, 0),
		},
		K8SObjectTrackerMock: k8sObjectsMock.NewK8SObjectTrackerMock(),
	}
}

func (s SimpleMockClient) WithAsset(clusterId string, assetMeta *api.AssetMeta, assetPayload []byte,
	issues *[]k8sObjectsTracker.ClusterObjectIssue,
) SimpleMockClient {
	switch assetMeta.Kind {
	case api.ASSETKIND_PVC:
		var asset api.PvcCreationRequest
		json.Unmarshal(assetPayload, &asset)
		return s.WithPvc(clusterId, &asset, issues)
	case api.ASSETKIND_PASSWORD:
		var asset api.PasswordCreationRequest
		json.Unmarshal(assetPayload, &asset)
		return s.WithPassword(clusterId, &asset, issues)
	case api.ASSETKIND_ACCESS_KEY:
		var asset api.AccessKeyCreationRequest
		json.Unmarshal(assetPayload, &asset)
		return s.WithAccessKey(clusterId, &asset, issues)
	case api.ASSETKIND_CONFIG_MAP:
		var asset api.ConfigMapCreationRequest
		json.Unmarshal(assetPayload, &asset)
		return s.WithConfigMap(clusterId, &asset, issues)
	case api.ASSETKIND_GENERIC_SECRET:
		var asset api.GenericSecretCreationRequest
		json.Unmarshal(assetPayload, &asset)
		return s.WithGenericSecret(clusterId, &asset, issues)
	case api.ASSETKIND_DOCKER_REGISTRY:
		var asset api.DockerRegistryCreationRequest
		json.Unmarshal(assetPayload, &asset)
		return s.WithDockerRegistry(clusterId, &asset, issues)
	default:
		panic("unsupported asset kind for simple mock client of k8sObjectsTracker")
	}
}

func (s SimpleMockClient) WithPvc(clusterId string, pvcAsset *api.PvcCreationRequest, issues *[]k8sObjectsTracker.ClusterObjectIssue) SimpleMockClient {
	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(pvcAsset.Meta.Scope), pvcAsset.Meta.ProjectId, pvcAsset.Meta.DepartmentId, pvcAsset.Meta.ClusterId, nil)
	if scopeInfo.Scope == scope.Tenant {
		scopeInfo.Scope = scope.Cluster // in cluster tenant scope resources are stored as cluster wide
	}

	pvc := k8sObjectsTracker.PVCListInfo{
		ClusterId: clusterId,
		ScopeType: string(scopeInfo.Scope),
		ScopeId:   scopeInfo.ScopeId,
		Name:      convutils.Deref(pvcAsset.Spec.ClaimName, ""),
	}

	pvcClusterResourceStatus := k8sObjectsTracker.ClusterObjectStatus{
		ResourceType: k8sObjectsTracker.CLUSTERRESOURCETYPE_PVC,
		ScopeType:    string(scopeInfo.Scope),
		ScopeId:      scopeInfo.ScopeId,
		ResourceName: convutils.Deref(pvcAsset.Spec.ClaimName, ""),
	}

	if issues != nil {
		pvcClusterResourceStatus.Issues = *issues
	}

	var origin k8sObjectsTracker.Origin

	if convutils.BoolIsFalse(pvcAsset.Spec.ExistingPvc) {
		origin = k8sObjectsTracker.ORIGIN_ASSETS
	} else {
		origin = k8sObjectsTracker.ORIGIN_EXISTING
	}

	pvc.Origin.Set(&origin)

	s.PvcResponse.Pvcs = append(s.PvcResponse.Pvcs, pvc)
	s.ClusterObjectsResponse.Entries = append(s.ClusterObjectsResponse.Entries, pvcClusterResourceStatus)
	return s
}

func (s SimpleMockClient) WithConfigMap(clusterId string, configMapAsset *api.ConfigMapCreationRequest, issues *[]k8sObjectsTracker.ClusterObjectIssue) SimpleMockClient {
	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(configMapAsset.Meta.Scope), configMapAsset.Meta.ProjectId, configMapAsset.Meta.DepartmentId, configMapAsset.Meta.ClusterId, nil)
	if scopeInfo.Scope == scope.Tenant {
		scopeInfo.Scope = scope.Cluster // in cluster tenant scope resources are stored as cluster wide
	}

	configMap := k8sObjectsTracker.ConfigMapListInfo{
		ClusterId: clusterId,
		ScopeType: string(scopeInfo.Scope),
		ScopeId:   scopeInfo.ScopeId,
		Name:      *configMapAsset.Spec.ConfigMap,
	}

	configMapClusterResourceStatus := k8sObjectsTracker.ClusterObjectStatus{
		ResourceType: k8sObjectsTracker.CLUSTERRESOURCETYPE_CONFIG_MAP,
		ScopeType:    string(scopeInfo.Scope),
		ScopeId:      scopeInfo.ScopeId,
		ResourceName: *configMapAsset.Spec.ConfigMap,
	}
	if issues != nil {
		configMapClusterResourceStatus.Issues = *issues
	}

	s.ConfigMapsResponse.ConfigMaps = append(s.ConfigMapsResponse.ConfigMaps, configMap)
	s.ClusterObjectsResponse.Entries = append(s.ClusterObjectsResponse.Entries, configMapClusterResourceStatus)
	return s
}

func (s SimpleMockClient) WithPassword(clusterId string, password *api.PasswordCreationRequest, issues *[]k8sObjectsTracker.ClusterObjectIssue) SimpleMockClient {
	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(password.Meta.Scope), password.Meta.ProjectId, password.Meta.DepartmentId, password.Meta.ClusterId, nil)
	if scopeInfo.Scope == scope.Tenant {
		scopeInfo.Scope = scope.Cluster // in cluster tenant scope resources are stored as cluster wide
	}

	origin, name := s.secretOriginAndName(api.ASSETKIND_PASSWORD, password.Meta.Name, password.Spec.ExistingSecretName)

	secret := k8sObjectsTracker.SecretListInfo{
		ClusterId: clusterId,
		ScopeType: string(scopeInfo.Scope),
		ScopeId:   scopeInfo.ScopeId,
		Type:      k8sObjectsTracker.SECRETTYPE_PASSWORD,
		Name:      name,
	}

	secretType := k8sObjectsTracker.SECRETTYPE_PASSWORD
	secretClusterResourceName := s.secretClusterResourceName(api.ASSETKIND_PASSWORD, password.Meta.Name, password.Spec.ExistingSecretName)
	secretClusterResourceStatus := k8sObjectsTracker.ClusterObjectStatus{
		ResourceType: k8sObjectsTracker.CLUSTERRESOURCETYPE_SECRET,
		ScopeType:    string(scopeInfo.Scope),
		ScopeId:      scopeInfo.ScopeId,
		ResourceName: secretClusterResourceName,
		SecretType:   &secretType,
	}
	if issues != nil {
		secretClusterResourceStatus.Issues = *issues
	}

	secret.Origin.Set(&origin)

	s.SecretsResponse.Secrets = append(s.SecretsResponse.Secrets, secret)
	s.ClusterObjectsResponse.Entries = append(s.ClusterObjectsResponse.Entries, secretClusterResourceStatus)
	return s
}

func (s SimpleMockClient) WithDockerRegistry(clusterId string, dockerRegistry *api.DockerRegistryCreationRequest, issues *[]k8sObjectsTracker.ClusterObjectIssue) SimpleMockClient {
	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(dockerRegistry.Meta.Scope), dockerRegistry.Meta.ProjectId, dockerRegistry.Meta.DepartmentId, dockerRegistry.Meta.ClusterId, nil)
	if scopeInfo.Scope == scope.Tenant {
		scopeInfo.Scope = scope.Cluster // in cluster tenant scope resources are stored as cluster wide
	}

	origin, name := s.secretOriginAndName(api.ASSETKIND_DOCKER_REGISTRY, dockerRegistry.Meta.Name, dockerRegistry.Spec.ExistingSecretName)

	secret := k8sObjectsTracker.SecretListInfo{
		ClusterId: clusterId,
		ScopeType: string(scopeInfo.Scope),
		ScopeId:   scopeInfo.ScopeId,
		Type:      k8sObjectsTracker.SECRETTYPE_DOCKER_REGISTRY,
		Name:      name,
	}

	secretType := k8sObjectsTracker.SECRETTYPE_DOCKER_REGISTRY
	secretClusterResourceName := s.secretClusterResourceName(api.ASSETKIND_DOCKER_REGISTRY, dockerRegistry.Meta.Name, dockerRegistry.Spec.ExistingSecretName)
	secretClusterResourceStatus := k8sObjectsTracker.ClusterObjectStatus{
		ResourceType: k8sObjectsTracker.CLUSTERRESOURCETYPE_SECRET,
		ScopeType:    string(scopeInfo.Scope),
		ScopeId:      scopeInfo.ScopeId,
		ResourceName: secretClusterResourceName,
		SecretType:   &secretType,
	}
	if issues != nil {
		secretClusterResourceStatus.Issues = *issues
	}

	secret.Origin.Set(&origin)

	s.SecretsResponse.Secrets = append(s.SecretsResponse.Secrets, secret)
	s.ClusterObjectsResponse.Entries = append(s.ClusterObjectsResponse.Entries, secretClusterResourceStatus)
	return s
}

func (s SimpleMockClient) WithGenericSecret(clusterId string, genericSecret *api.GenericSecretCreationRequest, issues *[]k8sObjectsTracker.ClusterObjectIssue) SimpleMockClient {
	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(genericSecret.Meta.Scope), genericSecret.Meta.ProjectId, genericSecret.Meta.DepartmentId, genericSecret.Meta.ClusterId, nil)
	if scopeInfo.Scope == scope.Tenant {
		scopeInfo.Scope = scope.Cluster // in cluster tenant scope resources are stored as cluster wide
	}

	origin, name := s.secretOriginAndName(api.ASSETKIND_GENERIC_SECRET, genericSecret.Meta.Name, genericSecret.Spec.ExistingSecretName)

	secret := k8sObjectsTracker.SecretListInfo{
		ClusterId: clusterId,
		ScopeType: string(scopeInfo.Scope),
		ScopeId:   scopeInfo.ScopeId,
		Type:      k8sObjectsTracker.SECRETTYPE_GENERIC,
		Name:      name,
	}

	secretType := k8sObjectsTracker.SECRETTYPE_GENERIC
	secretClusterResourceName := s.secretClusterResourceName(api.ASSETKIND_GENERIC_SECRET, genericSecret.Meta.Name, genericSecret.Spec.ExistingSecretName)
	secretClusterResourceStatus := k8sObjectsTracker.ClusterObjectStatus{
		ResourceType: k8sObjectsTracker.CLUSTERRESOURCETYPE_SECRET,
		ScopeType:    string(scopeInfo.Scope),
		ScopeId:      scopeInfo.ScopeId,
		ResourceName: secretClusterResourceName,
		SecretType:   &secretType,
	}
	if issues != nil {
		secretClusterResourceStatus.Issues = *issues
	}

	secret.Origin.Set(&origin)

	s.SecretsResponse.Secrets = append(s.SecretsResponse.Secrets, secret)
	s.ClusterObjectsResponse.Entries = append(s.ClusterObjectsResponse.Entries, secretClusterResourceStatus)
	return s
}

func (s SimpleMockClient) WithAccessKey(clusterId string, accessKey *api.AccessKeyCreationRequest, issues *[]k8sObjectsTracker.ClusterObjectIssue) SimpleMockClient {
	// nil tenantId: creation requests don't include tenantId, it's derived from the requester's context
	scopeInfo := scope.ScopeInfoOf(scope.Scope(accessKey.Meta.Scope), accessKey.Meta.ProjectId, accessKey.Meta.DepartmentId, accessKey.Meta.ClusterId, nil)
	if scopeInfo.Scope == scope.Tenant {
		scopeInfo.Scope = scope.Cluster // in cluster tenant scope resources are stored as cluster wide
	}

	origin, name := s.secretOriginAndName(api.ASSETKIND_ACCESS_KEY, accessKey.Meta.Name, accessKey.Spec.ExistingSecretName)

	secret := k8sObjectsTracker.SecretListInfo{
		ClusterId: clusterId,
		ScopeType: string(scopeInfo.Scope),
		ScopeId:   scopeInfo.ScopeId,
		Type:      k8sObjectsTracker.SECRETTYPE_ACCESS_KEY,
		Name:      name,
	}

	secretType := k8sObjectsTracker.SECRETTYPE_ACCESS_KEY
	secretClusterResourceName := s.secretClusterResourceName(api.ASSETKIND_ACCESS_KEY, accessKey.Meta.Name, accessKey.Spec.ExistingSecretName)
	secretClusterResourceStatus := k8sObjectsTracker.ClusterObjectStatus{
		ResourceType: k8sObjectsTracker.CLUSTERRESOURCETYPE_SECRET,
		ScopeType:    string(scopeInfo.Scope),
		ScopeId:      scopeInfo.ScopeId,
		ResourceName: secretClusterResourceName,
		SecretType:   &secretType,
	}
	if issues != nil {
		secretClusterResourceStatus.Issues = *issues
	}

	secret.Origin.Set(&origin)

	s.SecretsResponse.Secrets = append(s.SecretsResponse.Secrets, secret)
	s.ClusterObjectsResponse.Entries = append(s.ClusterObjectsResponse.Entries, secretClusterResourceStatus)
	return s
}

func (s SimpleMockClient) secretOriginAndName(assetKind api.AssetKind, assetName string, existingSecretName *string) (k8sObjectsTracker.Origin, string) {
	if existingSecretName != nil {
		return k8sObjectsTracker.ORIGIN_EXISTING, *existingSecretName
	}

	switch assetKind {
	case api.ASSETKIND_ACCESS_KEY:
		return k8sObjectsTracker.ORIGIN_ASSETS, fmt.Sprintf("%s-%s", k8sObjectsTracker.SECRETTYPE_ACCESS_KEY, assetName)
	case api.ASSETKIND_PASSWORD:
		return k8sObjectsTracker.ORIGIN_ASSETS, fmt.Sprintf("%s-%s", k8sObjectsTracker.SECRETTYPE_PASSWORD, assetName)
	case api.ASSETKIND_DOCKER_REGISTRY:
		return k8sObjectsTracker.ORIGIN_ASSETS, fmt.Sprintf("%s-%s", k8sObjectsTracker.SECRETTYPE_DOCKER_REGISTRY, assetName)
	case api.ASSETKIND_GENERIC_SECRET:
		return k8sObjectsTracker.ORIGIN_ASSETS, fmt.Sprintf("%s-%s", k8sObjectsTracker.SECRETTYPE_GENERIC, assetName)
	default:
		return k8sObjectsTracker.ORIGIN_ASSETS, fmt.Sprintf("unsupported-%s", assetName)
	}
}

func (s SimpleMockClient) secretClusterResourceName(assetKind api.AssetKind, assetName string, existingSecretName *string) string {
	if existingSecretName != nil {
		return *existingSecretName
	}

	switch assetKind {
	case api.ASSETKIND_ACCESS_KEY:
		return fmt.Sprintf("accesskey-%s", assetName)
	case api.ASSETKIND_PASSWORD:
		return fmt.Sprintf("password-%s", assetName)
	case api.ASSETKIND_DOCKER_REGISTRY:
		return fmt.Sprintf("dockerregistry-%s", assetName)
	case api.ASSETKIND_GENERIC_SECRET:
		return fmt.Sprintf("genericsecret-%s", assetName)
	default:
		return fmt.Sprintf("unsupported-%s", assetName)
	}
}

func (s SimpleMockClient) WithError(code int32, message string) SimpleMockClient {
	s.Error = k8sObjectsTracker.NewError(code, message)
	return s
}

func (s SimpleMockClient) GetSecretsOfCluster(ctx context.Context, clusterId string) (*k8sObjectsTracker.SecretListResponse, *k8sObjectsTracker.Error) {
	return s.SecretsResponse, s.Error
}

func (s SimpleMockClient) GetPvcsOfCluster(ctx context.Context, clusterId string) (*k8sObjectsTracker.PVCListResponse, *k8sObjectsTracker.Error) {
	return s.PvcResponse, s.Error
}

func (s SimpleMockClient) GetConfigMapsOfCluster(_ context.Context, _ string) (*k8sObjectsTracker.ConfigMapsListResponse, *k8sObjectsTracker.Error) {
	return s.ConfigMapsResponse, s.Error
}

func (s SimpleMockClient) GetClusterResourcesStatus(ctx context.Context, resourceType *k8sObjectsTracker.ClusterResourceType,
	scopeType *string, scopeId *string, secretType *k8sObjectsTracker.SecretType,
) (*k8sObjectsTracker.ClusterObjectsStatus, *k8sObjectsTracker.Error) {
	return s.ClusterObjectsResponse, s.Error
}

func (s SimpleMockClient) ListPvcsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*k8sObjectsTracker.PVCListResponse, *k8sObjectsTracker.Error) {
	result := k8sObjectsTracker.PVCListResponse{
		Pvcs: make([]k8sObjectsTracker.PVCListInfo, 0),
	}

	for _, pvc := range s.PvcResponse.Pvcs {
		if scopeType == pvc.ScopeType && scopeId == pvc.ScopeId && name == pvc.Name {
			result.Pvcs = append(result.Pvcs, pvc)
		}
	}

	return &result, nil
}

func (s SimpleMockClient) ListConfigMapsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*k8sObjectsTracker.ConfigMapsListResponse, *k8sObjectsTracker.Error) {
	result := k8sObjectsTracker.ConfigMapsListResponse{
		ConfigMaps: make([]k8sObjectsTracker.ConfigMapListInfo, 0),
	}

	for _, cm := range s.ConfigMapsResponse.ConfigMaps {
		if scopeType == cm.ScopeType && scopeId == cm.ScopeId && name == cm.Name {
			result.ConfigMaps = append(result.ConfigMaps, cm)
		}
	}

	return &result, nil
}

func (s SimpleMockClient) ListSecretsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*k8sObjectsTracker.SecretListResponse, *k8sObjectsTracker.Error) {
	result := k8sObjectsTracker.SecretListResponse{
		Secrets: make([]k8sObjectsTracker.SecretListInfo, 0),
	}

	for _, s := range s.SecretsResponse.Secrets {
		if scopeType == s.ScopeType && scopeId == s.ScopeId && name == s.Name {
			result.Secrets = append(result.Secrets, s)
		}
	}

	return &result, nil
}
