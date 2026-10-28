/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/stretchr/testify/mock"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

const (
	defaultNodePoolName = "default"
	defaultNodePoolId   = float32(1)
	nodePool2Name       = "np2"
	nodePool2Id         = float32(2)
)

var _ = Describe("Project API tests", Ordered, func() {
	clusterId := uuid.New().String()
	cluster2Id := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)
	departmentName1 := "parent1"
	departmentName2 := "parent2"
	departmentName3 := "parent3"

	var projectIds []string
	var departmentId1 string
	var departmentId2 string
	var departmentId3 string
	var projectApiService api.ProjectsAPIServicer
	var departmentApiService api.DepartmentsAPIServicer

	BeforeAll(func() {
		clusterServiceClient := mocks.IClusterClient{}
		clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
			Uuid:    clusterId,
			Version: *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
		}, nil)
		clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
			Nodepools: []clustersOpenapi.NodepoolMinimal{
				{
					Name:                   defaultNodePoolName,
					Id:                     fmt.Sprint(defaultNodePoolId),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
				{
					Name:                   nodePool2Name,
					Id:                     fmt.Sprint(nodePool2Id),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
			},
		}, nil)
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(services.ProjectDeletionService{}, &clusterServiceClient, nil)
		createDepartmentRequest := getFullDepartmentCreateRequest(departmentName1, clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId1 = response.Body.(api.Department).Id

		createDepartmentRequest = getFullDepartmentCreateRequest(departmentName2, clusterId)
		response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId2 = response.Body.(api.Department).Id

		createDepartmentRequest = getFullDepartmentCreateRequest(departmentName3, cluster2Id)
		response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId3 = response.Body.(api.Department).Id
	})

	AfterAll(func() {
		_, err := departmentApiService.DeleteDepartment(ctx, departmentId1)
		if err != nil {
			Fail("Failed deleting department")
		}
		_, err = departmentApiService.DeleteDepartment(ctx, departmentId2)
		if err != nil {
			Fail("Failed deleting department")
		}
		_, err = departmentApiService.DeleteDepartment(ctx, departmentId3)
		if err != nil {
			Fail("Failed deleting department")
		}
	})

	Context("test quota telemetry", func() {
		BeforeAll(func() {
			// proj on cluster 1, department 1
			// default nodepool: gpu: 0, cpu: 1, memory: 2
			// nodepool2: gpu: 3, cpu: 4, memory: 5
			request := getProjectRequest(createProjectName(), &departmentId1, clusterId, 0)
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectIds = append(projectIds, response.Body.(api.Project).Id)

			// proj on cluster 1, department 1
			// default nodepool: gpu: 1, cpu: 2, memory: 3
			// nodepool2: gpu: 4, cpu: 5, memory: 6
			request = getProjectRequest(createProjectName(), &departmentId1, clusterId, 1)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectIds = append(projectIds, response.Body.(api.Project).Id)

			// proj on cluster 1, department 2
			// default nodepool: gpu: 2, cpu: 3, memory: 4
			// nodepool2: gpu: 5, cpu: 6, memory: 7
			request = getProjectRequest(createProjectName(), &departmentId2, clusterId, 2)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectIds = append(projectIds, response.Body.(api.Project).Id)

			// proj on cluster 1, department 2
			// default nodepool: gpu: 3, cpu: 4, memory: 5
			// nodepool2: gpu: 6 cpu: 7, memory: 8
			request = getProjectRequest(createProjectName(), &departmentId2, clusterId, (3))
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectIds = append(projectIds, response.Body.(api.Project).Id)

			// proj on cluster 2, department 3
			// default nodepool: gpu: 4.5, cpu: 5.5, memory: 6.5
			// nodepool2: gpu: 7.5 cpu: 8.5, memory: 9.5
			request = getProjectRequest(createProjectName(), &departmentId3, cluster2Id, 4.5)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectIds = append(projectIds, response.Body.(api.Project).Id)
		})

		AfterAll(func() {
			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			projectIds = []string{}
		})

		It("no filters no groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("36.00"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_QUOTA, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("46.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("56.000"))
		})

		It("cluster filter no groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("24.00"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_QUOTA, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("32.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("40.000"))
		})

		It("no filters nodepool and cluster groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"6.00":  {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"18.00": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"4.50":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"7.50":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_QUOTA, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"10.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"22.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"5.500":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"8.500":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"14.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"26.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"6.500":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"9.500":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})
		})

		It("nodepool filter department groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"1.00": {departmentId1, departmentName1},
				"5.00": {departmentId2, departmentName2},
				"4.50": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_QUOTA, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"3.000": {departmentId1, departmentName1},
				"7.000": {departmentId2, departmentName2},
				"5.500": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"5.000": {departmentId1, departmentName1},
				"9.000": {departmentId2, departmentName2},
				"6.500": {departmentId3, departmentName3},
			})
		})

		It("project filter no groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, "", "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_QUOTA)))
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3.00"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_QUOTA, "", "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_QUOTA)))
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("5.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA, "", "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_QUOTA)))
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("7.000"))
		})

		It("project filter nodepool groups", func() {
			// projectIds[0]: default nodepool gpu=0, nodepool2 gpu=3
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, "", "", "", projectIds[0], []string{"NodepoolId"})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_QUOTA)))
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(2))
			expectedValues := map[string][]string{
				"0.00": {fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"3.00": {fmt.Sprint(nodePool2Id), nodePool2Name},
			}
			for expectedValue, expectedGroup := range expectedValues {
				found := false
				for _, tv := range telemetryValues {
					if expectedValue == tv.Value {
						Expect(tv.Groups[0].Key).To(Equal("NodepoolId"))
						Expect(tv.Groups[0].Value).To(Equal(expectedGroup[0]))
						Expect(tv.Groups[0].Name).To(Equal(expectedGroup[1]))
						found = true
						break
					}
				}
				Expect(found).To(BeTrue(), fmt.Sprintf("missing telemetry bucket for value %s", expectedValue))
			}
		})

		It("project and cluster filter no groups", func() {
			// projectIds[0] is on cluster 1 — combining both filters should return same result
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, clusterId, "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3.00"))
		})

		It("project-scoped permitted scopes - quota", func() {
			// Regression: pre-fix this query failed with "column reference 'id' is ambiguous"
			// because getPermittedScopesForProjects emitted unqualified `OR id IN (...)` and
			// the telemetry query JOINs project with resources (both have `id`).
			scopedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectIds[0]},
			})
			response, err := projectApiService.GetProjectsTelemetry(scopedCtx, api.ORGUNITTELEMETRYTYPE_GPU_QUOTA, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(len(telemetry.Values)).To(Equal(1))
			Expect(telemetry.Values[0].Value).To(Equal("3.00"))
		})

		It("PROJECTS_COUNT with no groups returns distinct project count", func() {
			response, err := projectApiService.GetInternalProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT, "", "", "", "", nil, []string{})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT)))
			Expect(len(telemetry.Values)).To(Equal(1))
			Expect(telemetry.Values[0].Value).To(Equal(fmt.Sprint(len(projectIds))))
		})

		It("PROJECTS_COUNT grouped by Rank", func() {
			response, err := projectApiService.GetInternalProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT, "", "", "", "", nil, []string{"Rank"})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(len(telemetry.Values)).To(Equal(1))
			Expect(telemetry.Values[0].Value).To(Equal(fmt.Sprint(len(projectIds))))
			Expect(telemetry.Values[0].Groups).To(HaveLen(1))
			Expect(telemetry.Values[0].Groups[0].Key).To(Equal("Rank"))
			Expect(telemetry.Values[0].Groups[0].Value).To(Equal("MediumLow"))
		})

		It("PROJECTS_COUNT grouped by Rank filtered by createdAt range", func() {
			start := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339)
			end := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
			response, err := projectApiService.GetInternalProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT, "", "", "", "", []string{
				"createdAt>=" + start,
				"createdAt<=" + end,
			}, []string{"Rank"})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(len(telemetry.Values)).To(Equal(1))
			Expect(telemetry.Values[0].Value).To(Equal(fmt.Sprint(len(projectIds))))
			Expect(telemetry.Values[0].Groups).To(HaveLen(1))
			Expect(telemetry.Values[0].Groups[0].Key).To(Equal("Rank"))
			Expect(telemetry.Values[0].Groups[0].Value).To(Equal("MediumLow"))
		})

		It("PROJECTS_COUNT grouped by Rank filtered by createdAt range with no matches", func() {
			response, err := projectApiService.GetInternalProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT, "", "", "", "", []string{
				"createdAt>=2000-01-01T00:00:00Z",
				"createdAt<=2000-01-02T00:00:00Z",
			}, []string{"Rank"})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Values).To(BeEmpty())
		})

		It("PROJECTS_COUNT with invalid createdAt filter returns bad request", func() {
			response, err := projectApiService.GetInternalProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT, "", "", "", "", []string{"createdAt>=not-a-date"}, []string{"Rank"})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		})

		It("PROJECTS_COUNT grouped by Rank splits projects of different ranks", func() {
			request := getProjectRequest(createProjectName(), &departmentId1, clusterId, 0)
			highRank := "High"
			request.Resources[0].Rank = &highRank
			request.Resources[1].Rank = &highRank
			createResponse, err := projectApiService.CreateProject(ctx, request)
			Expect(err).To(BeNil())
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			extraProjectId := createResponse.Body.(api.Project).Id
			defer func() {
				_, _ = projectApiService.DeleteProject(ctx, extraProjectId, "", false)
			}()

			response, err := projectApiService.GetInternalProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPEINTERNAL_PROJECTS_COUNT, "", "", "", "", nil, []string{"Rank"})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(len(telemetry.Values)).To(Equal(2))
			countsByRank := map[string]string{}
			for _, value := range telemetry.Values {
				countsByRank[value.Groups[0].Value] = value.Value
			}
			Expect(countsByRank["MediumLow"]).To(Equal(fmt.Sprint(len(projectIds))))
			Expect(countsByRank["High"]).To(Equal("1"))
		})
	})

	Context("test allocation telemetry", func() {
		BeforeAll(func() {
			// proj on cluster 1, department 1
			// default nodepool: gpu: 0, cpu: 0, memory: 0
			// nodepool2: gpu: 3, cpu: 3000, memory: 300
			request := getProjectRequest(createProjectName(), &departmentId1, clusterId, 0)
			response, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectId := response.Body.(api.Project).Id
			projectIds = append(projectIds, projectId)
			statusRequest := getProjectStatusRequest(0, projectId, clusterId)
			_, err = projectApiService.UpdateProjectStatus(ctx, statusRequest)
			if err != nil {
				Fail("Failed creating project")
			}

			// proj on cluster 1, department 1
			// default nodepool: gpu: 1, cpu: 1000, memory: 100
			// nodepool2: gpu: 4, cpu: 4000, memory: 400
			request = getProjectRequest(createProjectName(), &departmentId1, clusterId, 1)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectId = response.Body.(api.Project).Id
			projectIds = append(projectIds, projectId)
			statusRequest = getProjectStatusRequest(1, projectId, clusterId)
			_, err = projectApiService.UpdateProjectStatus(ctx, statusRequest)
			if err != nil {
				Fail("Failed creating project")
			}

			// proj on cluster 1, department 2
			// default nodepool: gpu: 2, cpu: 2000, memory: 200
			// nodepool2: gpu: 5, cpu: 5000, memory: 500
			request = getProjectRequest(createProjectName(), &departmentId2, clusterId, 2)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectId = response.Body.(api.Project).Id
			projectIds = append(projectIds, projectId)
			statusRequest = getProjectStatusRequest(2, projectId, clusterId)
			_, err = projectApiService.UpdateProjectStatus(ctx, statusRequest)
			if err != nil {
				Fail("Failed creating project")
			}

			// proj on cluster 1, department 2
			// default nodepool: gpu: 3, cpu: 3000, memory: 300
			// nodepool2: gpu: 6 cpu: 6000, memory: 600
			request = getProjectRequest(createProjectName(), &departmentId2, clusterId, 3)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectId = response.Body.(api.Project).Id
			projectIds = append(projectIds, projectId)
			statusRequest = getProjectStatusRequest(3, projectId, clusterId)
			_, err = projectApiService.UpdateProjectStatus(ctx, statusRequest)
			if err != nil {
				Fail("Failed creating project")
			}

			// proj on cluster 2, department 3
			// default nodepool: gpu: 4.5, cpu: 4500, memory: 450
			// nodepool2: gpu: 7.5 cpu: 7500, memory: 750
			request = getProjectRequest(createProjectName(), &departmentId3, cluster2Id, 4.5)
			response, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			projectId = response.Body.(api.Project).Id
			projectIds = append(projectIds, projectId)
			statusRequest = getProjectStatusRequest(4.5, projectId, clusterId)
			_, err = projectApiService.UpdateProjectStatus(ctx, statusRequest)
			if err != nil {
				Fail("Failed creating project")
			}
		})

		AfterAll(func() {
			for _, projectId := range projectIds {
				deleteResponse, err := projectApiService.DeleteProject(ctx, projectId, "", false)
				if err != nil || deleteResponse.Code != http.StatusNoContent {
					Fail("Failed deleting")
				}
			}
			projectIds = []string{}
		})

		It("no filters no groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("36.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("36000.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("3600.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("46.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("46000.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("4600.000"))
		})

		It("cluster filter no groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("24.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("24000.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("2400.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("32.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("32000.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE, clusterId, "", "", "", []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Groups).To(BeNil())
			Expect(telemetryValues[0].Value).To(Equal("3200.000"))
		})

		It("no filters nodepool and cluster groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"6.000":  {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"18.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"4.500":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"7.500":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"10.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"22.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"5.500":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"8.500":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"14.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"26.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"6.500":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"9.500":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"10.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"22.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"5.500":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"8.500":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"10000.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"22000.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"5500.000":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"8500.000":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE, "", "", "", "", []string{"ClusterId", "NodepoolId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry2Groups(telemetryValues, map[string][]string{
				"1000.000": {clusterId, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"2200.000": {clusterId, fmt.Sprint(nodePool2Id), nodePool2Name},
				"550.000":  {cluster2Id, fmt.Sprint(defaultNodePoolId), defaultNodePoolName},
				"850.000":  {cluster2Id, fmt.Sprint(nodePool2Id), nodePool2Name},
			})
		})

		It("nodepool filter department groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues := telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"1.000": {departmentId1, departmentName1},
				"5.000": {departmentId2, departmentName2},
				"4.500": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"3.000": {departmentId1, departmentName1},
				"7.000": {departmentId2, departmentName2},
				"5.500": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"5.000": {departmentId1, departmentName1},
				"9.000": {departmentId2, departmentName2},
				"6.500": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"3.000": {departmentId1, departmentName1},
				"7.000": {departmentId2, departmentName2},
				"5.500": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_CPU_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"3000.000": {departmentId1, departmentName1},
				"7000.000": {departmentId2, departmentName2},
				"5500.000": {departmentId3, departmentName3},
			})

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE, "", fmt.Sprint(defaultNodePoolId), "", "", []string{"ParentId"})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_MEMORY_ALLOCATION_NON_PREEMPTIBLE)))
			Expect(telemetry.Timestamp).NotTo(BeNil())
			telemetryValues = telemetry.Values
			validateTelemetry1Group(telemetryValues, map[string][]string{
				"300.000": {departmentId1, departmentName1},
				"700.000": {departmentId2, departmentName2},
				"550.000": {departmentId3, departmentName3},
			})
		})

		It("project filter no groups", func() {
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, "", "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION)))
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3.000"))

			response, err = projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE, "", "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry = httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION_NON_PREEMPTIBLE)))
			telemetryValues = telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("5.000"))
		})

		It("project filter nodepool groups", func() {
			// projectIds[0]: nodepool2 gpu_alloc=3
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, "", "", "", projectIds[0], []string{"NodepoolId"})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(telemetry.Type).To(Equal(string(api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION)))
			telemetryValues := telemetry.Values
			found := false
			for _, tv := range telemetryValues {
				if tv.Value == "3.000" {
					Expect(tv.Groups[0].Key).To(Equal("NodepoolId"))
					Expect(tv.Groups[0].Value).To(Equal(fmt.Sprint(nodePool2Id)))
					Expect(tv.Groups[0].Name).To(Equal(nodePool2Name))
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "missing telemetry bucket for nodepool2 gpu allocation")
		})

		It("project and cluster filter no groups", func() {
			// projectIds[0] is on cluster 1 — combining both filters should return same result
			response, err := projectApiService.GetProjectsTelemetry(ctx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, clusterId, "", "", projectIds[0], []string{})
			Expect(err).To(BeNil())
			telemetry := httpResponseBodyToTelemetryResponse(response)
			telemetryValues := telemetry.Values
			Expect(len(telemetryValues)).To(Equal(1))
			Expect(telemetryValues[0].Value).To(Equal("3.000"))
		})

		It("project-scoped permitted scopes - allocation", func() {
			// Regression: pre-fix this query failed with "column reference 'id' is ambiguous"
			// because getPermittedScopesForProjects emitted unqualified `OR id IN (...)` and
			// the telemetry query JOINs project with resources (both have `id`).
			scopedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectIds[0]},
			})
			response, err := projectApiService.GetProjectsTelemetry(scopedCtx, api.ORGUNITTELEMETRYTYPE_GPU_ALLOCATION, "", "", "", "", []string{})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			telemetry := httpResponseBodyToTelemetryResponse(response)
			Expect(len(telemetry.Values)).To(Equal(1))
			Expect(telemetry.Values[0].Value).To(Equal("3.000"))
		})
	})
})

