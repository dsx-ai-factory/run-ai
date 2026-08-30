package testing

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clustersOpenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/client/mocks"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/runaimetrics"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
	v1 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/v1"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

// TODO: uncheck when it is possible to mock metrics service more than once in the tests
var _ = Describe("Metrics v1 API tests", Ordered, func() {
	clusterId := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var departmentV1ApiService api.DepartmentsV1APIServicer
	var departmentApiService api.DepartmentsAPIServicer
	var projectV1ApiService api.ProjectsV1APIServicer
	var projectApiService api.ProjectsAPIServicer

	var defaultDepartmentId string
	var project1Id string
	var project1Name string
	var project2Id string
	var project2Name string
	// var project3Id string
	// var project3Name string
	// var department2Id string
	// var department2Name string
	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.17")),
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

	defaultNodePoolName := "default"
	defaultNodePoolId := float32(1)

	BeforeAll(func() {
		departmentV1ApiService = v1.NewDepartmentV1ApiService(&clusterServiceClient)
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		defaultDepartment, err := departmentApiService.CreateDepartment(ctx, getDefaultDepartmentV1Request(clusterId))
		if err != nil || defaultDepartment.Code != http.StatusCreated {
			Fail("Failed creating default department")
		}
		defaultDepartmentId = defaultDepartment.Body.(api.Department).Id

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
		projectV1ApiService = v1.NewProjectV1ApiService(&clusterServiceClient, services.ProjectDeletionService{})
		projectApiService = projects.NewProjectApiService(services.ProjectDeletionService{}, &clusterServiceClient, nil)
	})

	Context("metrics", func() {
		BeforeAll(func() {
			project1Name = createProjectName()
			request := getProjectV1Request(project1Name, defaultDepartmentId, clusterId)
			createResponse, err := projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project1Id = createResponse.Body.(api.Project).Id
			project2Name = createProjectName()
			request = getProjectV1Request(project2Name, defaultDepartmentId, clusterId)
			createResponse, err = projectApiService.CreateProject(ctx, request)
			if err != nil {
				Fail("Failed creating project")
			}
			Expect(createResponse.Code).To(Equal(http.StatusCreated))
			project2Id = createResponse.Body.(api.Project).Id
			//depReq := getDepartmentV1PutRequest(createDepartmentName())
			//createResponseDep, err := departmentV1ApiService.CreateDepartmentV1(ctx, clusterId, depReq)
			//if err != nil {
			//	Fail("Failed creating project")
			//}
			//Expect(createResponseDep.Code).To(Equal(http.StatusOK))
			//department2Id = strconv.Itoa(int(createResponseDep.Body.(api.DepartmentForGetV1).Id))
			//department2Name = createResponseDep.Body.(api.DepartmentForGetV1).Name
			//request = getProjectV1Request(createProjectName(), createResponseDep.Body.(api.DepartmentForGetV1).Id)
			//createResponse, err = projectV1ApiService.CreateProjectV1(ctx, clusterId, request, true)
			//if err != nil {
			//	Fail("Failed creating project")
			//}
			//Expect(createResponse.Code).To(Equal(http.StatusOK))
			//project3Id = strconv.Itoa(int(createResponse.Body.(api.ProjectV1).Id))
			//project3Name = createResponse.Body.(api.ProjectV1).Name
		})

		AfterAll(func() {
			project1, _ := strconv.ParseInt(project1Id, 10, 32)
			deleteResponse, err := projectApiService.DeleteProject(ctx, fmt.Sprint(project1), "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
			project2, _ := strconv.ParseInt(project2Id, 10, 32)
			deleteResponse, err = projectApiService.DeleteProject(ctx, fmt.Sprint(project2), "", false)
			if err != nil || deleteResponse.Code != http.StatusNoContent {
				Fail("Failed deleting")
			}
			//project3, _ := strconv.ParseInt(project3Id, 10, 32)
			//deleteResponse, err = projectV1ApiService.DeleteProjectV1(ctx, clusterId, int32(project3), true)
			//if err != nil || deleteResponse.Code != http.StatusOK {
			//	Fail("Failed deleting")
			//}
			//depId, _ := strconv.ParseInt(department2Id, 10, 32)
			//deleteResponse, err = departmentV1ApiService.DeleteDepartmentV1(ctx, clusterId, int32(depId))
			//if err != nil || deleteResponse.Code != http.StatusOK {
			//	Fail("Failed deleting")
			//}
		})

		//	Describe("single project metrics", func() {
		//		const baseTimestampUnix = 1708958503653
		//		const numOfSamples = 8
		//		BeforeAll(func() {
		//			metricMock := createCurrentMetricsMock(project1Name, "", "", "")
		//			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
		//				metrics.RunaiPendingWorkloadsPerQueueMetric{}.GetName():     GetMetricCurrentMockResult(metricMock),
		//				metrics.RunaiAllocatedGpuPerQueueMetric{}.GetName():         GetMetricCurrentMockResult(metricMock),
		//				metrics.RunaiAllocatedCpuPerQueueMetric{}.GetName():         GetMetricCurrentMockResult(metricMock),
		//				metrics.RunaiAllocatedMemoryMBPerQueueMetric{}.GetName(): GetMetricCurrentMockResult(metricMock),
		//
		//				metrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName(): GetTimerangeIntMetricsMock(1, baseTimestampUnix, numOfSamples, metrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName()),
		//				metrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.1, baseTimestampUnix, numOfSamples, metrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName()),
		//				metrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.2, baseTimestampUnix, numOfSamples, metrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName()),
		//				metrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.3, baseTimestampUnix, numOfSamples, metrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName()),
		//				metrics.RunaiDeservedCpuMetric{}.GetName():                    GetTimerangeFloatMetricsMock(0.4, baseTimestampUnix, numOfSamples, metrics.RunaiDeservedCpuMetric{}.GetName()),
		//				metrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.5, baseTimestampUnix, numOfSamples, metrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName()),
		//				metrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.6, baseTimestampUnix, numOfSamples, metrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName()),
		//				metrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName():      GetTimerangeFloatMetricsMock(1, baseTimestampUnix, numOfSamples, metrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName()),
		//				metrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(1.1, baseTimestampUnix, numOfSamples, metrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName()),
		//				metrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName():  GetTimerangeFloatMetricsMock(1.2, baseTimestampUnix, numOfSamples, metrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName()),
		//			})
		//		})
		//
		//		It("single project", func() {
		//			metricsGet, err := projectV1ApiService.V1K8sClustersClusterIdProjectsIdMetricsGet(ctx, clusterId, project1Id, time.Now(), time.Now(), 1, "") //clusterId string, id string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string
		//			if err != nil {
		//				Fail("Failed getting project metrics")
		//			}
		//			Expect(metricsGet.Code).To(Equal(http.StatusOK))
		//			response := metricsGet.Body.(api.ProjectMetricsV1Response)
		//			metadata := response.Metadata
		//			currentResources := response.Current.Resources
		//
		//			assertProjMetrics(project1Id, clusterId, defaultDepartmentId, metadata, currentResources, 2, 1)
		//
		//			overtimeResources := response.TimeRange.Resources
		//			validateOvertimeMetrics(overtimeResources, baseTimestampUnix, numOfSamples)
		//		})
		//	})

		Describe("cluster project metrics", func() {
			const baseTimestampUnix = 1708958503653
			const numOfSamples = 8
			BeforeAll(func() {
				metricMock := createCurrentMetricsMock(project1Name, project2Name, "", "")
				metrics_client.AddResponsesForMockMetricsServer(map[string]openapi.GetMetrics200Response{
					runaimetrics.RunaiPendingWorkloadsPerQueueMetric{}.GetName():  GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedGpuPerQueueMetric{}.GetName():      GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedCpuPerQueueMetric{}.GetName():      GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedMemoryMBPerQueueMetric{}.GetName(): GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedGpuCountPerProjectMetric{}.GetName(): GetMetricCurrentMockResult([]MetricsMockValue{
						{Project: project1Name, Value: "3"}, {Project: project2Name, Value: "1"},
					}),

					runaimetrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName(): GetTimerangeIntMetricsMock(1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.2, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.3, baseTimestampUnix, numOfSamples, runaimetrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName()),
					runaimetrics.RunaiDeservedCpuMetric{}.GetName():                    GetTimerangeFloatMetricsMock(0.4, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedCpuMetric{}.GetName()),
					runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.5, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.6, baseTimestampUnix, numOfSamples, runaimetrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName()),
					runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName():      GetTimerangeFloatMetricsMock(1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(1.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName():  GetTimerangeFloatMetricsMock(1.2, baseTimestampUnix, numOfSamples, runaimetrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName()),
				})
			})

			It("cluster projects", func() {
				metricsGet, err := projectV1ApiService.GetProjectsMetricsV1(ctx, clusterId, time.Now(), time.Now(), 1, "")
				if err != nil {
					Fail("Failed getting project metrics")
				}
				Expect(metricsGet.Code).To(Equal(http.StatusOK))
				response := metricsGet.Body.(api.ProjectsMetricsV1Response)
				Expect(len(response.Data)).To(Equal(2))
				firstProjectMetric := response.Data[0]
				secondProjectMetric := response.Data[1]
				if firstProjectMetric.Metadata.ProjectName == project2Name {
					firstProjectMetric = response.Data[1]
					secondProjectMetric = response.Data[0]
				}
				metadata := firstProjectMetric.Metadata
				currentResources := firstProjectMetric.Current.Resources
				assertProjMetrics(project1Id, clusterId, defaultDepartmentId, metadata, currentResources, 2, 1)

				metadata2 := secondProjectMetric.Metadata
				currentResources2 := secondProjectMetric.Current.Resources
				assertProjMetrics(project2Id, clusterId, defaultDepartmentId, metadata2, currentResources2, 3, 0)

				overtimeResources := response.TimeRange.Resources
				validateOvertimeMetrics(overtimeResources, baseTimestampUnix, numOfSamples, true)
			})

			It("project quota", func() {
				quotasGet, err := projectV1ApiService.GetProjectsQuotasV1(ctx, clusterId)
				Expect(err).To(BeNil())
				Expect(quotasGet.Code).To(Equal(http.StatusOK))
				quotas := quotasGet.Body.([]api.ProjectQuotasV1Inner)
				Expect(len(quotas)).To(Equal(2))
				proj1Quota := quotas[0]
				proj2Quota := quotas[1]
				if proj1Quota.Name == project2Name {
					proj1Quota = quotas[1]
					proj2Quota = quotas[0]
				}
				project1, _ := strconv.ParseInt(project1Id, 10, 32)
				Expect(proj1Quota.Id).To(Equal(int32(project1)))
				project2, _ := strconv.ParseInt(project2Id, 10, 32)
				Expect(proj2Quota.Id).To(Equal(int32(project2)))
				Expect(proj1Quota.Name).To(Equal(project1Name))
				Expect(proj2Quota.Name).To(Equal(project2Name))
				Expect(proj1Quota.ClusterUuid).To(Equal(clusterId))
				Expect(proj2Quota.ClusterUuid).To(Equal(clusterId))

				testTenant, _ := strconv.ParseInt(TestTenant, 10, 32)
				Expect(proj1Quota.TenantId).To(Equal(int32(testTenant)))
				Expect(proj2Quota.TenantId).To(Equal(int32(testTenant)))
				Expect(proj1Quota.DeservedGpus).To(Equal(float32(2)))
				Expect(proj2Quota.DeservedGpus).To(Equal(float32(2)))
				Expect(proj1Quota.AllocatedGpus).To(Equal(float32(3)))
				Expect(proj2Quota.AllocatedGpus).To(Equal(float32(1)))
			})
		})

		Describe("single department metrics", func() {
			const baseTimestampUnix = 1708958503653
			const numOfSamples = 8
			BeforeAll(func() {
				metricMock := createCurrentMetricsMock(project1Name, project2Name, "", "")
				metrics_client.AddResponsesForMockMetricsServer(map[string]openapi.GetMetrics200Response{
					runaimetrics.RunaiPendingWorkloadsPerQueueMetric{}.GetName():  GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedGpuPerQueueMetric{}.GetName():      GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedCpuPerQueueMetric{}.GetName():      GetMetricCurrentMockResult(metricMock),
					runaimetrics.RunaiAllocatedMemoryMBPerQueueMetric{}.GetName(): GetMetricCurrentMockResult(metricMock),

					runaimetrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName(): GetTimerangeIntMetricsMock(1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.2, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.3, baseTimestampUnix, numOfSamples, runaimetrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName()),
					runaimetrics.RunaiDeservedCpuMetric{}.GetName():                    GetTimerangeFloatMetricsMock(0.4, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedCpuMetric{}.GetName()),
					runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.5, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.6, baseTimestampUnix, numOfSamples, runaimetrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName()),
					runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName():      GetTimerangeFloatMetricsMock(1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(1.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName()),
					runaimetrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName():  GetTimerangeFloatMetricsMock(1.2, baseTimestampUnix, numOfSamples, runaimetrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName()),
				})
			})

			It("single department", func() {
				metricsGet, err := departmentV1ApiService.GetDepartmentMetricsV1(ctx, clusterId, defaultDepartmentId, time.Now(), time.Now(), 1, "") // clusterId string, id string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string
				if err != nil {
					Fail("Failed getting project metrics")
				}
				Expect(metricsGet.Code).To(Equal(http.StatusOK))
				response := metricsGet.Body.(api.DepartmentMetricsV1Response)
				metadata := response.Metadata
				currentResources := response.Current.Resources

				Expect(fmt.Sprint(metadata.DepartmentId)).To(Equal(defaultDepartmentId))
				Expect(metadata.ClusterId).To(Equal(clusterId))
				Expect(metadata.ClusterName).To(Equal(""))
				Expect(metadata.DepartmentName).To(Equal("default"))

				Expect(len(currentResources)).To(Equal(2))
				defaultNodePoolMetrics := currentResources[0]
				pool1NodePoolMetrics := currentResources[1]
				if defaultNodePoolMetrics.NodepoolName == "pool1" && pool1NodePoolMetrics.NodepoolName == "default" {
					defaultNodePoolMetrics = currentResources[1]
					pool1NodePoolMetrics = currentResources[0]
				}

				validateCurrentMetrics(defaultNodePoolMetrics, "default", 5, float32(5), float32(2))
				validateCurrentMetrics(pool1NodePoolMetrics, "pool1", 1, float32(1), float32(0))

				overtimeResources := response.TimeRange.Resources
				validateOvertimeMetrics(overtimeResources, baseTimestampUnix, numOfSamples, false)
			})
		})

		//Describe("cluster department metrics", func() {
		//	const baseTimestampUnix = 1708958503653
		//	const numOfSamples = 8
		//	BeforeAll(func() {
		//		metricMock := createCurrentMetricsMock(project1Name, project2Name, department2Name, project3Name)
		//		metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
		//			metrics.RunaiPendingWorkloadsPerQueueMetric{}.GetName():     GetMetricCurrentMockResult(metricMock),
		//			metrics.RunaiAllocatedGpuPerQueueMetric{}.GetName():         GetMetricCurrentMockResult(metricMock),
		//			metrics.RunaiAllocatedCpuPerQueueMetric{}.GetName():         GetMetricCurrentMockResult(metricMock),
		//			metrics.RunaiAllocatedMemoryMBPerQueueMetric{}.GetName(): GetMetricCurrentMockResult(metricMock),
		//
		//			metrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName(): GetTimerangeIntMetricsMock(1, baseTimestampUnix, numOfSamples, metrics.RunaiPendingWorkloadsCountPerQueuesMetric{}.GetName()),
		//			metrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.1, baseTimestampUnix, numOfSamples, metrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName()),
		//			metrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.2, baseTimestampUnix, numOfSamples, metrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName()),
		//			metrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.3, baseTimestampUnix, numOfSamples, metrics.RunaiGpuUtilizationPerNodepoolsMetric{}.GetName()),
		//			metrics.RunaiDeservedCpuMetric{}.GetName():                    GetTimerangeFloatMetricsMock(0.4, baseTimestampUnix, numOfSamples, metrics.RunaiDeservedCpuMetric{}.GetName()),
		//			metrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.5, baseTimestampUnix, numOfSamples, metrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName()),
		//			metrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.6, baseTimestampUnix, numOfSamples, metrics.RunaiCpuUtilizationPerNodepoolsMetric{}.GetName()),
		//			metrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName():      GetTimerangeFloatMetricsMock(1, baseTimestampUnix, numOfSamples, metrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName()),
		//			metrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(1.1, baseTimestampUnix, numOfSamples, metrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName()),
		//			metrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName():  GetTimerangeFloatMetricsMock(1.2, baseTimestampUnix, numOfSamples, metrics.RunaiMemoryUtilizationPerNodepoolsMetric{}.GetName()),
		//		})
		//	})
		//
		//	It("cluster department", func() {
		//		metricsGet, err := departmentV1ApiService.GetDepartmentsMetricsV1(ctx, clusterId, time.Now(), time.Now(), 1, "") //clusterId string, id string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string
		//		if err != nil {
		//			Fail("Failed getting project metrics")
		//		}
		//		Expect(metricsGet.Code).To(Equal(http.StatusOK))
		//		response := metricsGet.Body.(api.DepartmentsMetricsV1Response)
		//
		//		Expect(len(response.Data)).To(Equal(2))
		//		defaultDep := response.Data[0]
		//		firstDep := response.Data[1]
		//		if firstDep.Metadata.DepartmentName == "default" {
		//			defaultDep = response.Data[1]
		//			firstDep = response.Data[0]
		//		}
		//
		//		metadata := defaultDep.Metadata
		//		currentResources := defaultDep.Current.Resources
		//
		//		Expect(metadata.DepartmentId).To(Equal(defaultDepartmentId))
		//		Expect(metadata.ClusterId).To(Equal(clusterId))
		//		Expect(metadata.ClusterName).To(Equal(""))
		//		Expect(metadata.DepartmentName).To(Equal("default"))
		//
		//		Expect(len(currentResources)).To(Equal(2))
		//		defaultNodePoolMetrics := currentResources[0]
		//		pool1NodePoolMetrics := currentResources[1]
		//		if defaultNodePoolMetrics.NodepoolName == "pool1" && pool1NodePoolMetrics.NodepoolName == "default" {
		//			defaultNodePoolMetrics = currentResources[1]
		//			pool1NodePoolMetrics = currentResources[0]
		//		}
		//
		//		validateCurrentMetrics(defaultNodePoolMetrics, "default", 5, float32(5), float32(2))
		//		validateCurrentMetrics(pool1NodePoolMetrics, "pool1", 1, float32(1), float32(0))
		//
		//		metadata = firstDep.Metadata
		//		currentResources = firstDep.Current.Resources
		//
		//		Expect(fmt.Sprint(metadata.DepartmentId)).To(Equal(department2Id))
		//		Expect(metadata.ClusterId).To(Equal(clusterId))
		//		Expect(metadata.ClusterName).To(Equal(""))
		//		Expect(metadata.DepartmentName).To(Equal(department2Name))
		//
		//		Expect(len(currentResources)).To(Equal(2))
		//		defaultNodePoolMetrics = currentResources[0]
		//		pool1NodePoolMetrics = currentResources[1]
		//		if defaultNodePoolMetrics.NodepoolName == "pool1" && pool1NodePoolMetrics.NodepoolName == "default" {
		//			defaultNodePoolMetrics = currentResources[1]
		//			pool1NodePoolMetrics = currentResources[0]
		//		}
		//
		//		validateCurrentMetrics(defaultNodePoolMetrics, "default", 3, float32(3), float32(2))
		//		validateCurrentMetrics(pool1NodePoolMetrics, "pool1", 0, float32(0), float32(0))
		//
		//		overtimeResources := response.TimeRange.Resources
		//		validateOvertimeMetrics(overtimeResources, baseTimestampUnix, numOfSamples)
		//	})
		//})
	})
})

func assertProjMetrics(projectId string, clusterId string, departmentId string, metadata api.ProjectMetricsV1Metadata, currentResources []api.DepartmentMetricsV1CurrentResourcesInner, defaultPoolVal int32, pool1Val int32) {
	Expect(strconv.Itoa(int(metadata.ProjectId))).To(Equal(projectId))
	Expect(metadata.ClusterId).To(Equal(clusterId))
	Expect(metadata.ClusterName).To(Equal(""))
	Expect(fmt.Sprint(metadata.DepartmentId)).To(Equal(departmentId))
	Expect(metadata.DepartmentName).To(Equal("default"))
	Expect(len(currentResources)).To(Equal(2))
	defaultNodePoolMetrics := currentResources[0]
	pool1NodePoolMetrics := currentResources[1]
	if defaultNodePoolMetrics.NodepoolName == "pool1" && pool1NodePoolMetrics.NodepoolName == "default" {
		defaultNodePoolMetrics = currentResources[1]
		pool1NodePoolMetrics = currentResources[0]
	}
	validateCurrentMetrics(defaultNodePoolMetrics, "default", defaultPoolVal, float32(defaultPoolVal), float32(2))
	validateCurrentMetrics(pool1NodePoolMetrics, "pool1", pool1Val, float32(pool1Val), float32(0))
}

func validateOvertimeMetrics(overtimeResources []api.ResourcesDataOvertimeV1, baseTimestampUnix int, numOfSamples int, allClusterFetched bool) {
	Expect(len(overtimeResources)).To(Equal(numOfSamples))
	for i := 0; i < numOfSamples; i++ {
		resources := overtimeResources[i]
		Expect(resources.Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix) + int64(i*1000)))
		Expect(resources.NumberOfPendingWorkloads).To(Equal(int32(1)))
		Expect(resources.Gpu.Quota).To(Equal(getNextValue(0.1, i)))
		Expect(resources.Gpu.Allocated).To(Equal(getNextValue(0.2, i)))
		Expect(resources.Cpu.Quota).To(Equal((getNextValue(0.4, i)) * 1000))
		Expect(resources.Cpu.Allocated).To(Equal(getNextValue(0.5, i)))
		Expect(resources.Memory.Quota).To(Equal(getNextValue(1, i)))
		Expect(resources.Memory.Allocated).To(Equal(getNextValue(1.1, i)))
		if allClusterFetched {
			Expect(resources.Gpu.Utilization).To(Equal((getNextValue(0.3, i)) / 100))
			Expect(resources.Cpu.Utilization).To(Equal(getNextValue(0.6, i)))
			Expect(resources.Memory.Utilization).To(Equal(getNextValue(1.2, i)))
		} else {
			Expect(resources.Gpu.Utilization).To(Equal(float32(0)))
			Expect(resources.Cpu.Utilization).To(Equal(float32(0)))
			Expect(resources.Memory.Utilization).To(Equal(float32(0)))
		}
	}
}

