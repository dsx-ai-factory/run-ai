/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	workloadMetricsDefaultSamples = int32(20)
	workloadMetricsMaxSamples     = int32(1000)
	workloadMetricsDefaultWindow  = time.Hour
)

// Resource metrics every supported workload API version provides.
var workloadResourceMetricTypes = []workloads.WorkloadMetricType{
	workloads.WORKLOADMETRICTYPE_GPU_UTILIZATION,
	workloads.WORKLOADMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
	workloads.WORKLOADMETRICTYPE_CPU_USAGE_CORES,
	workloads.WORKLOADMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
}

// NVLink was added after the minimum API version supported by this tool. It is
// requested with baseline metrics first, with a baseline-only retry when the
// API reports that the combined metric selection is invalid or unsupported.
var workloadNVLinkMetricTypes = []workloads.WorkloadMetricType{
	workloads.WORKLOADMETRICTYPE_NVLINK_BANDWIDTH_TOTAL,
}

var workloadResourceAndNVLinkMetricTypes = append(
	append([]workloads.WorkloadMetricType{}, workloadResourceMetricTypes...),
	workloadNVLinkMetricTypes...,
)

var workloadNIMMetricTypes = []workloads.WorkloadMetricType{
	workloads.WORKLOADMETRICTYPE_NIM_NUM_REQUESTS_RUNNING,
	workloads.WORKLOADMETRICTYPE_NIM_REQUEST_SUCCESS_TOTAL,
	workloads.WORKLOADMETRICTYPE_NIM_REQUEST_FAILURE_TOTAL,
	workloads.WORKLOADMETRICTYPE_NIM_GPU_CACHE_USAGE_PERC,
	workloads.WORKLOADMETRICTYPE_NIM_E2_E_REQUEST_LATENCY_SECONDS,
}

var inferenceWorkloadMetricTypes = []workloads.InferenceWorkloadMetricType{
	workloads.INFERENCEWORKLOADMETRICTYPE_THROUGHPUT,
	workloads.INFERENCEWORKLOADMETRICTYPE_LATENCY,
}

type getWorkloadMetricsInput struct {
	WorkloadID      string `json:"workloadId"                jsonschema:"Workload ID (UUID)"`
	Start           string `json:"start,omitempty"           jsonschema:"Start of the metrics window (RFC3339); defaults to now - 1h"`
	End             string `json:"end,omitempty"             jsonschema:"End of the metrics window (RFC3339); defaults to now"`
	NumberOfSamples int32  `json:"numberOfSamples,omitempty" jsonschema:"Number of samples to return across the window; default 20, max 1000"`
}

func (in getWorkloadMetricsInput) Validate() error {
	if err := common.ValidateUUID("workloadId", in.WorkloadID); err != nil {
		return err
	}
	if err := common.ValidateTimeWindow(in.Start, in.End); err != nil {
		return err
	}
	if in.NumberOfSamples < 0 || in.NumberOfSamples > workloadMetricsMaxSamples {
		return fmt.Errorf("numberOfSamples must be between 0 and %d (got %d)", workloadMetricsMaxSamples, in.NumberOfSamples)
	}
	return nil
}

type getWorkloadMetricsOutput struct {
	WorkloadID   string            `json:"workloadId"   jsonschema:"Echo of the input workloadId"`
	Type         string            `json:"type"         jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types)"`
	Measurements []measurement     `json:"measurements" jsonschema:"Scalar time-series measurements for the window. Inference LATENCY is a scalar series; NIM_E2E_REQUEST_LATENCY_SECONDS is returned under histograms instead."`
	Histograms   []metricHistogram `json:"histograms,omitempty" jsonschema:"Histogram time series with timestamped bucket or percentile data. NIM_E2E_REQUEST_LATENCY_SECONDS is returned here, not under scalar measurements."`
	Warnings     []string          `json:"warnings,omitempty" jsonschema:"Non-fatal warnings that identify expected metric groups or metric types that could not be retrieved. When present, metrics are partial."`
}

