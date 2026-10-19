/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8s_objects_tracker_client

import (
	"context"
	"net/http"

	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/internal"
)

type IK8SObjectTrackerExternalClient interface {
	CreateClusterWideConfigmap(ctx context.Context, clusterId string, name string, issues []openapi.Issue) (*openapi.ConfigMap, *http.Response, error)
	CreateDepartmentsWideConfigmap(ctx context.Context, clusterId string, departmentId string, name string, issues []openapi.Issue) (*openapi.ConfigMap, *http.Response, error)
	CreateProjectsConfigmap(ctx context.Context, clusterId string, projectId string, name string, issues []openapi.Issue) (*openapi.ConfigMap, *http.Response, error)
	DeleteClusterWideConfigmap(ctx context.Context, clusterId string, name string) (*http.Response, error)
	DeleteDepartmentsWideConfigmap(ctx context.Context, clusterId string, departmentId string, name string) (*http.Response, error)
	DeleteProjectsConfigmap(ctx context.Context, clusterId string, projectId string, name string) (*http.Response, error)
	SyncClusterWideConfigmaps(ctx context.Context, clusterId string, configMaps []openapi.ConfigMap) (*openapi.ConfigMaps, *http.Response, error)
	SyncDepartmentsWideConfigmap(ctx context.Context, clusterId string, departmentId string, configMaps []openapi.ConfigMap) (*openapi.ConfigMaps, *http.Response, error)
	SyncProjectsConfigmap(ctx context.Context, clusterId string, projectId string, configMaps []openapi.ConfigMap) (*openapi.ConfigMaps, *http.Response, error)

	CreateClusterWidePvc(ctx context.Context, clusterId string, name string, request openapi.PVCCreationRequest) (*openapi.PVC, *http.Response, error)
	CreateDepartmentWidePvc(ctx context.Context, clusterId string, departmentId string, name string, request openapi.PVCCreationRequest) (*openapi.PVC, *http.Response, error)
	CreateProjectPvc(ctx context.Context, clusterId string, projectId string, name string, request openapi.PVCCreationRequest) (*openapi.PVC, *http.Response, error)
	DeleteClusterWidePvc(ctx context.Context, clusterId string, name string) (*http.Response, error)
	DeleteDepartmentWidePvc(ctx context.Context, clusterId string, departmentId string, name string) (*http.Response, error)
	DeleteProjectPvc(ctx context.Context, clusterId string, projectId string, name string) (*http.Response, error)
	SyncClusterWidePvcs(ctx context.Context, clusterId string, pvcs []openapi.PVC) (*openapi.PVCs, *http.Response, error)
	SyncDepartmentWidePvcs(ctx context.Context, clusterId string, departmentId string, pvcs []openapi.PVC) (*openapi.PVCs, *http.Response, error)
	SyncProjectPvcs(ctx context.Context, clusterId string, projectId string, pvcs []openapi.PVC) (*openapi.PVCs, *http.Response, error)

	CreateClusterWideSecret(ctx context.Context, clusterId string, name string, secretType openapi.SecretType, origin openapi.NullableOrigin, issues []openapi.Issue, private openapi.NullableBool) (*openapi.SecretCreationRequest, *http.Response, error)
	CreateDepartmentWideSecret(ctx context.Context, clusterId string, departmentId string, name string, secretType openapi.SecretType, origin openapi.NullableOrigin, issues []openapi.Issue) (*openapi.Secret, *http.Response, error)
	CreateProjectSecret(ctx context.Context, clusterId string, projectId string, name string, secretType openapi.SecretType, origin openapi.NullableOrigin, issues []openapi.Issue) (*openapi.Secret, *http.Response, error)
	DeleteClusterWideSecret(ctx context.Context, clusterId string, name string) (*http.Response, error)
	DeleteDepartmentWideSecret(ctx context.Context, clusterId string, departmentId string, name string) (*http.Response, error)
	DeleteProjectSecret(ctx context.Context, clusterId string, projectId string, name string) (*http.Response, error)
	SyncClusterWideSecrets(ctx context.Context, clusterId string, secrets []openapi.Secret) (*openapi.Secrets, *http.Response, error)
	SyncDepartmentWideSecrets(ctx context.Context, clusterId string, departmentId string, secrets []openapi.Secret) (*openapi.Secrets, *http.Response, error)
	SyncProjectSecrets(ctx context.Context, clusterId string, projectId string, secrets []openapi.Secret) (*openapi.Secrets, *http.Response, error)

	CreateStorageClass(ctx context.Context, storageClass openapi.StorageClass) (*openapi.StorageClass, *http.Response, error)
	DeleteStorageClass(ctx context.Context, clusterId string, name string) (*http.Response, error)
	UpdateStorageClass(ctx context.Context, storageClass openapi.StorageClass) (*openapi.StorageClass, *http.Response, error)
	SyncStorageClass(ctx context.Context, clusterId string, storageClasses []openapi.StorageClass) (*openapi.StorageClasses, *http.Response, error)

	CreateDeviceClass(ctx context.Context, deviceClass openapi.DeviceClass) (*openapi.DeviceClass, *http.Response, error)
	UpdateDeviceClass(ctx context.Context, deviceClass openapi.DeviceClass) (*openapi.DeviceClass, *http.Response, error)
	DeleteDeviceClass(ctx context.Context, clusterId string, name string) (*http.Response, error)
	SyncDeviceClass(ctx context.Context, clusterId string, deviceClasses []openapi.DeviceClass) (*openapi.DeviceClasses, *http.Response, error)
	// ListDeviceClasses returns every device class the tracker knows about for the given cluster.
	// Intended for validators that need to check whether a referenced deviceClassName exists
	// (e.g. the resource-claim-template validation path in assets-service).
	ListDeviceClasses(ctx context.Context, clusterId string) (*openapi.DeviceClasses, *http.Response, error)

	// ResourceClaimTemplate uses the same per-item CRUD + full-set sync shape
	// as DeviceClass. Per-item ops hit a single row (no pruning); SyncResourceClaimTemplates
	// is reserved for the periodic drift reconciliation that upserts the full
	// inventory and removes anything no longer in the cluster. clusterId
	// travels in the query because the RCT body schema does not carry it.
	CreateResourceClaimTemplate(ctx context.Context, clusterId string, rct openapi.ResourceClaimTemplate) (*openapi.ResourceClaimTemplate, *http.Response, error)
	UpdateResourceClaimTemplate(ctx context.Context, clusterId string, rct openapi.ResourceClaimTemplate) (*openapi.ResourceClaimTemplate, *http.Response, error)
	DeleteResourceClaimTemplate(ctx context.Context, clusterId string, scopeType openapi.ResourceScopeType, scopeId string, name string) (*http.Response, error)
	SyncResourceClaimTemplates(ctx context.Context, clusterId string, rcts []openapi.ResourceClaimTemplate) (*openapi.ResourceClaimTemplates, *http.Response, error)
	// ListResourceClaimTemplatesByName lists the templates of the cluster with the given name,
	// across every scope. An empty origin lists every origin.
	ListResourceClaimTemplatesByName(ctx context.Context, clusterId string, name string, origin string) (*openapi.ResourceClaimTemplates, *openapi.Error)

	ListSecrets(ctx context.Context, clusterId string) (*openapi.SecretListResponse, *http.Response, error)
	ListSecretsByScope(ctx context.Context, scopeType string, scopeId string) (*openapi.SecretListResponse, *openapi.Error)
	ListSecretsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*openapi.SecretListResponse, *openapi.Error)
	ListPvcs(ctx context.Context, clusterId string) (*openapi.PVCListResponse, *http.Response, error)
	ListPvcsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*openapi.PVCListResponse, *openapi.Error)
	ListConfigMaps(ctx context.Context, clusterId string) (*openapi.ConfigMapsListResponse, *http.Response, error)
	ListConfigMapsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*openapi.ConfigMapsListResponse, *openapi.Error)
	ListClusterObjectsStatus(ctx context.Context, resourceType *openapi.ClusterResourceType, scopeType *string, scopeId *string, secretType *openapi.SecretType) (*openapi.ClusterObjectsStatus, *http.Response, error)

	//  high Level client services

	GetSecretsOfCluster(ctx context.Context, clusterId string) (*openapi.SecretListResponse, *openapi.Error)
	GetPvcsOfCluster(ctx context.Context, clusterId string) (*openapi.PVCListResponse, *openapi.Error)
	GetConfigMapsOfCluster(ctx context.Context, clusterId string) (*openapi.ConfigMapsListResponse, *openapi.Error)
	GetDeviceClassesOfCluster(ctx context.Context, clusterId string) (*openapi.DeviceClasses, *openapi.Error)
	GetClusterResourcesStatus(ctx context.Context, resourceType *openapi.ClusterResourceType,
		scopeType *string, scopeId *string, secretType *openapi.SecretType) (*openapi.ClusterObjectsStatus, *openapi.Error)
}

type IK8SObjectTrackerClient interface {
	IK8SObjectTrackerExternalClient

	GetStorageClassConfigInt(ctx context.Context, clusterId string, storageClassName string) (*openapi.StorageClassConfiguration, *openapi.Error)
	GetStorageClassesConfigInt(ctx context.Context, clusterId string, storageClassNames []string) (map[string]openapi.StorageClassConfiguration, *openapi.Error)
}

func NewK8SObjectTrackerClient(baseURL string) IK8SObjectTrackerClient {
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	return NewK8SObjectTrackerClientWithConfig(config)
}

func NewK8SObjectTrackerClientWithConfig(config *openapi.Configuration) IK8SObjectTrackerClient {
	return &internal.K8SObjectTrackerClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewK8SObjectTrackerExternalClient(baseURL string) IK8SObjectTrackerExternalClient {
	return NewK8SObjectTrackerClient(client_utils.GetExternalUrl(baseURL))
}

func NewK8SObjectTrackerExternalClientWithConfig(config *openapi.Configuration) IK8SObjectTrackerExternalClient {
	return NewK8SObjectTrackerClientWithConfig(config)
}
