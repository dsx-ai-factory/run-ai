package services

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"
	testinghelpers "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

// MockTestWorkloadDao for testing - enhanced to support different scenarios
type MockTestWorkloadDao struct {
	additionalFieldsToReturn models.AdditionalFields
	errorToReturn            *api.Error
	additionalFieldsSet      bool
}

func (m *MockTestWorkloadDao) DbGetWorkloadAdditionalFields(ctx context.Context, tenantId uint, workloadId uuid.UUID) (models.AdditionalFields, *api.Error) {
	if m.errorToReturn != nil {
		return nil, m.errorToReturn
	}
	if m.additionalFieldsSet {
		return m.additionalFieldsToReturn, nil
	}
	return models.AdditionalFields{
		"nimWorkload": "true",
	}, nil
}

func (m *MockTestWorkloadDao) SetAdditionalFields(additionalFields models.AdditionalFields) {
	m.additionalFieldsToReturn = additionalFields
	m.errorToReturn = nil
	m.additionalFieldsSet = true
}

func (m *MockTestWorkloadDao) SetError(err *api.Error) {
	m.errorToReturn = err
	m.additionalFieldsToReturn = nil
	m.additionalFieldsSet = false
}

func (m *MockTestWorkloadDao) Reset() {
	m.additionalFieldsToReturn = nil
	m.errorToReturn = nil
	m.additionalFieldsSet = false
}

var _ = Describe("validateRequestedNIMWorkloadMetrics", func() {
	var ctx context.Context
	var mockDao *MockTestWorkloadDao
	var tenantId uint
	var workloadId string

	oldClusterId := testinghelpers.TestUuid5
	newClusterId := testinghelpers.TestUuid6

	BeforeEach(func() {
		ctx = testinghelpers.SetContext(testinghelpers.TestTenant, testinghelpers.TestUser, []string{})
		mockDao = &MockTestWorkloadDao{}
		tenantId = 1
		workloadId = testinghelpers.TestUuid1

		// configure two clusters: one below NIM min version, one at min version
		checkerClusterClient.SetMock(
			checkerClusterClient.NewSimpleMockClient().
				WithCluster(oldClusterId, testinghelpers.BuildDisplayedCluster(oldClusterId, "2.22.0", nil)).
				WithCluster(newClusterId, testinghelpers.BuildDisplayedCluster(newClusterId, "2.23.0", nil)),
		)
		checker.InitClusterCapabilitiesChecker("", testinghelpers.GetDefaultClustersCacheConfig(), "")
	})

	Describe("NIM metrics detection", func() {
		It("correctly detects various workload NIM metrics", func() {
			// Test different types of workload NIM metrics
			nimMetrics := []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_RUNNING,
				api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_WAITING,
				api.WORKLOADMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL,
				api.WORKLOADMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC,
				api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES,
			}

			for _, metricType := range nimMetrics {
				nimMetricsRequested := nimMetricStrings[string(metricType)]
				Expect(nimMetricsRequested).To(BeTrue(), "NIM metric %s should be detected correctly", metricType)
			}
		})

		It("correctly detects various pod NIM metrics", func() {
			// Test different types of pod NIM metrics
			nimMetrics := []api.PodMetricType{
				api.PODMETRICTYPE_NIM_NUM_REQUESTS_RUNNING,
				api.PODMETRICTYPE_NIM_NUM_REQUEST_MAX,
				api.PODMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL,
				api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS,
				api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS,
				api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES,
			}

			for _, metricType := range nimMetrics {
				nimMetricsRequested := nimMetricStrings[string(metricType)]
				Expect(nimMetricsRequested).To(BeTrue(), "NIM metric %s should be detected correctly", metricType)
			}
		})

		It("returns nil when no NIM metrics are requested", func() {
			id := uuid.MustParse(oldClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_CPU_USAGE_CORES}, mockDao, tenantId, workloadId)
			Expect(err).To(BeNil())
		})
	})

	Describe("cluster capability validation", func() {
		It("returns error for workload NIM metrics on unsupported cluster", func() {
			id := uuid.MustParse(oldClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_WAITING}, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
		})

		It("returns nil for workload NIM metrics on supported cluster", func() {
			id := uuid.MustParse(newClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.WorkloadMetricType{api.WORKLOADMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL}, mockDao, tenantId, workloadId)
			Expect(err).To(BeNil())
		})

		It("returns error for pod NIM metrics on unsupported cluster", func() {
			id := uuid.MustParse(oldClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.PodMetricType{api.PODMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC}, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
		})

		It("returns nil for pod NIM metrics on supported cluster", func() {
			id := uuid.MustParse(newClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.PodMetricType{api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS}, mockDao, tenantId, workloadId)
			Expect(err).To(BeNil())
		})

		It("returns error for multiple workload NIM metrics on unsupported cluster", func() {
			id := uuid.MustParse(oldClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.WorkloadMetricType{
				api.WORKLOADMETRICTYPE_NIM_NUM_REQUEST_MAX,
				api.WORKLOADMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL,
			}, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
		})

		It("returns nil for multiple pod NIM metrics on supported cluster", func() {
			id := uuid.MustParse(newClusterId)
			err := validateRequestedNIMWorkloadMetrics(ctx, id, []api.PodMetricType{
				api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS,
				api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES,
			}, mockDao, tenantId, workloadId)
			Expect(err).To(BeNil())
		})
	})

	Describe("validateNIMWorkload - workload validation", func() {
		It("succeeds when workload has nimWorkload: true", func() {
			mockDao.SetAdditionalFields(models.AdditionalFields{"nimWorkload": "true"})

			err := validateNIMWorkload(ctx, mockDao, tenantId, workloadId)
			Expect(err).To(BeNil())
		})

		It("fails when workload does not have nimWorkload field", func() {
			mockDao.SetAdditionalFields(models.AdditionalFields{"someOtherField": "value"})

			err := validateNIMWorkload(ctx, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("workload is not identified as NVIDIA NIM"))
		})

		It("fails when workload has nimWorkload: false", func() {
			mockDao.SetAdditionalFields(models.AdditionalFields{"nimWorkload": "false"})

			err := validateNIMWorkload(ctx, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("workload is not identified as NVIDIA NIM"))
		})

		It("fails when additionalFields is nil", func() {
			mockDao.SetAdditionalFields(nil)

			err := validateNIMWorkload(ctx, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("workload is not identified as NVIDIA NIM"))
		})

		It("fails when nimWorkload is not a string", func() {
			mockDao.SetAdditionalFields(models.AdditionalFields{"nimWorkload": true})

			err := validateNIMWorkload(ctx, mockDao, tenantId, workloadId)
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("workload is not identified as NVIDIA NIM"))
		})

		It("fails with invalid workload ID", func() {
			err := validateNIMWorkload(ctx, mockDao, tenantId, "invalid-uuid")
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("workload Id must be a valid uuid."))
		})
	})
})