func validateTelemetry2Groups(telemetryValues []api.TelemetryResponseValuesInner, values map[string][]string) {
	Expect(len(telemetryValues)).To(Equal(4))
	for expectedValue, expectedGroup := range values {
		for _, telemetryValue := range telemetryValues {
			if expectedValue == telemetryValue.Value {
				Expect(telemetryValue.Groups[0].Key).To(Equal("ClusterId"))
				Expect(telemetryValue.Groups[1].Key).To(Equal("NodepoolId"))
				Expect(telemetryValue.Groups[0].Value).To(Equal(expectedGroup[0]))
				Expect(telemetryValue.Groups[1].Value).To(Equal(expectedGroup[1]))
				Expect(telemetryValue.Groups[1].Name).To(Equal(expectedGroup[2]))
				break
			}
		}
	}
}

func validateTelemetry1Group(telemetryValues []api.TelemetryResponseValuesInner, values map[string][]string) {
	Expect(len(telemetryValues)).To(Equal(3))
	for expectedValue, expectedGroup := range values {
		for _, telemetryValue := range telemetryValues {
			if expectedValue == telemetryValue.Value {
				Expect(telemetryValue.Groups[0].Key).To(Equal("ParentId"))
				Expect(telemetryValue.Groups[0].Value).To(Equal(expectedGroup[0]))
				Expect(telemetryValue.Groups[0].Name).To(Equal(expectedGroup[1]))
				break
			}
		}
	}
}

