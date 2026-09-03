/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/reports"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Report filter tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var project *api.Project
	var department api.Department

	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.20")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     fmt.Sprint(defaultNodePoolId),
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
			},
		},
	}, nil)
	departmentApiService := departments.NewDepartmentApiService(&clusterServiceClient, nil)
	projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
	projectApiService := projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

	BeforeAll(func() {
		createDepartmentRequest := getFullDepartmentCreateRequest("dep1", clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated))
		department = response.Body.(api.Department)

		name := createProjectName()
		request := getProjectCreateRequest(name, name, department.Id, clusterId)
		createResponse, err := projectApiService.CreateProject(ctx, request)
		Expect(err).To(BeNil())

		Expect(createResponse.Code).To(Equal(http.StatusCreated))
		project = ptr.To(createResponse.Body.(api.Project))
	})
	AfterAll(func() {
		response, err := projectApiService.DeleteProject(ctx, project.Id, "", true)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))

		response, err = departmentApiService.DeleteDepartment(ctx, department.Id)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	})

	Context("filter by permissions", func() {
		It("cluster only", func() {
			userCtx := context.WithValue(context.Background(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{clusterId}})
			filtersMap := map[string][]string{}
			reports.SetPermissionFilters(userCtx, TestTenant, &filtersMap)

			Expect(len(filtersMap[reports.ClusterIdFilter])).To(Equal(1))
			Expect(filtersMap[reports.ClusterIdFilter][0]).To(Equal(clusterId))
			Expect(len(filtersMap[reports.DepartmentNameFilter])).To(Equal(0))
			Expect(len(filtersMap[reports.ProjectNameFilter])).To(Equal(0))
		})
		It("department only", func() {
			userCtx := context.WithValue(context.Background(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{department.Id}})
			filtersMap := map[string][]string{}
			reports.SetPermissionFilters(userCtx, TestTenant, &filtersMap)

			Expect(len(filtersMap[reports.ClusterIdFilter])).To(Equal(0))
			Expect(len(filtersMap[reports.DepartmentNameFilter])).To(Equal(1))
			Expect(filtersMap[reports.DepartmentNameFilter][0]).To(Equal(department.Name))
			Expect(len(filtersMap[reports.ProjectNameFilter])).To(Equal(0))
		})
		It("project only", func() {
			userCtx := context.WithValue(context.Background(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Projects: []string{project.Id}})
			filtersMap := map[string][]string{}
			reports.SetPermissionFilters(userCtx, TestTenant, &filtersMap)

			Expect(len(filtersMap[reports.ClusterIdFilter])).To(Equal(0))
			Expect(len(filtersMap[reports.DepartmentNameFilter])).To(Equal(0))
			Expect(len(filtersMap[reports.ProjectNameFilter])).To(Equal(1))
			Expect(filtersMap[reports.ProjectNameFilter][0]).To(Equal(project.Name))
		})
		It("multiple, parent department", func() {
			userCtx := context.WithValue(context.Background(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Departments: []string{department.Id}, Projects: []string{project.Id}})
			filtersMap := map[string][]string{}
			reports.SetPermissionFilters(userCtx, TestTenant, &filtersMap)

			Expect(len(filtersMap[reports.ClusterIdFilter])).To(Equal(0))
			Expect(len(filtersMap[reports.DepartmentNameFilter])).To(Equal(1))
			Expect(filtersMap[reports.DepartmentNameFilter][0]).To(Equal(department.Name))
			Expect(len(filtersMap[reports.ProjectNameFilter])).To(Equal(0))
		})
		It("multiple, non parent cluster", func() {
			randomId := uuid.NewString()
			userCtx := context.WithValue(context.Background(), authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{Clusters: []string{randomId}, Projects: []string{project.Id}})
			filtersMap := map[string][]string{}
			reports.SetPermissionFilters(userCtx, TestTenant, &filtersMap)

			Expect(len(filtersMap[reports.ClusterIdFilter])).To(Equal(1))
			Expect(filtersMap[reports.ClusterIdFilter][0]).To(Equal(randomId))
			Expect(len(filtersMap[reports.DepartmentNameFilter])).To(Equal(0))
			Expect(len(filtersMap[reports.ProjectNameFilter])).To(Equal(1))
			Expect(filtersMap[reports.ProjectNameFilter][0]).To(Equal(project.Name))
		})
	})

	Context("filter parsing", func() {
		It("single departmentName value", func() {
			filters, err := reports.ParseReportFilters([]string{"departmentName==depA"})
			Expect(err).To(BeNil())
			Expect(filters).To(HaveKey(reports.DepartmentNameFilter))
			Expect(filters[reports.DepartmentNameFilter]).To(Equal([]string{"depA"}))
		})

		It("multiple values for the same key preserve all values", func() {
			filters, err := reports.ParseReportFilters([]string{"departmentName==depA", "departmentName==depB", "departmentName==depC"})
			Expect(err).To(BeNil())
			Expect(filters[reports.DepartmentNameFilter]).To(Equal([]string{"depA", "depB", "depC"}))
		})

		It("multiple different keys aggregate correctly", func() {
			cluster := "11111111-1111-1111-1111-111111111111"
			filters, err := reports.ParseReportFilters([]string{
				"clusterId==" + cluster,
				"departmentName==depA",
				"projectName==projX",
				"projectName==projY",
			})
			Expect(err).To(BeNil())
			Expect(filters[reports.ClusterIdFilter]).To(Equal([]string{cluster}))
			Expect(filters[reports.DepartmentNameFilter]).To(Equal([]string{"depA"}))
			Expect(filters[reports.ProjectNameFilter]).To(Equal([]string{"projX", "projY"}))
		})

		It("unsupported key returns error", func() {
			_, err := reports.ParseReportFilters([]string{"unsupported==value"})
			Expect(err).To(Not(BeNil()))
		})

		It("invalid format returns error", func() {
			_, err := reports.ParseReportFilters([]string{"departmentName=depA"}) // missing second '='
			Expect(err).To(Not(BeNil()))
		})
	})
})