var _ = Describe("vLLM metrics", func() {
	var (
		ctx        context.Context
		mockDao    *MockTestWorkloadDao
		tenantId   uint
		workloadId string
	)

	oldClusterId := testinghelpers.TestUuid5
	newClusterId := testinghelpers.TestUuid6

	BeforeEach(func() {
		ctx = testinghelpers.SetContext(testinghelpers.TestTenant, testinghelpers.TestUser, []string{})
		mockDao = &MockTestWorkloadDao{}
		tenantId = 1
		workloadId = testinghelpers.TestUuid1
		checkerClusterClient.SetMock(
			checkerClusterClient.NewSimpleMockClient().
				WithCluster(oldClusterId, testinghelpers.BuildDisplayedCluster(oldClusterId, "2.99.9", nil)).
				WithCluster(newClusterId, testinghelpers.BuildDisplayedCluster(newClusterId, "3.0.0", nil)),
		)
		checker.InitClusterCapabilitiesChecker("", testinghelpers.GetDefaultClustersCacheConfig(), "")
	})

	It("maps every API metric to the recording-rule wrapper name", func() {
		expected := map[string]string{
			"VLLM_NUM_REQUESTS_RUNNING":                     "num_requests_running",
			"VLLM_NUM_REQUESTS_WAITING":                     "num_requests_waiting",
			"VLLM_REQUEST_SUCCESS":                          "request_success_per_second",
			"VLLM_REQUEST_SUCCESS_RATE":                     "request_success_rate",
			"VLLM_KV_CACHE_USAGE_PERC":                      "kv_cache_usage_perc",
			"VLLM_PROMPT_TOKENS_PER_SECOND":                 "prompt_tokens_per_second",
			"VLLM_GENERATION_TOKENS_PER_SECOND":             "generation_tokens_per_second",
			"VLLM_TIME_TO_FIRST_TOKEN_SECONDS":              "time_to_first_token_seconds_bucket",
			"VLLM_E2E_REQUEST_LATENCY_SECONDS":              "e2e_request_latency_seconds_bucket",
			"VLLM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES":  "time_to_first_token_seconds_bucket_percentile",
			"VLLM_E2E_REQUEST_LATENCY_SECONDS_PERCENTILES":  "e2e_request_latency_seconds_bucket_percentile",
			"VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES":   "request_queue_time_seconds_bucket_percentile",
			"VLLM_REQUEST_PREFILL_TIME_SECONDS_PERCENTILES": "request_prefill_time_seconds_bucket_percentile",
			"VLLM_REQUEST_DECODE_TIME_SECONDS_PERCENTILES":  "request_decode_time_seconds_bucket_percentile",
			"VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES":  "inter_token_latency_seconds_bucket_percentile",
		}

		Expect(vllmMetricNames).To(HaveLen(len(expected)))
		for metricType, metricName := range expected {
			Expect(string(vllmMetricNames[metricType])).To(Equal(metricName))
			Expect(runaimetrics.RunaiVLLMPerPodMetric{Name: vllmMetricNames[metricType]}.GetName()).
				To(Equal("runai_vllm_" + metricName + "_per_pod"))
			Expect(runaimetrics.RunaiVLLMPerWorkloadMetric{Name: vllmMetricNames[metricType]}.GetName()).
				To(Equal("runai_vllm_" + metricName + "_per_workload"))
		}
	})

	It("accepts the explicit vllmWorkload marker", func() {
		mockDao.SetAdditionalFields(models.AdditionalFields{"vllmWorkload": "true"})
		Expect(validateVLLMWorkload(ctx, mockDao, tenantId, workloadId)).To(BeNil())
	})

	It("rejects a missing vllmWorkload marker", func() {
		mockDao.SetAdditionalFields(models.AdditionalFields{})
		err := validateVLLMWorkload(ctx, mockDao, tenantId, workloadId)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("workload is not identified as vLLM"))
	})

	It("rejects a non-vLLM workload", func() {
		mockDao.SetAdditionalFields(models.AdditionalFields{"nimWorkload": "true"})
		err := validateVLLMWorkload(ctx, mockDao, tenantId, workloadId)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("workload is not identified as vLLM"))
	})

	It("rejects vLLM metrics immediately below the capability boundary", func() {
		mockDao.SetAdditionalFields(models.AdditionalFields{"vllmWorkload": "true"})
		err := validateMetricsClusterVersions(ctx, oldClusterId, []string{"VLLM_NUM_REQUESTS_RUNNING"}, supportedWorkloadsMetricsVersion)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("VLLM_NUM_REQUESTS_RUNNING metric is not supported"))
	})

	It("accepts vLLM metrics at the capability boundary", func() {
		err := validateMetricsClusterVersions(ctx, newClusterId, []string{"VLLM_NUM_REQUESTS_RUNNING"}, supportedWorkloadsMetricsVersion)
		Expect(err).To(BeNil())
	})

	It("does not validate ordinary metrics as vLLM", func() {
		err := validateRequestedVLLMWorkloadMetrics(ctx, []string{"CPU_USAGE_CORES"}, mockDao, tenantId, workloadId)
		Expect(err).To(BeNil())
	})

	It("gates every vLLM metric in workload and pod requests", func() {
		for metricType := range vllmMetricNames {
			Expect(supportedWorkloadsMetricsVersion).To(HaveKeyWithValue(metricType, checker.MinVersionVLLMMetrics))
			Expect(supportedPodMetricsVersion).To(HaveKeyWithValue(metricType, checker.MinVersionVLLMMetrics))
		}
	})
})

