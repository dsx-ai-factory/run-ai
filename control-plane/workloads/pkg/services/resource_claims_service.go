/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	k8s_objects_tracker "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
)

func (s *WorkloadsService) GetWorkloadResourceClaims(ctx context.Context, workloadId string, offset int32, limit int32) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	wid, err := uuid.Parse(workloadId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest,
			fmt.Sprintf("Invalid workload id %q", workloadId), nil)), nil
	}

	limit = setDefaultLimit(limit)

	workload, apiErr := s.WorkloadsDao.DbGetWorkloadScope(ctx, tenantId, wid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	pods, apiErr := dao.DbListPodsWithResourceClaims(ctx, tenantId, []uuid.UUID{wid}, "", offset, limit)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	templates, apiErr := resolveResourceClaimTemplates(ctx, *workload, referencedResourceClaimTemplateNames(pods))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	response := api.WorkloadResourceClaimsResponse{
		Pods: make([]api.WorkloadPodResourceClaims, 0, len(pods)),
	}
	for _, pod := range pods {
		response.Pods = append(response.Pods, podToWorkloadPodResourceClaims(pod, templates))
	}
	if int32(len(pods)) == limit {
		response.Next = offset + limit
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *WorkloadsService) InternalGetResourceClaimTemplateWorkloads(ctx context.Context,
	request api.InternalResourceClaimTemplateWorkloadsRequest, offset int32, limit int32,
) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	if apiErr = validateResourceClaimTemplateWorkloadsRequest(request); apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	limit = setDefaultLimit(limit)

	permittedScopes, err := authorization_client.ParsePermittedScopes(ctx)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusForbidden, "Insufficient permissions", nil)), nil
	}

	workloads, apiErr := dao.DbListWorkloadsReferencingResourceClaimTemplate(ctx, tenantId, request.Name,
		request.ScopeType, request.ScopeId, permittedScopes, offset, limit)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workloadIds := make([]uuid.UUID, 0, len(workloads))
	for _, workload := range workloads {
		workloadIds = append(workloadIds, workload.ID)
	}
	pods, apiErr := dao.DbListPodsWithResourceClaims(ctx, tenantId, workloadIds, request.Name, 0, 0)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	podsByWorkload := make(map[uuid.UUID][]api.ResourceClaimTemplateWorkloadPod, len(workloads))
	for _, pod := range pods {
		podsByWorkload[pod.WorkloadID] = append(podsByWorkload[pod.WorkloadID], api.ResourceClaimTemplateWorkloadPod{
			PodId:          pod.ID.String(),
			PodName:        pod.Name,
			ResourceClaims: templateResourceClaimBindings(pod, request.Name),
		})
	}

	response := api.ResourceClaimTemplateWorkloadsResponse{
		Workloads: make([]api.ResourceClaimTemplateWorkload, 0, len(workloads)),
	}
	for _, workload := range workloads {
		workloadPods := podsByWorkload[workload.ID]
		if workloadPods == nil {
			workloadPods = []api.ResourceClaimTemplateWorkloadPod{}
		}
		response.Workloads = append(response.Workloads, api.ResourceClaimTemplateWorkload{
			WorkloadId: workload.ID.String(),
			Name:       workload.Name,
			Type:       workload.Type,
			ProjectId:  workload.ProjectID,
			Pods:       workloadPods,
		})
	}
	if int32(len(workloads)) == limit {
		response.Next = offset + limit
	}

	return api.Response(http.StatusOK, response), nil
}

func validateResourceClaimTemplateWorkloadsRequest(request api.InternalResourceClaimTemplateWorkloadsRequest) *api.Error {
	if request.Name == "" {
		return api.NewError(http.StatusBadRequest, "name is required", nil)
	}
	if request.ScopeId == "" {
		return api.NewError(http.StatusBadRequest, "scopeId is required", nil)
	}

	switch request.ScopeType {
	case api.SCOPETYPE_CLUSTER:
		if _, err := uuid.Parse(request.ScopeId); err != nil {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("scopeId must be a valid cluster id. <%s>", request.ScopeId), nil)
		}
	case api.SCOPETYPE_DEPARTMENT, api.SCOPETYPE_PROJECT:
	default:
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("unsupported scopeType %q, expected cluster, department or project", request.ScopeType), nil)
	}

	return nil
}

