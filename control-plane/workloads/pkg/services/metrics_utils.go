package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"

	"github.com/google/uuid"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	clients "github.com/run-ai/runai/control-plane/workloads/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"
	"github.com/run-ai/runai/control-plane/workloads/pkg/setting_cache"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	maxAllowedMetricsInRequest = 10
	gpuMetricLabel             = "gpu"
	nodepoolLabel              = "nodepool"
	finishedReasonLabel        = "finished_reason"
)

var (
	nimWorkloadMetrics = map[api.WorkloadMetricType]bool{
		api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_RUNNING:                     true,
		api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_WAITING:                     true,
		api.WORKLOADMETRICTYPE_NIM_NUM_REQUEST_MAX:                          true,
		api.WORKLOADMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL:                    true,
		api.WORKLOADMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL:                    true,
		api.WORKLOADMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC:                     true,
		api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS:              true,
		api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS:             true,
		api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES:  true,
		api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES: true,
	}
	vllmMetricNames = map[string]runaimetrics.VLLMMetricName{
		string(api.WORKLOADMETRICTYPE_VLLM_NUM_REQUESTS_RUNNING):                     runaimetrics.VLLMNumRequestsRunning,
		string(api.WORKLOADMETRICTYPE_VLLM_NUM_REQUESTS_WAITING):                     runaimetrics.VLLMNumRequestsWaiting,
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS):                          runaimetrics.VLLMRequestSuccess,
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS_RATE):                     runaimetrics.VLLMRequestSuccessRate,
		string(api.WORKLOADMETRICTYPE_VLLM_KV_CACHE_USAGE_PERC):                      runaimetrics.VLLMKVCacheUsagePerc,
		string(api.WORKLOADMETRICTYPE_VLLM_PROMPT_TOKENS_PER_SECOND):                 runaimetrics.VLLMPromptTokensPerSecond,
		string(api.WORKLOADMETRICTYPE_VLLM_GENERATION_TOKENS_PER_SECOND):             runaimetrics.VLLMGenerationTokensPerSecond,
		string(api.WORKLOADMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS):              runaimetrics.VLLMTimeToFirstTokenSecondsBucket,
		string(api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS):             runaimetrics.VLLME2ERequestLatencySecondsBucket,
		string(api.WORKLOADMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES):  runaimetrics.VLLMTimeToFirstTokenSecondsPercentile,
		string(api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES): runaimetrics.VLLME2ERequestLatencySecondsPercentile,
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES):   runaimetrics.VLLMRequestQueueTimeSecondsPercentile,
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_PREFILL_TIME_SECONDS_PERCENTILES): runaimetrics.VLLMRequestPrefillTimeSecondsPercentile,
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_DECODE_TIME_SECONDS_PERCENTILES):  runaimetrics.VLLMRequestDecodeTimeSecondsPercentile,
		string(api.WORKLOADMETRICTYPE_VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES):  runaimetrics.VLLMInterTokenLatencySecondsPercentile,
	}

	allowedResponseLabelsForPodMetricTypes = map[string]map[string]bool{
		string(api.PODMETRICTYPE_GPU_UTILIZATION_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_GPU_SWAP_MEMORY_BYTES_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU): {
			gpuMetricLabel: true,
		},
		string(api.PODMETRICTYPE_VLLM_REQUEST_SUCCESS): {
			finishedReasonLabel: true,
		},
	}
	allowedResponseLabelsForWorkloadMetricTypes = map[string]map[string]bool{
		string(api.WORKLOADMETRICTYPE_GPU_ALLOCATION): {
			nodepoolLabel: true,
		},
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS): {
			finishedReasonLabel: true,
		},
	}

	advancedGpuMetrics = map[api.PodMetricType]bool{
		api.PODMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU:     true,
		api.PODMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU:                  true,
		api.PODMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU:                 true,
		api.PODMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU:              true,
		api.PODMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU:         true,
		api.PODMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU:         true,
		api.PODMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU:         true,
		api.PODMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU: true,
		api.PODMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU: true,
		api.PODMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU:    true,
		api.PODMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU:   true,
		api.PODMETRICTYPE_GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU:      true,
	}
)

// nimMetricStrings is a string-keyed set of all NIM metric names to allow
// generic validation across both workload and pod metric enums.
func newNimMetricStrings() map[string]bool {
	m := map[string]bool{}
	for t := range nimWorkloadMetrics {
		m[string(t)] = true
	}
	return m
}

var nimMetricStrings = newNimMetricStrings()

func isVLLMMetric(metricType string) bool {
	_, found := vllmMetricNames[metricType]
	return found
}

func withVLLMMetricsVersion(versions map[string]int) map[string]int {
	supportedVersions := make(map[string]int, len(versions)+len(vllmMetricNames))
	for metricType, version := range versions {
		supportedVersions[metricType] = version
	}
	for metricType := range vllmMetricNames {
		supportedVersions[metricType] = checker.MinVersionVLLMMetrics
	}
	return supportedVersions
}

