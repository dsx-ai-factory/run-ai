/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

// for backward compatibility only existing resources returned by default, specifying all returns also
// assets related resources
const originAll = "all"

type K8SObjectTrackerClient struct {
	ApiClient *openapi.APIClient
}

func (c *K8SObjectTrackerClient) CreateClusterWideConfigmap(ctx context.Context, clusterId string, name string, issues []openapi.Issue) (*openapi.ConfigMap, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.CreateClusterWideConfigmapV1(ctx, clusterId, name).ConfigMap(
		openapi.ConfigMap{
			Name:   name,
			Issues: issues,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateDepartmentsWideConfigmap(ctx context.Context, clusterId string, departmentId string, name string, issues []openapi.Issue) (*openapi.ConfigMap, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.CreateDepartmentsWideConfigmapV1(ctx, clusterId, departmentId, name).ConfigMap(
		openapi.ConfigMap{
			Name:   name,
			Issues: issues,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateProjectsConfigmap(ctx context.Context, clusterId string, projectId string, name string, issues []openapi.Issue) (*openapi.ConfigMap, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.CreateProjectsConfigmapV1(ctx, clusterId, projectId, name).ConfigMap(
		openapi.ConfigMap{
			Name:   name,
			Issues: issues,
		}).Execute()
}

func (c *K8SObjectTrackerClient) DeleteClusterWideConfigmap(ctx context.Context, clusterId string, name string) (*http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.DeleteClusterWideConfigmapV1(ctx, clusterId, name).Execute()
}

func (c *K8SObjectTrackerClient) DeleteDepartmentsWideConfigmap(ctx context.Context, clusterId string, departmentId string, name string) (*http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.DeleteDepartmentsWideConfigmapV1(ctx, clusterId, departmentId, name).Execute()
}

func (c *K8SObjectTrackerClient) DeleteProjectsConfigmap(ctx context.Context, clusterId string, projectId string, name string) (*http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.DeleteProjectsConfigmapV1(ctx, clusterId, projectId, name).Execute()
}

func (c *K8SObjectTrackerClient) SyncClusterWideConfigmaps(ctx context.Context, clusterId string, configMaps []openapi.ConfigMap) (*openapi.ConfigMaps, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.SyncClusterWideConfigmapsV1(ctx, clusterId).ConfigMaps(
		openapi.ConfigMaps{
			Configmaps: configMaps,
		}).Execute()
}

func (c *K8SObjectTrackerClient) SyncDepartmentsWideConfigmap(ctx context.Context, clusterId string, departmentId string, configMaps []openapi.ConfigMap) (*openapi.ConfigMaps, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.SyncDepartmentsWideConfigmapV1(ctx, clusterId, departmentId).ConfigMaps(
		openapi.ConfigMaps{
			Configmaps: configMaps,
		}).Execute()
}

func (c *K8SObjectTrackerClient) SyncProjectsConfigmap(ctx context.Context, clusterId string, projectId string, configMaps []openapi.ConfigMap) (*openapi.ConfigMaps, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.SyncProjectsConfigmapsV1(ctx, clusterId, projectId).ConfigMaps(
		openapi.ConfigMaps{
			Configmaps: configMaps,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateClusterWidePvc(ctx context.Context, clusterId string, name string, request openapi.PVCCreationRequest) (*openapi.PVC, *http.Response, error) {
	return c.ApiClient.PVCsAPI.CreateClusterWidePvcV2(ctx, clusterId, name).PVCCreationRequest(request).Execute()
}

func (c *K8SObjectTrackerClient) CreateDepartmentWidePvc(ctx context.Context, clusterId string, departmentId string, name string, request openapi.PVCCreationRequest) (*openapi.PVC, *http.Response, error) {
	return c.ApiClient.PVCsAPI.CreateDepartmentWidePvcV2(ctx, clusterId, departmentId, name).PVCCreationRequest(request).Execute()
}

func (c *K8SObjectTrackerClient) CreateProjectPvc(ctx context.Context, clusterId string, projectId string, name string, request openapi.PVCCreationRequest) (*openapi.PVC, *http.Response, error) {
	return c.ApiClient.PVCsAPI.CreateProjectPvcV2(ctx, clusterId, projectId, name).PVCCreationRequest(request).Execute()
}

func (c *K8SObjectTrackerClient) DeleteClusterWidePvc(ctx context.Context, clusterId string, name string) (*http.Response, error) {
	return c.ApiClient.PVCsAPI.DeleteClusterWidePvcV2(ctx, clusterId, name).Execute()
}

func (c *K8SObjectTrackerClient) DeleteDepartmentWidePvc(ctx context.Context, clusterId string, departmentId string, name string) (*http.Response, error) {
	return c.ApiClient.PVCsAPI.DeleteDepartmentWidePvcV2(ctx, clusterId, departmentId, name).Execute()
}

func (c *K8SObjectTrackerClient) DeleteProjectPvc(ctx context.Context, clusterId string, projectId string, name string) (*http.Response, error) {
	return c.ApiClient.PVCsAPI.DeleteProjectPvcV2(ctx, clusterId, projectId, name).Execute()
}

func (c *K8SObjectTrackerClient) SyncClusterWidePvcs(ctx context.Context, clusterId string, pvcs []openapi.PVC) (*openapi.PVCs, *http.Response, error) {
	return c.ApiClient.PVCsAPI.SyncClusterWidePvcsV2(ctx, clusterId).PVCs(
		openapi.PVCs{
			Pvcs: pvcs,
		}).Execute()
}

func (c *K8SObjectTrackerClient) SyncDepartmentWidePvcs(ctx context.Context, clusterId string, departmentId string, pvcs []openapi.PVC) (*openapi.PVCs, *http.Response, error) {
	return c.ApiClient.PVCsAPI.SyncDepartmentWidePvcsV2(ctx, clusterId, departmentId).PVCs(
		openapi.PVCs{
			Pvcs: pvcs,
		}).Execute()
}

func (c *K8SObjectTrackerClient) SyncProjectPvcs(ctx context.Context, clusterId string, projectId string, pvcs []openapi.PVC) (*openapi.PVCs, *http.Response, error) {
	return c.ApiClient.PVCsAPI.SyncProjectPvcsV2(ctx, clusterId, projectId).PVCs(
		openapi.PVCs{
			Pvcs: pvcs,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateClusterWideSecret(ctx context.Context, clusterId string, name string, secretType openapi.SecretType, origin openapi.NullableOrigin, issues []openapi.Issue, private openapi.NullableBool) (*openapi.SecretCreationRequest, *http.Response, error) {
	return c.ApiClient.SecretsAPI.CreateClusterWideSecret(ctx, clusterId, name).SecretCreationRequest(
		openapi.SecretCreationRequest{
			Type:    secretType,
			Origin:  origin,
			Issues:  issues,
			Private: private,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateDepartmentWideSecret(ctx context.Context, clusterId string, departmentId string, name string, secretType openapi.SecretType, origin openapi.NullableOrigin, issues []openapi.Issue) (*openapi.Secret, *http.Response, error) {
	return c.ApiClient.SecretsAPI.CreateDepartmentWideSecret(ctx, clusterId, departmentId, name).SecretCreationRequest(
		openapi.SecretCreationRequest{
			Type:   secretType,
			Origin: origin,
			Issues: issues,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateProjectSecret(ctx context.Context, clusterId string, projectId string, name string, secretType openapi.SecretType, origin openapi.NullableOrigin, issues []openapi.Issue) (*openapi.Secret, *http.Response, error) {
	return c.ApiClient.SecretsAPI.CreateProjectSecret(ctx, clusterId, projectId, name).SecretCreationRequest(
		openapi.SecretCreationRequest{
			Type:   secretType,
			Origin: origin,
			Issues: issues,
		}).Execute()
}

func (c *K8SObjectTrackerClient) DeleteClusterWideSecret(ctx context.Context, clusterId string, name string) (*http.Response, error) {
	return c.ApiClient.SecretsAPI.DeleteClusterWideSecret(ctx, clusterId, name).Execute()
}

func (c *K8SObjectTrackerClient) DeleteDepartmentWideSecret(ctx context.Context, clusterId string, departmentId string, name string) (*http.Response, error) {
	return c.ApiClient.SecretsAPI.DeleteDepartmentWideSecret(ctx, clusterId, departmentId, name).Execute()
}

func (c *K8SObjectTrackerClient) DeleteProjectSecret(ctx context.Context, clusterId string, projectId string, name string) (*http.Response, error) {
	return c.ApiClient.SecretsAPI.DeleteProjectSecret(ctx, clusterId, projectId, name).Execute()
}

func (c *K8SObjectTrackerClient) SyncClusterWideSecrets(ctx context.Context, clusterId string, secrets []openapi.Secret) (*openapi.Secrets, *http.Response, error) {
	return c.ApiClient.SecretsAPI.SyncClusterWideSecrets(ctx, clusterId).Secrets(
		openapi.Secrets{
			Secrets: secrets,
		}).Execute()
}

func (c *K8SObjectTrackerClient) SyncDepartmentWideSecrets(ctx context.Context, clusterId string, departmentId string, secrets []openapi.Secret) (*openapi.Secrets, *http.Response, error) {
	return c.ApiClient.SecretsAPI.SyncDepartmentWideSecrets(ctx, clusterId, departmentId).Secrets(
		openapi.Secrets{
			Secrets: secrets,
		}).Execute()
}

func (c *K8SObjectTrackerClient) SyncProjectSecrets(ctx context.Context, clusterId string, projectId string, secrets []openapi.Secret) (*openapi.Secrets, *http.Response, error) {
	return c.ApiClient.SecretsAPI.SyncProjectSecrets(ctx, clusterId, projectId).Secrets(
		openapi.Secrets{
			Secrets: secrets,
		}).Execute()
}

func (c *K8SObjectTrackerClient) CreateStorageClass(ctx context.Context, storageClass openapi.StorageClass) (*openapi.StorageClass, *http.Response, error) {
	return c.ApiClient.StorageClassesAPI.CreateStorageClass(ctx).StorageClass(storageClass).Execute()
}

func (c *K8SObjectTrackerClient) DeleteStorageClass(ctx context.Context, clusterId string, name string) (*http.Response, error) {
	return c.ApiClient.StorageClassesAPI.DeleteStorageClass(ctx).ClusterId(clusterId).Name(name).Execute()
}

func (c *K8SObjectTrackerClient) UpdateStorageClass(ctx context.Context, storageClass openapi.StorageClass) (*openapi.StorageClass, *http.Response, error) {
	return c.ApiClient.StorageClassesAPI.UpdateStorageClass(ctx).StorageClass(storageClass).Execute()
}

func (c *K8SObjectTrackerClient) SyncStorageClass(ctx context.Context, clusterId string, storageClasses []openapi.StorageClass) (*openapi.StorageClasses, *http.Response, error) {
	return c.ApiClient.StorageClassesAPI.SyncStorageClass(ctx).StorageClasses(
		openapi.StorageClasses{
			Items: storageClasses,
		}).ClusterId(clusterId).Execute()
}

func (c *K8SObjectTrackerClient) CreateDeviceClass(ctx context.Context, deviceClass openapi.DeviceClass) (*openapi.DeviceClass, *http.Response, error) {
	return c.ApiClient.DeviceClassesAPI.CreateDeviceClass(ctx).DeviceClass(deviceClass).Execute()
}

func (c *K8SObjectTrackerClient) UpdateDeviceClass(ctx context.Context, deviceClass openapi.DeviceClass) (*openapi.DeviceClass, *http.Response, error) {
	return c.ApiClient.DeviceClassesAPI.UpdateDeviceClass(ctx).DeviceClass(deviceClass).Execute()
}

func (c *K8SObjectTrackerClient) DeleteDeviceClass(ctx context.Context, clusterId string, name string) (*http.Response, error) {
	return c.ApiClient.DeviceClassesAPI.DeleteDeviceClass(ctx).ClusterId(clusterId).Name(name).Execute()
}

func (c *K8SObjectTrackerClient) SyncDeviceClass(ctx context.Context, clusterId string, deviceClasses []openapi.DeviceClass) (*openapi.DeviceClasses, *http.Response, error) {
	return c.ApiClient.DeviceClassesAPI.SyncDeviceClass(ctx).DeviceClasses(
		openapi.DeviceClasses{
			Items: deviceClasses,
		}).ClusterId(clusterId).Execute()
}

func (c *K8SObjectTrackerClient) ListDeviceClasses(ctx context.Context, clusterId string) (*openapi.DeviceClasses, *http.Response, error) {
	return c.ApiClient.DeviceClassesAPI.GetDeviceClasses(ctx).ClusterId(clusterId).Execute()
}

func (c *K8SObjectTrackerClient) CreateResourceClaimTemplate(ctx context.Context, clusterId string, rct openapi.ResourceClaimTemplate) (*openapi.ResourceClaimTemplate, *http.Response, error) {
	return c.ApiClient.ResourceClaimTemplatesAPI.CreateResourceClaimTemplate(ctx).ClusterId(clusterId).ResourceClaimTemplate(rct).Execute()
}

func (c *K8SObjectTrackerClient) UpdateResourceClaimTemplate(ctx context.Context, clusterId string, rct openapi.ResourceClaimTemplate) (*openapi.ResourceClaimTemplate, *http.Response, error) {
	return c.ApiClient.ResourceClaimTemplatesAPI.UpdateResourceClaimTemplate(ctx).ClusterId(clusterId).ResourceClaimTemplate(rct).Execute()
}

func (c *K8SObjectTrackerClient) DeleteResourceClaimTemplate(ctx context.Context, clusterId string, scopeType openapi.ResourceScopeType, scopeId string, name string) (*http.Response, error) {
	return c.ApiClient.ResourceClaimTemplatesAPI.DeleteResourceClaimTemplate(ctx).
		ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).Name(name).Execute()
}

func (c *K8SObjectTrackerClient) SyncResourceClaimTemplates(ctx context.Context, clusterId string, rcts []openapi.ResourceClaimTemplate) (*openapi.ResourceClaimTemplates, *http.Response, error) {
	return c.ApiClient.ResourceClaimTemplatesAPI.SyncResourceClaimTemplates(ctx).ResourceClaimTemplates(
		openapi.ResourceClaimTemplates{
			Items: rcts,
		}).ClusterId(clusterId).Execute()
}

func (c *K8SObjectTrackerClient) ListResourceClaimTemplatesByName(ctx context.Context, clusterId string, name string, origin string) (*openapi.ResourceClaimTemplates, *openapi.Error) {
	paramsAsStr := fmt.Sprintf("resource claim templates named %s in cluster %s", name, clusterId)

	request := c.ApiClient.ResourceClaimTemplatesAPI.ListResourceClaimTemplates(ctx).ClusterId(clusterId).Name(name)
	if origin != "" {
		request = request.Origin(origin)
	}

	resp, r, err := request.Execute()
	if r != nil {
		if err == nil && resp != nil && r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			data = []byte(fmt.Sprintf("<failed to read body: %v>", readErr))
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list %s: %s", paramsAsStr, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to list %s", paramsAsStr))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request list %s: %v", paramsAsStr, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list %s", paramsAsStr))
}

func (c *K8SObjectTrackerClient) ListSecrets(ctx context.Context, clusterId string) (*openapi.SecretListResponse, *http.Response, error) {
	return c.ApiClient.SecretsAPI.ListSecrets(ctx).ClusterId(clusterId).Origin(originAll).Execute()
}

func (c *K8SObjectTrackerClient) ListSecretsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*openapi.SecretListResponse, *openapi.Error) {
	paramsAsStr := fmt.Sprintf("secret named %s of scope %s:%s", name, scopeType, scopeId)

	resp, r, err := c.ApiClient.SecretsAPI.ListSecrets(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list %s: %v", paramsAsStr, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to list %s", paramsAsStr))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request list %s: %v", paramsAsStr, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list %s", paramsAsStr))
}

// ListSecretsByScope lists all secrets of a given scope, including ancestors, all origins
func (c *K8SObjectTrackerClient) ListSecretsByScope(ctx context.Context, scopeType string, scopeId string) (*openapi.SecretListResponse, *openapi.Error) {
	paramsAsStr := fmt.Sprintf("secret of scope %s:%s", scopeType, scopeId)

	resp, r, err := c.ApiClient.SecretsAPI.ListSecrets(ctx).ScopeType(scopeType).ScopeId(scopeId).IncludeAncestors(true).Origin(originAll).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list %s: %v", paramsAsStr, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to list %s", paramsAsStr))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request list %s: %v", paramsAsStr, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list %s", paramsAsStr))
}

func (c *K8SObjectTrackerClient) ListPvcs(ctx context.Context, clusterId string) (*openapi.PVCListResponse, *http.Response, error) {
	return c.ApiClient.PVCsAPI.ListPvcs(ctx).ClusterId(clusterId).Origin(originAll).Execute()
}

func (c *K8SObjectTrackerClient) ListPvcsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*openapi.PVCListResponse, *openapi.Error) {
	paramsAsStr := fmt.Sprintf("pvcs named %s of scope %s:%s", name, scopeType, scopeId)

	resp, r, err := c.ApiClient.PVCsAPI.ListPvcs(ctx).Origin(originAll).ScopeType(scopeType).ScopeId(scopeId).Name(name).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list %s: %v", paramsAsStr, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to list %s", paramsAsStr))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request list %s: %v", paramsAsStr, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list %s", paramsAsStr))
}

func (c *K8SObjectTrackerClient) ListConfigMaps(ctx context.Context, clusterId string) (*openapi.ConfigMapsListResponse, *http.Response, error) {
	return c.ApiClient.ConfigMapsAPI.ListConfigMaps(ctx).ClusterId(clusterId).Execute()
}

func (c *K8SObjectTrackerClient) ListConfigMapsByScopeAndName(ctx context.Context, scopeType string, scopeId string, name string) (*openapi.ConfigMapsListResponse, *openapi.Error) {
	paramsAsStr := fmt.Sprintf("config map named %s of scope %s:%s", name, scopeType, scopeId)

	resp, r, err := c.ApiClient.ConfigMapsAPI.ListConfigMaps(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list %s: %v", paramsAsStr, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to list %s", paramsAsStr))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request list %s: %v", paramsAsStr, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list %s", paramsAsStr))
}