func registerGetWorkloadMetrics(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadMetricsInput, getWorkloadMetricsOutput]{
		Name: "get_workload_metrics",
		Description: "Call when the user asks how a workload's (job's) resource usage changed over time — such as GPU " +
			"utilization, GPU memory, CPU usage, or memory consumption — or, for inference workloads, request throughput " +
			"and latency. Returns scalar measurements as lists of timestamped values; NIM end-to-end request latency " +
			"is returned separately under histograms with timestamped bucket or percentile data. Uses one combined " +
			"resource request on supported clusters and automatically retries baseline CPU/GPU metrics without NVLink " +
			"when the API reports that the combined metric selection is unsupported. Accepts an optional time range " +
			"(start/end, default the last hour) and number of samples. " +
			"If workloadId is unknown, use list_workloads to resolve it. " +
			"Use this tool for one workload; use get_cluster_metrics for a cluster/node pool and get_org_unit_metrics for a project/department. Use list_workloads for allocated or requested GPU-count filters. " +
			"Do not compare utilization across workloads by default. If the user insists, clarify the time window and latest/average/maximum statistic, then call this tool only when the complete candidate set has 10 or fewer workloads; otherwise ask the user to narrow the scope. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
			"When resolving a name, follow all list_workloads pages and ask for project or cluster context if multiple matches remain. " +
			"Do NOT call for per-pod metrics — use get_workload_pods.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getWorkloadMetricsInput) (getWorkloadMetricsOutput, error) {
			return getWorkloadMetricsRun(ctx, client, in)
		},
	})
}

// getWorkloadMetricsRun assembles a workload's metrics:
//  1. identify the workload so we know its type;
//  2. pin the query window (defaults: now-1h..now, 20 samples);
//  3. fetch the core resource measurements (GPU / CPU / NVLink / NIM), retrying
//     baseline resource metrics only if the API rejects the optional NVLink selection;
//  4. for inference workloads, also fetch serving (throughput / latency) and merge the two.
func getWorkloadMetricsRun(ctx context.Context, client *runaiclient.Client, in getWorkloadMetricsInput) (getWorkloadMetricsOutput, error) {
	// Step 1 — identify the workload. Type decides whether we also call the serving endpoint.
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, in.WorkloadID).Execute()
	if err != nil {
		return getWorkloadMetricsOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	if wl == nil {
		return getWorkloadMetricsOutput{}, fmt.Errorf("get workload %q returned an empty response: %w", in.WorkloadID, runaimcp.ErrUpstream)
	}

	// Step 2 — resolve the metrics window, defaulting to the last 1h with 20 samples.
	start, end, samples := resolveMetricsWindow(in)

	// Step 3 — use one request on clusters that support NVLink. A 400/422 may
	// mean NVLink is unsupported, so retry baseline resource metrics alone and
	// retain the original validation failure as an NVLink warning. Other errors
	// are not retried because authentication, transport, or server failures are
	// not evidence that only NVLink is unavailable.
	resourceAndNVLinkMetrics := fetchMetricsBatch(ctx, client, in.WorkloadID, workloadResourceAndNVLinkMetricTypes, start, end, samples)
	var batches []workloadMetricBatch
	if !resourceAndNVLinkMetrics.retryWithoutOptionalMetrics {
		batches = []workloadMetricBatch{{
			name:              "resource and NVLink",
			metricFetchResult: resourceAndNVLinkMetrics,
		}}
	} else {
		resourceMetrics := fetchMetricsBatch(ctx, client, in.WorkloadID, workloadResourceMetricTypes, start, end, samples)
		batches = []workloadMetricBatch{
			{name: "resource", metricFetchResult: resourceMetrics},
			{name: "NVLink", metricFetchResult: resourceAndNVLinkMetrics},
		}
	}

	// NIM counters are expected only when the workload is marked as NVIDIA NIM.
	if isNIMWorkload(wl) {
		nimMetrics := fetchMetricsBatch(ctx, client, in.WorkloadID, workloadNIMMetricTypes, start, end, samples)
		batches = append(batches, workloadMetricBatch{
			name:              "NIM",
			metricFetchResult: nimMetrics,
		})
	}

	// Step 4 — inference serving metrics (throughput/latency) live on a separate endpoint.
	if isInferenceType(wl.Type) {
		servingMetrics := fetchInferenceWorkloadMetrics(ctx, client, in.WorkloadID, start, end, samples)
		batches = append(batches, workloadMetricBatch{
			name:              "inference serving",
			metricFetchResult: servingMetrics,
		})
	}

	return combineMetricBatches(ctx, in.WorkloadID, wl.Type, batches)
}

