/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	podsDefaultLimit         = int32(20)
	podsMaxLimit             = int32(500)
	podsMetricsMaxLimit      = int32(20)
	podMetricsMaxConcurrency = 8
)

var basicPodMetricTypes = []workloads.PodMetricType{
	workloads.PODMETRICTYPE_GPU_UTILIZATION,
	workloads.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
	workloads.PODMETRICTYPE_CPU_USAGE_CORES,
	workloads.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
}

// Advanced GPU metrics require an optional tenant setting and some require a
// newer cluster. Isolate them so they cannot block the baseline metrics.
var advancedPodMetricTypes = []workloads.PodMetricType{
	workloads.PODMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU,
	workloads.PODMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU,
	workloads.PODMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU,
	workloads.PODMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU,
	workloads.PODMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU,
	workloads.PODMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU,
}

var basicAndAdvancedPodMetricTypes = append(
	append([]workloads.PodMetricType{}, basicPodMetricTypes...),
	advancedPodMetricTypes...,
)

var inferencePodMetricTypes = []workloads.InferencePodMetricType{
	workloads.INFERENCEPODMETRICTYPE_THROUGHPUT,
	workloads.INFERENCEPODMETRICTYPE_LATENCY,
}

type getWorkloadPodsInput struct {
	WorkloadID             string `json:"workloadId"                       jsonschema:"Workload ID (UUID)"`
	Limit                  int32  `json:"limit,omitempty"                  jsonschema:"Maximum number of pods to return. Defaults to 20, max 500."`
	Offset                 int32  `json:"offset,omitempty"                 jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
	IncludeMetrics         bool   `json:"includeMetrics,omitempty"         jsonschema:"When true, fetch the latest baseline CPU, GPU, memory, and inference measurements for every returned pod. Defaults to false."`
	IncludeAdvancedMetrics bool   `json:"includeAdvancedMetrics,omitempty" jsonschema:"When true, also request advanced per-GPU profiling metrics for non-inference pods. Requires includeMetrics=true and may return a warning when the cluster or tenant does not support these metrics. Defaults to false."`
}

func (in getWorkloadPodsInput) Validate() error {
	if err := common.ValidateUUID("workloadId", in.WorkloadID); err != nil {
		return err
	}
	if in.Limit < 0 || in.Limit > podsMaxLimit {
		return fmt.Errorf("limit must be between 0 and %d", podsMaxLimit)
	}
	if in.IncludeMetrics && in.Limit > podsMetricsMaxLimit {
		return fmt.Errorf("limit must be between 0 and %d when includeMetrics is true", podsMetricsMaxLimit)
	}
	if in.IncludeAdvancedMetrics && !in.IncludeMetrics {
		return fmt.Errorf("includeAdvancedMetrics requires includeMetrics=true")
	}
	if in.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}

type getWorkloadPodsOutput struct {
	WorkloadID string      `json:"workloadId"           jsonschema:"Echo of the input workloadId"`
	Type       string      `json:"type"                 jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types)"`
	Pods       []podOutput `json:"pods"                 jsonschema:"Requested page of pods belonging to the workload. Measurements are included only when includeMetrics is true."`
	HasMore    bool        `json:"hasMore"              jsonschema:"True when more pods exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset int32       `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch more pods. Only set when hasMore is true."`
}

func registerGetWorkloadPods(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadPodsInput, getWorkloadPodsOutput]{
		Name: "get_workload_pods",
		Description: "Call when the user asks for the list of pods belonging to a workload (job) — what node each pod " +
			"is on, how much CPU/GPU each pod requested or was allocated, or the status of individual pods. Returns a " +
			"paginated list of pods (default 20, max 500 without metrics and max 20 with metrics); pass nextOffset back as offset while hasMore is true. For each " +
			"pod: name, node, node pool, current status, and requested and allocated resources. Set includeMetrics=true " +
			"only when the user also needs the latest baseline per-pod GPU/CPU/memory or inference measurements. " +
			"For non-inference pods, set includeAdvancedMetrics=true only when advanced per-GPU profiling is explicitly needed; " +
			"it requires includeMetrics=true and may be unavailable on some clusters or tenants. " +
			"If workloadId is unknown, use list_workloads to resolve it. " +
			"Do NOT call for workload-level status — use get_workload_status. " +
			"Do NOT call to chart utilization over time — use get_workload_metrics. " +
			"Do NOT call to see which pods are on a given node — use get_node_pods.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getWorkloadPodsInput) (getWorkloadPodsOutput, error) {
			return getWorkloadPodsRun(ctx, client, in)
		},
	})
}

