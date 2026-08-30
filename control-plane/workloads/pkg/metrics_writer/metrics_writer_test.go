package metrics_writer

import (
	"fmt"
	ginkgoTesting "testing"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

const (
	TrainingType    = "Training"
	DefaultK8sPhase = "Running"
)

func TestMetricsWriter(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	config.Init()
	conf := config.Database()
	conf.SkipDefaultTransaction = true
	dao_utils.OpenPostgresForTesting(t, conf)

	RunSpecs(t, "metrics writer test")
	t.Cleanup(func() {
		testing.CleanTables()
	})
}

var _ = Describe("Metrics Writer tests", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadDao := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]())
	metricValue := float32(2.5)
	var workloadNotRunning uuid.UUID
	var workloadRunning1 uuid.UUID
	var workloadRunning2 uuid.UUID
	var workloadNotRunning2 uuid.UUID
	var workloadRunning3 uuid.UUID

	Context("Test metrics updater", func() {
		BeforeEach(func() {
			clusterId, err := uuid.NewRandom()
			Expect(err).To(BeNil())

			workloadNotRunning = uuid.New()
			workload := composeWorkload(api.PHASE_STOPPED, time.Now().Add(-3*time.Hour), clusterId, workloadNotRunning, 1)
			upsertErr := workloadDao.DbUpsertWorkload(ctx, workload)
			Expect(upsertErr).To(BeNil())

			workloadRunning1 = uuid.New()
			workload = composeWorkload(api.PHASE_RUNNING, time.Now().Add(-2*time.Hour), clusterId, workloadRunning1, 1)
			upsertErr = workloadDao.DbUpsertWorkload(ctx, workload)
			Expect(upsertErr).To(BeNil())

			workloadRunning2 = uuid.New()
			workload = composeWorkload(api.PHASE_RUNNING, time.Now().Add(-1*time.Hour), clusterId, workloadRunning2, 1)
			upsertErr = workloadDao.DbUpsertWorkload(ctx, workload)
			Expect(upsertErr).To(BeNil())

			workloadNotRunning2 = uuid.New()
			workload = composeWorkload(api.PHASE_STOPPED, time.Now().Add(-3*time.Hour), clusterId, workloadNotRunning2, 2)
			upsertErr = workloadDao.DbUpsertWorkload(ctx, workload)
			Expect(upsertErr).To(BeNil())

			workloadRunning3 = uuid.New()
			workload = composeWorkload(api.PHASE_RUNNING, time.Now().Add(-2*time.Hour), clusterId, workloadRunning3, 2)
			upsertErr = workloadDao.DbUpsertWorkload(ctx, workload)
			Expect(upsertErr).To(BeNil())

			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiIdleGpuCountPerWorkloadMetric{}.GetName(): getMetricResponse([]uuid.UUID{workloadRunning1, workloadRunning3}, fmt.Sprint(metricValue)),
			})
		})

		AfterEach(func() {
			// close(stopper)
			// cancel()
			testing.CleanTables()
		})

		It("test workloads", func() {
			results, apiErr := workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			Expect(len(*results)).To(Equal(3))
			for _, res := range *results {
				Expect(res.IdleAllocatedGpus).To(BeNil())
				Expect(res.IdleSince).To(BeNil())
			}
			results, apiErr = workloadDao.DbGetWorkloads(ctx, 2, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			Expect(len(*results)).To(Equal(2))
			for _, res := range *results {
				Expect(res.IdleAllocatedGpus).To(BeNil())
				Expect(res.IdleSince).To(BeNil())
			}

			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1", "2"})

			results, apiErr = workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			Expect(len(*results)).To(Equal(3))
			Expect((*results)[0].ID).To(Equal(workloadNotRunning))
			Expect((*results)[0].IdleAllocatedGpus).To(BeNil())
			Expect((*results)[0].IdleSince).To(BeNil())
			Expect((*results)[1].ID).To(Equal(workloadRunning1))
			Expect(*(*results)[1].IdleAllocatedGpus).To(Equal(metricValue))
			Expect((*results)[1].IdleSince).NotTo(BeNil())
			Expect((*results)[2].ID).To(Equal(workloadRunning2))
			Expect(*(*results)[2].IdleAllocatedGpus).To(Equal(float32(0)))
			Expect((*results)[2].IdleSince).To(BeNil())

			results, apiErr = workloadDao.DbGetWorkloads(ctx, 2, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			Expect(len(*results)).To(Equal(2))
			Expect((*results)[0].ID).To(Equal(workloadNotRunning2))
			Expect((*results)[0].IdleAllocatedGpus).To(BeNil())
			Expect((*results)[0].IdleSince).To(BeNil())
			Expect((*results)[1].ID).To(Equal(workloadRunning3))
			Expect(*(*results)[1].IdleAllocatedGpus).To(Equal(metricValue))
			Expect((*results)[1].IdleSince).NotTo(BeNil())
		})

		It("test idle_since is preserved when workload stays idle", func() {
			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr := workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			Expect(len(*results)).To(Equal(3))

			var idleWorkload *models.WorkloadView
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			Expect(idleWorkload).NotTo(BeNil())
			Expect(idleWorkload.IdleSince).NotTo(BeNil())
			firstIdleSince := *idleWorkload.IdleSince

			// Second update - workload still idle
			time.Sleep(100 * time.Millisecond)
			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr = workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			Expect(idleWorkload.IdleSince).NotTo(BeNil())
			// Verify idle_since was preserved (not updated)
			Expect(*idleWorkload.IdleSince).To(Equal(firstIdleSince))
		})

		It("test idle_since is cleared when workload stops being idle", func() {

			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr := workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			var idleWorkload *models.WorkloadView
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			Expect(idleWorkload).NotTo(BeNil())
			Expect(idleWorkload.IdleSince).NotTo(BeNil())

			// Change metrics to show workload is no longer idle
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiIdleGpuCountPerWorkloadMetric{}.GetName(): getMetricResponse([]uuid.UUID{workloadRunning1}, "0"),
			})

			// Second update - workload no longer idle
			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr = workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			// Verify idle_since was cleared
			Expect(idleWorkload.IdleSince).To(BeNil())
		})

		It("test idle_since tracks new idle period after becoming active", func() {

			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr := workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			var idleWorkload *models.WorkloadView
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			Expect(idleWorkload.IdleSince).NotTo(BeNil())
			firstIdleSince := *idleWorkload.IdleSince

			// Workload becomes active
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiIdleGpuCountPerWorkloadMetric{}.GetName(): getMetricResponse([]uuid.UUID{workloadRunning1}, "0"),
			})
			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr = workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			Expect(idleWorkload.IdleSince).To(BeNil())

			// Workload becomes idle again with a different metric response (new timestamp)
			time.Sleep(100 * time.Millisecond)
			newTimestamp := time.Now()
			metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
				runaimetrics.RunaiIdleGpuCountPerWorkloadMetric{}.GetName(): getMetricResponseWithTimestamp([]uuid.UUID{workloadRunning1}, fmt.Sprint(metricValue), newTimestamp),
			})
			updateWorkloadIdleGpus(ctx, dao_utils.DB(), []string{"1"})

			results, apiErr = workloadDao.DbGetWorkloads(ctx, 1, nil, false, 0, 0, "asc", "created_at", nil, "")
			Expect(apiErr).To(BeNil())
			for i := range *results {
				if (*results)[i].ID == workloadRunning1 {
					idleWorkload = &(*results)[i]
					break
				}
			}
			Expect(idleWorkload.IdleSince).NotTo(BeNil())
			// Verify idle_since was set again (workload became idle in a new period)
			Expect(idleWorkload.IdleSince).NotTo(BeNil())
			// Verify the timestamp is after the first idle period
			Expect((*idleWorkload.IdleSince).After(firstIdleSince)).To(BeTrue())
		})
	})
})