type workloadMetricBatch struct {
	name string
	metricFetchResult
}

type metricFetchResult struct {
	measurements                []measurement
	histograms                  []metricHistogram
	partial                     bool
	unavailableMetricTypes      []string
	err                         error
	retryWithoutOptionalMetrics bool
}

// combineMetricBatches returns every successful group and names failed expected
// groups in warnings. If every expected group fails, the hard error names every
// group and wraps the first error to preserve its classification.
func combineMetricBatches(ctx context.Context, workloadID, workloadType string, batches []workloadMetricBatch) (getWorkloadMetricsOutput, error) {
	out := getWorkloadMetricsOutput{
		WorkloadID:   workloadID,
		Type:         workloadType,
		Measurements: []measurement{},
	}
	successfulBatches := 0
	failedGroups := make([]string, 0, len(batches))
	var firstErr error
	for _, batch := range batches {
		if batch.err == nil {
			successfulBatches++
			out.Measurements = append(out.Measurements, batch.measurements...)
			out.Histograms = append(out.Histograms, batch.histograms...)
			if batch.partial {
				warning := fmt.Sprintf("%s metrics were partially retrieved", batch.name)
				if len(batch.unavailableMetricTypes) > 0 {
					warning += "; unavailable metric types: " + strings.Join(batch.unavailableMetricTypes, ", ")
				}
				out.Warnings = append(out.Warnings, warning)
				zerolog.Ctx(ctx).Warn().Str("workload_id", workloadID).Str("metric_group", batch.name).
					Strs("unavailable_metric_types", batch.unavailableMetricTypes).
					Msgf("get_workload_metrics: %s metrics partially available", batch.name)
			}
			continue
		}
		if firstErr == nil {
			firstErr = batch.err
		}
		failedGroups = append(failedGroups, batch.name)
		zerolog.Ctx(ctx).Warn().Err(batch.err).Str("workload_id", workloadID).Str("metric_group", batch.name).
			Msgf("get_workload_metrics: %s metrics unavailable, dropping from result", batch.name)
		warning := fmt.Sprintf("%s metrics could not be retrieved; returned metrics are partial", batch.name)
		if len(batch.unavailableMetricTypes) > 0 {
			warning += "; unavailable metric types: " + strings.Join(batch.unavailableMetricTypes, ", ")
		}
		out.Warnings = append(out.Warnings, warning)
	}
	if successfulBatches == 0 {
		return getWorkloadMetricsOutput{}, fmt.Errorf(
			"all expected metric groups failed (%s): %w",
			strings.Join(failedGroups, ", "), firstErr)
	}
	return out, nil
}

func isNIMWorkload(wl *workloads.WorkloadDetailed) bool {
	if wl == nil {
		return false
	}
	value, ok := wl.AdditionalFields["nimWorkload"].(string)
	return ok && value == "true"
}

// Fetches workload-level measurements for one set of metric types — used by both the resource and NIM batches.
func fetchMetricsBatch(ctx context.Context, client *runaiclient.Client, workloadID string, types []workloads.WorkloadMetricType, start, end time.Time, samples int32) metricFetchResult {
	resp, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkloadMetrics(ctx, workloadID).
		MetricType(types).
		Start(start).
		End(end).
		NumberOfSamples(samples).
		Execute()
	if err != nil {
		return metricFetchResult{
			err:                         fmt.Errorf("get workload metrics %q: %w", workloadID, runaimcp.MapHTTPError(httpResp, err)),
			retryWithoutOptionalMetrics: isOptionalMetricsValidationFailure(httpResp),
		}
	}
	if resp == nil {
		return metricFetchResult{err: fmt.Errorf("get workload metrics %q returned an empty response: %w", workloadID, runaimcp.ErrUpstream)}
	}
	return mapMetricFetchResult(resp, httpResp)
}

