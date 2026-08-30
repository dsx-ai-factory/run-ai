package workloads

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"

	workloads "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const (
	workloadMetricsDefaultSamples = int32(20)
	workloadMetricsMaxSamples     = int32(1000)
	workloadMetricsDefaultWindow  = time.Hour
)

// Resource metrics every workload supports
var workloadResourceMetricTypes = []workloads.WorkloadMetricType{
	workloads.WORKLOADMETRICTYPE_GPU_UTILIZATION,
	workloads.WORKLOADMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
	workloads.WORKLOADMETRICTYPE_CPU_USAGE_CORES,
	workloads.WORKLOADMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
	workloads.WORKLOADMETRICTYPE_NVLINK_BANDWIDTH_TOTAL,
}

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
	WorkloadID   string        `json:"workloadId"   jsonschema:"Echo of the input workloadId"`
	Type         string        `json:"type"         jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types)"`
	Measurements []measurement `json:"measurements" jsonschema:"Time-series measurements for the window. For inference workloads, includes throughput/latency measurements appended after GPU/CPU measurements"`
}

func registerGetWorkloadMetrics(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadMetricsInput, getWorkloadMetricsOutput]{
		Name: "get_workload_metrics",
		Description: "Call when the user asks how a workload's (job's) resource usage changed over time — such as GPU " +
			"utilization, GPU memory, CPU usage, or memory consumption — or, for inference workloads, request throughput " +
			"and latency. Returns these metrics as time series: each metric is a list of timestamped values over the " +
			"requested window. Accepts an optional time range (start/end, default the last hour) and number of samples. " +
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
//  3. fetch the core resource measurements (GPU / CPU / NVLink / NIM);
//  4. for inference workloads, also fetch serving (throughput / latency) and merge the two.
func getWorkloadMetricsRun(ctx context.Context, client *runaiclient.Client, in getWorkloadMetricsInput) (getWorkloadMetricsOutput, error) {
	// Step 1 — identify the workload. Type decides whether we also call the serving endpoint.
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, in.WorkloadID).Execute()
	if err != nil {
		return getWorkloadMetricsOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}

	// Step 2 — resolve the metrics window, defaulting to the last 1h with 20 samples.
	start, end, samples := resolveMetricsWindow(in)

	// Step 3 — fetch core measurements. For non-inference workloads this is the whole answer.
	coreMeas, coreErr := fetchWorkloadMetrics(ctx, client, in.WorkloadID, start, end, samples)
	out := getWorkloadMetricsOutput{WorkloadID: in.WorkloadID, Type: wl.Type}
	if !isInferenceType(wl.Type) {
		if coreErr != nil {
			return getWorkloadMetricsOutput{}, coreErr
		}
		out.Measurements = coreMeas
		return out, nil
	}

	// Step 4 — for inference, serving (throughput/latency) lives on a separate endpoint; fetch + merge.
	servingMeas, servingErr := fetchInferenceWorkloadMetrics(ctx, client, in.WorkloadID, start, end, samples)
	out.Measurements, err = combineInferenceMetrics(ctx, in.WorkloadID, coreMeas, coreErr, servingMeas, servingErr)
	if err != nil {
		return getWorkloadMetricsOutput{}, err
	}
	return out, nil
}

// fetchWorkloadMetrics returns resource (GPU/CPU/NVLink) measurements, plus NIM serving
// counters when the workload is NIM-tagged.
func fetchWorkloadMetrics(ctx context.Context, client *runaiclient.Client, workloadID string, start, end time.Time, samples int32) ([]measurement, error) {
	resource, err := fetchMetricsBatch(ctx, client, workloadID, workloadResourceMetricTypes, start, end, samples)
	if err != nil {
		return nil, err
	}
	return append(resource, fetchNIMMetrics(ctx, client, workloadID, start, end, samples)...), nil
}

// Fetches workload-level measurements for one set of metric types — used by both the resource and NIM batches.
func fetchMetricsBatch(ctx context.Context, client *runaiclient.Client, workloadID string, types []workloads.WorkloadMetricType, start, end time.Time, samples int32) ([]measurement, error) {
	resp, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkloadMetrics(ctx, workloadID).
		MetricType(types).
		Start(start).
		End(end).
		NumberOfSamples(samples).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("get workload metrics %q: %w", workloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	return mapMeasurements(resp.Measurements), nil
}

func fetchNIMMetrics(ctx context.Context, client *runaiclient.Client, workloadID string, start, end time.Time, samples int32) []measurement {
	nim, err := fetchMetricsBatch(ctx, client, workloadID, workloadNIMMetricTypes, start, end, samples)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Str("workload_id", workloadID).
			Msg("get_workload_metrics: NIM metrics unavailable")
		return nil
	}
	return nim
}

func fetchInferenceWorkloadMetrics(ctx context.Context, client *runaiclient.Client, workloadID string, start, end time.Time, samples int32) ([]measurement, error) {
	resp, httpResp, err := client.Workloads.InferencesAPI.GetInferenceWorkloadMetrics(ctx, workloadID).
		MetricType(inferenceWorkloadMetricTypes).
		Start(start).
		End(end).
		NumberOfSamples(samples).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("get inference workload metrics %q: %w", workloadID, runaimcp.MapHTTPError(httpResp, err))
	}
	return mapMeasurements(resp.Measurements), nil
}

// Returns resource + serving for inference workloads; drops a failing side and surfaces the core error if both fail.
func combineInferenceMetrics(ctx context.Context, workloadID string, coreMeas []measurement, coreErr error, servingMeas []measurement, servingErr error) ([]measurement, error) {
	if coreErr != nil && servingErr != nil {
		return nil, coreErr
	}
	for _, side := range []struct {
		name string
		err  error
		meas *[]measurement
	}{{"core", coreErr, &coreMeas}, {"serving", servingErr, &servingMeas}} {
		if side.err != nil {
			zerolog.Ctx(ctx).Warn().Err(side.err).Str("workload_id", workloadID).Str("side", side.name).Msgf("get_workload_metrics: %s metrics unavailable, dropping from result", side.name)
			*side.meas = nil
		}
	}

	out := append(coreMeas, servingMeas...)
	if out == nil {
		out = []measurement{}
	}
	return out, nil
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
