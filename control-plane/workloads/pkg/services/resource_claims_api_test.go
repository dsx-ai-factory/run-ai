/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	k8sObjectsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	k8sObjectsMocks "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/mocks"
)

var _ = Describe("Resource claims API", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	podsService := NewPodsService(&bc_mocks.TenantSettingsClient{})

	createWorkload := func(name string, projectId string, departmentId string, clusterId string) string {
		id := uuid.NewString()
		workload := SetupBaseWorkload(id, projectId, clusterId)
		workload.Name = name
		workload.DepartmentId = departmentId
		apiResponse, _ := workloadsService.UpsertWorkload(ctx, id, workload)
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		return id
	}

	createPod := func(workloadId string, projectId string, clusterId string, name string, claims ...api.PodResourceClaim) string {
		pod := SetupBasePod(uuid.NewString(), workloadId, projectId, clusterId)
		pod.Name = name
		if len(claims) > 0 {
			pod.ResourceClaims = &claims
		}
		apiResponse, _ := podsService.UpsertPod(ctx, pod.Id, pod)
		Expect(apiResponse.Code).To(Equal(http.StatusNoContent))
		return pod.Id
	}

	templateClaim := func(name string, template string, allocated *string) api.PodResourceClaim {
		return api.PodResourceClaim{Name: name, ResourceClaimTemplateName: ptr.To(template), AllocatedResourceClaimName: allocated}
	}

	trackerRct := func(name string, scopeType k8sObjectsApi.ResourceScopeType, scopeId string, assetId *string) k8sObjectsApi.ResourceClaimTemplate {
		rct := k8sObjectsApi.ResourceClaimTemplate{Name: name, Namespace: "runai", ScopeType: scopeType, ScopeId: scopeId}
		if assetId != nil {
			rct.Origin = *k8sObjectsApi.NewNullableOrigin(ptr.To(k8sObjectsApi.ORIGIN_ASSETS))
			rct.AssetId = *k8sObjectsApi.NewNullableString(assetId)
		} else {
			rct.Origin = *k8sObjectsApi.NewNullableOrigin(ptr.To(k8sObjectsApi.ORIGIN_EXISTING))
		}
		return rct
	}

	BeforeAll(func() {
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock())
	})

	AfterAll(func() {
		k8sObjectsClient.SetMock(k8sObjectsMocks.NewK8SObjectTrackerMock())
		testing.CleanTables()
	})

	Context("Get workload resource claims", func() {
		var workloadId string
		projectAssetId := uuid.NewString()
		departmentAssetId := uuid.NewString()
		clusterAssetId := uuid.NewString()
		shadowedAssetId := uuid.NewString()
		siblingAssetId := uuid.NewString()

		BeforeAll(func() {
			workloadId = createWorkload("claims-forward", DefaultProjectId, DefaultDepartmentId, testing.ClusterUuid1)
			createPod(workloadId, DefaultProjectId, testing.ClusterUuid1, "a-pod",
				templateClaim("gpu", "project-rct", nil),
				templateClaim("net", "department-rct", nil),
				api.PodResourceClaim{Name: "ib", ResourceClaimName: ptr.To("shared-ib"), AllocatedResourceClaimName: ptr.To("shared-ib")},
			)
			createPod(workloadId, DefaultProjectId, testing.ClusterUuid1, "b-pod",
				templateClaim("gpu", "cluster-rct", ptr.To("b-pod-gpu-x7k2p")),
				templateClaim("scratch", "unmanaged-rct", nil),
				templateClaim("nic", "sibling-rct", nil),
			)
			createPod(workloadId, DefaultProjectId, testing.ClusterUuid1, "c-pod")
		})

		BeforeEach(func() {
			k8sObjectsClient.SetMock(k8sObjectsMocks.NewK8SObjectTrackerMock().WithResourceClaimTemplates(testing.ClusterUuid1, []k8sObjectsApi.ResourceClaimTemplate{
				trackerRct("project-rct", k8sObjectsApi.RESOURCESCOPETYPE_PROJECT, DefaultProjectId, &projectAssetId),
				trackerRct("project-rct", k8sObjectsApi.RESOURCESCOPETYPE_CLUSTER, testing.ClusterUuid1, &shadowedAssetId),
				trackerRct("department-rct", k8sObjectsApi.RESOURCESCOPETYPE_DEPARTMENT, DefaultDepartmentId, &departmentAssetId),
				trackerRct("department-rct", k8sObjectsApi.RESOURCESCOPETYPE_CLUSTER, testing.ClusterUuid1, &shadowedAssetId),
				trackerRct("cluster-rct", k8sObjectsApi.RESOURCESCOPETYPE_CLUSTER, testing.ClusterUuid1, &clusterAssetId),
				trackerRct("unmanaged-rct", k8sObjectsApi.RESOURCESCOPETYPE_PROJECT, DefaultProjectId, nil),
				trackerRct("sibling-rct", k8sObjectsApi.RESOURCESCOPETYPE_PROJECT, AnotherProjectId, &siblingAssetId),
				trackerRct("sibling-rct", k8sObjectsApi.RESOURCESCOPETYPE_DEPARTMENT, "2", &siblingAssetId),
			}))
		})

		It("returns the claims of the pods that have any, resolving each template", func() {
			apiResponse, _ := workloadsService.GetWorkloadResourceClaims(ctx, workloadId, 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.WorkloadResourceClaimsResponse)
			Expect(response.Next).To(BeZero())
			Expect(response.Pods).To(HaveLen(2))

			Expect(response.Pods[0].PodName).To(Equal("a-pod"))
			Expect(response.Pods[0].ResourceClaims).To(Equal([]api.WorkloadResourceClaim{
				{Name: "gpu", Template: &api.ResourceClaimTemplateRef{Name: "project-rct", Managed: true, Id: &projectAssetId}},
				{Name: "net", Template: &api.ResourceClaimTemplateRef{Name: "department-rct", Managed: true, Id: &departmentAssetId}},
				{Name: "ib", ClaimName: ptr.To("shared-ib")},
			}))

			Expect(response.Pods[1].PodName).To(Equal("b-pod"))
			Expect(response.Pods[1].ResourceClaims).To(Equal([]api.WorkloadResourceClaim{
				{Name: "gpu", ClaimName: ptr.To("b-pod-gpu-x7k2p"), Template: &api.ResourceClaimTemplateRef{Name: "cluster-rct", Managed: true, Id: &clusterAssetId}},
				{Name: "scratch", Template: &api.ResourceClaimTemplateRef{Name: "unmanaged-rct"}},
				{Name: "nic", Template: &api.ResourceClaimTemplateRef{Name: "sibling-rct"}},
			}))
		})

		It("pages the pods", func() {
			apiResponse, _ := workloadsService.GetWorkloadResourceClaims(ctx, workloadId, 0, 1)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.WorkloadResourceClaimsResponse)
			Expect(response.Pods).To(HaveLen(1))
			Expect(response.Pods[0].PodName).To(Equal("a-pod"))
			Expect(response.Next).To(Equal(int32(1)))

			apiResponse, _ = workloadsService.GetWorkloadResourceClaims(ctx, workloadId, 1, 5)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response = apiResponse.Body.(api.WorkloadResourceClaimsResponse)
			Expect(response.Pods).To(HaveLen(1))
			Expect(response.Pods[0].PodName).To(Equal("b-pod"))
			Expect(response.Next).To(BeZero())
		})

		It("returns 503 when the templates cannot be resolved", func() {
			k8sObjectsClient.SetMock(k8sObjectsMocks.NewK8SObjectTrackerMock().
				WithResourceClaimTemplatesError(k8sObjectsApi.NewError(http.StatusServiceUnavailable, "unavailable")))

			apiResponse, _ := workloadsService.GetWorkloadResourceClaims(ctx, workloadId, 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusServiceUnavailable))
		})

		It("does not consult the tracker when no pod references a template", func() {
			k8sObjectsClient.SetMock(k8sObjectsMocks.NewK8SObjectTrackerMock().
				WithResourceClaimTemplatesError(k8sObjectsApi.NewError(http.StatusServiceUnavailable, "unavailable")))
			standaloneWorkloadId := createWorkload("claims-standalone", DefaultProjectId, DefaultDepartmentId, testing.ClusterUuid1)
			createPod(standaloneWorkloadId, DefaultProjectId, testing.ClusterUuid1, "standalone-pod",
				api.PodResourceClaim{Name: "ib", ResourceClaimName: ptr.To("shared-ib"), AllocatedResourceClaimName: ptr.To("shared-ib")})

			apiResponse, _ := workloadsService.GetWorkloadResourceClaims(ctx, standaloneWorkloadId, 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			response := apiResponse.Body.(api.WorkloadResourceClaimsResponse)
			Expect(response.Pods).To(HaveLen(1))
			Expect(response.Pods[0].ResourceClaims[0].Template).To(BeNil())
		})

		It("returns 404 for an unknown workload", func() {
			apiResponse, _ := workloadsService.GetWorkloadResourceClaims(ctx, uuid.NewString(), 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 404 for a deleted workload", func() {
			deletedWorkloadId := createWorkload("claims-deleted", DefaultProjectId, DefaultDepartmentId, testing.ClusterUuid1)
			createPod(deletedWorkloadId, DefaultProjectId, testing.ClusterUuid1, "deleted-pod", templateClaim("gpu", "project-rct", nil))
			apiResponse, _ := workloadsService.DeleteWorkload(ctx, deletedWorkloadId)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			apiResponse, _ = workloadsService.GetWorkloadResourceClaims(ctx, deletedWorkloadId, 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("returns 400 when the workload id is not a uuid", func() {
			apiResponse, _ := workloadsService.GetWorkloadResourceClaims(ctx, "not-a-uuid", 0, 0)
			Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Internal get resource claim template workloads", func() {
		const templateName = "reverse-rct"
		var projectOneWorkloadId, projectTwoWorkloadId, otherDepartmentWorkloadId string

		listWorkloads := func(callCtx context.Context, scopeType api.ScopeType, scopeId string, offset int32, limit int32) api.ResourceClaimTemplateWorkloadsResponse {
			apiResponse, _ := workloadsService.InternalGetResourceClaimTemplateWorkloads(callCtx,
				api.InternalResourceClaimTemplateWorkloadsRequest{Name: templateName, ScopeType: scopeType, ScopeId: scopeId}, offset, limit)
			Expect(apiResponse.Code).To(Equal(http.StatusOK))
			return apiResponse.Body.(api.ResourceClaimTemplateWorkloadsResponse)
		}

		workloadIdsOf := func(response api.ResourceClaimTemplateWorkloadsResponse) []string {
			ids := []string{}
			for _, workload := range response.Workloads {
				ids = append(ids, workload.WorkloadId)
			}
			return ids
		}

		BeforeAll(func() {
			projectOneWorkloadId = createWorkload("reverse-a", DefaultProjectId, DefaultDepartmentId, testing.ClusterUuid1)
			createPod(projectOneWorkloadId, DefaultProjectId, testing.ClusterUuid1, "a-worker-0",
				templateClaim("gpu", templateName, ptr.To("a-worker-0-gpu-q8d1z")),
				templateClaim("nic", "other-rct", nil),
			)
			createPod(projectOneWorkloadId, DefaultProjectId, testing.ClusterUuid1, "a-worker-1",
				templateClaim("gpu-0", templateName, nil),
				templateClaim("gpu-1", templateName, nil),
			)
			createPod(projectOneWorkloadId, DefaultProjectId, testing.ClusterUuid1, "a-launcher")
			deletedPodId := createPod(projectOneWorkloadId, DefaultProjectId, testing.ClusterUuid1, "a-worker-2",
				templateClaim("gpu", templateName, nil))
			apiResponse, _ := podsService.DeletePod(ctx, deletedPodId)
			Expect(apiResponse.Code).To(Equal(http.StatusNoContent))

			projectTwoWorkloadId = createWorkload("reverse-b", AnotherProjectId, DefaultDepartmentId, testing.ClusterUuid1)
			createPod(projectTwoWorkloadId, AnotherProjectId, testing.ClusterUuid1, "b-0", templateClaim("gpu", templateName, nil))

			otherDepartmentWorkloadId = createWorkload("reverse-c", "3", "2", testing.ClusterUuid1)
			createPod(otherDepartmentWorkloadId, "3", testing.ClusterUuid1, "c-0", templateClaim("gpu", templateName, nil))

			otherClusterWorkloadId := createWorkload("reverse-d", "4", "4", testing.ClusterUuid2)
			createPod(otherClusterWorkloadId, "4", testing.ClusterUuid2, "d-0", templateClaim("gpu", templateName, nil))

			otherTemplateWorkloadId := createWorkload("reverse-e", DefaultProjectId, DefaultDepartmentId, testing.ClusterUuid1)
			createPod(otherTemplateWorkloadId, DefaultProjectId, testing.ClusterUuid1, "e-0", templateClaim("gpu", "other-rct", nil))
		})

		It("lists the workloads of the cluster with the pods and claims that use the template", func() {
			response := listWorkloads(ctx, api.SCOPETYPE_CLUSTER, testing.ClusterUuid1, 0, 0)
			Expect(response.Next).To(BeZero())
			Expect(workloadIdsOf(response)).To(Equal([]string{projectOneWorkloadId, projectTwoWorkloadId, otherDepartmentWorkloadId}))

			workload := response.Workloads[0]
			Expect(workload.Name).To(Equal("reverse-a"))
			Expect(workload.Type).To(Equal(DefaultType))
			Expect(workload.ProjectId).To(Equal(DefaultProjectId))
			Expect(workload.Pods).To(HaveLen(2))
			Expect(workload.Pods[0].PodName).To(Equal("a-worker-0"))
			Expect(workload.Pods[0].ResourceClaims).To(Equal([]api.PodResourceClaimBinding{
				{Name: "gpu", ClaimName: ptr.To("a-worker-0-gpu-q8d1z")},
			}))
			Expect(workload.Pods[1].PodName).To(Equal("a-worker-1"))
			Expect(workload.Pods[1].ResourceClaims).To(Equal([]api.PodResourceClaimBinding{
				{Name: "gpu-0"},
				{Name: "gpu-1"},
			}))
		})

		It("narrows the lookup to a department", func() {
			response := listWorkloads(ctx, api.SCOPETYPE_DEPARTMENT, DefaultDepartmentId, 0, 0)
			Expect(workloadIdsOf(response)).To(Equal([]string{projectOneWorkloadId, projectTwoWorkloadId}))
		})

		It("narrows the lookup to a project", func() {
			response := listWorkloads(ctx, api.SCOPETYPE_PROJECT, AnotherProjectId, 0, 0)
			Expect(workloadIdsOf(response)).To(Equal([]string{projectTwoWorkloadId}))
		})

		It("returns only the workloads the caller may read", func() {
			projectCtx := testing.SetProjectContext(testing.TestTenant, testing.TestUser, []string{}, AnotherProjectId)
			response := listWorkloads(projectCtx, api.SCOPETYPE_CLUSTER, testing.ClusterUuid1, 0, 0)
			Expect(workloadIdsOf(response)).To(Equal([]string{projectTwoWorkloadId}))
		})

		It("pages the workloads", func() {
			response := listWorkloads(ctx, api.SCOPETYPE_CLUSTER, testing.ClusterUuid1, 0, 2)
			Expect(workloadIdsOf(response)).To(Equal([]string{projectOneWorkloadId, projectTwoWorkloadId}))
			Expect(response.Next).To(Equal(int32(2)))

			response = listWorkloads(ctx, api.SCOPETYPE_CLUSTER, testing.ClusterUuid1, 2, 2)
			Expect(workloadIdsOf(response)).To(Equal([]string{otherDepartmentWorkloadId}))
			Expect(response.Next).To(BeZero())
		})

		DescribeTable("rejects an invalid request",
			func(request api.InternalResourceClaimTemplateWorkloadsRequest) {
				apiResponse, _ := workloadsService.InternalGetResourceClaimTemplateWorkloads(ctx, request, 0, 0)
				Expect(apiResponse.Code).To(Equal(http.StatusBadRequest))
			},
			Entry("missing name", api.InternalResourceClaimTemplateWorkloadsRequest{ScopeType: api.SCOPETYPE_PROJECT, ScopeId: DefaultProjectId}),
			Entry("missing scope id", api.InternalResourceClaimTemplateWorkloadsRequest{Name: templateName, ScopeType: api.SCOPETYPE_PROJECT}),
			Entry("tenant scope", api.InternalResourceClaimTemplateWorkloadsRequest{Name: templateName, ScopeType: api.SCOPETYPE_TENANT, ScopeId: testing.TestTenant}),
			Entry("system scope", api.InternalResourceClaimTemplateWorkloadsRequest{Name: templateName, ScopeType: api.SCOPETYPE_SYSTEM, ScopeId: "system"}),
			Entry("cluster scope with a non-uuid id", api.InternalResourceClaimTemplateWorkloadsRequest{Name: templateName, ScopeType: api.SCOPETYPE_CLUSTER, ScopeId: "cluster-1"}),
		)
	})
})
