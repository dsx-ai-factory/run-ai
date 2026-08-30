package workloads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	clusterservice "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	workloads "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const (
	nodePodsListLimit = int32(500)
	nodesListLimit    = int32(2)
	nodePodsCap       = 10
)

type getNodePodsInput struct {
	NodeName    string `json:"nodeName"              jsonschema:"Node name as it appears on the cluster"`
	ClusterUUID string `json:"clusterUuid,omitempty" jsonschema:"Cluster UUID. Optional — when omitted, the tool derives it from the pods it finds on the node. Pass it explicitly when the node is idle (no pods to derive from) or when pods on the node span multiple clusters (the tool returns an ambiguous_node validation error pointing back here)"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func (in getNodePodsInput) Validate() error {
	if strings.TrimSpace(in.NodeName) == "" {
		return fmt.Errorf("nodeName is required")
	}
	if err := common.ValidateNoCommas("nodeName", in.NodeName); err != nil {
		return err
	}
	if v := strings.TrimSpace(in.ClusterUUID); v != "" {
		return common.ValidateUUID("clusterUuid", v)
	}
	return nil
}

type nodeGpuInfo struct {
	GpuType  string  `json:"gpuType,omitempty"  jsonschema:"GPU model on the node (e.g. \"Tesla-V100\")"`
	GpuCount float64 `json:"gpuCount,omitempty" jsonschema:"Number of GPUs on the node"`
}

type nodePodOutput struct {
	podOutput
	WorkloadID   string `json:"workloadId"             jsonschema:"Workload ID (UUID) the pod belongs to"`
	WorkloadName string `json:"workloadName,omitempty" jsonschema:"Workload name. May be empty when the per-workload enrichment lookup failed and the pod listing didn't carry it"`
	WorkloadType string `json:"workloadType,omitempty" jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types). May be empty when the per-workload enrichment lookup failed"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