func getMetricResponse(workloadIds []uuid.UUID, metricValue string) openapi.GetMetrics200Response {
	return getMetricResponseWithTimestamp(workloadIds, metricValue, time.UnixMilli(1708958503653))
}

func getMetricResponseWithTimestamp(workloadIds []uuid.UUID, metricValue string, timestamp time.Time) openapi.GetMetrics200Response {
	results := []openapi.Result{}
	for _, workloadId := range workloadIds {
		measurements := []openapi.Measurement{
			{
				Timestamp: timestamp,
				Value:     metricValue,
			},
		}
		labels := map[string]string{
			"workload_id": workloadId.String(),
		}
		results = append(results, openapi.Result{
			Measurements: measurements,
			MetricLabels: labels,
		})

	}
	return openapi.GetMetrics200Response{
		Status: "success",
		Data:   results,
	}
}

func composeWorkload(phase api.Phase, createdAt time.Time, clusterId uuid.UUID, id uuid.UUID, tenantId uint) models.Workload {
	now := time.Now()
	return models.Workload{
		ID:                id,
		TenantID:          tenantId,
		UpdatedAt:         time.Now(),
		Phase:             phase,
		PhaseUpdatedAt:    time.Now(),
		Urls:              []string{"http://test.com"},
		CreatedAt:         createdAt,
		SubmittedBy:       testing.TestUser,
		Name:              "name",
		Source:            "source",
		Type:              TrainingType,
		ClusterID:         clusterId,
		ProjectName:       "project",
		ProjectID:         "project-id",
		Namespace:         "namespace",
		SupportDelete:     true,
		SupportSuspend:    true,
		K8sPhase:          DefaultK8sPhase,
		K8sPhaseUpdatedAt: now,
		AdditionalFields:  models.AdditionalFields{},
	}
}