func getNextValue(value float32, i int) float32 {
	return float32(math.Round(float64(value+float32(i)*0.1)*100) / 100)
}

func validateCurrentMetrics(resources api.DepartmentMetricsV1CurrentResourcesInner, nodePool string, pendingWorkloads int32, allocated float32, quota float32) {
	expected := float32(0)
	Expect(resources.NodepoolName).To(Equal(nodePool))
	Expect(resources.NumberOfPendingWorkloads).To(Equal(pendingWorkloads))
	Expect(resources.Gpu.Allocated).To(Equal(allocated))
	Expect(resources.Gpu.Quota).To(Equal(quota))
	Expect(resources.Cpu.Allocated).To(Equal(allocated))
	Expect(resources.Cpu.Quota).To(Equal(expected))
	Expect(resources.Memory.Allocated).To(Equal(allocated))
	Expect(resources.Memory.Quota).To(Equal(expected))
}

func createCurrentMetricsMock(proj1 string, proj2 string, dep string, depProj string) []MetricsMockValue {
	metricMock := []MetricsMockValue{{
		Project:    proj1,
		Department: "default",
		NodePool:   "default",
		Value:      "2",
	}, {
		Project:    proj1,
		Department: "default",
		NodePool:   "pool1",
		Value:      "1",
	}}
	if proj2 != "" {
		metricMock = append(metricMock, MetricsMockValue{
			Project:    proj2,
			Department: "default",
			NodePool:   "default",
			Value:      "3",
		}, MetricsMockValue{
			Project:    proj2,
			Department: "default",
			NodePool:   "pool1",
			Value:      "0",
		})
	}
	if dep != "" && depProj != "" {
		metricMock = append(metricMock, MetricsMockValue{
			Project:    depProj,
			Department: dep,
			NodePool:   "default",
			Value:      "3",
		}, MetricsMockValue{
			Project:    depProj,
			Department: dep,
			NodePool:   "pool1",
			Value:      "0",
		})
	}
	return metricMock
}