func getProjectStatusRequest(initialAllocation float64, projectId string, clusterId string) api.ProjectStatusUpdateRequest {
	quotaStatuses := []api.QuotaStatusNodePool{{
		Allocated: api.QuotaStatusResource{
			Gpu:    initialAllocation,
			Cpu:    initialAllocation * 1000,
			Memory: initialAllocation * 100,
		},
		AllocatedNonPreemptible: api.QuotaStatusResource{
			Gpu:    initialAllocation + 1,
			Cpu:    (initialAllocation + 1) * 1000,
			Memory: (initialAllocation + 1) * 100,
		},
		Requested: api.QuotaStatusResource{
			Gpu:    initialAllocation + 2,
			Cpu:    (initialAllocation + 2) * 1000,
			Memory: (initialAllocation + 2) * 100,
		},
		NodePoolName: defaultNodePoolName,
		NodePoolId:   fmt.Sprint(defaultNodePoolId),
	}, {
		Allocated: api.QuotaStatusResource{
			Gpu:    initialAllocation + 3,
			Cpu:    (initialAllocation + 3) * 1000,
			Memory: (initialAllocation + 3) * 100,
		},
		AllocatedNonPreemptible: api.QuotaStatusResource{
			Gpu:    initialAllocation + 4,
			Cpu:    (initialAllocation + 4) * 1000,
			Memory: (initialAllocation + 4) * 100,
		},
		Requested: api.QuotaStatusResource{
			Gpu:    initialAllocation + 5,
			Cpu:    (initialAllocation + 5) * 1000,
			Memory: (initialAllocation + 5) * 100,
		},
		NodePoolName: nodePool2Name,
		NodePoolId:   fmt.Sprint(nodePool2Id),
	}}
	return api.ProjectStatusUpdateRequest{
		Id:        projectId,
		ClusterId: clusterId,
		Status: api.ProjectStatus{
			Phase:                 "Ready",
			NodePoolQuotaStatuses: quotaStatuses,
		},
	}
}