// getWorkloadPodsRun assembles a workload's pod list:
//  1. identify the workload so we know its type;
//  2. list the requested page of its pods;
//  3. map each pod and, when requested, enrich it with its latest per-pod measurements.
func getWorkloadPodsRun(ctx context.Context, client *runaiclient.Client, in getWorkloadPodsInput) (getWorkloadPodsOutput, error) {
	// Step 1 — identify the workload. Type is echoed in the output and picks the per-pod metrics endpoint.
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, in.WorkloadID).Execute()
	if err != nil {
		return getWorkloadPodsOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	if wl == nil {
		return getWorkloadPodsOutput{}, fmt.Errorf("get workload %q returned an empty response: %w", in.WorkloadID, runaimcp.ErrUpstream)
	}

	// Step 2 — list the requested page of pods. A pods-list failure fails the whole tool.
	limit := in.Limit
	if limit == 0 {
		limit = podsDefaultLimit
	}
	queryLimit := limit
	if limit < podsMaxLimit {
		queryLimit++
	}
	pods, httpResp, err := client.Workloads.PodsAPI.GetWorkloadPods(ctx, in.WorkloadID).
		Limit(queryLimit).
		Offset(in.Offset).
		Execute()
	if err != nil {
		return getWorkloadPodsOutput{}, fmt.Errorf("list pods for workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}

	if pods == nil {
		return getWorkloadPodsOutput{}, fmt.Errorf("nil pods response for workload %q: %w", in.WorkloadID, runaimcp.ErrUpstream)
	}

	if pods.Next != nil && *pods.Next != 0 && *pods.Next <= in.Offset {
		return getWorkloadPodsOutput{}, fmt.Errorf(
			"list pods for workload %q returned a non-advancing next offset %d: %w",
			in.WorkloadID, *pods.Next, runaimcp.ErrUpstream)
	}

	pagePods := pods.Pods
	hasMore := len(pagePods) > int(limit)
	if hasMore {
		pagePods = pagePods[:limit]
	}

	// limit=500 cannot use an in-page lookahead because 500 is the upstream
	// maximum. Probe one record only when the page is full.
	if limit == podsMaxLimit && !hasMore && len(pagePods) == int(limit) {
		lookaheadOffset := int64(in.Offset) + int64(limit)
		if lookaheadOffset > int64(^uint32(0)>>1) {
			return getWorkloadPodsOutput{}, fmt.Errorf("next pod offset exceeds the supported range: %w", runaimcp.ErrValidation)
		}
		lookahead, lookaheadHTTPResp, lookaheadErr := client.Workloads.PodsAPI.GetWorkloadPods(ctx, in.WorkloadID).
			Limit(1).
			Offset(int32(lookaheadOffset)).
			Execute()
		if lookaheadErr != nil {
			return getWorkloadPodsOutput{}, fmt.Errorf("check for more pods for workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(lookaheadHTTPResp, lookaheadErr))
		}
		if lookahead == nil {
			return getWorkloadPodsOutput{}, fmt.Errorf("nil pods lookahead response for workload %q: %w", in.WorkloadID, runaimcp.ErrUpstream)
		}
		hasMore = len(lookahead.Pods) > 0
	}

	var nextOffset int32
	if hasMore {
		nextOffset64 := int64(in.Offset) + int64(limit)
		if nextOffset64 > int64(^uint32(0)>>1) {
			return getWorkloadPodsOutput{}, fmt.Errorf("next pod offset exceeds the supported range: %w", runaimcp.ErrValidation)
		}
		nextOffset = int32(nextOffset64)
	}

	// Step 3 — map each pod and optionally enrich it with its latest per-pod measurements.
	return getWorkloadPodsOutput{
		WorkloadID: in.WorkloadID,
		Type:       wl.Type,
		Pods:       buildPodOutputs(ctx, client, in.WorkloadID, wl.Type, pagePods, in.IncludeMetrics, in.IncludeAdvancedMetrics),
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}, nil
}

// buildPodOutputs maps upstream pods to output entries and optionally attaches
// each pod's latest measurements snapshot using bounded concurrency.
func buildPodOutputs(ctx context.Context, client *runaiclient.Client, workloadID, workloadType string, pods []workloads.Pod, includeMetrics, includeAdvancedMetrics bool) []podOutput {
	out := make([]podOutput, len(pods))
	for i, p := range pods {
		out[i] = mapPodOutput(p)
	}
	if !includeMetrics {
		return out
	}
	end := time.Now().UTC()
	start := end.Add(-10 * time.Minute)

	semaphore := make(chan struct{}, podMetricsMaxConcurrency)
	var wg sync.WaitGroup
	wg.Add(len(out))
	for i := range out {
		go func() {
			defer wg.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}
			attachPodMeasurements(ctx, client, workloadID, out[i].ID, isInferenceType(workloadType), includeAdvancedMetrics, start, end, &out[i])
		}()
	}
	wg.Wait()
	return out
}

// Fetches the pod's latest measurements.
func attachPodMeasurements(ctx context.Context, client *runaiclient.Client, workloadID, podID string, isInference, includeAdvancedMetrics bool, start, end time.Time, entry *podOutput) {
	batches := fetchPodMetricBatches(ctx, client, workloadID, podID, isInference, includeAdvancedMetrics, start, end)
	mapped := make([]measurement, 0)
	successfulBatches := 0
	for _, batch := range batches {
		if batch.err != nil {
			warning := fmt.Sprintf("%s metrics could not be retrieved", batch.name)
			if len(batch.unavailableMetricTypes) > 0 {
				warning += "; unavailable metric types: " + strings.Join(batch.unavailableMetricTypes, ", ")
			}
			entry.MetricWarnings = append(entry.MetricWarnings, warning)
			zerolog.Ctx(ctx).Warn().Err(batch.err).Str("workload_id", workloadID).Str("pod_id", podID).
				Str("metric_group", batch.name).Msg("get_workload_pods: per-pod metric group unavailable")
			continue
		}
		successfulBatches++
		mapped = append(mapped, batch.measurements...)
		if batch.partial {
			warning := fmt.Sprintf("%s metrics were partially retrieved", batch.name)
			if len(batch.unavailableMetricTypes) > 0 {
				warning += "; unavailable metric types: " + strings.Join(batch.unavailableMetricTypes, ", ")
			}
			entry.MetricWarnings = append(entry.MetricWarnings, warning)
		}
	}
	if successfulBatches == 0 {
		zerolog.Ctx(ctx).Warn().Str("workload_id", workloadID).Str("pod_id", podID).
			Msg("get_workload_pods: per-pod metrics fetch failed, returning measurements=null")
		return
	}
	entry.Measurements = &mapped
}

type podMetricBatch struct {
	name string
	metricFetchResult
}

// Latest snapshot (10-min window, numberOfSamples=2) from the per-pod endpoint matching the workload type.
func fetchPodMetricBatches(ctx context.Context, client *runaiclient.Client, workloadID, podID string, isInference, includeAdvancedMetrics bool, start, end time.Time) []podMetricBatch {
	if isInference {
		resp, httpResp, err := client.Workloads.InferencesAPI.GetInferenceWorkloadPodMetrics(ctx, workloadID, podID).
			MetricType(inferencePodMetricTypes).
			Start(start).
			End(end).
			NumberOfSamples(2).
			Execute()
		if err != nil {
			return []podMetricBatch{{
				name: "inference serving",
				metricFetchResult: metricFetchResult{err: fmt.Errorf(
					"get inference pod metrics for workload %q pod %q: %w", workloadID, podID, runaimcp.MapHTTPError(httpResp, err))},
			}}
		}
		return []podMetricBatch{{name: "inference serving", metricFetchResult: mapMeasurementFetchResult(resp, httpResp)}}
	}

	if !includeAdvancedMetrics {
		return []podMetricBatch{{
			name:              "resource",
			metricFetchResult: fetchPodMetricsBatch(ctx, client, workloadID, podID, "resource", basicPodMetricTypes, start, end),
		}}
	}

	combined := fetchPodMetricsBatch(ctx, client, workloadID, podID, "resource and advanced GPU", basicAndAdvancedPodMetricTypes, start, end)
	if !combined.retryWithoutOptionalMetrics {
		// A 207 response is intentionally not retried: it already contains every
		// successful baseline series and names unavailable advanced metric types.
		return []podMetricBatch{{name: "resource and advanced GPU", metricFetchResult: combined}}
	}

	// Optional advanced metrics can be rejected when the feature is disabled or
	// the cluster is too old. Retry baseline metrics alone only for that 400/422
	// response and preserve the original failure as the advanced-metrics warning.
	return []podMetricBatch{
		{
			name:              "resource",
			metricFetchResult: fetchPodMetricsBatch(ctx, client, workloadID, podID, "resource", basicPodMetricTypes, start, end),
		},
		{name: "advanced GPU", metricFetchResult: combined},
	}
}

func fetchPodMetricsBatch(ctx context.Context, client *runaiclient.Client, workloadID, podID, name string, types []workloads.PodMetricType, start, end time.Time) metricFetchResult {
	resp, httpResp, err := client.Workloads.PodsAPI.GetWorkloadPodMetrics(ctx, workloadID, podID).
		MetricType(types).
		Start(start).
		End(end).
		NumberOfSamples(2).
		Execute()
	if err != nil {
		return metricFetchResult{
			err:                         fmt.Errorf("get %s pod metrics for workload %q pod %q: %w", name, workloadID, podID, runaimcp.MapHTTPError(httpResp, err)),
			retryWithoutOptionalMetrics: isOptionalMetricsValidationFailure(httpResp),
		}
	}
	if resp == nil {
		return metricFetchResult{err: fmt.Errorf("get %s pod metrics for workload %q pod %q returned an empty response: %w", name, workloadID, podID, runaimcp.ErrUpstream)}
	}
	return mapMetricFetchResult(resp, httpResp)
}