type getNodePodsOutput struct {
	Name     string          `json:"name"               jsonschema:"Echo of the input nodeName"`
	NodePool string          `json:"nodePool,omitempty" jsonschema:"Node pool the node belongs to. Omitted when node metadata could not be resolved"`
	Status   string          `json:"status,omitempty"   jsonschema:"Calculated node status. Omitted when node metadata could not be resolved"`
	GpuInfo  *nodeGpuInfo    `json:"gpuInfo,omitempty"  jsonschema:"GPU model and count for the node. Omitted when node metadata could not be resolved or when the node has no GPUs"`
	Pods     []nodePodOutput `json:"pods"               jsonschema:"Up to 10 pods running on this node (sorted by pod name, capped to 10). Each entry has the same shape as get_workload_pods.pods plus workload-context fields (workloadId/workloadName/workloadType). The measurements field is never populated here — use get_workload_pods for per-pod metrics"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func registerGetNodePods(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getNodePodsInput, getNodePodsOutput]{
		Name: "get_node_pods",
		Description: "Call when the user names a specific node and asks which pods are running on it, which workloads " +
			"(jobs) those pods belong to, or how the node's GPUs are allocated across those pods. Returns up to 10 " +
			"pods on that node; for each pod, its name, status, requested and allocated resources, and the workload " +
			"it belongs to (id, name, and type), plus the node's own node pool, status, and GPU type and count. " +
			"Do NOT call to see which pods a given workload has — use get_workload_pods.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getNodePodsInput) (getNodePodsOutput, error) {
			return getNodePodsRun(ctx, client, in)
		},
	})
}

// getNodePodsRun returns the pods on a node along with the node's own metadata:
//  1. fetch pods on this node — fails the tool on error;
//  2. sort and cap to the 10 we'll return so later steps work off the trimmed list;
//  3. pick the cluster — caller input wins, otherwise taken from the pods;
//  4. fetch node metadata (soft fail — drops nodePool/status/gpuInfo on error);
//  5. map each pod and fill in missing workloadName/workloadType in parallel (soft fail per workload).
func getNodePodsRun(ctx context.Context, client *runaiclient.Client, in getNodePodsInput) (getNodePodsOutput, error) {
	// Step 1 — list pods on the node (hard fail: the pod list is the core return value).
	podsResp, httpResp, err := client.Workloads.PodsAPI.ListPods(ctx).
		FilterBy([]string{common.FilterEq("nodeName", in.NodeName)}).
		Limit(nodePodsListLimit).
		Execute()
	if err != nil {
		return getNodePodsOutput{}, fmt.Errorf("list pods on node %q: %w", in.NodeName, runaimcp.MapHTTPError(httpResp, err))
	}

	// Step 2 — sort and cap so later steps work off the 10 pods we return.
	sortUpstreamPods(podsResp.Pods)
	visiblePods := podsResp.Pods[:min(len(podsResp.Pods), nodePodsCap)]

	// Step 3 — resolve clusterUuid (caller input → derived from visible pods).
	// No pods + no cluster input ⇒ we have nothing to call listNodes against
	if len(visiblePods) == 0 && in.ClusterUUID == "" {
		return getNodePodsOutput{}, fmt.Errorf("no pods found on node %q — check the node name; if the node is real but idle, retry with clusterUuid to verify it exists: %w", in.NodeName, runaimcp.ErrValidation)
	}
	clusterUUID, err := resolveNodeCluster(in.ClusterUUID, visiblePods)
	if err != nil {
		return getNodePodsOutput{}, err
	}

	out := getNodePodsOutput{Name: in.NodeName}

	// Step 4 — attach node metadata.
	// Pods prove the node exists; without them, listNodes must — hence requireExists.
	if err := attachNodeMetadata(ctx, client, clusterUUID, in.NodeName, &out, len(visiblePods) == 0); err != nil {
		return getNodePodsOutput{}, err
	}

	// Step 5 — map pods, enrich.
	out.Pods = buildNodePodOutputs(ctx, client, visiblePods)
	return out, nil
}

func sortUpstreamPods(pods []workloads.Pod) {
	sort.SliceStable(pods, func(i, j int) bool {
		if pods[i].Name != pods[j].Name {
			return pods[i].Name < pods[j].Name
		}
		return pods[i].Id < pods[j].Id
	})
}

// resolveNodeCluster picks the cluster: caller input wins, else pods[].clusterId
func resolveNodeCluster(input string, pods []workloads.Pod) (string, error) {
	if input != "" {
		for _, p := range pods {
			if p.ClusterId != "" && p.ClusterId != input {
				return "", fmt.Errorf("clusterUuid %q does not match the cluster of pods on this node (%q): %w", input, p.ClusterId, runaimcp.ErrValidation)
			}
		}
		return input, nil
	}
	var found string
	for _, p := range pods {
		if p.ClusterId == "" || p.ClusterId == found {
			continue
		}
		if found != "" {
			return "", fmt.Errorf("pods on this node belong to multiple clusters — pass clusterUuid explicitly to disambiguate: %w", runaimcp.ErrValidation)
		}
		found = p.ClusterId
	}
	return found, nil
}

func attachNodeMetadata(ctx context.Context, client *runaiclient.Client, clusterUUID, nodeName string, out *getNodePodsOutput, requireExists bool) error {
	logger := zerolog.Ctx(ctx).With().Str("cluster_uuid", clusterUUID).Str("node_name", nodeName).Logger()
	resp, httpResp, err := client.Cluster.NodesAPI.GetTenantNodes(ctx).
		FilterBy([]string{common.FilterEq("name", nodeName), common.FilterEq("clusterUuid", clusterUUID)}).
		Limit(nodesListLimit).
		Execute()
	if err != nil {
		if requireExists {
			return fmt.Errorf("verify node %q in cluster %q: %w", nodeName, clusterUUID, runaimcp.MapHTTPError(httpResp, err))
		}
		evt := logger.Warn().Err(err).Str("call", "getNodes")
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_node_pods: listNodes failed, omitting node metadata")
		return nil
	}
	for _, n := range resp.Nodes {
		if n.Name == nodeName {
			out.NodePool = n.NodePool
			out.Status = n.Status
			out.GpuInfo = mapNodeGpuInfo(n, func() string { return legacyGpuTypeForNode(httpResp, nodeName) })
			return nil
		}
	}
	if requireExists {
		return fmt.Errorf("node %q not found in cluster %q: %w", nodeName, clusterUUID, runaimcp.ErrValidation)
	}
	logger.Warn().Msg("get_node_pods: node not in listNodes response, omitting node metadata")
	return nil
}

func mapNodeGpuInfo(node clusterservice.Node, legacyGpuType func() string) *nodeGpuInfo {
	out := nodeGpuInfo{}
	if gi := node.GpuInfo.Get(); gi != nil {
		out.GpuType = gi.GpuType
		out.GpuCount = float64(gi.GpuCount)
	}
	if out.GpuType == "" {
		out.GpuType = legacyGpuType()
	}
	if out.GpuType == "" && out.GpuCount == 0 {
		return nil
	}
	return &out
}

// legacyGpuTypeForNode recovers nodes[].gpuType — dropped by the typed client since it isn't in the modern schema. Best-effort, returns "" on any failure.
func legacyGpuTypeForNode(httpResp *http.Response, nodeName string) string {
	if httpResp == nil || httpResp.Body == nil {
		return ""
	}
	var raw struct {
		Nodes []struct {
			Name    string `json:"name"`
			GpuType string `json:"gpuType"`
		} `json:"nodes"`
	}
	_ = json.NewDecoder(httpResp.Body).Decode(&raw)
	for _, n := range raw.Nodes {
		if n.Name == nodeName {
			return n.GpuType
		}
	}
	return ""
}

func buildNodePodOutputs(ctx context.Context, client *runaiclient.Client, pods []workloads.Pod) []nodePodOutput {
	out := make([]nodePodOutput, len(pods))
	for i, p := range pods {
		out[i] = nodePodOutput{podOutput: mapPodOutput(p), WorkloadID: p.WorkloadId}
		if p.WorkloadName != nil {
			out[i].WorkloadName = *p.WorkloadName
		}
	}
	enrichWorkloadContext(ctx, client, out)
	return out
}

// enrichWorkloadContext fills in pods[].workloadName/workloadType
func enrichWorkloadContext(ctx context.Context, client *runaiclient.Client, pods []nodePodOutput) {
	podsToEnrich := map[string][]int{}
	for i, p := range pods {
		if p.WorkloadID != "" && (p.WorkloadName == "" || p.WorkloadType == "") {
			podsToEnrich[p.WorkloadID] = append(podsToEnrich[p.WorkloadID], i)
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(podsToEnrich))
	for id, idxs := range podsToEnrich {
		go func() {
			defer wg.Done()
			applyWorkloadContext(ctx, client, pods, id, idxs)
		}()
	}
	wg.Wait()
}

// applyWorkloadContext fills missing workloadName/Type
func applyWorkloadContext(ctx context.Context, client *runaiclient.Client, pods []nodePodOutput, id string, idxs []int) {
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, id).Execute()
	if err != nil {
		evt := zerolog.Ctx(ctx).Warn().Err(err).Str("workload_id", id).Str("call", "getWorkload")
		if httpResp != nil {
			evt = evt.Int("status_code", httpResp.StatusCode)
		}
		evt.Msg("get_node_pods: workload enrichment failed")
		return
	}
	for _, i := range idxs {
		if pods[i].WorkloadName == "" {
			pods[i].WorkloadName = wl.Name
		}
		if pods[i].WorkloadType == "" {
			pods[i].WorkloadType = wl.Type
		}
	}
}