func (c *K8SObjectTrackerClient) ListClusterObjectsStatus(ctx context.Context, resourceType *openapi.ClusterResourceType, scopeType *string, scopeId *string, secretType *openapi.SecretType) (*openapi.ClusterObjectsStatus, *http.Response, error) {
	return c.ApiClient.ClusterObjectsStatusAPI.ListClusterObjectsStatus(ctx).
		ResourceType(ptr.Deref(resourceType, "")).
		ScopeType(string(ptr.Deref(scopeType, ""))).
		ScopeId(ptr.Deref(scopeId, "")).
		SecretType(string(ptr.Deref(secretType, ""))).
		Execute()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
//  High level client services
////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *K8SObjectTrackerClient) GetSecretsOfCluster(ctx context.Context, clusterId string) (*openapi.SecretListResponse, *openapi.Error) {
	resp, r, err := c.ListSecrets(ctx, clusterId)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get workspace secrets of cluster %s: %s", clusterId, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get secrets of cluster %s", clusterId))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request secrets of cluster %s: %v", clusterId, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request secrets of cluster %s", clusterId))
}

func (c *K8SObjectTrackerClient) GetPvcsOfCluster(ctx context.Context, clusterId string) (*openapi.PVCListResponse, *openapi.Error) {
	resp, r, err := c.ListPvcs(ctx, clusterId)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get workspace pvcs of cluster %s: %s", clusterId, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pvcs of cluster %s", clusterId))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request pvcs of cluster %s: %v", clusterId, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request pvcs of cluster %s", clusterId))
}