type histogramMetricKind int

const (
	histogramMetricNone histogramMetricKind = iota
	histogramMetricPercentiles
	histogramMetricBuckets
)

// getHistogramMetricKind returns the histogram handling kind for a given metric type string.
// It supports both pod and workload metric enums by comparing their string values.
func getHistogramMetricKind(metricType string) histogramMetricKind {
	switch metricType {
	case string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_QUEUE_TIME_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_PREFILL_TIME_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_DECODE_TIME_SECONDS_PERCENTILES),
		string(api.WORKLOADMETRICTYPE_VLLM_INTER_TOKEN_LATENCY_SECONDS_PERCENTILES):
		return histogramMetricPercentiles
	case string(api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS),
		string(api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS),
		string(api.WORKLOADMETRICTYPE_VLLM_TIME_TO_FIRST_TOKEN_SECONDS),
		string(api.WORKLOADMETRICTYPE_VLLM_E2_E_REQUEST_LATENCY_SECONDS):
		return histogramMetricBuckets
	default:
		return histogramMetricNone
	}
}

func buildCsvFile[T ~string](metricTypes []T, workloadServiceResponse api.MetricsCompositeResponse, csvFileName string) runaihttp.CsvFile {
	types := make([]string, 0, len(metricTypes))
	for _, metricType := range metricTypes {
		types = append(types, string(metricType))
	}
	data := GetMetricsCsvComposite(&workloadServiceResponse, types)
	response := runaihttp.CsvFile{Data: data, FileName: csvFileName}
	return response
}

func buildMetricsComposite(
	ctx context.Context,
	metricType string,
	workloadServiceResponse *api.MetricsCompositeResponse,
	workloadServiceResponseCode *int,
	metricServiceResponsePerMetricType *openapi.GetMetrics200Response,
	isPodMetric bool,
) {
	if metricServiceResponsePerMetricType == nil {
		*workloadServiceResponseCode = http.StatusMultiStatus
		workloadServiceResponse.Measurements = append(workloadServiceResponse.Measurements, api.MeasurementResponse{
			Type: metricType,
		})
		return
	}

	if _, ok := allowedResponseLabelsForWorkloadMetricTypes[string(metricType)]; ok && !isPodMetric {
		groupingLabel := nodepoolLabel
		if metricType == string(api.WORKLOADMETRICTYPE_VLLM_REQUEST_SUCCESS) {
			groupingLabel = finishedReasonLabel
		}
		measurementsByLabels := parseWorkloadUMetricsResultToMeasurementsByLabel(ctx, string(metricType), groupingLabel, *metricServiceResponsePerMetricType)
		workloadServiceResponse.Measurements = append(workloadServiceResponse.Measurements, measurementsByLabels...)
		return
	}

	switch getHistogramMetricKind(metricType) {
	case histogramMetricPercentiles:
		hs := parsePercentileResultsToHistogram(ctx, string(metricType), *metricServiceResponsePerMetricType)
		if len(hs) > 0 {
			if workloadServiceResponse.Histogram == nil {
				var h []api.HistogramSeries
				workloadServiceResponse.Histogram = &h
			}
			*workloadServiceResponse.Histogram = append(*workloadServiceResponse.Histogram, hs...)
		}
	case histogramMetricBuckets:
		hs := parseBucketResultsToHistogram(ctx, string(metricType), *metricServiceResponsePerMetricType)
		if len(hs) > 0 {
			if workloadServiceResponse.Histogram == nil {
				var h []api.HistogramSeries
				workloadServiceResponse.Histogram = &h
			}
			*workloadServiceResponse.Histogram = append(*workloadServiceResponse.Histogram, hs...)
		}
	default:
		if isPodMetric {
			ms := parsePodMetricResultToMeasurements(ctx, string(metricType), *metricServiceResponsePerMetricType)
			if len(ms) > 0 {
				workloadServiceResponse.Measurements = append(workloadServiceResponse.Measurements, ms...)
			}
		} else {
			measurement := parseWorkloadMetricResultToMeasurement(ctx, string(metricType), *metricServiceResponsePerMetricType)
			workloadServiceResponse.Measurements = append(workloadServiceResponse.Measurements, measurement)
		}
	}
}