func fetchInferenceWorkloadMetrics(ctx context.Context, client *runaiclient.Client, workloadID string, start, end time.Time, samples int32) metricFetchResult {
	resp, httpResp, err := client.Workloads.InferencesAPI.GetInferenceWorkloadMetrics(ctx, workloadID).
		MetricType(inferenceWorkloadMetricTypes).
		Start(start).
		End(end).
		NumberOfSamples(samples).
		Execute()
	if err != nil {
		return metricFetchResult{err: fmt.Errorf("get inference workload metrics %q: %w", workloadID, runaimcp.MapHTTPError(httpResp, err))}
	}
	if resp == nil {
		return metricFetchResult{err: fmt.Errorf("get inference workload metrics %q returned an empty response: %w", workloadID, runaimcp.ErrUpstream)}
	}
	return mapMeasurementFetchResult(resp, httpResp)
}

func isOptionalMetricsValidationFailure(resp *http.Response) bool {
	return resp != nil && (resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity)
}

// A 207 response contains one values-less measurement for each metric type that
// could not be retrieved. Preserve successful series and name each unavailable
// type so callers can distinguish partial data from a valid empty series.
func mapMetricFetchResult(resp *workloads.MetricsCompositeResponse, httpResp *http.Response) metricFetchResult {
	if resp == nil {
		return metricFetchResult{err: fmt.Errorf("empty metrics response: %w", runaimcp.ErrUpstream)}
	}
	return mapMetricResponseParts(resp.Measurements, resp.Histogram, httpResp)
}

func mapMeasurementFetchResult(resp *workloads.MetricsResponse, httpResp *http.Response) metricFetchResult {
	if resp == nil {
		return metricFetchResult{err: fmt.Errorf("empty metrics response: %w", runaimcp.ErrUpstream)}
	}
	return mapMetricResponseParts(resp.Measurements, nil, httpResp)
}

func mapMetricResponseParts(src []workloads.MeasurementResponse, histograms []workloads.HistogramSeries, httpResp *http.Response) metricFetchResult {
	if httpResp == nil || httpResp.StatusCode != http.StatusMultiStatus {
		return metricFetchResult{
			measurements: mapMeasurements(src),
			histograms:   mapMetricHistograms(histograms),
		}
	}

	available := make([]workloads.MeasurementResponse, 0, len(src))
	unavailable := make([]string, 0)
	seenUnavailable := make(map[string]struct{})
	for _, item := range src {
		if item.Values != nil {
			available = append(available, item)
			continue
		}
		if item.Type == "" {
			continue
		}
		if _, seen := seenUnavailable[item.Type]; seen {
			continue
		}
		seenUnavailable[item.Type] = struct{}{}
		unavailable = append(unavailable, item.Type)
	}

	result := metricFetchResult{
		measurements:           mapMeasurements(available),
		histograms:             mapMetricHistograms(histograms),
		partial:                len(unavailable) > 0,
		unavailableMetricTypes: unavailable,
	}
	if len(result.measurements) == 0 && len(result.histograms) == 0 && len(unavailable) > 0 {
		result.err = fmt.Errorf(
			"all returned metric types were unavailable (%s): %w",
			strings.Join(unavailable, ", "), runaimcp.ErrUpstream)
	}
	return result
}

func resolveMetricsWindow(in getWorkloadMetricsInput) (time.Time, time.Time, int32) {
	end := time.Now().UTC()
	if t, set, _ := common.ParseRFC3339IfSet("end", in.End); set {
		end = t.UTC()
	}
	start := end.Add(-workloadMetricsDefaultWindow)
	if t, set, _ := common.ParseRFC3339IfSet("start", in.Start); set {
		start = t.UTC()
	}
	return start, end, cmp.Or(in.NumberOfSamples, workloadMetricsDefaultSamples)
}