// GetConfigMapsOfCluster queries the k8s tracker for the config maps of the cluster
func (c *K8SObjectTrackerClient) GetConfigMapsOfCluster(ctx context.Context, clusterId string) (*openapi.ConfigMapsListResponse, *openapi.Error) {
	resp, r, err := c.ListConfigMaps(ctx, clusterId)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get workspace config-maps of cluster %s: %s", clusterId, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get config-maps of cluster %s", clusterId))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request config-maps of cluster %s: %v", clusterId, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request config-maps of cluster %s", clusterId))
}

// GetDeviceClassesOfCluster queries the k8s tracker for the device classes of the cluster.
// Used by validators (e.g. the resource-claim-template path in assets-service) that need to
// verify a referenced deviceClassName exists in the target cluster.
func (c *K8SObjectTrackerClient) GetDeviceClassesOfCluster(ctx context.Context, clusterId string) (*openapi.DeviceClasses, *openapi.Error) {
	resp, r, err := c.ListDeviceClasses(ctx, clusterId)
	if r != nil {
		// Treat as success only when the SDK itself reported no error AND
		// handed us a decoded response. A 200 with a JSON-decode failure
		// comes back as (nil-or-partial, http.Response{200}, err); without
		// this guard the caller gets a nil resp and panics on resp.Items.
		if err == nil && resp != nil && r.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Surface read failures in the log rather than swallowing them —
		// an empty "data" otherwise makes the error path indistinguishable
		// from a non-200 with an empty body.
		data, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			data = []byte(fmt.Sprintf("<failed to read body: %v>", readErr))
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get device classes of cluster %s: %s", clusterId, data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get device classes of cluster %s", clusterId))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request device classes of cluster %s: %v", clusterId, err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request device classes of cluster %s", clusterId))
}