func getProjectRequest(name string, parentId *string, clusterId string, initialDeserved float64) api.ProjectCreationRequest {
	return api.ProjectCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		ParentId:  parentId,
		Resources: []api.ResourcesNullable{
			createApiResources(initialDeserved, initialDeserved+1, initialDeserved+2, strconv.FormatFloat(float64(defaultNodePoolId), 'f', -1, 32), defaultNodePoolName),
			createApiResources(initialDeserved+3, initialDeserved+4, initialDeserved+5, strconv.FormatFloat(float64(nodePool2Id), 'f', -1, 32), nodePool2Name),
		},
	}
}

func createApiResources(gpuDeserved float64, cpuDeserved float64, memoryDeserved float64, nodePoolId string, nodePoolName string) api.ResourcesNullable {
	resources := api.ResourcesNullable{
		NodePool: &api.ResourcesNodePool{
			Id:   nodePoolId,
			Name: nodePoolName,
		},
		Gpu: api.ResourcesNullableGpu{
			Limit:           &gpuDeserved,
			Deserved:        ptr.To(gpuDeserved),
			OverQuotaWeight: &gpuDeserved,
		},
		Cpu: &api.ResourcesCpu{
			Limit:           &cpuDeserved,
			Deserved:        &cpuDeserved,
			OverQuotaWeight: &cpuDeserved,
		},
		Memory: &api.ResourcesMemory{
			Limit:           &memoryDeserved,
			Deserved:        &memoryDeserved,
			OverQuotaWeight: &memoryDeserved,
		},
	}
	return resources
}

func httpResponseBodyToTelemetryResponse(httpResponse api.ImplResponse) api.TelemetryResponse {
	bytes, err := json.Marshal(httpResponse.Body)
	Expect(err).To(BeNil())

	response := &api.TelemetryResponse{}
	err = json.Unmarshal(bytes, response)
	Expect(err).To(BeNil())
	return *response
}