var _ = Describe("histogram bucket ordering", func() {
	It("orders numeric keys ascending with +Inf last per timestamp", func() {
		// Build a HistogramSeries with two timestamps and unsorted bucket keys
		ts1 := time.Unix(1000, 0)
		ts2 := time.Unix(2000, 0)
		series := []api.HistogramSeries{
			{
				Type: "NIM_TIME_TO_FIRST_TOKEN_SECONDS",
				Values: []api.HistogramValue{
					{Timestamp: &ts1, Data: map[string]string{"0.5": "3", "+Inf": "8", "0.01": "2", "0.2": "1"}},
					{Timestamp: &ts2, Data: map[string]string{"+Inf": "6", "0.2": "1", "0.01": "0"}},
				},
			},
		}

		ordered := convertHistogramSeriesOrdered(series)
		Expect(ordered).NotTo(BeNil())
		Expect(len(*ordered)).To(Equal(1))
		vals := (*ordered)[0].Values
		Expect(len(vals)).To(Equal(2))

		// Since map iteration is random, validate order by reconstructing keys via CSV builder which uses sorted headers
		resp := api.MetricsCompositeResponse{Histogram: ordered}
		csv := GetMetricsCsvComposite(&resp, []string{"NIM_TIME_TO_FIRST_TOKEN_SECONDS"})

		// Header should expand to ordered bucket columns
		Expect(csv[0]).To(Equal([]string{"Timestamp", "NIM_TIME_TO_FIRST_TOKEN_SECONDS[0.01]", "NIM_TIME_TO_FIRST_TOKEN_SECONDS[0.2]", "NIM_TIME_TO_FIRST_TOKEN_SECONDS[0.5]", "NIM_TIME_TO_FIRST_TOKEN_SECONDS[+Inf]"}))
		// Validate only the ordered values
		Expect(csv[1][1:]).To(Equal([]string{"2", "1", "3", "8"}))
		Expect(csv[2][1:]).To(Equal([]string{"0", "1", "", "6"}))
	})
})