func parseWorkloadUMetricsResultToMeasurementsByLabel(ctx context.Context, metricType string, label string, response openapi.GetMetrics200Response) []api.MeasurementResponse {
	measurements := []api.MeasurementResponse{}
	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.MeasurementResponseValuesInner{},
	}
	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return []api.MeasurementResponse{measurement}
	}

	labelToMeasurement := make(map[string]api.MeasurementResponse)

	for _, result := range results {
		if _, ok := result.MetricLabels[label]; !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("No matching label %s in metrics result for %s", label, metricType)
			continue
		}
		if _, ok := labelToMeasurement[result.MetricLabels[label]]; ok {
			*labelToMeasurement[result.MetricLabels[label]].Values = append(*labelToMeasurement[result.MetricLabels[label]].Values, *parseMetricSingleResultToMeasurement(result)...)
			continue
		}

		labelToMeasurement[result.MetricLabels[label]] = api.MeasurementResponse{
			Type:   metricType,
			Labels: getLabelsForMetricResult(result, metricType, allowedResponseLabelsForWorkloadMetricTypes),
			Values: parseMetricSingleResultToMeasurement(result),
		}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("parsed metric %s measurements %+v by label %s, measurements by labels: %+v", metricType, results, label, labelToMeasurement)
	for _, labelMeasurement := range labelToMeasurement {
		measurements = append(measurements, labelMeasurement)
	}
	return measurements
}

func parseWorkloadMetricResultToMeasurement(ctx context.Context, metricType string, response openapi.GetMetrics200Response) api.MeasurementResponse {
	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.MeasurementResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return measurement
	}

	*measurement.Values = parseMetricResultsToMeasurement(results)
	return measurement
}

func parsePodMetricResultToMeasurements(ctx context.Context, metricType string, response openapi.GetMetrics200Response) []api.MeasurementResponse {
	measurement := api.MeasurementResponse{
		Type:   metricType,
		Values: &[]api.MeasurementResponseValuesInner{},
	}

	results := response.GetData()
	if len(results) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("No result return for metric type %s", metricType)
		return []api.MeasurementResponse{measurement}
	}

	measurementsResponse := []api.MeasurementResponse{}
	for _, result := range results {
		measurementsResponse = append(measurementsResponse, api.MeasurementResponse{
			Type:   metricType,
			Values: parseMetricSingleResultToMeasurement(result),
			Labels: getLabelsForMetricResult(result, metricType, allowedResponseLabelsForPodMetricTypes),
		})
	}
	return measurementsResponse
}

func parseMetricSingleResultToMeasurement(result openapi.Result) *[]api.MeasurementResponseValuesInner {
	values := &[]api.MeasurementResponseValuesInner{}
	for _, m := range result.Measurements {
		*values = append(*values, api.MeasurementResponseValuesInner{
			Value:     m.Value,
			Timestamp: convutils.Ref(m.Timestamp),
		})
	}

	return values
}

func parseMetricResultsToMeasurement(results []openapi.Result) []api.MeasurementResponseValuesInner {
	values := []api.MeasurementResponseValuesInner{}
	for _, result := range results {
		values = append(values, *parseMetricSingleResultToMeasurement(result)...)
	}

	return values
}

func getLabelsForMetricResult(result openapi.Result, metricType string, allowedResponseLabelsForMetricTypes map[string]map[string]bool) *map[string]string {
	allowedResponseLabels, found := allowedResponseLabelsForMetricTypes[metricType]
	if !found {
		return nil
	}

	metricResponseLabels := result.MetricLabels
	for label := range metricResponseLabels {
		if _, found = allowedResponseLabels[label]; !found {
			delete(metricResponseLabels, label)
		}
	}
	return &metricResponseLabels
}

// parsePercentileResultsToHistogram aggregates percentile-series (quantile_label) into a single HistogramSeries per metric type.
// Each timestamp will contain keys like p50, p90, p99 with their corresponding values as strings.
func parsePercentileResultsToHistogram(ctx context.Context, metricType string, response openapi.GetMetrics200Response) []api.HistogramSeries {
	results := response.GetData()
	if len(results) == 0 {
		return []api.HistogramSeries{}
	}

	// Map of timestamp -> percentile-key-to-value map (e.g., at each timestamp: {"p50":"...","p90":"..."})
	timestampToData := map[time.Time]map[string]string{}

	for _, result := range results {
		quantileLabel, ok := result.MetricLabels["quantile_label"]
		if !ok {
			if q, ok2 := result.MetricLabels["quantile"]; ok2 {
				quantileLabel = q
			} else {
				continue
			}
		}
		percentileKey := fmt.Sprintf("p%s", quantileLabel)
		for _, m := range result.Measurements {
			if _, ok := timestampToData[m.Timestamp]; !ok {
				timestampToData[m.Timestamp] = map[string]string{}
			}
			timestampToData[m.Timestamp][percentileKey] = m.Value
		}
	}

	if len(timestampToData) == 0 {
		return []api.HistogramSeries{}
	}

	// Sort timestamps for deterministic output
	timestamps := make([]time.Time, 0, len(timestampToData))
	for ts := range timestampToData {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })

	// Carry forward NaN values per percentile key across sorted timestamps
	initValue := "0"
	lastValidPerKey := map[string]string{}
	for _, ts := range timestamps {
		data := timestampToData[ts]
		for k, v := range data {
			if v == "NaN" {
				if lv, ok := lastValidPerKey[k]; ok {
					data[k] = lv
				} else {
					data[k] = initValue
				}
			} else {
				lastValidPerKey[k] = v
			}
		}
	}

	values := make([]api.HistogramValue, 0, len(timestamps))
	for _, ts := range timestamps {
		data := timestampToData[ts]
		if len(data) == 0 {
			// Drop timestamps that have no valid percentile values
			continue
		}
		tsCopy := ts
		values = append(values, api.HistogramValue{
			Timestamp: &tsCopy,
			Data:      data,
		})
	}

	return []api.HistogramSeries{
		{
			Type:   metricType,
			Values: values,
		},
	}
}

