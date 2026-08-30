package testing

import (
	"fmt"
	"net/http"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cluster_service "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/runaimetrics"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services/projects"
)

var _ = Describe("Project and Department metrics API tests", Ordered, func() {
	const (
		baseTimestampUnix = 1708958503653
		numOfSamples      = 8
	)

	var (
		clusterId            = uuid.New().String()
		ctx                  = SetContext(TestTenant, TestUser)
		projectId            string
		departmentId         string
		projectApiService    api.ProjectsAPIServicer
		departmentApiService api.DepartmentsAPIServicer

		dNodePoolName = "default"
		dNodePoolId   = float32(1)
	)

	BeforeAll(func() {
		metrics_client.AddResponsesForMockMetricsServer(map[string]openapi.GetMetrics200Response{
			runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedGpusCountPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiDeservedCpuMillicoresPerQueuesMetric{}.GetName(): GetTimerangeFloatMetricsMock(0.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedCpuMillicoresPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName():      GetTimerangeFloatMetricsMock(1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiDeservedMemoryMbPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.2, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedGpuCountPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(0.5, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedCpuCountPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(1.1, baseTimestampUnix, numOfSamples, runaimetrics.RunaiAllocatedMemoryMbPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiGpuMemoryUtilizationPerQueuesMetric{}.GetName():  GetTimerangeFloatMetricsMock(75, baseTimestampUnix, numOfSamples, runaimetrics.RunaiGpuMemoryUtilizationPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiGpuUtilizationPerQueuesMetric{}.GetName():        GetTimerangeFloatMetricsMock(80, baseTimestampUnix, numOfSamples, runaimetrics.RunaiGpuUtilizationPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiCpuUtilizationPerQueuesMetric{}.GetName():        GetTimerangeFloatMetricsMock(60, baseTimestampUnix, numOfSamples, runaimetrics.RunaiCpuUtilizationPerQueuesMetric{}.GetName()),
			runaimetrics.RunaiMemoryUtilizationPerQueuesMetric{}.GetName():     GetTimerangeFloatMetricsMock(55, baseTimestampUnix, numOfSamples, runaimetrics.RunaiMemoryUtilizationPerQueuesMetric{}.GetName()),
			"runai_workload_count_per_project":                                 GetTimerangeFloatMetricsMock(10, baseTimestampUnix, numOfSamples, "runai_workload_count_per_project"),
			"runai_workload_count_per_department":                              GetTimerangeFloatMetricsMock(15, baseTimestampUnix, numOfSamples, "runai_workload_count_per_department"),
			"runai_workload_count_per_category":                                GetTimerangeFloatMetricsMock(5, baseTimestampUnix, numOfSamples, "runai_workload_count_per_category"),
		})
	})

	BeforeAll(func() {
		clusterServiceClient := mocks.IClusterClient{}
		clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&cluster_service.ClustersMinimalData{
			Uuid:    clusterId,
			Version: *cluster_service.NewNullableString(ptr.To(services.GpuMemoryUtilizationPerOrgUnitClusterMinVersion)),
		}, nil)
		clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, mock.Anything).Return(&cluster_service.GetNodepoolsMinimal200Response{
			Nodepools: []cluster_service.NodepoolMinimal{
				{
					Name:                   dNodePoolName,
					Id:                     fmt.Sprint(dNodePoolId),
					SchedulingCapabilities: NpSchedCapabilitiesDefault(true, false),
				},
			},
		}, nil)
		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
		createDepartmentRequest := getFullDepartmentCreateRequest("dep1", clusterId)
		response, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		if err != nil {
			Fail("Failed creating parent department")
		}
		Expect(response.Code).To(Equal(http.StatusCreated))
		departmentId = response.Body.(api.Department).Id

		projectApiService = projects.NewProjectApiService(services.ProjectDeletionService{}, &clusterServiceClient, nil)
		request := getProjectCreateRequest(createProjectName(), "", departmentId, clusterId)
		createResponse, err := projectApiService.CreateProject(ctx, request)
		if err != nil {
			Fail("Failed creating project")
		}

		Expect(createResponse.Code).To(Equal(http.StatusCreated))
		created := createResponse.Body.(api.Project)
		projectId = created.Id
	})

	AfterAll(func() {
		_, err := departmentApiService.DeleteDepartment(ctx, departmentId)
		if err != nil {
			Fail("Failed deleting department")
		}

		_, err = projectApiService.DeleteProject(ctx, projectId, "", false)
		if err != nil {
			Fail("Failed deleting project")
		}
	})

	Context("Test metrics API", func() {
		It("Test project metrics sanity", func() {
			// get project metrics
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))

			metrics, ok := result.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(2))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_ALLOCATION))
			Expect(metrics.Measurements[0].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("0.2"))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION))
			Expect(metrics.Measurements[1].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[1].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("75"))
		})

		It("Test project metrics 2", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_QUOTA,
				api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))

			metrics, ok := result.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(6))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_ALLOCATION))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_QUOTA))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES))
			Expect(metrics.Measurements[3].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES))
			Expect(metrics.Measurements[4].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB))
			Expect(metrics.Measurements[5].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB))

			Expect(metrics.Measurements[0].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("0.2"))

			Expect(metrics.Measurements[1].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[1].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("0.1"))

			Expect(metrics.Measurements[2].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[2].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[2].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[2].Values)[0].Value).To(Equal("0.5"))

			Expect(metrics.Measurements[3].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[3].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[3].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[3].Values)[0].Value).To(Equal("0.1"))

			Expect(metrics.Measurements[4].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[4].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[4].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[4].Values)[0].Value).To(Equal("1.1"))

			Expect(metrics.Measurements[5].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[5].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[5].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[5].Values)[0].Value).To(Equal("1"))
		})

		It("Test project utilization and workloads count metrics", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_UTILIZATION,
				api.ORGUNITMETRICTYPE_CPU_UTILIZATION,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION,
				api.ORGUNITMETRICTYPE_WORKLOADS_COUNT,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))

			metrics, ok := result.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(4))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_UTILIZATION))
			Expect(metrics.Measurements[0].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("80"))

			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_UTILIZATION))
			Expect(metrics.Measurements[1].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("60"))

			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION))
			Expect(metrics.Measurements[2].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[2].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[2].Values)[0].Value).To(Equal("55"))

			Expect(metrics.Measurements[3].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_WORKLOADS_COUNT))
			Expect(metrics.Measurements[3].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[3].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[3].Values)[0].Value).To(Equal("10"))
		})

		It("Test department metrics sanity", func() {
			// get department metrics
			result, err := departmentApiService.GetDepartmentMetrics(ctx, departmentId, []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_GPU_QUOTA}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))

			metrics, ok := result.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(1))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_QUOTA))
			Expect(metrics.Measurements[0].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("0.1"))
		})

		It("Test department metrics 2", func() {
			result, err := departmentApiService.GetDepartmentMetrics(ctx, departmentId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_QUOTA, api.ORGUNITMETRICTYPE_GPU_ALLOCATION,
				api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))

			metrics, ok := result.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(6))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_QUOTA))
			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_ALLOCATION))
			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES))
			Expect(metrics.Measurements[3].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES))
			Expect(metrics.Measurements[4].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB))
			Expect(metrics.Measurements[5].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB))

			Expect(metrics.Measurements[0].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("0.1"))

			Expect(metrics.Measurements[1].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[1].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("0.2"))

			Expect(metrics.Measurements[2].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[2].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[2].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[2].Values)[0].Value).To(Equal("0.5"))

			Expect(metrics.Measurements[3].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[3].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[3].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[3].Values)[0].Value).To(Equal("0.1"))

			Expect(metrics.Measurements[4].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[4].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[4].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[4].Values)[0].Value).To(Equal("1.1"))

			Expect(metrics.Measurements[5].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[5].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[5].Values)[0].Timestamp.UnixMilli()).To(Equal(int64(baseTimestampUnix)))
			Expect((*metrics.Measurements[5].Values)[0].Value).To(Equal("1"))
		})

		It("Test department utilization and workloads count metrics", func() {
			result, err := departmentApiService.GetDepartmentMetrics(ctx, departmentId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_UTILIZATION,
				api.ORGUNITMETRICTYPE_CPU_UTILIZATION,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION,
				api.ORGUNITMETRICTYPE_WORKLOADS_COUNT,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))

			metrics, ok := result.Body.(api.MetricsResponse)
			Expect(ok).To(BeTrue())
			Expect(len(metrics.Measurements)).To(Equal(4))
			Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_GPU_UTILIZATION))
			Expect(metrics.Measurements[0].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[0].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[0].Values)[0].Value).To(Equal("80"))

			Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_UTILIZATION))
			Expect(metrics.Measurements[1].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[1].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[1].Values)[0].Value).To(Equal("60"))

			Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_CPU_MEMORY_UTILIZATION))
			Expect(metrics.Measurements[2].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[2].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[2].Values)[0].Value).To(Equal("55"))

			Expect(metrics.Measurements[3].Type).To(BeEquivalentTo(api.ORGUNITMETRICTYPE_WORKLOADS_COUNT))
			Expect(metrics.Measurements[3].Values).ToNot(BeNil())
			Expect(len(*metrics.Measurements[3].Values)).To(Equal(numOfSamples))
			Expect((*metrics.Measurements[3].Values)[0].Value).To(Equal("15"))
		})
	})

	Context("Test metrics API - groupBy", func() {
		It("groupBy Category with WORKLOADS_COUNT should succeed", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_WORKLOADS_COUNT,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "Category")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))
		})

		It("groupBy Category with non-WORKLOADS_COUNT metric should return BadRequest", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_ALLOCATION,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "Category")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))

			result, err = departmentApiService.GetDepartmentMetrics(ctx, departmentId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_UTILIZATION,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "Category")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))
		})

		It("invalid groupBy value should return BadRequest", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_WORKLOADS_COUNT,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "InvalidValue")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Test metrics API - failre test", func() {
		It("Too many metrics requested", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_QUOTA,
				api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB,
				api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_QUOTA,
				api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))

			result, err = departmentApiService.GetDepartmentMetrics(ctx, departmentId, []api.OrgUnitMetricType{
				api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_QUOTA,
				api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB,
				api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_QUOTA,
				api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_CPU_QUOTA_MILLICORES,
				api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB, api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB,
			}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))
		})

		It("Unexisting project", func() {
			result, err := projectApiService.GetProjectMetrics(ctx, "222", []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_GPU_ALLOCATION}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusNotFound))
		})

		It("Unexisting department", func() {
			result, err := departmentApiService.GetDepartmentMetrics(ctx, "222", []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_GPU_ALLOCATION}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusNotFound))
		})

		It("Too low cluster version for specific metric", func() {
			clusterServiceClient := mocks.IClusterClient{}
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&cluster_service.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *cluster_service.NewNullableString(ptr.To("2.20")),
			}, nil)
			departmentApiService2 := departments.NewDepartmentApiService(&clusterServiceClient, nil)

			result, err := departmentApiService2.GetDepartmentMetrics(ctx, departmentId, []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_CPU_MEMORY_QUOTA_MB, api.ORGUNITMETRICTYPE_CPU_ALLOCATION_MILLICORES, api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))

			projectApiService2 := projects.NewProjectApiService(services.ProjectDeletionService{}, &clusterServiceClient, nil)

			result, err = projectApiService2.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_GPU_MEMORY_UTILIZATION, api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusBadRequest))
			result, err = projectApiService2.GetProjectMetrics(ctx, projectId, []api.OrgUnitMetricType{api.ORGUNITMETRICTYPE_GPU_ALLOCATION, api.ORGUNITMETRICTYPE_CPU_MEMORY_ALLOCATION_MB}, time.Now().Add(-time.Hour), time.Now(), 10, "", "")
			Expect(err).To(BeNil())
			Expect(result.Code).To(Equal(http.StatusOK))
		})
	})
})
