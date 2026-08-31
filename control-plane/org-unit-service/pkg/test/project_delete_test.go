/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustersOpenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	async_events "github.com/run-ai/runai/runai-application-packages/async-events"
	openapi2 "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	dvMocks "github.com/run-ai/runai/runai-clients/datavolumes-service-client/mocks"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsMocks "github.com/run-ai/runai/runai-clients/workloads-client/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	pubsubmocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

var _ = Describe("Project delete API tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var project *api.Project
	var department api.Department
	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)

	var projectApiService api.ProjectsAPIServicer
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To(services.ProjectDeletionMinVersion)),
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

	BeforeAll(func() {
		createDepartmentRequest := getFullDepartmentCreateRequest("parent", clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusCreated))
		department = response.Body.(api.Department)
	})
	AfterAll(func() {
		response, err := departmentApiService.DeleteDepartment(ctx, department.Id)
		Expect(err).To(BeNil())
		Expect(response.Code).To(Equal(http.StatusNoContent))
	})

	Context("test deletion types, new cluster", func() {
		var contextClusterServiceClient mocks.IClusterClient
		dvsClient := dvMocks.IDatavolumesClient{}

		BeforeAll(func() {
			contextClusterServiceClient = mocks.IClusterClient{}
			contextClusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.20")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			contextClusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			dvsClient.On("GetDatavolumes", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&openapi2.GetDatavolumes200Response{
				Datavolumes: []openapi2.DatavolumeNoSharedScopes{},
			}, nil)
			projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, &dvsClient)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		})
		BeforeEach(func() {
			if project == nil {
				name := createProjectName()
				request := getProjectCreateRequest(name, name, department.Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				Expect(err).To(BeNil())

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				project = ptr.To(createResponse.Body.(api.Project))
			}
		})
		AfterEach(func() {
			if project != nil {
				response, err := projectApiService.DeleteProject(ctx, project.Id, "", true)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
				project = nil
			}
		})

		It("orphan no force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_ORPHAN, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			project := ptr.To(response.Body.(api.Project))

			Expect(project.Status.Phase).To(Equal(api.PROJECTPHASE_DELETING))
			Expect(project.Deletion).To(Not(BeNil()))
			Expect(project.Deletion.Type).To(Equal(api.DELETIONTYPES_ORPHAN))
			Expect(project.Deletion.Force).To(Equal(false))
			Expect(project.Deletion.Timestamp.UnixMilli()).To(Not(Equal(time.Time{}.UnixMilli())))
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("cascade no force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_CASCADE, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			project := ptr.To(response.Body.(api.Project))

			Expect(project.Status.Phase).To(Equal(api.PROJECTPHASE_DELETING))
			Expect(project.Deletion).To(Not(BeNil()))
			Expect(project.Deletion.Type).To(Equal(api.DELETIONTYPES_CASCADE))
			Expect(project.Deletion.Force).To(Equal(false))
			Expect(project.Deletion.Timestamp.UnixMilli()).To(Not(Equal(time.Time{}.UnixMilli())))
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("orphan force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_ORPHAN, true)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("cascade force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_CASCADE, true)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("only empty, not empty", func() {
			projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, clearWorkloadsClient(project.Id, []openapi.Workload{{Name: "workload-a"}}), &dvsClient)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_IF_EMPTY, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body).To(ContainSubstring("workload-a"))

			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("only empty, empty", func() {
			projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, clearWorkloadsClient(project.Id, []openapi.Workload{}), &dvsClient)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_IF_EMPTY, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
	})
	Context("test deletion types, old cluster", func() {
		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.19")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient, nil, nil), &clusterServiceClient, nil)
		})
		BeforeEach(func() {
			if project == nil {
				name := createProjectName()
				request := getProjectCreateRequest(name, name, department.Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				Expect(err).To(BeNil())

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				project = ptr.To(createResponse.Body.(api.Project))
			}
		})
		AfterEach(func() {
			if project != nil {
				response, err := projectApiService.DeleteProject(ctx, project.Id, "", true)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
				project = nil
			}
		})

		It("orphan no force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_ORPHAN, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("cascade no force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_CASCADE, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("orphan force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_ORPHAN, true)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("cascade force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_CASCADE, true)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
	})
	Context("test delete multiple times", Ordered, func() {
		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.20")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient, nil, nil), &clusterServiceClient, nil)

			if project == nil {
				name := createProjectName()
				request := getProjectCreateRequest(name, name, department.Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				Expect(err).To(BeNil())

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				project = ptr.To(createResponse.Body.(api.Project))
			}
		})
		AfterAll(func() {
			if project != nil {
				response, err := projectApiService.DeleteProject(ctx, project.Id, "", false)
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusNoContent))
				project = nil
			}
		})

		It("orphan no force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_ORPHAN, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			project := ptr.To(response.Body.(api.Project))

			Expect(project.Status.Phase).To(Equal(api.PROJECTPHASE_DELETING))
			Expect(project.Deletion).To(Not(BeNil()))
			Expect(project.Deletion.Type).To(Equal(api.DELETIONTYPES_ORPHAN))
			Expect(project.Deletion.Force).To(Equal(false))
			Expect(project.Deletion.Timestamp.UnixMilli()).To(Not(Equal(time.Time{}.UnixMilli())))
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("cascade no force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_CASCADE, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusBadRequest))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			project := ptr.To(response.Body.(api.Project))

			Expect(project.Status.Phase).To(Equal(api.PROJECTPHASE_DELETING))
			Expect(project.Deletion).To(Not(BeNil()))
			Expect(project.Deletion.Type).To(Equal(api.DELETIONTYPES_ORPHAN))
			Expect(project.Deletion.Force).To(Equal(false))
			Expect(project.Deletion.Timestamp.UnixMilli()).To(Not(Equal(time.Time{}.UnixMilli())))
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
		It("orphan force", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_ORPHAN, true)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNoContent))

			response, err = projectApiService.GetProject(ctx, project.Id)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusNotFound))
			project = nil

			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})
	})
	Context("test delete sync from cluster", Ordered, func() {
		createdProjects := []api.Project{}
		projCount := 3

		BeforeAll(func() {
			clusterServiceClient = mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.20")),
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
					},
				},
			}, nil)
			projectApiService = projects.NewProjectApiService(services.NewProjectDeletionService(&clusterServiceClient, nil, nil), &clusterServiceClient, nil)

			for i := 0; i < projCount; i++ {
				name := createProjectName()
				request := getProjectCreateRequest(name, name, department.Id, clusterId)
				createResponse, err := projectApiService.CreateProject(ctx, request)
				Expect(err).To(BeNil())

				Expect(createResponse.Code).To(Equal(http.StatusCreated))
				createdProjects = append(createdProjects, createResponse.Body.(api.Project))
			}

			project = ptr.To(createdProjects[0])
		})
		AfterAll(func() {
			for _, createdProject := range createdProjects {
				projectApiService.DeleteProject(ctx, createdProject.Id, "", false)
			}

			createdProjects = []api.Project{}
		})

		It("sync with un existing projects that are in creating phase", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.SyncProjects(ctx, api.SyncProjectsRequest{
				ClusterId: clusterId,
				Projects:  []api.ProjectInSyncRequest{},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			getProjResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			projects := ptr.To(getProjResponse.Body.(api.GetProjects200Response))

			Expect(len(projects.Projects)).To(Equal(projCount))
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})

		It("sync with project in cascade, but with wrong cluster id", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.DeleteProject(ctx, project.Id, api.DELETIONTYPES_CASCADE, false)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusAccepted))

			response, err = projectApiService.SyncProjects(ctx, api.SyncProjectsRequest{
				ClusterId: uuid.NewString(),
				Projects:  []api.ProjectInSyncRequest{},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			getProjResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			projects := ptr.To(getProjResponse.Body.(api.GetProjects200Response))

			Expect(len(projects.Projects)).To(Equal(projCount))
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())
		})

		It("sync with project in cascade", func() {
			pubSuber := clearPubSuber()

			response, err := projectApiService.SyncProjects(ctx, api.SyncProjectsRequest{
				ClusterId: clusterId,
				Projects:  []api.ProjectInSyncRequest{},
			})
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			getProjResponse, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			projects := ptr.To(getProjResponse.Body.(api.GetProjects200Response))

			Expect(len(projects.Projects)).To(Equal(projCount - 1))
			Expect(pubSuber.AssertCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleted{}).Name(), mock.Anything)).To(BeTrue())
			Expect(pubSuber.AssertNotCalled(GinkgoT(), "Publish", reflect.TypeOf(async_events.ProjectDeleteRequest{}).Name(), mock.Anything)).To(BeTrue())

			project = nil
		})
	})
})

func clearPubSuber() *pubsubmocks.PubSuber {
	pubSuber := &pubsubmocks.PubSuber{}
	pubSuber.On("Publish", mock.Anything, mock.Anything).Return(nil)
	pubSuber.On("SubscribeAsync", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	runaipubsub.SetMock(pubSuber)

	return pubSuber
}

func clearWorkloadsClient(projectId string, workloads []openapi.Workload) *workloadsMocks.IWorkloadsClient {
	workloadsClient := &workloadsMocks.IWorkloadsClient{}
	workloadsClient.On("GetWorkloads", mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(arg interface{}) bool {
		filters := arg.([]string)

		return len(filters) == 1 && filters[0] == fmt.Sprintf("projectId==%s", projectId)
	}), mock.Anything, mock.Anything, mock.Anything).Return(&openapi.GetWorkloads200Response{
		Workloads: workloads,
	}, nil)

	return workloadsClient
}