// parseBucketResultsToHistogram aggregates bucket-series (le label) into a single HistogramSeries per metric type.
// Each timestamp will contain keys of bucket upper-bounds (e.g., "0.08", "+Inf") mapped to counts as strings.
func parseBucketResultsToHistogram(ctx context.Context, metricType string, response openapi.GetMetrics200Response) []api.HistogramSeries {
	results := response.GetData()
	if len(results) == 0 {
		return []api.HistogramSeries{}
	}

	timestampToData := map[time.Time]map[string]string{}
	for _, result := range results {
		le, ok := result.MetricLabels["le"]
		if !ok {
			continue
		}
		for _, m := range result.Measurements {
			if _, ok := timestampToData[m.Timestamp]; !ok {
				timestampToData[m.Timestamp] = map[string]string{}
			}
			timestampToData[m.Timestamp][le] = m.Value
		}
	}

	if len(timestampToData) == 0 {
		return []api.HistogramSeries{}
	}

	timestamps := make([]time.Time, 0, len(timestampToData))
	for ts := range timestampToData {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })

	values := make([]api.HistogramValue, 0, len(timestamps))
	for _, ts := range timestamps {
		tsCopy := ts
		values = append(values, api.HistogramValue{
			Timestamp: &tsCopy,
			Data:      timestampToData[ts],
		})
	}

	return []api.HistogramSeries{
		{
			Type:   metricType,
			Values: values,
		},
	}
}

// convertHistogramSeriesOrdered returns a pointer to the provided histogram series slice after
// normalizing ordering for deterministic clients consumption.
// - Timestamps are assumed already sorted by parse functions
// - For bucket series (numeric keys and/or +Inf), ensure numeric ascending order with "+Inf" last
// - For non-bucket maps, fall back to lexicographic order
func convertHistogramSeriesOrdered(series []api.HistogramSeries) *[]api.HistogramSeries {
	for i := range series {
		values := series[i].Values
		for j := range values {
			data := values[j].Data
			if len(data) == 0 {
				continue
			}
			keys := make([]string, 0, len(data))
			isBucketLike := false
			for k := range data {
				keys = append(keys, k)
				if k == HistogramPlusInfBucketKey {
					isBucketLike = true
					continue
				}
				if _, err := strconv.ParseFloat(k, 64); err == nil {
					isBucketLike = true
				}
			}

			if isBucketLike {
				sort.Slice(keys, func(a, b int) bool {
					if keys[a] == HistogramPlusInfBucketKey {
						return false
					}
					if keys[b] == HistogramPlusInfBucketKey {
						return true
					}
					fa, errA := strconv.ParseFloat(keys[a], 64)
					fb, errB := strconv.ParseFloat(keys[b], 64)
					if errA == nil && errB == nil {
						return fa < fb
					}
					return keys[a] < keys[b]
				})
			} else {
				sort.Strings(keys)
			}

			// Rebuild map in chosen order (order may be used by some json encoders)
			ordered := make(map[string]string, len(keys))
			for _, k := range keys {
				ordered[k] = data[k]
			}
			series[i].Values[j].Data = ordered
		}
	}
	return &series
}

