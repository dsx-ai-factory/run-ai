package workloads

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	workloads "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const podsPageLimit = 20

// The 10 metrics we ask for per pod (upstream caps each request at 10): 4 resource basics + 6 GPU profiling counters.
var nonInferencePodMetricTypes = []workloads.PodMetricType{
	workloads.PODMETRICTYPE_GPU_UTILIZATION,
	workloads.PODMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
	workloads.PODMETRICTYPE_CPU_USAGE_CORES,
	workloads.PODMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
	workloads.PODMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU,
	workloads.PODMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU,
	workloads.PODMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU,
	workloads.PODMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU,
	workloads.PODMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU,
	workloads.PODMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU,
}

var inferencePodMetricTypes = []workloads.InferencePodMetricType{
	workloads.INFERENCEPODMETRICTYPE_THROUGHPUT,
	workloads.INFERENCEPODMETRICTYPE_LATENCY,
}

type getWorkloadPodsInput struct {
	WorkloadID string `json:"workloadId" jsonschema:"Workload ID (UUID)"`
}

func (in getWorkloadPodsInput) Validate() error {
	return common.ValidateUUID("workloadId", in.WorkloadID)
}

type getWorkloadPodsOutput struct {
	WorkloadID string      `json:"workloadId" jsonschema:"Echo of the input workloadId"`
	Type       string      `json:"type"       jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types)"`
	Pods       []podOutput `json:"pods"       jsonschema:"Up to 20 pods belonging to the workload (offset=0, limit=20)"`
}

func registerGetWorkloadPods(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadPodsInput, getWorkloadPodsOutput]{
		Name: "get_workload_pods",
		Description: "Call when the user asks for the list of pods belonging to a workload (job) — what node each pod " +
			"is on, how much CPU/GPU each pod requested or was allocated, or the status of individual pods. Returns up " +
			"to 20 pods. For each pod: name, node, node pool, current status, and its requested and allocated resources, " +
			"plus optional GPU/CPU/memory metrics. " +
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
//  2. list the first page of its pods (offset=0, limit=20);
//  3. map each pod and enrich it with its latest per-pod measurements.
func getWorkloadPodsRun(ctx context.Context, client *runaiclient.Client, in getWorkloadPodsInput) (getWorkloadPodsOutput, error) {
	// Step 1 — identify the workload. Type is echoed in the output and picks the per-pod metrics endpoint.
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, in.WorkloadID).Execute()
	if err != nil {
		return getWorkloadPodsOutput{}, fmt.Errorf("get workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}

	// Step 2 — list the first page of pods. A pods-list failure fails the whole tool.
	pods, httpResp, err := client.Workloads.PodsAPI.GetWorkloadPods(ctx, in.WorkloadID).
		Limit(podsPageLimit).
		Offset(0).
		Execute()
	if err != nil {
		return getWorkloadPodsOutput{}, fmt.Errorf("list pods for workload %q: %w", in.WorkloadID, runaimcp.MapHTTPError(httpResp, err))
	}

	// Step 3 — map each pod and enrich it with its latest per-pod measurements.
	return getWorkloadPodsOutput{
		WorkloadID: in.WorkloadID,
		Type:       wl.Type,
		Pods:       buildPodOutputs(ctx, client, in.WorkloadID, wl.Type, pods.Pods),
	}, nil
}

// buildPodOutputs maps upstream pods to output entries and attaches each pod's
// latest measurements snapshot. Per-pod measurement fetches run concurrently so
// total latency is bound by the slowest single call, not by the pod count.
func buildPodOutputs(ctx context.Context, client *runaiclient.Client, workloadID, workloadType string, pods []workloads.Pod) []podOutput {
	out := make([]podOutput, len(pods))
	for i, p := range pods {
		out[i] = mapPodOutput(p)
	}

	var wg sync.WaitGroup
	wg.Add(len(out))
	for i := range out {
		go func() {
			defer wg.Done()
			attachPodMeasurements(ctx, client, workloadID, out[i].ID, isInferenceType(workloadType), &out[i])
		}()
	}
	wg.Wait()
	return out
}

// Fetches the pod's latest measurements.
func attachPodMeasurements(ctx context.Context, client *runaiclient.Client, workloadID, podID string, isInference bool, entry *podOutput) {
	measurements, err := fetchPodMeasurements(ctx, client, workloadID, podID, isInference)
	if err != nil {
		zerolog.Ctx(ctx).Warn().Err(err).Str("workload_id", workloadID).Str("pod_id", podID).Msg("get_workload_pods: per-pod metrics fetch failed, returning measurements=null")
		return
	}
	mapped := mapMeasurements(measurements)
	entry.Measurements = &mapped
}

// Latest snapshot (10-min window, numberOfSamples=2) from the per-pod endpoint matching the workload type.
func fetchPodMeasurements(ctx context.Context, client *runaiclient.Client, workloadID, podID string, isInference bool) ([]workloads.MeasurementResponse, error) {
	end := time.Now().UTC()
	start := end.Add(-10 * time.Minute)
	if isInference {
		resp, _, err := client.Workloads.InferencesAPI.GetInferenceWorkloadPodMetrics(ctx, workloadID, podID).
			MetricType(inferencePodMetricTypes).
			Start(start).
			End(end).
			NumberOfSamples(2).
			Execute()
		if err != nil {
			return nil, err
		}
		return resp.Measurements, nil
	}

	resp, _, err := client.Workloads.PodsAPI.GetWorkloadPodMetrics(ctx, workloadID, podID).
		MetricType(nonInferencePodMetricTypes).
		Start(start).
		End(end).
		NumberOfSamples(2).
		Execute()
	if err != nil {
		return nil, err
	}
	return resp.Measurements, nil
}
