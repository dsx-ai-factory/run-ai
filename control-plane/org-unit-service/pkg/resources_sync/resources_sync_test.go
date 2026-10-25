/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources_sync

import (
	"fmt"
	"net/http"
	"strconv"
	ginkgoTesting "testing"

	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	clusterService "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/settings"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/mocks"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

func TestResourcesSync(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	// Register this test to prevent premature cleanup
	testing.RegisterTest()

	testing.SharedOpenPostgresForTesting(t)
	// Initialize the settings package against the test DB so the migrated LimitQuotaOverSubscription
	// setting can be read during quota validation (defaults to disabled when no row exists).
	setting_pkg.Init(&setting_pkg.SettingConfig{}, nil)
	_ = setting_pkg.Register[*settings.LimitQuotaOverSubscription]()
	testing.SetAuthMock()
	pubSuber := mocks.PubSuber{}
	pubSuber.On("Publish", mock.Anything, mock.Anything).Return(nil)
	pubSuber.On("SubscribeAsync", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	runaipubsub.SetMock(&pubSuber)

	RunSpecs(t, "resources sync test")

	t.Cleanup(func() {
		testing.UnregisterTest()
	})
}

var _ = Describe("test", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser)
	cluster1Id := uuid.New()
	cluster2Id := uuid.New()

	var (
		projectApiService    api.ProjectsAPIServicer
		departmentApiService api.DepartmentsAPIServicer
		nodeTypeApiService   api.NodeTypesAPIServicer
		defaultNodePoolName  = "default"
		defaultNodePoolId    = "1"
		nodePool2Name        = "pool2"
		nodePool2Id          = "2"
		nodePool3Name        = "pool3"
		nodePool3Id          = "3"
	)

	BeforeAll(func() {
		clusterServiceClient := clusterService.IClusterClient{}
		nodePools := []clustersOpenapi.NodepoolMinimal{
			{Name: defaultNodePoolName, Id: defaultNodePoolId, Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: testing.NpSchedCapabilitiesDefault(true, false)},
			{Name: nodePool2Name, Id: nodePool2Id, Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: testing.NpSchedCapabilitiesDefault(true, false)},
			{Name: nodePool3Name, Id: nodePool3Id, Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY), SchedulingCapabilities: testing.NpSchedCapabilitiesDefault(true, false)},
		}
		clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
			Uuid:    cluster1Id.String(),
			Version: *clustersOpenapi.NewNullableString(ptr.To(services.ResourcePriorityMinVersion)),
		}, nil)
		clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
			Nodepools: nodePools,
		}, nil)

		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		projectApiService = projects.NewProjectApiService(
			services.ProjectDeletionService{}, &clusterServiceClient, nil)
		nodeTypeApiService = services.NewNodeTypeApiService()
	})

	Context("test sync job", Ordered, func() {
		var (
			defaultDepartmentCluster1 api.Department
			department1Cluster1       api.Department
			department1Cluster2       api.Department
			project1Cluster1          api.Project
			project2Cluster1          api.Project
			project3Cluster1          api.Project
			project4Cluster1          api.Project
			project1Cluster2          api.Project
			nodeType1                 api.NodeType
			nodeType2                 api.NodeType
		)

		BeforeAll(func() {
			nodeType, err := nodeTypeApiService.CreateNodeType(ctx, api.NodeTypeCreateRequest{
				Name:      "type1",
				ClusterId: cluster1Id.String(),
			})
			Expect(err).To(BeNil())
			nodeType1 = nodeType.Body.(api.NodeType)

			nodeType, err = nodeTypeApiService.CreateNodeType(ctx, api.NodeTypeCreateRequest{
				Name:      "type2",
				ClusterId: cluster1Id.String(),
			})
			Expect(err).To(BeNil())
			nodeType2 = nodeType.Body.(api.NodeType)

			name := services.DefaultDepartmentName
			createDepartmentRequest := getDepartmentCreateRequest(name, cluster1Id.String())
			response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			defaultDepartmentCluster1 = response.Body.(api.Department)

			name = "dep-" + uuid.New().String()
			createDepartmentRequest = getDepartmentCreateRequest(name, cluster1Id.String())
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department1Cluster1 = response.Body.(api.Department)

			name = "dep-" + uuid.New().String()
			createDepartmentRequest = getDepartmentCreateRequest(name, cluster2Id.String())
			response, err = departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
			if err != nil {
				Fail("Failed creating parent department")
			}
			Expect(response.Code).To(Equal(http.StatusCreated))
			department1Cluster2 = response.Body.(api.Department)

			projName := "proj-" + uuid.New().String()
			createProjectRequest := getProjectCreateRequest(projName, cluster1Id.String(), department1Cluster1.Id)
			createResponse, err := projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project1Cluster1 = createResponse.Body.(api.Project)

			// Create Project without setting priority
			projName = "proj-" + uuid.New().String()
			createProjectRequest = getProjectCreateRequest(projName, cluster1Id.String(), department1Cluster1.Id)
			createProjectRequest.Resources[0].Rank = nil
			createResponse, err = projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project1Cluster1 = createResponse.Body.(api.Project)
			Expect(*(project1Cluster1.Resources[0].Rank)).To(Equal(services.DefaultResourceRank))

			projName = "proj-" + uuid.New().String()
			createProjectRequest = getProjectCreateRequest(projName, cluster1Id.String(), department1Cluster1.Id)
			createResponse, err = projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project2Cluster1 = createResponse.Body.(api.Project)

			projName = "proj-" + uuid.New().String()
			createProjectRequest = getProjectCreateRequest(projName, cluster1Id.String(), defaultDepartmentCluster1.Id)
			createResponse, err = projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project3Cluster1 = createResponse.Body.(api.Project)

			projName = "proj-" + uuid.New().String()
			createProjectRequest = getProjectCreateRequest(projName, cluster1Id.String(), defaultDepartmentCluster1.Id)
			createResponse, err = projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project4Cluster1 = createResponse.Body.(api.Project)

			projName = "proj-" + uuid.New().String()
			createProjectRequest = getProjectCreateRequest(projName, cluster2Id.String(), department1Cluster2.Id)
			createResponse, err = projectApiService.CreateProject(ctx, createProjectRequest)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project1Cluster2 = createResponse.Body.(api.Project)
		})

		AfterAll(func() {
			_, err := projectApiService.DeleteProject(ctx, project1Cluster1.Id, "", false)
			Expect(err).To(BeNil())
			_, err = projectApiService.DeleteProject(ctx, project1Cluster2.Id, "", false)
			Expect(err).To(BeNil())
			_, err = projectApiService.DeleteProject(ctx, project2Cluster1.Id, "", false)
			Expect(err).To(BeNil())
			_, err = projectApiService.DeleteProject(ctx, project3Cluster1.Id, "", false)
			Expect(err).To(BeNil())
			_, err = projectApiService.DeleteProject(ctx, project4Cluster1.Id, "", false)
			Expect(err).To(BeNil())
			_, err = departmentApiService.DeleteDepartment(ctx, department1Cluster1.Id)
			Expect(err).To(BeNil())
			_, err = departmentApiService.DeleteDepartment(ctx, department1Cluster2.Id)
			Expect(err).To(BeNil())
			_, err = departmentApiService.DeleteDepartment(ctx, defaultDepartmentCluster1.Id)
			Expect(err).To(BeNil())
			_, err = nodeTypeApiService.DeleteNodeType(ctx, nodeType1.Id)
			Expect(err).To(BeNil())
			_, err = nodeTypeApiService.DeleteNodeType(ctx, nodeType2.Id)
			Expect(err).To(BeNil())
		})

		Describe("test missing resources", func() {
			BeforeEach(func() {
				resourcesDelete := []models.Resources{}
				projectId, err := strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project1ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				for i := range project1ById.OrgUnitData.Resources {
					resources := project1ById.OrgUnitData.Resources[i]
					if resources.NodePoolId == fmt.Sprint(nodePool2Id) {
						resourcesDelete = append(resourcesDelete, resources)
					}
				}

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				for i := range project2ById.OrgUnitData.Resources {
					resources := project2ById.OrgUnitData.Resources[i]
					if resources.NodePoolId == fmt.Sprint(nodePool2Id) {
						resourcesDelete = append(resourcesDelete, resources)
					}
				}

				err = dao.DeleteResources(ctx, resourcesDelete)
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project1ById, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project1ById.OrgUnitData.Resources)).To(Equal(2))

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project2ById.OrgUnitData.Resources)).To(Equal(2))

				projectId, err = strconv.ParseInt(project2Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project3ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project3ById.OrgUnitData.Resources)).To(Equal(3))

				resourcesDelete = []models.Resources{}
				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				for i := range department.OrgUnitData.Resources {
					resources := department.OrgUnitData.Resources[i]
					if resources.NodePoolId == fmt.Sprint(nodePool3Id) {
						resourcesDelete = append(resourcesDelete, resources)
					}
				}

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				for i := range department.OrgUnitData.Resources {
					resources := department.OrgUnitData.Resources[i]
					if resources.NodePoolId == fmt.Sprint(nodePool2Id) || resources.NodePoolId == fmt.Sprint(defaultNodePoolId) {
						resourcesDelete = append(resourcesDelete, resources)
					}
				}

				err = dao.DeleteResources(ctx, resourcesDelete)
				Expect(err).To(BeNil())

				departmentId, err = strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(2))

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(1))

				departmentId, err = strconv.ParseInt(department1Cluster1.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(3))
			})

			It("projects", func() {
				err := handleMissingResources(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				projectId, err := strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project1ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				validateAllResources(project1ById.OrgUnitData.Resources)

				err = handleMissingResources(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				validateAllResources(project2ById.OrgUnitData.Resources)
			})

			It("departments", func() {
				err := handleMissingResources(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				validateAllResources(department.OrgUnitData.Resources)

				err = handleMissingResources(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				validateAllResources(department.OrgUnitData.Resources)
			})
		})

		Describe("test deleted resources", func() {
			BeforeEach(func() {
				dummyNodePoolId := "11"
				dummyNodePoolName := "dummy"
				dummy2NodePoolId := "22"
				dummy2NodePoolName := "dummy2"
				resourcesAdd := []*models.Resources{}

				projectId, err := strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				resourcesAdd = append(resourcesAdd, services.GetProjectResourcesToAdd(project.OrgUnitData, dummyNodePoolId, dummyNodePoolName))

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				resourcesAdd = append(resourcesAdd, services.GetProjectResourcesToAdd(project.OrgUnitData, dummyNodePoolId, dummyNodePoolName))
				resourcesAdd = append(resourcesAdd, services.GetProjectResourcesToAdd(project.OrgUnitData, dummy2NodePoolId, dummy2NodePoolName))

				err = dao.AddResources(ctx, resourcesAdd)
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project.OrgUnitData.Resources)).To(Equal(4))

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project.OrgUnitData.Resources)).To(Equal(5))

				resourcesAdd = []*models.Resources{}
				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				resourcesAdd = append(resourcesAdd, services.GetDepartmentResourcesToAdd(department.OrgUnitData, true, dummyNodePoolId, dummyNodePoolName))

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				resourcesAdd = append(resourcesAdd, services.GetDepartmentResourcesToAdd(department.OrgUnitData, false, dummyNodePoolId, dummyNodePoolName))
				resourcesAdd = append(resourcesAdd, services.GetDepartmentResourcesToAdd(department.OrgUnitData, false, dummy2NodePoolId, dummy2NodePoolName))

				err = dao.AddResources(ctx, resourcesAdd)
				Expect(err).To(BeNil())

				departmentId, err = strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(4))

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(5))

				departmentId, err = strconv.ParseInt(department1Cluster1.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(3))
			})

			It("projects", func() {
				err := handleDeletedResources(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name})
				Expect(err).To(BeNil())

				projectId, err := strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project1ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				validateAllResources(project1ById.OrgUnitData.Resources)

				err = handleDeletedResources(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name})
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				validateAllResources(project2ById.OrgUnitData.Resources)
			})

			It("departments", func() {
				err := handleDeletedResources(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name})
				Expect(err).To(BeNil())

				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				validateAllResources(department.OrgUnitData.Resources)

				err = handleDeletedResources(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name})
				Expect(err).To(BeNil())

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				validateAllResources(department.OrgUnitData.Resources)
			})
		})

		Describe("test no resources", func() {
			BeforeEach(func() {
				resourcesSoftDelete := []models.Resources{}
				projectId, err := strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project1ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				for i := range project1ById.OrgUnitData.Resources {
					resources := project1ById.OrgUnitData.Resources[i]
					resourcesSoftDelete = append(resourcesSoftDelete, resources)
				}

				err = dao.DeleteResources(ctx, resourcesSoftDelete)
				Expect(err).To(BeNil())

				resourcesHardDelete := []string{}
				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				for i := range project2ById.OrgUnitData.Resources {
					resources := project2ById.OrgUnitData.Resources[i]
					resourcesHardDelete = append(resourcesHardDelete, resources.ID.String())
				}
				_, err = runaidb.DB().Where("id IN ?", resourcesHardDelete).Delete(&[]models.Resources{})
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster1.Id, 10, 64)
				project1ById, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project1ById.OrgUnitData.Resources)).To(Equal(0))

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(len(project2ById.OrgUnitData.Resources)).To(Equal(0))

				resourcesSoftDelete = []models.Resources{}
				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				for i := range department.OrgUnitData.Resources {
					resources := department.OrgUnitData.Resources[i]
					resourcesSoftDelete = append(resourcesSoftDelete, resources)
				}

				err = dao.DeleteResources(ctx, resourcesSoftDelete)
				Expect(err).To(BeNil())

				resourcesHardDelete = []string{}
				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				for i := range department.OrgUnitData.Resources {
					resources := department.OrgUnitData.Resources[i]
					resourcesHardDelete = append(resourcesHardDelete, resources.ID.String())
				}
				_, err = runaidb.DB().Where("id IN ?", resourcesHardDelete).Delete(&[]models.Resources{})
				Expect(err).To(BeNil())

				departmentId, err = strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(0))

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(len(department.OrgUnitData.Resources)).To(Equal(0))
			})

			It("projects", func() {
				err := handleNoResources(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				projectId, err := strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project1ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				validateAllResources(project1ById.OrgUnitData.Resources)

				err = handleNoResources(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project2ById, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				validateAllResources(project2ById.OrgUnitData.Resources)
			})

			It("departments", func() {
				err := handleNoResources(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				validateAllResources(department.OrgUnitData.Resources)

				err = handleNoResources(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name}, false)
				Expect(err).To(BeNil())

				departmentId, err = strconv.ParseInt(department1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				validateAllResources(department.OrgUnitData.Resources)
			})
		})

		Describe("test default node pools", func() {
			/*
				fix the following: (dummy and dummy2 does not exist)
					cluster1:
						defaultDepartmentCluster1: defaultNodePools: {np2, dummy} // test department change
							project3Cluster1: defaultNodePools: null, effectiveDefaultNodePools: {np2, dummy}, effectiveOrigin: defaultDepartmentCluster1 	// test effective change
							project4Cluster1: defaultNodePools: {dummy2}, effectiveDefaultNodePools: {dummy2}, effectiveOrigin: null 						// test empty list and get parent's list to be effective
						department1Cluster1: defaultNodePools: {dummy}
							project1Cluster1: defaultNodePools: {np3}, effectiveDefaultNodePools: {np3}, effectiveOrigin: null 								// test keep valid list
							project2Cluster1: defaultNodePools: null, effectiveDefaultNodePools: {dummy}, effectiveOrigin: department1Cluster1				// test effective change to {default}
					cluster2:
						department1Cluster2: defaultNodePools: null
							project1Cluster2: defaultNodePools: {dummy2}, effectiveDefaultNodePools: {dummy2}, effectiveOrigin: null 						// test list change to {default}
			*/
			BeforeEach(func() {
				dummyNodePoolName := "dummy"
				dummy2NodePoolName := "dummy2"

				// default department with invalid node pool list
				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				array := pq.StringArray([]string{dummyNodePoolName, nodePool2Name})
				department.OrgUnitData.DefaultNodePoolList = &array
				department.OrgUnitData.WorkspaceNodeTypes = &[]models.NodeType{nodeTypeRequestFromApi(nodeType1, cluster1Id)}
				_, err = dao.UpdateDepartment(ctx, department)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(*department.OrgUnitData.DefaultNodePoolList).To(Equal(array))

				// default department's project inherits invalid node pool list
				projectId, err := strconv.ParseInt(project3Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective = &array
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin = &department.OrgUnitDataID
				project.OrgUnitData.WorkspaceNodeTypes = &[]models.NodeType{nodeTypeRequestFromApi(nodeType1, cluster1Id)}
				project.OrgUnitData.TrainingNodeTypes = &[]models.NodeType{nodeTypeRequestFromApi(nodeType1, cluster1Id), nodeTypeRequestFromApi(nodeType2, cluster1Id)}
				_, err = dao.UpdateProject(ctx, project, false, false)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(project.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(array))
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(Equal(department.OrgUnitDataID))

				// default department's project with invalid node pool list and should change it's parent
				projectId, err = strconv.ParseInt(project4Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				array = []string{dummy2NodePoolName}
				project.OrgUnitData.DefaultNodePoolList = &array
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective = &array
				project.OrgUnitData.WorkspaceNodeTypes = &[]models.NodeType{nodeTypeRequestFromApi(nodeType1, cluster1Id)}
				project.OrgUnitData.TrainingNodeTypes = &[]models.NodeType{nodeTypeRequestFromApi(nodeType1, cluster1Id), nodeTypeRequestFromApi(nodeType2, cluster1Id)}
				_, err = dao.UpdateProject(ctx, project, false, false)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(*project.OrgUnitData.DefaultNodePoolList).To(Equal(array))
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(array))
				Expect(project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(BeNil())

				// department1Cluster1 department with invalid node pool list
				departmentId, err = strconv.ParseInt(department1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				array = []string{dummyNodePoolName}
				department.OrgUnitData.DefaultNodePoolList = &array
				_, err = dao.UpdateDepartment(ctx, department)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(*department.OrgUnitData.DefaultNodePoolList).To(Equal(array))

				// department1Cluster1 department's project inherits invalid node pool list
				projectId, err = strconv.ParseInt(project2Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective = &array
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin = &department.OrgUnitDataID
				_, err = dao.UpdateProject(ctx, project, false, false)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(project.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(array))
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(Equal(department.OrgUnitDataID))

				// department1Cluster1 child project1Cluster1 with valid node pool list (should not change)
				projectId, err = strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				array = []string{nodePool3Name}
				project.OrgUnitData.DefaultNodePoolList = &array
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective = &array
				_, err = dao.UpdateProject(ctx, project, false, false)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(*project.OrgUnitData.DefaultNodePoolList).To(Equal(array))

				// project1Cluster2 with invalid node pool list
				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				array = []string{dummy2NodePoolName}
				project.OrgUnitData.DefaultNodePoolList = &array
				project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective = &array
				_, err = dao.UpdateProject(ctx, project, false, false)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(*project.OrgUnitData.DefaultNodePoolList).To(Equal(array))
			})

			It("departments and projects", func() {
				err := handleDefaultNodePools(ctx, runaidb.DB(), testing.TestTenant, cluster1Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name})
				Expect(err).To(BeNil())

				departmentId, err := strconv.ParseInt(defaultDepartmentCluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err := dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(*department.OrgUnitData.DefaultNodePoolList).To(Equal(pq.StringArray{nodePool2Name}))
				Expect(department.OrgUnitData.TrainingNodeTypes).To(BeNil())
				Expect(len(*department.OrgUnitData.WorkspaceNodeTypes)).To(Equal(1))
				Expect(fmt.Sprint((*department.OrgUnitData.WorkspaceNodeTypes)[0].ID)).To(Equal(nodeType1.Id))

				projectId, err := strconv.ParseInt(project3Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err := dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(project.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(pq.StringArray{nodePool2Name}))
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(Equal(department.OrgUnitDataID))
				Expect(len(*project.OrgUnitData.WorkspaceNodeTypes)).To(Equal(1))
				Expect(fmt.Sprint((*project.OrgUnitData.WorkspaceNodeTypes)[0].ID)).To(Equal(nodeType1.Id))
				Expect(len(*project.OrgUnitData.TrainingNodeTypes)).To(Equal(2))
				Expect(fmt.Sprint((*project.OrgUnitData.TrainingNodeTypes)[0].ID)).To(Equal(nodeType1.Id))
				Expect(fmt.Sprint((*project.OrgUnitData.TrainingNodeTypes)[1].ID)).To(Equal(nodeType2.Id))

				projectId, err = strconv.ParseInt(project4Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(project.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(pq.StringArray{nodePool2Name}))
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(Equal(department.OrgUnitDataID))
				Expect(len(*project.OrgUnitData.WorkspaceNodeTypes)).To(Equal(1))
				Expect(fmt.Sprint((*project.OrgUnitData.WorkspaceNodeTypes)[0].ID)).To(Equal(nodeType1.Id))
				Expect(len(*project.OrgUnitData.TrainingNodeTypes)).To(Equal(2))
				Expect(fmt.Sprint((*project.OrgUnitData.TrainingNodeTypes)[0].ID)).To(Equal(nodeType1.Id))
				Expect(fmt.Sprint((*project.OrgUnitData.TrainingNodeTypes)[1].ID)).To(Equal(nodeType2.Id))

				departmentId, err = strconv.ParseInt(department1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				department, err = dao.GetDepartmentById(ctx, departmentId, "1", models.SlimVerbosity, nil)
				Expect(err).To(BeNil())
				Expect(department.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(department.OrgUnitData.WorkspaceNodeTypes).To(BeNil())
				Expect(department.OrgUnitData.TrainingNodeTypes).To(BeNil())

				projectId, err = strconv.ParseInt(project2Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(project.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(pq.StringArray{defaultNodePoolName}))
				Expect(project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(BeNil())
				Expect(project.OrgUnitData.WorkspaceNodeTypes).To(BeNil())
				Expect(project.OrgUnitData.TrainingNodeTypes).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster1.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(*project.OrgUnitData.DefaultNodePoolList).To(Equal(pq.StringArray{nodePool3Name}))
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(pq.StringArray{nodePool3Name}))
				Expect(project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(BeNil())

				err = handleDefaultNodePools(ctx, runaidb.DB(), testing.TestTenant, cluster2Id, []string{fmt.Sprint(defaultNodePoolId), fmt.Sprint(nodePool2Id), fmt.Sprint(nodePool3Id)}, []string{defaultNodePoolName, nodePool2Name, nodePool3Name})
				Expect(err).To(BeNil())

				projectId, err = strconv.ParseInt(project1Cluster2.Id, 10, 64)
				Expect(err).To(BeNil())
				project, err = dao.GetProjectById(ctx, projectId, "1", models.SlimVerbosity)
				Expect(err).To(BeNil())
				Expect(project.OrgUnitData.DefaultNodePoolList).To(BeNil())
				Expect(*project.OrgUnitData.EffectiveData.DefaultNodePoolListEffective).To(Equal(pq.StringArray{defaultNodePoolName}))
				Expect(project.OrgUnitData.EffectiveData.DefaultNodePoolListEffectiveOrigin).To(BeNil())
			})
		})
	})
})

func nodeTypeRequestFromApi(nodeType api.NodeType, clusterId uuid.UUID) models.NodeType {
	id, _ := strconv.ParseInt(nodeType.Id, 10, 64)
	return models.NodeType{
		IntBaseModel: models.IntBaseModel{
			ID: id,
		},
		Name:      nodeType.Name,
		ClusterId: clusterId,
		TenantId:  "1",
	}
}

func validateAllResources(resources []models.Resources) {
	Expect(len(resources)).To(Equal(3))
	nodePools := map[string]bool{}
	for _, resource := range resources {
		nodePools[resource.NodePoolId] = true
	}
	Expect(nodePools["1"]).To(BeTrue())
	Expect(nodePools["2"]).To(BeTrue())
	Expect(nodePools["3"]).To(BeTrue())
}

func getDepartmentCreateRequest(name string, clusterId string) api.DepartmentCreationRequest {
	return api.DepartmentCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		Resources: getResources(),
	}
}

func getProjectCreateRequest(name string, clusterId string, departmentId string) api.ProjectCreationRequest {
	return api.ProjectCreationRequest{
		Resources: getResources(),
		Name:      name,
		ClusterId: clusterId,
		ParentId:  &departmentId,
	}
}

func getResources() []api.ResourcesNullable {
	return []api.ResourcesNullable{{
		Gpu: api.ResourcesNullableGpu{
			Deserved:        ptr.To(float64(2)),
			Limit:           ptr.To(float64(2)),
			OverQuotaWeight: ptr.To(float64(2)),
		},
		Cpu: &api.ResourcesCpu{
			Deserved:        ptr.To(float64(20)),
			Limit:           ptr.To(float64(20)),
			OverQuotaWeight: ptr.To(float64(20)),
		},
		Memory: &api.ResourcesMemory{
			Deserved:        ptr.To(float64(200)),
			Limit:           ptr.To(float64(200)),
			OverQuotaWeight: ptr.To(float64(200)),
			Units:           api.UNITS_MB,
		},
		Rank: ptr.To("High"),
	}}
}