func queryWorkloadMetricType(ctx context.Context, workloadId string, clusterId uuid.UUID, jobIdFilter []string,
	start time.Time, end time.Time, metricType api.WorkloadMetricType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	var metric metrics_client.Metric
	workloadFilter := []string{workloadId}
	clusterFilter := []uuid.UUID{clusterId}

	switch metricType {
	case api.WORKLOADMETRICTYPE_GPU_UTILIZATION:
		metric = runaimetrics.RunaiGpuUtilizationPerWorkloadMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_GPU_MEMORY_USAGE_BYTES:
		metric = runaimetrics.RunaiGpuMemoryUsedBytesPerWorkloadMetric{
			Filters: runaimetrics.RunaiGpuMemoryUsedBytesPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_GPU_MEMORY_REQUEST_BYTES:
		metric = runaimetrics.RunaiGpuMemoryRequestedBytesPerWorkloadMetric{
			Filters: runaimetrics.RunaiGpuMemoryRequestedBytesPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_CPU_USAGE_CORES:
		metric = runaimetrics.RunaiCpuUsedCoresPerWorkloadMetric{
			Filters: runaimetrics.RunaiCpuUsedCoresPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_CPU_REQUEST_CORES:
		metric = runaimetrics.RunaiCpuRequestedCoresPerWorkloadMetric{
			Filters: runaimetrics.RunaiCpuRequestedCoresPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_CPU_LIMIT_CORES:
		metric = runaimetrics.RunaiCpuLimitCoresPerWorkloadMetric{
			Filters: runaimetrics.RunaiCpuLimitCoresPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_CPU_MEMORY_USAGE_BYTES:
		metric = runaimetrics.RunaiMemoryUsedBytesPerWorkloadMetric{
			Filters: runaimetrics.RunaiMemoryUsedBytesPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_CPU_MEMORY_REQUEST_BYTES:
		metric = runaimetrics.RunaiMemoryRequestedBytesPerWorkloadMetric{
			Filters: runaimetrics.RunaiMemoryRequestedBytesPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_CPU_MEMORY_LIMIT_BYTES:
		metric = runaimetrics.RunaiMemoryLimitBytesPerWorkloadMetric{
			Filters: runaimetrics.RunaiMemoryLimitBytesPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_POD_COUNT:
		metric = runaimetrics.RunaiPodCountPerWorkloadMetric{
			Filters: runaimetrics.RunaiPodCountPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_RUNNING_POD_COUNT:
		metric = runaimetrics.RunaiRunningPodCountPerWorkloadMetric{
			Filters: runaimetrics.RunaiRunningPodCountPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_GPU_ALLOCATION:
		metric = runaimetrics.RunaiAllocatedGpuCountPerWorkloadMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
				JobId:      jobIdFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_RUNNING:
		metric = runaimetrics.RunaiNimNumRequestsRunningPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimNumRequestsRunningPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_WAITING:
		metric = runaimetrics.RunaiNimNumRequestsWaitingPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimNumRequestsWaitingPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_NUM_REQUEST_MAX:
		metric = runaimetrics.RunaiNimNumRequestMaxPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimNumRequestMaxPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL:
		metric = runaimetrics.RunaiNimRequestSuccessTotalPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimRequestSuccessTotalPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL:
		metric = runaimetrics.RunaiNimRequestFailureTotalPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimRequestFailureTotalPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC:
		metric = runaimetrics.RunaiNimGpuCacheUsagePercPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimGpuCacheUsagePercPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS:
		metric = runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS:
		metric = runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerWorkloadMetric{
			Filters: runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES:
		metric = runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadMetric{
			Filters: runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES:
		metric = runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadMetric{
			Filters: runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.WORKLOADMETRICTYPE_NVLINK_BANDWIDTH_TOTAL:
		metric = runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesMetric{
			Filters: runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesFilters{
				ClusterId: clusterFilter,
				Workload:  workloadFilter,
			},
		}
	default:
		if name, found := vllmMetricNames[string(metricType)]; found {
			metric = runaimetrics.RunaiVLLMPerWorkloadMetric{
				Name: name,
				Filters: runaimetrics.RunaiVLLMPerWorkloadFilters{
					ClusterId:  clusterFilter,
					WorkloadId: workloadFilter,
				},
			}
			break
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func queryPodMetricType(ctx context.Context, podId string, clusterId uuid.UUID,
	start time.Time, end time.Time, metricType api.PodMetricType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	var metric metrics_client.Metric
	podFilter := []string{podId}
	clusterFilter := []uuid.UUID{clusterId}

	switch metricType {
	case api.PODMETRICTYPE_GPU_UTILIZATION_PER_GPU:
		metric = runaimetrics.RunaiGpuUtilizationPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerPodPerGpuFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_UTILIZATION:
		metric = runaimetrics.RunaiGpuUtilizationPerPodMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU:
		metric = runaimetrics.RunaiGpuMemoryUsedBytesPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuMemoryUsedBytesPerPodPerGpuFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES:
		metric = runaimetrics.RunaiGpuMemoryUsedBytesPerPodMetric{
			Filters: runaimetrics.RunaiGpuMemoryUsedBytesPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_CPU_USAGE_CORES:
		metric = runaimetrics.RunaiCpuUsedCoresPerPodMetric{
			Filters: runaimetrics.RunaiCpuUsedCoresPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES:
		metric = runaimetrics.RunaiMemoryUsedBytesPerPodMetric{
			Filters: runaimetrics.RunaiMemoryUsedBytesPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuGREngineActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuGREngineActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuSmActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuSmActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU:
		metric = runaimetrics.RunaiGpuSmOccupancyPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuSmOccupancyPerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeTensorActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeTensorActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeFp64ActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeFp64ActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeFp32ActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeFp32ActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeFp16ActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeFp16ActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU:
		metric = runaimetrics.RunaiGpuDramActivePerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuDramActivePerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuNvlinkTxBytesPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuNvlinkTxBytesPerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuNvlinkRxBytesPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuNvlinkRxBytesPerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuPcieTxBytesPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPcieTxBytesPerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuPcieRxBytesPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPcieRxBytesPerPodPerGpuFilters{
				PodId: podFilter,
			},
		}
	case api.PODMETRICTYPE_GPU_SWAP_MEMORY_BYTES_PER_GPU:
		metric = runaimetrics.RunaiGpuSwapMemoryBytesPerPodPerGpuMetric{
			Filters: runaimetrics.RunaiGpuSwapMemoryBytesPerPodPerGpuFilters{
				PodId: podFilter,
			},
		}

	case api.PODMETRICTYPE_NIM_NUM_REQUESTS_RUNNING:
		metric = runaimetrics.RunaiNimNumRequestsRunningPerPodMetric{
			Filters: runaimetrics.RunaiNimNumRequestsRunningPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_NUM_REQUESTS_WAITING:
		metric = runaimetrics.RunaiNimNumRequestsWaitingPerPodMetric{
			Filters: runaimetrics.RunaiNimNumRequestsWaitingPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_NUM_REQUEST_MAX:
		metric = runaimetrics.RunaiNimNumRequestMaxPerPodMetric{
			Filters: runaimetrics.RunaiNimNumRequestMaxPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL:
		metric = runaimetrics.RunaiNimRequestSuccessTotalPerPodMetric{
			Filters: runaimetrics.RunaiNimRequestSuccessTotalPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL:
		metric = runaimetrics.RunaiNimRequestFailureTotalPerPodMetric{
			Filters: runaimetrics.RunaiNimRequestFailureTotalPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC:
		metric = runaimetrics.RunaiNimGpuCacheUsagePercPerPodMetric{
			Filters: runaimetrics.RunaiNimGpuCacheUsagePercPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	// histogram metrics
	case api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS:
		metric = runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerPodMetric{
			Filters: runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS:
		metric = runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerPodMetric{
			Filters: runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES:
		metric = runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodMetric{
			Filters: runaimetrics.RunaiNimTimeToFirstTokenSecondsBucketPercentilePerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS_PERCENTILES:
		metric = runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodMetric{
			Filters: runaimetrics.RunaiNimE2eRequestLatencySecondsBucketPercentilePerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU:
		metric = runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesMetric{
			Filters: runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesFilters{
				ClusterId: clusterFilter,
				Pod:       podFilter,
			},
			GroupBy: metrics_client.GroupByGpu,
		}
	case api.PODMETRICTYPE_NVLINK_BANDWIDTH_TOTAL:
		metric = runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesMetric{
			Filters: runaimetrics.RunaiDcgmNvlinkBandwidthTotalBytesFilters{
				ClusterId: clusterFilter,
				Pod:       podFilter,
			},
		}
	default:
		if name, found := vllmMetricNames[string(metricType)]; found {
			metric = runaimetrics.RunaiVLLMPerPodMetric{
				Name: name,
				Filters: runaimetrics.RunaiVLLMPerPodFilters{
					ClusterId: clusterFilter,
					PodId:     podFilter,
				},
			}
			break
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func queryInferenceWorkloadMetricType(ctx context.Context, workloadId string, clusterId uuid.UUID,
	start time.Time, end time.Time, metricType api.InferenceWorkloadMetricType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	var metric metrics_client.Metric
	workloadFilter := []string{workloadId}
	clusterFilter := []uuid.UUID{clusterId}

	switch metricType {
	case api.INFERENCEWORKLOADMETRICTYPE_THROUGHPUT:
		metric = runaimetrics.RunaiInferenceThroughputPerWorkloadMetric{
			Filters: runaimetrics.RunaiInferenceThroughputPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	case api.INFERENCEWORKLOADMETRICTYPE_LATENCY:
		metric = runaimetrics.RunaiInferenceLatencyPerWorkloadMetric{
			Filters: runaimetrics.RunaiInferenceLatencyPerWorkloadFilters{
				ClusterId:  clusterFilter,
				WorkloadId: workloadFilter,
			},
		}
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func queryInferencePodMetricType(ctx context.Context, podId string, clusterId uuid.UUID,
	start time.Time, end time.Time, metricType api.InferencePodMetricType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	var metric metrics_client.Metric
	podFilter := []string{podId}
	clusterFilter := []uuid.UUID{clusterId}

	switch metricType {
	case api.INFERENCEPODMETRICTYPE_THROUGHPUT:
		metric = runaimetrics.RunaiInferenceThroughputPerPodMetric{
			Filters: runaimetrics.RunaiInferenceThroughputPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	case api.INFERENCEPODMETRICTYPE_LATENCY:
		metric = runaimetrics.RunaiInferenceLatencyPerPodMetric{
			Filters: runaimetrics.RunaiInferenceLatencyPerPodFilters{
				ClusterId: clusterFilter,
				PodId:     podFilter,
			},
		}
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func queryElementMetricType(ctx context.Context, elementId string, clusterId uuid.UUID,
	start time.Time, end time.Time, metricType api.ElementMetricType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	elementFilter := []string{elementId}
	clusterFilter := []uuid.UUID{clusterId}

	var name runaimetrics.ElementMetricName
	switch metricType {
	case api.ELEMENTMETRICTYPE_GPU_UTILIZATION:
		name = runaimetrics.ElementGpuUtilization
	case api.ELEMENTMETRICTYPE_GPU_ALLOCATION:
		name = runaimetrics.ElementGpuAllocation
	case api.ELEMENTMETRICTYPE_GPU_MEMORY_USAGE_BYTES:
		name = runaimetrics.ElementGpuMemoryUsedBytes
	case api.ELEMENTMETRICTYPE_GPU_MEMORY_REQUEST_BYTES:
		name = runaimetrics.ElementGpuMemoryRequestedBytes
	case api.ELEMENTMETRICTYPE_CPU_USAGE_CORES:
		name = runaimetrics.ElementCpuUsedCores
	case api.ELEMENTMETRICTYPE_CPU_REQUEST_CORES:
		name = runaimetrics.ElementCpuRequestedCores
	case api.ELEMENTMETRICTYPE_CPU_LIMIT_CORES:
		name = runaimetrics.ElementCpuLimitCores
	case api.ELEMENTMETRICTYPE_CPU_MEMORY_USAGE_BYTES:
		name = runaimetrics.ElementMemoryUsedBytes
	case api.ELEMENTMETRICTYPE_CPU_MEMORY_REQUEST_BYTES:
		name = runaimetrics.ElementMemoryRequestedBytes
	case api.ELEMENTMETRICTYPE_CPU_MEMORY_LIMIT_BYTES:
		name = runaimetrics.ElementMemoryLimitBytes
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown element metric type %s.", metricType)
		return nil
	}

	metric := runaimetrics.NewPerElementMetric(name, clusterFilter, elementFilter)
	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func submitMetric(ctx context.Context, metric metrics_client.Metric, metricType string, start time.Time, end time.Time, numberOfSamples int32) *openapi.GetMetrics200Response {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("submitting metric %+v, type %s, start %v, end %v, samples %v", metric, metricType, start, end, numberOfSamples)
	q, err := metrics_client.NewMetricQuery(metric)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to construct query for metric type %s. err: %v", metricType, err)
		return nil
	}

	q.SetRange(start, end)
	q.SetDefaultNumOfSamples(numberOfSamples)
	res, err := q.Query(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to query metric type %s. err: %v", metricType, err)
		return nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("submitted metric %+v, type %s, start %v, end %v, samples %v. result: %+v", metric, metricType, start, end, numberOfSamples, res)

	return res
}

func wasAdvancedMetricRequested(metricTypes []api.PodMetricType) bool {
	for _, metricType := range metricTypes {
		if _, found := advancedGpuMetrics[metricType]; found {
			return true
		}
	}

	return false
}

func isAdvancedGpuMetricsFFOn(ctx context.Context, client clients.TenantSettingsClient) bool {
	tenantId := convutils.Deref(runaictx.GetString(ctx, runaictx.CfTenantId), "")
	if tenantId == "" {
		return isAdvancedGpuMetricsFFOnSync(ctx, client)
	}

	cachedValue, err := cache.Get[setting_cache.AdvancedGpuMetricsSettingCacheValue](ctx, tenantId)
	if err != nil {
		logging.Warnf("Failed to get advanced GPU metrics setting from cache for tenant %s: %v", tenantId, err)
	}
	if cachedValue != nil {
		return cachedValue.IsOn
	}

	isOn := isAdvancedGpuMetricsFFOnSync(ctx, client)
	err = cache.Set[setting_cache.AdvancedGpuMetricsSettingCacheValue](ctx, tenantId, setting_cache.AdvancedGpuMetricsSettingCacheValue{IsOn: isOn})
	if err != nil {
		logging.Warnf("Failed to set advanced GPU metrics setting in cache for tenant %s: %v", tenantId, err)
	}
	return isOn
}

func isAdvancedGpuMetricsFFOnSync(ctx context.Context, client clients.TenantSettingsClient) bool {
	enabled, err := client.GetAdvancedGpuMetricsSetting(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get advanced gpu metrics setting, assuming the FF is Off. err: %v", err)
		return false
	}

	return enabled
}

func validateRequestedMetrics(metricTypesLen int) *api.Error {
	if metricTypesLen > maxAllowedMetricsInRequest {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("Too many metrics requested. Max allowed is %d", maxAllowedMetricsInRequest), nil)
	}

	return nil
}

func validateRequestedAdvancedMetrics(ctx context.Context, tenantSettingsClient clients.TenantSettingsClient, metricTypes []api.PodMetricType) *api.Error {
	if !wasAdvancedMetricRequested(metricTypes) {
		return nil
	}

	if !isAdvancedGpuMetricsFFOn(ctx, tenantSettingsClient) {
		return api.NewError(http.StatusBadRequest, "Advanced GPU metrics are not enabled", nil)
	}

	return nil
}

func validateRequestedInferenceWorkloadMetrics(ctx context.Context, clusterId uuid.UUID) *api.Error {
	// validate that the requested metrics are supported in the cluster version

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionInferenceMetricsThroughputLatency)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "The selected inference metrics are not supported in the cluster", nil)
	}

	return nil
}

func validateMetricsClusterVersions[T ~string](ctx context.Context, clusterId string, metricsTypes []T, supportedMetricsVersion map[string]int) *api.Error {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	for _, metricType := range metricsTypes {
		version, exists := supportedMetricsVersion[string(metricType)]
		if exists && version != 0 {
			supported, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, version)
			if err != nil {
				return api.NewError(http.StatusInternalServerError, err.Error(), nil)
			}
			if !supported {
				return api.NewError(http.StatusBadRequest, fmt.Sprintf("%s metric is not supported for cluster %s", metricType, clusterId), nil)
			}
		}
	}
	return nil
}

func validateRequestedNIMWorkloadMetrics[T ~string](ctx context.Context, clusterId uuid.UUID, metricTypes []T, workloadDao WorkloadAdditionalFieldsGetter, tenantId uint, workloadId string) *api.Error {
	nimMetricsRequested := false
	for _, metricType := range metricTypes {
		if nimMetricStrings[string(metricType)] {
			nimMetricsRequested = true
			break
		}
	}

	if !nimMetricsRequested {
		return nil
	}

	if err := validateNIMWorkload(ctx, workloadDao, tenantId, workloadId); err != nil {
		return err
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId.String())
	clusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionNIMMetrics)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if !clusterSupport {
		return api.NewError(http.StatusBadRequest, "NIM metrics are not supported in the cluster", nil)
	}

	return nil
}

func validateRequestedVLLMWorkloadMetrics[T ~string](ctx context.Context, metricTypes []T, workloadDao WorkloadAdditionalFieldsGetter, tenantId uint, workloadId string) *api.Error {
	vllmMetricsRequested := false
	for _, metricType := range metricTypes {
		if isVLLMMetric(string(metricType)) {
			vllmMetricsRequested = true
			break
		}
	}

	if !vllmMetricsRequested {
		return nil
	}

	if err := validateVLLMWorkload(ctx, workloadDao, tenantId, workloadId); err != nil {
		return err
	}

	return nil
}

type WorkloadAdditionalFieldsGetter interface {
	DbGetWorkloadAdditionalFields(ctx context.Context, tenantId uint, workloadId uuid.UUID) (models.AdditionalFields, *api.Error)
}

func validateNIMWorkload(ctx context.Context, workloadDao WorkloadAdditionalFieldsGetter, tenantId uint, workloadId string) *api.Error {
	workloadUuid, err := uuid.Parse(workloadId)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)
	}

	additionalFields, dbErr := workloadDao.DbGetWorkloadAdditionalFields(ctx, tenantId, workloadUuid)
	if dbErr != nil {
		return dbErr
	}

	if nimWorkloadStr, ok := additionalFields[clusterCommonConstants.AdditionalNIMWorkload].(string); !ok || nimWorkloadStr != "true" {
		return api.NewError(http.StatusBadRequest, "workload is not identified as NVIDIA NIM", nil)
	}

	return nil
}

func validateVLLMWorkload(ctx context.Context, workloadDao WorkloadAdditionalFieldsGetter, tenantId uint, workloadId string) *api.Error {
	workloadUuid, err := uuid.Parse(workloadId)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("workload Id must be a valid uuid. <%s>", workloadId), err)
	}

	additionalFields, dbErr := workloadDao.DbGetWorkloadAdditionalFields(ctx, tenantId, workloadUuid)
	if dbErr != nil {
		return dbErr
	}

	if vllmWorkloadStr, ok := additionalFields[clusterCommonConstants.AdditionalVLLMWorkload].(string); !ok || vllmWorkloadStr != "true" {
		return api.NewError(http.StatusBadRequest, "workload is not identified as vLLM", nil)
	}

	return nil
}