// resolveResourceClaimTemplates looks each template name up in the workload's cluster. The same
// name may be used by sibling projects and departments, so only the workload's own project,
// department and cluster count, closest first. Names matching none of them are unmanaged.
func resolveResourceClaimTemplates(ctx context.Context, workload models.Workload, names []string) (map[string]api.ResourceClaimTemplateRef, *api.Error) {
	clusterId := workload.ClusterID.String()
	refs := make(map[string]api.ResourceClaimTemplateRef, len(names))
	for _, name := range names {
		rcts, trackerErr := k8s_objects_tracker.Get().ListResourceClaimTemplatesByName(ctx, clusterId, name, string(k8sObjectsApi.ORIGIN_ASSETS))
		if trackerErr != nil {
			return nil, api.NewError(http.StatusServiceUnavailable,
				fmt.Sprintf("Failed to resolve the resource claim template %s", name), nil)
		}

		refs[name] = api.ResourceClaimTemplateRef{Name: name}
		if assetId := closestResourceClaimTemplateAssetId(rcts.Items, workload); assetId != nil {
			refs[name] = api.ResourceClaimTemplateRef{Name: name, Managed: true, Id: assetId}
		}
	}

	return refs, nil
}

// we use this function to handle the case that multiple claims found with the same name. we want to find
// the closest match, so we start comparing project, then department, and finally the cluster. the first
// match is the closest.
func closestResourceClaimTemplateAssetId(rcts []k8sObjectsApi.ResourceClaimTemplate, workload models.Workload) *string {
	var departmentAssetId, clusterAssetId *string
	for _, rct := range rcts {
		assetId := rct.AssetId.Get()
		if assetId == nil {
			continue
		}
		switch rct.ScopeType {
		case k8sObjectsApi.RESOURCESCOPETYPE_PROJECT:
			if rct.ScopeId == workload.ProjectID {
				return assetId
			}
		case k8sObjectsApi.RESOURCESCOPETYPE_DEPARTMENT:
			if rct.ScopeId == workload.DepartmentID {
				departmentAssetId = assetId
			}
		case k8sObjectsApi.RESOURCESCOPETYPE_CLUSTER:
			clusterAssetId = assetId
		}
	}

	if departmentAssetId != nil {
		return departmentAssetId
	}
	return clusterAssetId
}

func referencedResourceClaimTemplateNames(pods []models.Pod) []string {
	names := datastructs.NewSet[string]()
	for _, pod := range pods {
		for _, claim := range pod.ResourceClaims {
			if claim.ResourceClaimTemplateName != nil {
				names.Add(*claim.ResourceClaimTemplateName)
			}
		}
	}
	return names.Keys()
}

func podToWorkloadPodResourceClaims(pod models.Pod, templates map[string]api.ResourceClaimTemplateRef) api.WorkloadPodResourceClaims {
	claims := make([]api.WorkloadResourceClaim, 0, len(pod.ResourceClaims))
	for _, claim := range pod.ResourceClaims {
		apiClaim := api.WorkloadResourceClaim{
			Name:      claim.Name,
			ClaimName: claim.AllocatedResourceClaimName,
		}
		if claim.ResourceClaimTemplateName != nil {
			template := templates[*claim.ResourceClaimTemplateName]
			apiClaim.Template = &template
		}
		claims = append(claims, apiClaim)
	}

	return api.WorkloadPodResourceClaims{
		PodId:          pod.ID.String(),
		PodName:        pod.Name,
		Task:           pod.Task,
		ResourceClaims: claims,
	}
}

func templateResourceClaimBindings(pod models.Pod, templateName string) []api.PodResourceClaimBinding {
	bindings := []api.PodResourceClaimBinding{}
	for _, claim := range pod.ResourceClaims {
		if claim.ResourceClaimTemplateName == nil || *claim.ResourceClaimTemplateName != templateName {
			continue
		}
		bindings = append(bindings, api.PodResourceClaimBinding{
			Name:      claim.Name,
			ClaimName: claim.AllocatedResourceClaimName,
		})
	}
	return bindings
}
