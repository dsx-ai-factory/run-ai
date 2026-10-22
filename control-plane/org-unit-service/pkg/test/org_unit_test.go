/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	v1 "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Org unit tests", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)
	cluster1Id := uuid.New().String()
	cluster2Id := uuid.New().String()
	cluster3Id := uuid.New().String()

	var orgUnitApiService api.OrgUnitAPIServicer
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var projectIds []string
	var departmentIds []string

	BeforeAll(func() {
		clusterServiceClient := mocks.IClusterClient{}
		clusterServiceClient.On("GetInternalClusters", mock.Anything, mock.Anything).Return([]clustersOpenapi.DisplayedCluster{
			{
				Uuid:      cluster1Id,
				Name:      "deltst-cluster1",
				Version:   *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				DeletedAt: *clustersOpenapi.NewNullableTime(nil),
			},
			{
				Uuid:      cluster2Id,
				Name:      "deltst-cluster2",
				Version:   *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				DeletedAt: *clustersOpenapi.NewNullableTime(nil),
			},
			{
				Uuid:      cluster3Id,
				Name:      "deltst-cluster3",
				Version:   *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				DeletedAt: *clustersOpenapi.NewNullableTime(ptr.To(time.Now())),
			},
		}, nil)
		clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
			Nodepools: []clustersOpenapi.NodepoolMinimal{
				{
					Name:                   defaultNodePoolName,
					Id:                     fmt.Sprint(defaultNodePoolId),
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false),
				},
			},
		}, nil)
		clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
			Uuid:    cluster1Id,
			Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
		}, nil)
		orgUnitApiService = services.NewOrgUnitApiService(
			&SimpleMockAuthService{},
			&clusterServiceClient, nil, &clusterServiceClient)
		projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
	})

	Context("test get priorities", func() {
		var ranks []api.Rank

		It("Get Ranks", func() {
			response, err := orgUnitApiService.GetRanks(ctx)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			ranksResponse := response.Body.(api.GetRanks200Response)
			ranks = ranksResponse.Ranks
			Expect(ranks).To(Not(BeNil()))
		})

		It("test default rank", func() {
			var defaultRank api.Rank

			for _, rank := range ranks {
				if rank.IsDefault {
					defaultRank = rank
				}
			}

			Expect(defaultRank).To(Not(BeNil()))
		})
	})

	Context("test get deleted org units", func() {
		var deletedProjectIds []string
		var deletedDepartmentIds []string
		deletedAt := time.Now().Add(time.Duration(-4) * time.Hour)

		BeforeAll(func() {
			// departments:
			createDepartmentRequest := getDepartmentCreateRequest("deltst-dep-keep1", cluster1Id)
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department1Id := response.Body.(api.Department).Id
			departmentIds = append(departmentIds, department1Id)

			createDepartmentRequest = getDepartmentCreateRequest("deltst-dep-keep2", cluster2Id)
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department2Id := response.Body.(api.Department).Id
			departmentIds = append(departmentIds, department2Id)

			createDepartmentRequest = getDepartmentCreateRequest("deltst-dep-delete1", cluster1Id)
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department3Id := response.Body.(api.Department).Id
			deletedDepartmentIds = append(deletedDepartmentIds, department3Id)

			createDepartmentRequest = getDepartmentCreateRequest("deltst-dep-delete2", cluster2Id)
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department4Id := response.Body.(api.Department).Id
			deletedDepartmentIds = append(deletedDepartmentIds, department4Id)

			// projects
			name := fmt.Sprint("deltst-proj-keep1-", uuid.New())
			request := getProjectCreateRequest(name, name, department1Id, cluster1Id)
			createProjResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			proj1Id := createProjResponse.Body.(api.Project).Id
			projectIds = append(projectIds, proj1Id)

			name = fmt.Sprint("deltst-proj-delete1", uuid.New())
			request = getProjectCreateRequest(name, name, department3Id, cluster1Id)
			createProjResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			proj2Id := createProjResponse.Body.(api.Project).Id
			deletedProjectIds = append(deletedProjectIds, proj2Id)

			name = fmt.Sprint("deltst-proj-keep2", uuid.New())
			request = getProjectCreateRequest(name, name, department2Id, cluster2Id)
			createProjResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			proj3Id := createProjResponse.Body.(api.Project).Id
			projectIds = append(projectIds, proj3Id)

			name = fmt.Sprint("deltst-proj-delete2", uuid.New())
			request = getProjectCreateRequest(name, name, department4Id, cluster2Id)
			createProjResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			proj4Id := createProjResponse.Body.(api.Project).Id
			deletedProjectIds = append(deletedProjectIds, proj4Id)

			unitsRes, err := orgUnitApiService.GetInternalDeletedOrgUnits(ctx, deletedAt)
			Expect(err).To(BeNil())
			Expect(unitsRes.Code).To(Equal(http.StatusOK))
			units := unitsRes.Body.(api.GetInternalDeletedOrgUnits200Response)
			for _, unit := range units.OrgUnits {
				Expect(strings.HasPrefix(unit.Name, "deltst-dep")).To(BeFalse())
				Expect(strings.HasPrefix(unit.Name, "deltst-proj")).To(BeFalse())
			}

			createDepartmentRequest = getDepartmentCreateRequest("deltst-dep-delete-c3", cluster3Id)
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department5Id := response.Body.(api.Department).Id
			departmentIds = append(departmentIds, department5Id)

			createDepartmentRequest = getDepartmentCreateRequest("deltst-dep-delete-keep-c3", cluster3Id)
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department6Id := response.Body.(api.Department).Id
			deletedDepartmentIds = append(deletedDepartmentIds, department6Id)

			name = fmt.Sprint("deltst-proj-delete-c3", uuid.New())
			request = getProjectCreateRequest(name, name, department5Id, cluster3Id)
			createProjResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			proj5Id := createProjResponse.Body.(api.Project).Id
			projectIds = append(projectIds, proj5Id)

			name = fmt.Sprint("deltst-proj-delete-keep-c3", uuid.New())
			request = getProjectCreateRequest(name, name, department5Id, cluster3Id)
			createProjResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			proj6Id := createProjResponse.Body.(api.Project).Id
			deletedProjectIds = append(deletedProjectIds, proj6Id)
		})

		AfterAll(func() {
			for _, id := range projectIds {
				_, _ = projectApiService.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
			}
			projectIds = []string{}
			for _, id := range departmentIds {
				_, _ = departmentApiService.DeleteDepartment(ctx, id)
			}
			projectIds = []string{}
		})

		It("test deleted projects and departments", func() {
			// delete projects:
			for _, id := range deletedProjectIds {
				deleteRes, err := projectApiService.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
				Expect(err).To(BeNil())
				Expect(deleteRes.Code).To(Equal(http.StatusNoContent))
			}
			// delete departments:
			for _, id := range deletedDepartmentIds {
				deleteRes, err := departmentApiService.DeleteDepartment(ctx, id)
				Expect(err).To(BeNil())
				Expect(deleteRes.Code).To(Equal(http.StatusNoContent))
			}
			unitsRes, err := orgUnitApiService.GetInternalDeletedOrgUnits(ctx, deletedAt)
			Expect(err).To(BeNil())
			Expect(unitsRes.Code).To(Equal(http.StatusOK))
			units := unitsRes.Body.(api.GetInternalDeletedOrgUnits200Response)
			delProjUnits := []api.DeletedOrgUnit{}
			for _, unit := range units.OrgUnits {
				if strings.HasPrefix(unit.Name, "deltst-proj") {
					delProjUnits = append(delProjUnits, unit)
				}
			}
			Expect(len(delProjUnits)).To(Equal(4))
			for _, delUnit := range delProjUnits {
				Expect(strings.HasPrefix(delUnit.Name, "deltst-proj-delete")).To(BeTrue())
			}

			delDepUnits := []api.DeletedOrgUnit{}
			for _, unit := range units.OrgUnits {
				if strings.HasPrefix(unit.Name, "deltst-dep") {
					delDepUnits = append(delDepUnits, unit)
				}
			}
			Expect(len(delDepUnits)).To(Equal(4))
			for _, delUnit := range delDepUnits {
				Expect(strings.HasPrefix(delUnit.Name, "deltst-dep-delete")).To(BeTrue())
			}

			delClusterUnits := []api.DeletedOrgUnit{}
			for _, unit := range units.OrgUnits {
				if strings.HasPrefix(unit.Name, "deltst-cluster") {
					delClusterUnits = append(delClusterUnits, unit)
				}
			}
			Expect(len(delClusterUnits)).To(Equal(1))
			Expect(delClusterUnits[0].Name).To(Equal("deltst-cluster3"))
		})
	})

	Context("test orgs nodepool summary", func() {
		var (
			clusterServiceClient mocks.IClusterClient
			clusterId            string

			depQuotaNoRank_np2NoQuotaNoRank, depRankNoQuota_np2QuotaNoRank, depRankAndQuota_np2RankNoQuota      api.Department
			projQuotaNoRank_np2RankAndQuota, projRankNoQuota_np2NoQuotaNoRank, projQuotaNoRank_np2RankAndQuota2 api.Project

			// additional nodepool to validate per-nodepool behavior
			nodePoolName2 string
			nodePoolId2   string
		)

		BeforeAll(func() {
			clusterId = uuid.New().String()
			nodePoolName2 = "pool2"
			nodePoolId2 = "20"

			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetInternalClusters", mock.Anything, mock.Anything).Return([]clustersOpenapi.DisplayedCluster{
				{Uuid: clusterId, Name: "sum-cluster", Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")), DeletedAt: *clustersOpenapi.NewNullableTime(nil)},
			}, nil)
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{{
					Name:                   defaultNodePoolName,
					Id:                     fmt.Sprint(defaultNodePoolId),
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				}, {
					Name:                   nodePoolName2,
					Id:                     nodePoolId2,
					Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				}},
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName, ClusterId: clusterId},
					{Id: nodePoolId2, Name: nodePoolName2, ClusterId: clusterId},
				},
			}, nil)

			orgUnitApiService = services.NewOrgUnitApiService(
				&SimpleMockAuthService{},
				&clusterServiceClient, nil, &clusterServiceClient)

			projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})

		BeforeAll(func() {
			// create a department with quota, default rank
			// in nodepool2 - no quota, no rank
			creationReq := getDepartmentCreateRequest("sum-dep-quota", clusterId)
			creationReq.Resources = append(creationReq.Resources, CreateNodePoolApiResourcesNoQuota(nodePoolId2, nodePoolName2))
			depRes, err := departmentApiService.CreateDepartment(ctx, creationReq)
			Expect(err).To(BeNil())
			Expect(depRes.Code).To(Equal(http.StatusCreated), fmt.Sprint("failed creating department: ", depRes.Body))
			depQuotaNoRank_np2NoQuotaNoRank = depRes.Body.(api.Department)
			departmentIds = append(departmentIds, depQuotaNoRank_np2NoQuotaNoRank.Id)

			// add another department without quota, with non-default rank
			// in nodepool2 - with quota, no rank
			creationReq = getDepartmentCreateRequest("sum-dep-noquota", clusterId)
			creationReq.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			creationReq.Resources[0].Gpu.Limit = nil
			creationReq.Resources[0].Gpu.OverQuotaWeight = nil
			creationReq.Resources[0].Cpu = &api.ResourcesCpu{
				Deserved: ptr.To(float64(-1)),
			}
			creationReq.Resources[0].Memory = &api.ResourcesMemory{
				Deserved: ptr.To(float64(-1)),
			}
			creationReq.Resources[0].Rank = ptr.To("Low") // set non-default priority
			creationReq.Resources = append(creationReq.Resources, CreateNodePoolApiResources(false, nodePoolId2, nodePoolName2))
			depNoQuotaRes, err := departmentApiService.CreateDepartment(ctx, creationReq)
			Expect(err).To(BeNil())
			Expect(depNoQuotaRes.Code).To(Equal(http.StatusCreated))
			depRankNoQuota_np2QuotaNoRank = depNoQuotaRes.Body.(api.Department)
			departmentIds = append(departmentIds, depRankNoQuota_np2QuotaNoRank.Id)

			// add another department with both cpu quota and non-default rank
			// in nodepool2 - no quota, with non-default rank
			creationReq = getDepartmentCreateRequest("sum-dep-both", clusterId)
			creationReq.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			creationReq.Resources[0].Gpu.Limit = nil
			creationReq.Resources[0].Gpu.OverQuotaWeight = nil
			creationReq.Resources[0].Cpu = &api.ResourcesCpu{
				Deserved: ptr.To(float64(4)),
				Limit:    ptr.To(float64(4)),
			}
			creationReq.Resources[0].Rank = ptr.To("High") // set non-default priority
			creationReq.Resources = append(creationReq.Resources, CreateNodePoolApiResourcesNoQuota(nodePoolId2, nodePoolName2))
			creationReq.Resources[1].Rank = ptr.To("VeryLow") // set non-default rank for np2
			depBothRes, err := departmentApiService.CreateDepartment(ctx, creationReq)
			Expect(err).To(BeNil())
			Expect(depBothRes.Code).To(Equal(http.StatusCreated), fmt.Sprint("failed creating department: ", depBothRes.Body))
			depRankAndQuota_np2RankNoQuota = depBothRes.Body.(api.Department)
			departmentIds = append(departmentIds, depRankAndQuota_np2RankNoQuota.Id)

			// create 2 projects and set CPU memory quota on default nodepool
			// in nodepool2 - with quota, with rank
			pNames := []string{fmt.Sprint("sum-proj-quota-", uuid.New()), fmt.Sprint("sum-proj-quota-", uuid.New())}
			for i, pName := range pNames {
				projReq := getProjectCreateRequest(pName, pName, depRankNoQuota_np2QuotaNoRank.Id, clusterId)
				projReq.Resources[0].Gpu.Deserved = ptr.To(float64(0))
				projReq.Resources[0].Gpu.Limit = nil
				projReq.Resources[0].Gpu.OverQuotaWeight = nil
				projReq.Resources[0].Memory = &api.ResourcesMemory{
					Deserved: ptr.To(float64(4)),
				}
				projReq.Resources = append(projReq.Resources, CreateNodePoolApiResources(false, nodePoolId2, nodePoolName2))
				projReq.Resources[1].Rank = ptr.To("VeryLow") // set non-default priority for np2
				projRes, err := projectApiService.CreateProject(ctx, projReq)
				Expect(err).To(BeNil())
				Expect(projRes.Code).To(Equal(http.StatusCreated))
				if i == 0 {
					projQuotaNoRank_np2RankAndQuota = projRes.Body.(api.Project)
					projectIds = append(projectIds, projQuotaNoRank_np2RankAndQuota.Id)
				} else {
					projQuotaNoRank_np2RankAndQuota2 = projRes.Body.(api.Project)
					projectIds = append(projectIds, projQuotaNoRank_np2RankAndQuota2.Id)
				}
			}

			// create another project without quota, with rank
			// in nodepool2 - no quota, no rank
			pNoQuotaName := fmt.Sprint("sum-proj-noquota-", uuid.New())
			projNoQuotaReq := getProjectCreateRequest(pNoQuotaName, pNoQuotaName, depQuotaNoRank_np2NoQuotaNoRank.Id, clusterId)
			projNoQuotaReq.Resources[0].Gpu.Deserved = ptr.To(float64(0))
			projNoQuotaReq.Resources[0].Gpu.Limit = nil
			projNoQuotaReq.Resources[0].Gpu.OverQuotaWeight = nil
			projNoQuotaReq.Resources[0].Rank = ptr.To("Low") // set non-default rank
			projNoQuotaReq.Resources = append(projNoQuotaReq.Resources, CreateNodePoolApiResourcesNoQuota(nodePoolId2, nodePoolName2))
			projNoQuotaRes, err := projectApiService.CreateProject(ctx, projNoQuotaReq)
			Expect(err).To(BeNil())
			Expect(projNoQuotaRes.Code).To(Equal(http.StatusCreated))
			projRankNoQuota_np2NoQuotaNoRank = projNoQuotaRes.Body.(api.Project)
			projectIds = append(projectIds, projRankNoQuota_np2NoQuotaNoRank.Id)
		})

		AfterAll(func() {
			// Cleanup any created entities
			for _, id := range projectIds {
				_, _ = projectApiService.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
			}
			projectIds = []string{}
			for _, id := range departmentIds {
				_, _ = departmentApiService.DeleteDepartment(ctx, id)
			}
			departmentIds = []string{}
		})

		It("test nodepool hasQuota summary", func() {
			// call OrgsNodePoolSummary for hasQuota
			sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, fmt.Sprint(defaultNodePoolId), api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA)
			Expect(err).To(BeNil())
			Expect(sumRes.Code).To(Equal(http.StatusOK))
			orgUnitsNodepoolSummary := sumRes.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary

			Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
			Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))

			// call OrgsNodePoolSummary for nodepool 2 and validate per-nodepool behavior
			sumResNp2, err := orgUnitApiService.OrgsNodePoolSummary(ctx, nodePoolId2, api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA)
			Expect(err).To(BeNil())
			Expect(sumResNp2.Code).To(Equal(http.StatusOK))
			orgUnitsNodepoolSummaryNp2 := sumResNp2.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary
			Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
			Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))
		})

		It("test nodepool nonDefaultRank summary", func() {
			// call OrgsNodePoolSummary for NON_DEFAULT_RANK
			sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, fmt.Sprint(defaultNodePoolId), api.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK)
			Expect(err).To(BeNil())
			Expect(sumRes.Code).To(Equal(http.StatusOK))
			orgUnitsNodepoolSummary := sumRes.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary

			Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
			Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota2.Id, Name: projQuotaNoRank_np2RankAndQuota2.Name, Type: "project"}))
			Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))

			// validate nodepool 2
			sumResNp2, err := orgUnitApiService.OrgsNodePoolSummary(ctx, nodePoolId2, api.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK)
			Expect(err).To(BeNil())
			Expect(sumResNp2.Code).To(Equal(http.StatusOK))
			orgUnitsNodepoolSummaryNp2 := sumResNp2.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary
			Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
			Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota2.Id, Name: projQuotaNoRank_np2RankAndQuota2.Name, Type: "project"}))
			Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
			Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))

			// now delete one of the projects to see it is not included in the summary anymore
			delRes, err := projectApiService.DeleteProject(ctx, projQuotaNoRank_np2RankAndQuota2.Id, api.DELETIONTYPES_ORPHAN, true)
			Expect(err).To(BeNil())
			Expect(delRes.Code).To(Equal(http.StatusNoContent))
			nodePoolsToCheck := []string{fmt.Sprint(defaultNodePoolId), nodePoolId2}
			summaryTypesToCheck := []api.OrgUnitsNodepoolSummaryType{api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA, api.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK}
			for _, npId := range nodePoolsToCheck {
				for _, sumType := range summaryTypesToCheck {
					sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, npId, sumType)
					Expect(err).To(BeNil())
					Expect(sumRes.Code).To(Equal(http.StatusOK))
					orgUnitsNodepoolSummary := sumRes.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary
					Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota2.Id, Name: projQuotaNoRank_np2RankAndQuota2.Name, Type: "project"}))
				}
			}
		})

		It("test wrong summary type requested", func() {
			// call OrgsNodePoolSummary for hasQuota
			sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, fmt.Sprint(defaultNodePoolId), "wrongType")
			Expect(err).To(BeNil())
			Expect(sumRes).ToNot(BeNil())
			Expect(sumRes.Code).To(Equal(http.StatusBadRequest))
		})

		It("test unexisting nodepool", func() {
			// call OrgsNodePoolSummary for hasQuota
			sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, "100", api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA)
			Expect(err).To(BeNil())
			Expect(sumRes.Code).To(Equal(http.StatusOK))
			orgUnitsNodepoolSummary := sumRes.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary
			Expect(len(orgUnitsNodepoolSummary)).To(Equal(0))
		})

		Context("test summary after 'fixing' it using ResetResources", func() {
			var ()

			BeforeAll(func() {
				// use the ResetResources API to reset resources

				// for default nodepool - ProjectsResetResources with QUOTAS_TO_DEFAULT
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT,
				}
				res, err := projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNoContent))

				// for nodepool2 - DepartmentsResetResources with RANKS_TO_DEFAULT
				depReq := api.DepartmentsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{nodePoolId2},
					Action:      api.RESOURCESRESETACTION_RANKS_TO_DEFAULT,
				}
				depRes, err := departmentApiService.DepartmentsResetResources(ctx, depReq)
				Expect(err).To(BeNil())
				Expect(depRes.Code).To(Equal(http.StatusNoContent))
			})

			AfterAll(func() {
				departmentsToRestore := []api.Department{depQuotaNoRank_np2NoQuotaNoRank, depRankNoQuota_np2QuotaNoRank, depRankAndQuota_np2RankNoQuota}
				projectsToRestore := []api.Project{projQuotaNoRank_np2RankAndQuota, projRankNoQuota_np2NoQuotaNoRank}

				// restore original resources
				for _, d := range departmentsToRestore {
					updateReq := getDepartmentUpdateRequest(d)
					updateRes, err := departmentApiService.UpdateDepartment(ctx, d.Id, updateReq)
					Expect(err).To(BeNil())
					Expect(updateRes.Code).To(Equal(http.StatusOK))
				}

				for _, p := range projectsToRestore {
					updateReq := getProjectUpdateRequest(p)
					updateRes, err := projectApiService.UpdateProject(ctx, p.Id, updateReq)
					Expect(err).To(BeNil())
					Expect(updateRes.Code).To(Equal(http.StatusOK))
				}
			})

			It("after ProjectsResetResources with QUOTAS_TO_DEFAULT - validate summary is correct", func() {
				// call OrgsNodePoolSummary for hasQuota
				sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, fmt.Sprint(defaultNodePoolId), api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA)
				Expect(err).To(BeNil())
				Expect(sumRes.Code).To(Equal(http.StatusOK))
				orgUnitsNodepoolSummary := sumRes.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary

				Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
				// this is the changed one - should not be present anymore
				Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
				Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))

				// nodepool 2 - should stay the same
				sumResNp2, err := orgUnitApiService.OrgsNodePoolSummary(ctx, nodePoolId2, api.ORGUNITSNODEPOOLSUMMARYTYPE_HAS_QUOTA)
				Expect(err).To(BeNil())
				Expect(sumResNp2.Code).To(Equal(http.StatusOK))
				orgUnitsNodepoolSummaryNp2 := sumResNp2.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
				Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))
			})

			It("after DepartmentsResetResources with RANKS_TO_DEFAULT - validate summary is correct", func() {
				// call OrgsNodePoolSummary for NON_DEFAULT_RANK
				// default np - should stay the same
				sumRes, err := orgUnitApiService.OrgsNodePoolSummary(ctx, fmt.Sprint(defaultNodePoolId), api.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK)
				Expect(err).To(BeNil())
				Expect(sumRes.Code).To(Equal(http.StatusOK))
				orgUnitsNodepoolSummary := sumRes.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary

				Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummary).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
				Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummary).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))

				// validate nodepool 2
				sumResNp2, err := orgUnitApiService.OrgsNodePoolSummary(ctx, nodePoolId2, api.ORGUNITSNODEPOOLSUMMARYTYPE_NON_DEFAULT_RANK)
				Expect(err).To(BeNil())
				Expect(sumResNp2.Code).To(Equal(http.StatusOK))
				orgUnitsNodepoolSummaryNp2 := sumResNp2.Body.(api.OrgsNodePoolSummary200Response).OrgUnitsNodepoolSummary
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depQuotaNoRank_np2NoQuotaNoRank.Id, Name: depQuotaNoRank_np2NoQuotaNoRank.Name, Type: "department"}))
				// this is the changed one - should not be present anymore
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankAndQuota_np2RankNoQuota.Id, Name: depRankAndQuota_np2RankNoQuota.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummaryNp2).To(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projQuotaNoRank_np2RankAndQuota.Id, Name: projQuotaNoRank_np2RankAndQuota.Name, Type: "project"}))
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: depRankNoQuota_np2QuotaNoRank.Id, Name: depRankNoQuota_np2QuotaNoRank.Name, Type: "department"}))
				Expect(orgUnitsNodepoolSummaryNp2).ToNot(ContainElement(api.OrgUnitsNodepoolSummaryInner{Id: projRankNoQuota_np2NoQuotaNoRank.Id, Name: projRankNoQuota_np2NoQuotaNoRank.Name, Type: "project"}))
			})
		})
	})

	Context("test CRUD projects/departments with nodepools that have different scheduling capabilities", func() {
		var (
			clusterId              string
			depIds                 []string
			projIds                []string
			clusterServiceClient   mocks.IClusterClient
			projectV1ApiService    api.ProjectsV1APIServicer
			departmentV1ApiService api.DepartmentsV1APIServicer
		)

		BeforeAll(func() {
			clusterId = uuid.New().String()
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   "default",
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
					{
						Name:                   "pool2",
						Id:                     "2",
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false),
					},
				},
			}, nil)

			projectApiService = projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient, nil, nil), &clusterServiceClient, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, services.ProjectDeletionService{})
			departmentV1ApiService = v1.NewDepartmentV1ApiService(&clusterServiceClient)
		})

		AfterAll(func() {
			for _, id := range projIds {
				_, _ = projectApiService.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
			}
			projIds = []string{}
			for _, id := range depIds {
				_, _ = departmentApiService.DeleteDepartment(ctx, id)
			}
			depIds = []string{}
		})

		It("creates department with mixed capabilities (cpu/weights on default, disabled on pool2)", func() {
			req := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			// ensure default has full resources and custom weights
			fullDefault := CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), "default")
			fullDefault.Gpu.OverQuotaWeight = ptr.To(float64(3))
			fullDefault.Cpu.OverQuotaWeight = ptr.To(float64(4))
			fullDefault.Memory.OverQuotaWeight = ptr.To(float64(5))
			// pool2 without cpu/memory and with no explicit oq weight (disabled => equals deserved)
			pool2 := CreateNodePoolApiResources(false, "2", "pool2")
			pool2.Gpu.OverQuotaWeight = nil
			req.Resources = []api.ResourcesNullable{fullDefault, pool2}

			res, err := departmentApiService.CreateDepartment(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusCreated))
			dep := res.Body.(api.Department)
			depIds = append(depIds, dep.Id)

			// validate default keeps weights and cpu/memory, pool2 has only gpu and weight equals deserved
			var defR, np2R api.Resources
			for _, r := range dep.Resources {
				if r.NodePool.Name == "default" {
					defR = r
				}
				if r.NodePool.Name == "pool2" {
					np2R = r
				}
			}
			Expect(defR.Cpu).ToNot(BeNil())
			Expect(defR.Memory).ToNot(BeNil())
			Expect(defR.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*defR.Gpu.OverQuotaWeight).To(Equal(float64(3)))
			Expect(*defR.Cpu.OverQuotaWeight).To(Equal(float64(4)))
			Expect(*defR.Memory.OverQuotaWeight).To(Equal(float64(5)))

			Expect(np2R.Cpu).To(BeNil())
			Expect(np2R.Memory).To(BeNil())
			Expect(np2R.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*np2R.Gpu.OverQuotaWeight).To(Equal(np2R.Gpu.Deserved))
		})

		It("department Get and Update happy flow", func() {
			// create
			depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			depReq.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), "default"),
				CreateNodePoolApiResources(false, "2", "pool2"),
			}
			createRes, err := departmentApiService.CreateDepartment(ctx, depReq)
			Expect(err).To(BeNil())
			Expect(createRes.Code).To(Equal(http.StatusCreated))
			dep := createRes.Body.(api.Department)
			depIds = append(depIds, dep.Id)

			// get by id
			getRes, err := departmentApiService.GetDepartment(ctx, dep.Id)
			Expect(err).To(BeNil())
			Expect(getRes.Code).To(Equal(http.StatusOK))
			got := getRes.Body.(api.Department)
			Expect(got.Id).To(Equal(dep.Id))
			// validate resources on get
			var getDefR, getNp2R api.Resources
			for _, r := range got.Resources {
				if r.NodePool.Name == "default" {
					getDefR = r
				}
				if r.NodePool.Name == "pool2" {
					getNp2R = r
				}
			}
			Expect(getDefR.Cpu).ToNot(BeNil())
			Expect(getDefR.Memory).ToNot(BeNil())
			Expect(getDefR.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(getNp2R.Cpu).To(BeNil())
			Expect(getNp2R.Memory).To(BeNil())
			Expect(getNp2R.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*getNp2R.Gpu.OverQuotaWeight).To(Equal(getNp2R.Gpu.Deserved))

			// validate Sync departments contains consistent per-nodepool resources
			syncRes, err := departmentApiService.GetSyncDepartments(ctx, clusterId, time.Time{}, 0, 0)
			Expect(err).To(BeNil())
			Expect(syncRes.Code).To(Equal(http.StatusOK))
			depsSync := syncRes.Body.(api.GetSyncDepartments200Response).Departments
			var depSync api.DepartmentSync
			for i := range depsSync {
				if depsSync[i].Id == dep.Id {
					depSync = depsSync[i]
				}
			}
			var sDef, sNp2 api.ResourcesSync
			for _, r := range depSync.Resources {
				if r.NodePool != nil && r.NodePool.Name == "default" {
					sDef = r
				}
				if r.NodePool != nil && r.NodePool.Name == "pool2" {
					sNp2 = r
				}
			}
			Expect(sDef.Cpu).ToNot(BeNil())
			Expect(sDef.Memory).ToNot(BeNil())
			Expect(sDef.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(sNp2.Cpu).To(BeNil())
			Expect(sNp2.Memory).To(BeNil())
			Expect(sNp2.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*sNp2.Gpu.OverQuotaWeight).To(Equal(sNp2.Gpu.Deserved))

			// validate V1 list contains consistent per-nodepool resources
			getV1, err := departmentV1ApiService.GetDepartmentsV1(ctx, clusterId, true, false)
			Expect(err).To(BeNil())
			Expect(getV1.Code).To(Equal(http.StatusOK))
			depsV1 := getV1.Body.([]api.DepartmentForGetV1)
			var depV1 api.DepartmentForGetV1
			for i := range depsV1 {
				if fmt.Sprint(depsV1[i].Id) == dep.Id {
					depV1 = depsV1[i]
				}
			}
			Expect(depV1.Name).ToNot(BeEmpty())
			var v1Def, v1Np2 api.NodePoolAssignedResourcesV1Response
			for _, r := range depV1.NodePoolsResources {
				if r.NodePool.Name == "default" {
					v1Def = r
				}
				if r.NodePool.Name == "pool2" {
					v1Np2 = r
				}
			}
			Expect(v1Def.Cpu.Deserved).ToNot(BeNil())
			Expect(v1Def.Memory.Deserved).ToNot(BeNil())
			Expect(v1Def.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(v1Np2.Cpu.Deserved).To(BeNil())
			Expect(v1Np2.Memory.Deserved).To(BeNil())
			Expect(v1Np2.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*v1Np2.Gpu.OverQuotaWeight).To(Equal(*v1Np2.Gpu.Deserved))

			// list
			listRes, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(listRes.Code).To(Equal(http.StatusOK))
			deps := listRes.Body.(api.GetDepartments200Response).Departments
			Expect(len(deps) > 0).To(BeTrue())

			// update default gpu OQ weight and CPU limit
			upd := getDepartmentUpdateRequest(got)
			for i := range upd.Resources {
				if upd.Resources[i].NodePool.Name == "default" {
					upd.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(9))
					upd.Resources[i].Cpu.Limit = ptr.To(float64(5))
				}
			}
			updateRes, err := departmentApiService.UpdateDepartment(ctx, dep.Id, upd)
			Expect(err).To(BeNil())
			Expect(updateRes.Code).To(Equal(http.StatusOK))
			updated := updateRes.Body.(api.Department)
			// validate applied
			var defR api.Resources
			for _, r := range updated.Resources {
				if r.NodePool.Name == "default" {
					defR = r
				}
			}
			Expect(defR.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*defR.Gpu.OverQuotaWeight).To(Equal(float64(9)))
			Expect(defR.Cpu).ToNot(BeNil())
			Expect(*defR.Cpu.Limit).To(Equal(float64(5)))
		})

		It("department Update returns errors for invalid changes per-nodepool", func() {
			depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			depReq.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default"),
				CreateNodePoolApiResources(false, "2", "pool2"),
			}
			createRes, err := departmentApiService.CreateDepartment(ctx, depReq)
			Expect(err).To(BeNil())
			Expect(createRes.Code).To(Equal(http.StatusCreated))
			dep := createRes.Body.(api.Department)
			depIds = append(depIds, dep.Id)

			upd := getDepartmentUpdateRequest(dep)
			// attempt CPU on pool2 where CpuQuota=false
			for i := range upd.Resources {
				if upd.Resources[i].NodePool.Name == "pool2" {
					limit := float64(1)
					upd.Resources[i].Cpu = &api.ResourcesCpu{Deserved: &limit, Limit: &limit}
				}
			}
			res, err := departmentApiService.UpdateDepartment(ctx, dep.Id, upd)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			Expect(res.Body.(api.Error).Message).To(ContainSubstring("CPU memory feature is off"))

			// attempt again with only deserved=-1 - should pass since it's not setting a quota
			upd = getDepartmentUpdateRequest(dep)
			for i := range upd.Resources {
				if upd.Resources[i].NodePool.Name == "pool2" {
					limit := float64(-1)
					upd.Resources[i].Cpu = &api.ResourcesCpu{Deserved: &limit, Limit: nil}
				}
			}
			res, err = departmentApiService.UpdateDepartment(ctx, dep.Id, upd)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusOK))

			// attempt OQ weight change on pool2 where AdjustableWeights=false
			upd = getDepartmentUpdateRequest(dep)
			for i := range upd.Resources {
				if upd.Resources[i].NodePool.Name == "pool2" {
					upd.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(123))
				}
			}
			res, err = departmentApiService.UpdateDepartment(ctx, dep.Id, upd)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			Expect(res.Body.(api.Error).Message).To(ContainSubstring("over quota feature is off"))
		})

		It("project CRUD with mixed capabilities", func() {
			// create dep as parent
			depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			depReq.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), "default"),
				CreateNodePoolApiResources(false, "2", "pool2"),
			}
			depRes, err := departmentApiService.CreateDepartment(ctx, depReq)
			Expect(err).To(BeNil())
			Expect(depRes.Code).To(Equal(http.StatusCreated))
			dep := depRes.Body.(api.Department)
			depIds = append(depIds, dep.Id)

			// create project obeying capabilities
			pReq := getProjectCreateRequest(createProjectName(), "", dep.Id, clusterId)
			// default full, pool2 gpu-only
			pReq.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), "default"),
				CreateNodePoolApiResources(false, "2", "pool2"),
			}
			pRes, err := projectApiService.CreateProject(ctx, pReq)
			Expect(err).To(BeNil())
			Expect(pRes.Code).To(Equal(http.StatusCreated))
			proj := pRes.Body.(api.Project)
			projIds = append(projIds, proj.Id)

			// get by id
			gp, err := projectApiService.GetProject(ctx, proj.Id)
			Expect(err).To(BeNil())
			Expect(gp.Code).To(Equal(http.StatusOK))
			projGot := gp.Body.(api.Project)
			Expect(projGot.Id).To(Equal(proj.Id))
			// validate resources on get
			var pDefR, pNp2R api.Resources
			for _, r := range projGot.Resources {
				if r.NodePool.Name == "default" {
					pDefR = r
				}
				if r.NodePool.Name == "pool2" {
					pNp2R = r
				}
			}
			Expect(pDefR.Cpu).ToNot(BeNil())
			Expect(pDefR.Memory).ToNot(BeNil())
			Expect(pDefR.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(pNp2R.Cpu).To(BeNil())
			Expect(pNp2R.Memory).To(BeNil())
			Expect(pNp2R.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*pNp2R.Gpu.OverQuotaWeight).To(Equal(pNp2R.Gpu.Deserved))

			// validate V1 list contains consistent per-nodepool resources
			gpV1, err := projectV1ApiService.GetProjectsV1(ctx, clusterId, true, true)
			Expect(err).To(BeNil())
			Expect(gpV1.Code).To(Equal(http.StatusOK))
			projsV1 := gpV1.Body.([]api.ProjectV1)
			var projV1 api.ProjectV1
			for i := range projsV1 {
				if fmt.Sprint(projsV1[i].Id) == proj.Id {
					projV1 = projsV1[i]
				}
			}
			var pv1Def, pv1Np2 api.NodePoolAssignedResourcesV1Response
			for _, r := range projV1.NodePoolsResources {
				if r.NodePool.Name == "default" {
					pv1Def = r
				}
				if r.NodePool.Name == "pool2" {
					pv1Np2 = r
				}
			}
			Expect(pv1Def.Cpu.Deserved).ToNot(BeNil())
			Expect(pv1Def.Memory.Deserved).ToNot(BeNil())
			Expect(pv1Def.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(pv1Np2.Cpu.Deserved).To(BeNil())
			Expect(pv1Np2.Memory.Deserved).To(BeNil())
			Expect(pv1Np2.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*pv1Np2.Gpu.OverQuotaWeight).To(Equal(*pv1Np2.Gpu.Deserved))

			// validate Sync projects contains consistent per-nodepool resources
			syncP, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, false)
			Expect(err).To(BeNil())
			Expect(syncP.Code).To(Equal(http.StatusOK))
			projsSync := syncP.Body.(api.GetSyncProjects200Response).Projects
			var projSync api.ProjectSync
			for i := range projsSync {
				if projsSync[i].Id == proj.Id {
					projSync = projsSync[i]
				}
			}
			var psDef, psNp2 api.ResourcesSync
			for _, r := range projSync.Resources {
				if r.NodePool != nil && r.NodePool.Name == "default" {
					psDef = r
				}
				if r.NodePool != nil && r.NodePool.Name == "pool2" {
					psNp2 = r
				}
			}
			Expect(psDef.Cpu).ToNot(BeNil())
			Expect(psDef.Memory).ToNot(BeNil())
			Expect(psDef.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(psNp2.Cpu).To(BeNil())
			Expect(psNp2.Memory).To(BeNil())
			Expect(psNp2.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*psNp2.Gpu.OverQuotaWeight).To(Equal(psNp2.Gpu.Deserved))

			// list
			gl, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(gl.Code).To(Equal(http.StatusOK))
			Expect(len(gl.Body.(api.GetProjects200Response).Projects) > 0).To(BeTrue())

			// update happy path - change default gpu oq weight
			pu := getProjectUpdateRequest(proj)
			for i := range pu.Resources {
				if pu.Resources[i].NodePool.Name == "default" {
					pu.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(11))
				}
			}
			puRes, err := projectApiService.UpdateProject(ctx, proj.Id, pu)
			Expect(err).To(BeNil())
			Expect(puRes.Code).To(Equal(http.StatusOK))
			updated := puRes.Body.(api.Project)
			var dR api.Resources
			for _, r := range updated.Resources {
				if r.NodePool.Name == "default" {
					dR = r
				}
			}
			Expect(*dR.Gpu.OverQuotaWeight).To(Equal(float64(11)))

			// update error: CPU on pool2
			pu = getProjectUpdateRequest(updated)
			for i := range pu.Resources {
				if pu.Resources[i].NodePool.Name == "pool2" {
					val := float64(1)
					pu.Resources[i].Cpu = &api.ResourcesCpu{Deserved: &val, Limit: &val}
				}
			}
			puErrRes, err := projectApiService.UpdateProject(ctx, proj.Id, pu)
			Expect(err).To(BeNil())
			Expect(puErrRes.Code).To(Equal(http.StatusBadRequest))
			Expect(puErrRes.Body.(api.Error).Message).To(ContainSubstring("CPU memory feature is off"))

			// update error: OQ weight change on pool2
			pu = getProjectUpdateRequest(updated)
			for i := range pu.Resources {
				if pu.Resources[i].NodePool.Name == "pool2" {
					pu.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(77))
				}
			}
			puErrRes, err = projectApiService.UpdateProject(ctx, proj.Id, pu)
			Expect(err).To(BeNil())
			Expect(puErrRes.Code).To(Equal(http.StatusBadRequest))
			Expect(puErrRes.Body.(api.Error).Message).To(ContainSubstring("over quota feature is off"))
		})

		It("create dep fails when assigning CPU on a nodepool with CpuQuota disabled", func() {
			req := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			// default ok, pool2 mistakenly includes cpu
			def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
			pool2 := CreateNodePoolApiResources(true, "2", "pool2") // CPU present where disabled
			req.Resources = []api.ResourcesNullable{def, pool2}

			res, err := departmentApiService.CreateDepartment(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			Expect(res.Body.(api.Error).Message).To(ContainSubstring("CPU memory feature is off"))
		})

		It("create dep fails when setting over-quota weight on a nodepool with AdjustableWeights disabled", func() {
			req := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
			pool2 := CreateNodePoolApiResources(false, "2", "pool2")
			// set OQ weight different than deserved on np2 where adjustable weights are disabled
			pool2.Gpu.OverQuotaWeight = ptr.To(float64(999))
			req.Resources = []api.ResourcesNullable{def, pool2}

			res, err := departmentApiService.CreateDepartment(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusBadRequest))
			Expect(res.Body.(api.Error).Message).To(ContainSubstring("over quota feature is off"))
		})

		It("create project fails when assigning CPU on a nodepool with CpuQuota disabled", func() {
			// create parent dep
			depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			depReq.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), "default"),
				CreateNodePoolApiResources(false, "2", "pool2"),
			}
			depRes, err := departmentApiService.CreateDepartment(ctx, depReq)
			Expect(err).To(BeNil())
			Expect(depRes.Code).To(Equal(http.StatusCreated))
			dep := depRes.Body.(api.Department)
			depIds = append(depIds, dep.Id)

			pReq := getProjectCreateRequest(createProjectName(), "", dep.Id, clusterId)
			def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
			pool2 := CreateNodePoolApiResources(true, "2", "pool2") // CPU present where disabled
			pReq.Resources = []api.ResourcesNullable{def, pool2}

			pRes, err := projectApiService.CreateProject(ctx, pReq)
			Expect(err).To(BeNil())
			Expect(pRes.Code).To(Equal(http.StatusBadRequest))
			Expect(pRes.Body.(api.Error).Message).To(ContainSubstring("CPU memory feature is off"))
		})

		It("create project fails when setting over-quota weight on a nodepool with AdjustableWeights disabled", func() {
			// create parent dep
			depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			depReq.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), "default"),
				CreateNodePoolApiResources(false, "2", "pool2"),
			}
			depRes, err := departmentApiService.CreateDepartment(ctx, depReq)
			Expect(err).To(BeNil())
			Expect(depRes.Code).To(Equal(http.StatusCreated))
			dep := depRes.Body.(api.Department)
			depIds = append(depIds, dep.Id)

			pReq := getProjectCreateRequest(createProjectName(), "", dep.Id, clusterId)
			def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
			pool2 := CreateNodePoolApiResources(false, "2", "pool2")
			pool2.Gpu.OverQuotaWeight = ptr.To(float64(999))
			pReq.Resources = []api.ResourcesNullable{def, pool2}

			pRes, err := projectApiService.CreateProject(ctx, pReq)
			Expect(err).To(BeNil())
			Expect(pRes.Code).To(Equal(http.StatusBadRequest))
			Expect(pRes.Body.(api.Error).Message).To(ContainSubstring("over quota feature is off"))
		})

		Context("test quota/ranks capabilities", func() {
			var (
				clusterId2            string
				projectApiService2    api.ProjectsAPIServicer
				departmentApiService2 api.DepartmentsAPIServicer
				defaultDepId          string
			)

			BeforeAll(func() {
				clusterId2 = uuid.New().String()
				clusterServiceClient2 := mocks.IClusterClient{}
				clusterServiceClient2.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					Uuid:    clusterId2,
					Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				}, nil)
				clusterServiceClient2.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   "default",
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: npSchedCapabilities(false, false, true, true),
						},
						{
							Name:                   "pool2",
							Id:                     "2",
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: npSchedCapabilities(false, false, false, false),
						},
					},
				}, nil)

				projectApiService2 = projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient2, nil, nil), &clusterServiceClient2, nil)
				departmentApiService2 = departments.NewDepartmentApiService(&clusterServiceClient2, nil)

				// create parent dep
				depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				depReq.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default"),
					CreateNodePoolApiResourcesNoQuota("2", "pool2"),
				}
				depRes, err := departmentApiService.CreateDepartment(ctx, depReq)
				Expect(err).To(BeNil())
				Expect(depRes.Code).To(Equal(http.StatusCreated))
				dep := depRes.Body.(api.Department)
				depIds = append(depIds, dep.Id)
				defaultDepId = dep.Id
			})

			AfterAll(func() {
				for _, id := range projIds {
					_, _ = projectApiService2.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
				}
				projIds = []string{}
				for _, id := range depIds {
					_, _ = departmentApiService2.DeleteDepartment(ctx, id)
				}
				depIds = []string{}
			})

			It("create department/project fails when quota is disabled on pool2 but GPU quota is set", func() {
				// default ok, pool3 has Quota=false in mock and should reject gpu quota
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				pool2 := CreateNodePoolApiResources(false, "2", "pool2")

				// department
				req := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				req.Resources = []api.ResourcesNullable{def, pool2}

				res, err := departmentApiService2.CreateDepartment(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(res.Body.(api.Error).Message).To(ContainSubstring("quota is not enabled on nodepool id <2>"))

				// project
				reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				reqP.Resources = []api.ResourcesNullable{def, pool2}

				resP, err := projectApiService2.CreateProject(ctx, reqP)
				Expect(err).To(BeNil())
				Expect(resP.Code).To(Equal(http.StatusBadRequest))
				Expect(resP.Body.(api.Error).Message).To(ContainSubstring("quota is not enabled on nodepool id <2>"))
			})

			It("create department/project fails when quota is enabled and gpu deserved is omitted", func() {
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				def.Gpu.Deserved = nil
				pool2 := CreateNodePoolApiResourcesNoQuota("2", "pool2")
				expectedMessage := fmt.Sprintf("quota is enabled on nodepool id <%s>. GPU deserved is required", fmt.Sprint(defaultNodePoolId))

				req := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				req.Resources = []api.ResourcesNullable{def, pool2}
				res, err := departmentApiService2.CreateDepartment(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(res.Body.(api.Error).Message).To(ContainSubstring(expectedMessage))

				reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				reqP.Resources = []api.ResourcesNullable{def, pool2}
				resP, err := projectApiService2.CreateProject(ctx, reqP)
				Expect(err).To(BeNil())
				Expect(resP.Code).To(Equal(http.StatusBadRequest))
				Expect(resP.Body.(api.Error).Message).To(ContainSubstring(expectedMessage))
			})

			It("create project accepts an omitted gpu on a node pool with quota disabled", func() {
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				pool2 := CreateNodePoolApiResourcesNoQuota("2", "pool2")
				pool2.Gpu = api.ResourcesNullableGpu{}

				reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				reqP.Resources = []api.ResourcesNullable{def, pool2}
				resP, err := projectApiService2.CreateProject(ctx, reqP)
				Expect(err).To(BeNil())
				Expect(resP.Code).To(Equal(http.StatusCreated), fmt.Sprintf("got body: %+v", resP.Body))
				proj := resP.Body.(api.Project)
				projIds = append(projIds, proj.Id)
				var pool2Resources *api.Resources
				for i := range proj.Resources {
					if proj.Resources[i].NodePool != nil && proj.Resources[i].NodePool.Id == "2" {
						pool2Resources = &proj.Resources[i]
					}
				}
				Expect(pool2Resources).ToNot(BeNil())
				Expect(pool2Resources.Gpu.Deserved).To(Equal(float64(0)))
			})

			It("update project and put and patch project resources fail when quota is enabled and gpu deserved is omitted", func() {
				reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				reqP.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default"),
					CreateNodePoolApiResourcesNoQuota("2", "pool2"),
				}
				resP, err := projectApiService2.CreateProject(ctx, reqP)
				Expect(err).To(BeNil())
				Expect(resP.Code).To(Equal(http.StatusCreated))
				proj := resP.Body.(api.Project)
				projIds = append(projIds, proj.Id)

				noDeserved := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				noDeserved.Gpu.Deserved = nil
				expectedMessage := fmt.Sprintf("quota is enabled on nodepool id <%s>. GPU deserved is required", fmt.Sprint(defaultNodePoolId))

				patchRes, err := projectApiService2.PatchProjectResources(ctx, proj.Id, []api.ResourcesNullable{noDeserved})
				Expect(err).To(BeNil())
				Expect(patchRes.Code).To(Equal(http.StatusBadRequest))
				Expect(patchRes.Body.(api.Error).Message).To(ContainSubstring(expectedMessage))

				putRes, err := projectApiService2.UpdateProjectResources(ctx, proj.Id, []api.ResourcesNullable{noDeserved, CreateNodePoolApiResourcesNoQuota("2", "pool2")})
				Expect(err).To(BeNil())
				Expect(putRes.Code).To(Equal(http.StatusBadRequest))
				Expect(putRes.Body.(api.Error).Details).To(ContainSubstring(expectedMessage))

				updateReq := getProjectUpdateRequest(proj)
				for i := range updateReq.Resources {
					if updateReq.Resources[i].NodePool.Id == fmt.Sprint(defaultNodePoolId) {
						updateReq.Resources[i].Gpu.Deserved = nil
					}
				}
				updateRes, err := projectApiService2.UpdateProject(ctx, proj.Id, updateReq)
				Expect(err).To(BeNil())
				Expect(updateRes.Code).To(Equal(http.StatusBadRequest))
				Expect(updateRes.Body.(api.Error).Message).To(ContainSubstring(expectedMessage))
			})

			It("update department and put and patch department resources fail when quota is enabled and gpu deserved is omitted", func() {
				depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				depReq.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default"),
					CreateNodePoolApiResourcesNoQuota("2", "pool2"),
				}
				depRes, err := departmentApiService2.CreateDepartment(ctx, depReq)
				Expect(err).To(BeNil())
				Expect(depRes.Code).To(Equal(http.StatusCreated))
				dep := depRes.Body.(api.Department)
				depIds = append(depIds, dep.Id)

				noDeserved := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				noDeserved.Gpu.Deserved = nil
				expectedMessage := fmt.Sprintf("quota is enabled on nodepool id <%s>. GPU deserved is required", fmt.Sprint(defaultNodePoolId))

				patchRes, err := departmentApiService2.PatchDepartmentResources(ctx, dep.Id, []api.ResourcesNullable{noDeserved})
				Expect(err).To(BeNil())
				Expect(patchRes.Code).To(Equal(http.StatusBadRequest))
				Expect(patchRes.Body.(api.Error).Details).To(ContainSubstring(expectedMessage))

				putRes, err := departmentApiService2.UpdateDepartmentResources(ctx, dep.Id, []api.ResourcesNullable{noDeserved, CreateNodePoolApiResourcesNoQuota("2", "pool2")})
				Expect(err).To(BeNil())
				Expect(putRes.Code).To(Equal(http.StatusBadRequest))
				Expect(putRes.Body.(api.Error).Details).To(ContainSubstring(expectedMessage))

				updateReq := getDepartmentUpdateRequest(dep)
				for i := range updateReq.Resources {
					if updateReq.Resources[i].NodePool.Id == fmt.Sprint(defaultNodePoolId) {
						updateReq.Resources[i].Gpu.Deserved = nil
					}
				}
				updateRes, err := departmentApiService2.UpdateDepartment(ctx, dep.Id, updateReq)
				Expect(err).To(BeNil())
				Expect(updateRes.Code).To(Equal(http.StatusBadRequest))
				Expect(updateRes.Body.(api.Error).Message).To(ContainSubstring(expectedMessage))
			})

			It("batch patch without resources keeps the stored gpu deserved", func() {
				reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				reqP.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default"),
					CreateNodePoolApiResourcesNoQuota("2", "pool2"),
				}
				resP, err := projectApiService2.CreateProject(ctx, reqP)
				Expect(err).To(BeNil())
				Expect(resP.Code).To(Equal(http.StatusCreated))
				proj := resP.Body.(api.Project)
				projIds = append(projIds, proj.Id)

				batchRes, err := projectApiService2.BatchPatchProjectNodePoolResources(ctx, api.BatchPatchNodePoolResourcesRequest{
					PatchRequests: []api.NodePoolResourcesPatchRequest{{
						OrgUnitId:  proj.Id,
						NodePoolId: fmt.Sprint(defaultNodePoolId),
						Rank:       ptr.To("High"),
					}},
				})
				Expect(err).To(BeNil())
				Expect(batchRes.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", batchRes.Body))
				body := batchRes.Body.(api.BatchPatchProjectResourcesResponse)
				Expect(body.TotalFailed).To(Equal(int32(0)), fmt.Sprintf("failed: %+v", body.Failed))
				Expect(body.Succeeded).To(HaveLen(1))
				Expect(body.Succeeded[0].Resources.Gpu.Deserved).ToNot(BeNil())
				Expect(*body.Succeeded[0].Resources.Gpu.Deserved).To(Equal(FullResourcesDeservedGpu))
			})

			Context("batch patch with gpu but without gpu deserved", func() {
				var projId string

				BeforeEach(func() {
					reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
					reqP.Resources = []api.ResourcesNullable{
						CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default"),
						CreateNodePoolApiResourcesNoQuota("2", "pool2"),
					}
					resP, err := projectApiService2.CreateProject(ctx, reqP)
					Expect(err).To(BeNil())
					Expect(resP.Code).To(Equal(http.StatusCreated))
					projId = resP.Body.(api.Project).Id
					projIds = append(projIds, projId)
				})

				batchPatchProjectGpu := func(gpu api.ResourcesPerNodePoolNullableGpu) (int, api.BatchPatchProjectResourcesResponse) {
					batchRes, err := projectApiService2.BatchPatchProjectNodePoolResources(ctx, api.BatchPatchNodePoolResourcesRequest{
						PatchRequests: []api.NodePoolResourcesPatchRequest{{
							OrgUnitId:  projId,
							NodePoolId: fmt.Sprint(defaultNodePoolId),
							Resources:  &api.ResourcesPerNodePoolNullable{Gpu: &gpu},
						}},
					})
					Expect(err).To(BeNil())
					return batchRes.Code, batchRes.Body.(api.BatchPatchProjectResourcesResponse)
				}

				It("keeps the stored gpu deserved when only the limit is sent", func() {
					code, body := batchPatchProjectGpu(api.ResourcesPerNodePoolNullableGpu{Limit: ptr.To(float64(2))})
					Expect(code).To(Equal(http.StatusOK))
					Expect(body.TotalFailed).To(Equal(int32(0)), fmt.Sprintf("failed: %+v", body.Failed))
					Expect(body.Succeeded).To(HaveLen(1))
					Expect(body.Succeeded[0].Resources.Gpu.Deserved).ToNot(BeNil())
					Expect(*body.Succeeded[0].Resources.Gpu.Deserved).To(Equal(FullResourcesDeservedGpu))
				})

				It("checks the over quota weight against the stored gpu deserved", func() {
					code, body := batchPatchProjectGpu(api.ResourcesPerNodePoolNullableGpu{OverQuotaWeight: ptr.To(FullResourcesDeservedGpu)})
					Expect(code).To(Equal(http.StatusOK))
					Expect(body.TotalFailed).To(Equal(int32(0)), fmt.Sprintf("failed: %+v", body.Failed))
					Expect(*body.Succeeded[0].Resources.Gpu.Deserved).To(Equal(FullResourcesDeservedGpu))

					// a batch where every request fails returns 500
					code, body = batchPatchProjectGpu(api.ResourcesPerNodePoolNullableGpu{OverQuotaWeight: ptr.To(services.MinOverQuotaWeight)})
					Expect(code).To(Equal(http.StatusInternalServerError))
					Expect(body.TotalFailed).To(Equal(int32(1)))
					Expect(body.Failed[0].Message).To(ContainSubstring("OverQuotaWeight should be deserved or empty"))
				})

				It("checks a department's over quota weight against its stored gpu deserved", func() {
					batchRes, err := departmentApiService2.BatchPatchDepartmentNodePoolResources(ctx, api.BatchPatchNodePoolResourcesRequest{
						PatchRequests: []api.NodePoolResourcesPatchRequest{{
							OrgUnitId:  defaultDepId,
							NodePoolId: fmt.Sprint(defaultNodePoolId),
							Resources: &api.ResourcesPerNodePoolNullable{
								Gpu: &api.ResourcesPerNodePoolNullableGpu{OverQuotaWeight: ptr.To(FullResourcesDeservedGpu)},
							},
						}},
					})
					Expect(err).To(BeNil())
					Expect(batchRes.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", batchRes.Body))
					body := batchRes.Body.(api.BatchPatchDepartmentResourcesResponse)
					Expect(body.TotalFailed).To(Equal(int32(0)), fmt.Sprintf("failed: %+v", body.Failed))
					Expect(*body.Succeeded[0].Resources.Gpu.Deserved).To(Equal(FullResourcesDeservedGpu))
				})
			})

			It("create department/project fails when ranks are disabled on pool2 but non-default priority provided", func() {
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				pool2 := CreateNodePoolApiResourcesNoQuota("2", "pool2")

				// department
				req := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				pool2.Rank = ptr.To("High")
				req.Resources = []api.ResourcesNullable{def, pool2}

				res, err := departmentApiService2.CreateDepartment(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(res.Body.(api.Error).Message).To(ContainSubstring("ranks is not enabled on nodepool id <2>"))

				// project
				reqP := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				pool2.Rank = ptr.To("High")
				reqP.Resources = []api.ResourcesNullable{def, pool2}

				resP, err := projectApiService2.CreateProject(ctx, reqP)
				Expect(err).To(BeNil())
				Expect(resP.Code).To(Equal(http.StatusBadRequest))
				Expect(resP.Body.(api.Error).Message).To(ContainSubstring("ranks is not enabled on nodepool id <2>"))
			})

			It("update returns errors when quota/ranks are disabled on pool2", func() {
				// create a valid department (default has quota+ranks enabled, pool2 has both disabled)
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				pool2NoQuotaNoRanks := CreateNodePoolApiResourcesNoQuota("2", "pool2")
				depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				depReq.Resources = []api.ResourcesNullable{def, pool2NoQuotaNoRanks}
				depCreate, err := departmentApiService2.CreateDepartment(ctx, depReq)
				Expect(err).To(BeNil())
				Expect(depCreate.Code).To(Equal(http.StatusCreated))
				dep := depCreate.Body.(api.Department)
				depIds = append(depIds, dep.Id)

				// Act+Assert: try to enable quota on pool2 via update -> expect quota disabled error
				depUpd := getDepartmentUpdateRequest(dep)
				for i := range depUpd.Resources {
					if depUpd.Resources[i].NodePool.Name == "pool2" {
						// attempt to set GPU quota on a nodepool with quota disabled
						depUpd.Resources[i].Gpu.Deserved = ptr.To(float64(1))
					}
				}
				depUpdRes, err := departmentApiService2.UpdateDepartment(ctx, dep.Id, depUpd)
				Expect(err).To(BeNil())
				Expect(depUpdRes.Code).To(Equal(http.StatusBadRequest))
				Expect(depUpdRes.Body.(api.Error).Message).To(ContainSubstring("quota is not enabled on nodepool id <2>"))

				// Act+Assert: try to set non-default rank on pool2 via update -> expect ranks disabled error
				depUpd = getDepartmentUpdateRequest(dep)
				for i := range depUpd.Resources {
					if depUpd.Resources[i].NodePool.Name == "pool2" {
						depUpd.Resources[i].Rank = ptr.To("High")
					}
				}
				depUpdRes2, err := departmentApiService2.UpdateDepartment(ctx, dep.Id, depUpd)
				Expect(err).To(BeNil())
				Expect(depUpdRes2.Code).To(Equal(http.StatusBadRequest))
				Expect(depUpdRes2.Body.(api.Error).Message).To(ContainSubstring("ranks is not enabled on nodepool id <2>"))

				// create a valid project under the defaultDepId
				projReq := getProjectCreateRequest(createProjectName(), "", defaultDepId, clusterId2)
				projReq.Resources = []api.ResourcesNullable{def, pool2NoQuotaNoRanks}
				projCreate, err := projectApiService2.CreateProject(ctx, projReq)
				Expect(err).To(BeNil())
				Expect(projCreate.Code).To(Equal(http.StatusCreated))
				proj := projCreate.Body.(api.Project)
				projIds = append(projIds, proj.Id)

				// Act+Assert: project update - enable quota on pool2 -> expect error
				projUpd := getProjectUpdateRequest(proj)
				for i := range projUpd.Resources {
					if projUpd.Resources[i].NodePool.Name == "pool2" {
						projUpd.Resources[i].Gpu.Deserved = ptr.To(float64(1))
					}
				}
				projUpdRes, err := projectApiService2.UpdateProject(ctx, proj.Id, projUpd)
				Expect(err).To(BeNil())
				Expect(projUpdRes.Code).To(Equal(http.StatusBadRequest))
				Expect(projUpdRes.Body.(api.Error).Message).To(ContainSubstring("quota is not enabled on nodepool id <2>"))

				// Act+Assert: project update - set non-default rank on pool2 -> expect error
				projUpd = getProjectUpdateRequest(proj)
				for i := range projUpd.Resources {
					if projUpd.Resources[i].NodePool.Name == "pool2" {
						projUpd.Resources[i].Rank = ptr.To("High")
					}
				}
				projUpdRes2, err := projectApiService2.UpdateProject(ctx, proj.Id, projUpd)
				Expect(err).To(BeNil())
				Expect(projUpdRes2.Code).To(Equal(http.StatusBadRequest))
				Expect(projUpdRes2.Body.(api.Error).Message).To(ContainSubstring("ranks is not enabled on nodepool id <2>"))
			})

			It("department create+update succeed when quota and ranks are enabled", func() {
				// Create with quota+ranks on default and neither on pool2
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				// set non-default priority to validate ranks on enabled nodepool
				depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				depReq.Resources = []api.ResourcesNullable{def, CreateNodePoolApiResourcesNoQuota("2", "pool2")}
				depReq.Resources[0].Rank = ptr.To("High")

				createRes, err := departmentApiService2.CreateDepartment(ctx, depReq)
				Expect(err).To(BeNil())
				Expect(createRes.Code).To(Equal(http.StatusCreated))
				dep := createRes.Body.(api.Department)
				depIds = append(depIds, dep.Id)

				// Update priority on default (ranks enabled) -> expect success
				upd := getDepartmentUpdateRequest(dep)
				for i := range upd.Resources {
					if upd.Resources[i].NodePool.Name == "default" {
						upd.Resources[i].Rank = ptr.To("VeryLow")
					}
				}
				updateRes, err := departmentApiService2.UpdateDepartment(ctx, dep.Id, upd)
				Expect(err).To(BeNil())
				Expect(updateRes.Code).To(Equal(http.StatusOK))
				updated := updateRes.Body.(api.Department)
				var defR api.Resources
				for _, r := range updated.Resources {
					if r.NodePool.Name == "default" {
						defR = r
					}
				}
				Expect(defR.Rank).ToNot(BeNil())
				Expect(*defR.Rank).To(Equal("VeryLow"))
			})

			It("project create+update succeed when quota and ranks are enabled", func() {
				// Create parent dep
				def := CreateNodePoolApiResources(false, fmt.Sprint(defaultNodePoolId), "default")
				depReq := getDepartmentCreateRequest(createDepartmentName(), clusterId2)
				depReq.Resources = []api.ResourcesNullable{def, CreateNodePoolApiResourcesNoQuota("2", "pool2")}
				depRes, err := departmentApiService2.CreateDepartment(ctx, depReq)
				Expect(err).To(BeNil())
				Expect(depRes.Code).To(Equal(http.StatusCreated))
				dep := depRes.Body.(api.Department)
				depIds = append(depIds, dep.Id)

				// Create project with non-default priority on default (ranks enabled)
				pReq := getProjectCreateRequest(createProjectName(), "", dep.Id, clusterId2)
				pReq.Resources = []api.ResourcesNullable{def, CreateNodePoolApiResourcesNoQuota("2", "pool2")}
				pReq.Resources[0].Rank = ptr.To("High")
				pRes, err := projectApiService2.CreateProject(ctx, pReq)
				Expect(err).To(BeNil())
				Expect(pRes.Code).To(Equal(http.StatusCreated))
				proj := pRes.Body.(api.Project)
				projIds = append(projIds, proj.Id)

				// Update priority on default -> expect success
				upd := getProjectUpdateRequest(proj)
				for i := range upd.Resources {
					if upd.Resources[i].NodePool.Name == "default" {
						upd.Resources[i].Rank = ptr.To("VeryLow")
					}
				}
				updRes, err := projectApiService2.UpdateProject(ctx, proj.Id, upd)
				Expect(err).To(BeNil())
				Expect(updRes.Code).To(Equal(http.StatusOK))
				updated := updRes.Body.(api.Project)
				var defR api.Resources
				for _, r := range updated.Resources {
					if r.NodePool.Name == "default" {
						defR = r
					}
				}
				Expect(defR.Rank).ToNot(BeNil())
				Expect(*defR.Rank).To(Equal("VeryLow"))
			})
		})

		Context("multi-cluster per-nodepool combinations", func() {
			var (
				clusterB           string
				localDepIds        []string
				localProjIds       []string
				localClusterClient mocks.IClusterClient
			)

			BeforeAll(func() {
				clusterB = uuid.New().String()
				localClusterClient = mocks.IClusterClient{}
				// return version for any cluster
				localClusterClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					Uuid:    clusterB,
					Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				}, nil)
				// cluster A nodepools: (true,false) and (false,true)
				localClusterClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{Name: "a-default", Id: "1", Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false)},
						{Name: "a-pool2", Id: "2", Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: NpSchedCapabilitiesDefault(false, true)},
					},
				}, nil)
				// cluster B nodepools: (true,true) and (false,false) [and for all other clusters as well]
				localClusterClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.MatchedBy(func(reqClusterId string) bool {
					return reqClusterId != clusterId
				})).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{Name: "b-default", Id: "3", Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true)},
						{Name: "b-pool2", Id: "4", Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: NpSchedCapabilitiesDefault(false, false)},
					},
				}, nil)

				// bind services to local mock
				projectApiService = projects.NewProjectApiService(services.NewProjectDeletionService(&localClusterClient, nil, nil), &localClusterClient, nil)
				departmentApiService = departments.NewDepartmentApiService(&localClusterClient, nil)
			})

			AfterAll(func() {
				for _, id := range localProjIds {
					_, _ = projectApiService.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
				}
				for _, id := range localDepIds {
					_, _ = departmentApiService.DeleteDepartment(ctx, id)
				}
				localProjIds = []string{}
				localDepIds = []string{}
			})

			It("validates list responses per cluster and per-nodepool capabilities", func() {
				// Create departments and projects in cluster A
				depReqA := getDepartmentCreateRequest(createDepartmentName(), clusterId)
				depReqA.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(true, "1", "a-default"), // cpu=true, weights=false
					CreateNodePoolApiResources(false, "2", "a-pool2"),  // cpu=false, weights=true
				}
				depResA, err := departmentApiService.CreateDepartment(ctx, depReqA)
				Expect(err).To(BeNil())
				Expect(depResA.Code).To(Equal(http.StatusCreated))
				depA := depResA.Body.(api.Department)
				localDepIds = append(localDepIds, depA.Id)

				projReqA := getProjectCreateRequest(createProjectName(), "", depA.Id, clusterId)
				projReqA.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(true, "1", "a-default"),
					CreateNodePoolApiResources(false, "2", "a-pool2"),
				}
				projResA, err := projectApiService.CreateProject(ctx, projReqA)
				Expect(err).To(BeNil())
				Expect(projResA.Code).To(Equal(http.StatusCreated))
				projA := projResA.Body.(api.Project)
				localProjIds = append(localProjIds, projA.Id)

				// Create departments and projects in cluster B
				depReqB := getDepartmentCreateRequest(createDepartmentName(), clusterB)
				depReqB.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(true, "3", "b-default"), // cpu=true, weights=true
					CreateNodePoolApiResources(false, "4", "b-pool2"),  // cpu=false, weights=false
				}
				depResB, err := departmentApiService.CreateDepartment(ctx, depReqB)
				Expect(err).To(BeNil())
				Expect(depResB.Code).To(Equal(http.StatusCreated))
				depB := depResB.Body.(api.Department)
				localDepIds = append(localDepIds, depB.Id)

				projReqB := getProjectCreateRequest(createProjectName(), "", depB.Id, clusterB)
				projReqB.Resources = []api.ResourcesNullable{
					CreateNodePoolApiResources(true, "3", "b-default"),
					CreateNodePoolApiResources(false, "4", "b-pool2"),
				}
				projResB, err := projectApiService.CreateProject(ctx, projReqB)
				Expect(err).To(BeNil())
				Expect(projResB.Code).To(Equal(http.StatusCreated))
				projB := projResB.Body.(api.Project)
				localProjIds = append(localProjIds, projB.Id)

				// Validate departments list without filters (all clusters)
				depListAll, err := departmentApiService.GetDepartments(ctx, []string{}, "", "", "", 0, 0)
				Expect(err).To(BeNil())
				Expect(depListAll.Code).To(Equal(http.StatusOK))
				depsAll := depListAll.Body.(api.GetDepartments200Response).Departments
				var depAList, depBList api.DepartmentForList
				for _, d := range depsAll {
					if d.Id == depA.Id {
						depAList = d
					}
					if d.Id == depB.Id {
						depBList = d
					}
				}
				var aDef, aNp2 api.Resources
				for _, r := range depAList.Resources {
					if r.NodePool.Name == "a-default" {
						aDef = r
					}
					if r.NodePool.Name == "a-pool2" {
						aNp2 = r
					}
				}
				// a-default: cpu on, weights off
				Expect(aDef.Cpu).ToNot(BeNil())
				Expect(aDef.Memory).ToNot(BeNil())
				Expect(aDef.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*aDef.Gpu.OverQuotaWeight).To(Equal(aDef.Gpu.Deserved))
				// a-pool2: cpu off, weights on
				Expect(aNp2.Cpu).To(BeNil())
				Expect(aNp2.Memory).To(BeNil())
				Expect(aNp2.Gpu.OverQuotaWeight).ToNot(BeNil())

				var bDef, bNp2 api.Resources
				for _, r := range depBList.Resources {
					if r.NodePool.Name == "b-default" {
						bDef = r
					}
					if r.NodePool.Name == "b-pool2" {
						bNp2 = r
					}
				}
				// b-default: cpu on, weights on
				Expect(bDef.Cpu).ToNot(BeNil())
				Expect(bDef.Memory).ToNot(BeNil())
				Expect(bDef.Gpu.OverQuotaWeight).ToNot(BeNil())
				// b-pool2: cpu off, weights off
				Expect(bNp2.Cpu).To(BeNil())
				Expect(bNp2.Memory).To(BeNil())
				Expect(bNp2.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*bNp2.Gpu.OverQuotaWeight).To(Equal(bNp2.Gpu.Deserved))

				// Validate projects list without filters (all clusters)
				projListAll, err := projectApiService.GetProjects(ctx, []string{}, "", "", 0, 0, "")
				Expect(err).To(BeNil())
				Expect(projListAll.Code).To(Equal(http.StatusOK))
				projsAll := projListAll.Body.(api.GetProjects200Response).Projects
				var projAList, projBList api.Project
				for _, p := range projsAll {
					if p.Id == projA.Id {
						projAList = p
					}
					if p.Id == projB.Id {
						projBList = p
					}
				}
				var paDef, paNp2 api.Resources
				for _, r := range projAList.Resources {
					if r.NodePool.Name == "a-default" {
						paDef = r
					}
					if r.NodePool.Name == "a-pool2" {
						paNp2 = r
					}
				}
				Expect(paDef.Cpu).ToNot(BeNil())
				Expect(paDef.Memory).ToNot(BeNil())
				Expect(paDef.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*paDef.Gpu.OverQuotaWeight).To(Equal(paDef.Gpu.Deserved))
				Expect(paNp2.Cpu).To(BeNil())
				Expect(paNp2.Memory).To(BeNil())
				Expect(paNp2.Gpu.OverQuotaWeight).ToNot(BeNil())

				var pbDef, pbNp2 api.Resources
				for _, r := range projBList.Resources {
					if r.NodePool.Name == "b-default" {
						pbDef = r
					}
					if r.NodePool.Name == "b-pool2" {
						pbNp2 = r
					}
				}
				Expect(pbDef.Cpu).ToNot(BeNil())
				Expect(pbDef.Memory).ToNot(BeNil())
				Expect(pbDef.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(pbNp2.Cpu).To(BeNil())
				Expect(pbNp2.Memory).To(BeNil())
				Expect(pbNp2.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*pbNp2.Gpu.OverQuotaWeight).To(Equal(pbNp2.Gpu.Deserved))
			})
		})
	})
})

func CreateNodePoolApiResourcesNoQuota(nodePoolId string, nodePoolName string) api.ResourcesNullable {
	nodePool := api.ResourcesNodePool{
		Id:   nodePoolId,
		Name: nodePoolName,
	}
	resources := api.ResourcesNullable{
		NodePool: &nodePool,
		Gpu: api.ResourcesNullableGpu{
			Limit:           nil,
			Deserved:        ptr.To(float64(0)),
			OverQuotaWeight: nil,
		},
		Cpu: &api.ResourcesCpu{
			Limit: ptr.To(float64(-1)),
		},
		Memory: &api.ResourcesMemory{
			Limit: ptr.To(float64(-1)),
		},
	}
	return resources
}