func (c *K8SObjectTrackerClient) GetClusterResourcesStatus(ctx context.Context, resourceType *openapi.ClusterResourceType, scopeType *string,
	scopeId *string, secretType *openapi.SecretType,
) (*openapi.ClusterObjectsStatus, *openapi.Error) {
	resp, r, err := c.ListClusterObjectsStatus(ctx, resourceType, scopeType, scopeId, secretType)
	if r != nil {
		if r.StatusCode == http.StatusOK {
			return resp, nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get cluster resources status: %s", data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get cluster resources status"))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request cluster resources status: %v", err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request cluster resources status"))
}

func (c *K8SObjectTrackerClient) GetStorageClassConfigInt(ctx context.Context, clusterId string, storageClassName string) (*openapi.StorageClassConfiguration, *openapi.Error) {
	// filter by name and clusterId
	filters := []string{
		fmt.Sprintf("scopeId==%s", clusterId),
		fmt.Sprintf("name==%s", storageClassName),
	}

	resp, r, err := c.ApiClient.StorageClassConfigurationAPI.ListStorageClassConfigurationInt(ctx).FilterBy(filters).Execute()
	if r != nil {
		if r.StatusCode == http.StatusOK {
			if len(resp.StorageClassConfigurations) == 0 {
				return nil, openapi.NewError(http.StatusNotFound, fmt.Sprintf("storage class %s not found in cluster %s", storageClassName, clusterId))
			}

			// we are expecting one result; if more are returned, take the first one and log the unexpected state
			if len(resp.StorageClassConfigurations) > 1 {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%d storage classes named %s found in cluster %s", len(resp.StorageClassConfigurations), storageClassName, clusterId)
			}

			return &resp.StorageClassConfigurations[0], nil
		}

		data, _ := io.ReadAll(r.Body)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get storage class configuration: %s", data)

		return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get storage class configuration"))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to request storage class configuration: %v", err)
	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request storage class configuration"))
}

func (c *K8SObjectTrackerClient) GetStorageClassesConfigInt(ctx context.Context, clusterId string, storageClassNames []string) (map[string]openapi.StorageClassConfiguration, *openapi.Error) {
	// filter by name and clusterId
	filters := []string{
		fmt.Sprintf("scopeId==%s", clusterId),
		fmt.Sprintf("name==%s", strings.Join(storageClassNames, "||")),
	}

	nextOffset := int32(0)

	req := c.ApiClient.StorageClassConfigurationAPI.ListStorageClassConfigurationInt(ctx).FilterBy(filters)

	configs := make(map[string]openapi.StorageClassConfiguration)

	for true {
		resp, hResp, err := req.Offset(nextOffset).Execute()
		if hResp != nil && hResp.StatusCode >= http.StatusBadRequest {
			return nil, openapi.NewError(int32(hResp.StatusCode), fmt.Sprintf("Failed to get configurations offset %d", nextOffset))
		}
		if err != nil {
			return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
		}
		if resp == nil {
			return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("Failed to obtain response of get storage configurations offset %d", nextOffset))
		}

		if len(resp.StorageClassConfigurations) == 0 {
			break
		}

		for _, config := range resp.StorageClassConfigurations {
			configs[config.Name] = config
		}

		if resp.GetNext() == 0 {
			break
		}
		if resp.GetNext() <= nextOffset {
			return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("received next offset %d while current offset is %d", resp.GetNext(), nextOffset))
		}
		nextOffset = resp.GetNext()
	}

	return configs, nil
}
