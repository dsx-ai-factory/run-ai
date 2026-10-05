/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	nodePodsListLimit      = int32(500)
	nodePodsDefaultLimit   = int32(10)
	nodePodEnrichmentLimit = 8
	nodesListLimit         = int32(2)
)

type getNodePodsInput struct {
	NodeName    string `json:"nodeName"              jsonschema:"Node name as it appears on the cluster"`
	ClusterUUID string `json:"clusterUuid,omitempty" jsonschema:"Cluster UUID. Optional — when omitted, the tool derives it from the pods it finds on the node. Pass it explicitly when the node is idle (no pods to derive from) or when pods on the node span multiple clusters (the tool returns an ambiguous_node validation error pointing back here)"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Limit       int32  `json:"limit,omitempty"       jsonschema:"Maximum number of pods to return. Defaults to 10, capped at 200."`
	Offset      int32  `json:"offset,omitempty"      jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in getNodePodsInput) Validate() error {
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	if in.Offset > math.MaxInt32-effectiveNodePodsLimit(in.Limit) {
		return fmt.Errorf("offset and limit exceed the supported 32-bit pagination range")
	}
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
	Name               string                 `json:"name"                         jsonschema:"Pod name"`
	ID                 string                 `json:"id"                           jsonschema:"Pod ID (UUID)"`
	NodeName           string                 `json:"nodeName,omitempty"           jsonschema:"Name of the node the pod runs on"`
	CurrentNodePool    string                 `json:"currentNodePool,omitempty"    jsonschema:"Node pool the pod is currently scheduled to"`
	K8sPhase           string                 `json:"k8sPhase"                     jsonschema:"Kubernetes pod phase (Pending | Running | Succeeded | Failed | Unknown)"`
	CreatedAt          string                 `json:"createdAt"                    jsonschema:"Pod creation timestamp (RFC3339)"`
	RevisionID         string                 `json:"revisionId,omitempty"         jsonschema:"Inference revision ID (UUID); set only for inference workload pods"`
	ClusterID          string                 `json:"clusterId,omitempty"          jsonschema:"Cluster ID the pod belongs to"`
	RequestedResources *podRequestedResources `json:"requestedResources,omitempty" jsonschema:"What the workload asked for: each dimension is a {request, limit} pair. Compare against allocatedResources to see whether the pod got what it asked for"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	AllocatedResources *podAllocatedResources `json:"allocatedResources,omitempty" jsonschema:"What the scheduler granted: each dimension is a single number / quantity (no request/limit pair)"`
	WorkloadID         string                 `json:"workloadId"                   jsonschema:"Workload ID (UUID) the pod belongs to"`
	WorkloadName       string                 `json:"workloadName,omitempty"       jsonschema:"Workload name. May be empty when the per-workload enrichment lookup failed and the pod listing didn't carry it"`
	WorkloadType       string                 `json:"workloadType,omitempty"       jsonschema:"Workload type (Training | Workspace | Distributed | Inference | DistributedInference | non-native types). May be empty when the per-workload enrichment lookup failed"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

type getNodePodsOutput struct {
	Name       string          `json:"name"                 jsonschema:"Echo of the input nodeName"`
	NodePool   string          `json:"nodePool,omitempty"   jsonschema:"Node pool the node belongs to. Omitted when node metadata could not be resolved"`
	Status     string          `json:"status,omitempty"     jsonschema:"Calculated node status. Omitted when node metadata could not be resolved"`
	GpuInfo    *nodeGpuInfo    `json:"gpuInfo,omitempty"    jsonschema:"GPU model and count for the node. Omitted when node metadata could not be resolved or when the node has no GPUs"`
	Pods       []nodePodOutput `json:"pods"                 jsonschema:"Requested page of pods running on this node, sorted by pod name. Each entry includes pod identity, status, resources, and workload context; use get_workload_pods for per-pod metrics"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	HasMore    bool            `json:"hasMore"              jsonschema:"True when more pods exist beyond the returned page; pass nextOffset back as offset"`
	NextOffset int32           `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch more pods. Only set when hasMore is true"`
}

func registerGetNodePods(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getNodePodsInput, getNodePodsOutput]{
		Name: "get_node_pods",
		Description: "Call when the user names a specific node and asks which pods are running on it, which workloads " +
			"(jobs) those pods belong to, or how the node's GPUs are allocated across those pods. Returns a paginated " +
			"list of pods on that node; for each pod, its name, status, requested and allocated resources, and the workload " +
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
//  1. fetch the requested page when clusterUuid is known, otherwise scan every pod to resolve it;
//  2. pick the cluster from the complete fetched set — caller input wins, otherwise taken from the pods;
//  3. sort and select the requested output page when a full scan was required;
//  4. fetch node metadata (soft fail — drops nodePool/status/gpuInfo on error);
//  5. map each pod and fill in missing workloadName/workloadType in parallel (soft fail per workload).
func getNodePodsRun(ctx context.Context, client *runaiclient.Client, in getNodePodsInput) (getNodePodsOutput, error) {
	if in.ClusterUUID != "" {
		return getNodePodsForCluster(ctx, client, in)
	}

	// Step 1 — list every pod on the node (hard fail: both the list and cluster
	// resolution must be based on a complete result set).
	allPods, err := fetchAllNodePods(ctx, client, in.NodeName, "")
	if err != nil {
		return getNodePodsOutput{}, err
	}

	// Step 2 — resolve clusterUuid from all matched pods, including pods beyond
	// the requested output page.
	// No pods + no cluster input ⇒ we have nothing to call listNodes against
	if len(allPods) == 0 && in.ClusterUUID == "" {
		return getNodePodsOutput{}, fmt.Errorf("no pods found on node %q — check the node name; if the node is real but idle, retry with clusterUuid to verify it exists: %w", in.NodeName, runaimcp.ErrValidation)
	}
	clusterUUID, err := resolveNodeCluster("", allPods)
	if err != nil {
		return getNodePodsOutput{}, err
	}

	// Step 3 — sort the complete set before local pagination so page boundaries
	// are deterministic.
	sortUpstreamPods(allPods)
	visiblePods, hasMore, nextOffset := common.Paginate(allPods, in.Offset, effectiveNodePodsLimit(in.Limit))
	out := getNodePodsOutput{
		Name:       in.NodeName,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}

	// Step 4 — attach node metadata.
	// Pods prove the node exists; without them, listNodes must — hence requireExists.
	if err := attachNodeMetadata(ctx, client, clusterUUID, in.NodeName, &out, len(allPods) == 0); err != nil {
		return getNodePodsOutput{}, err
	}

	// Step 5 — map pods, enrich.
	out.Pods = buildNodePodOutputs(ctx, client, visiblePods)
	return out, nil
}

// getNodePodsForCluster can use the API's pagination directly because the
// caller already supplied the cluster needed for node metadata and there is no
// cross-cluster ambiguity to resolve.
func getNodePodsForCluster(ctx context.Context, client *runaiclient.Client, in getNodePodsInput) (getNodePodsOutput, error) {
	limit := effectiveNodePodsLimit(in.Limit)
	pods, hasMore, nextOffset, err := fetchNodePodsPage(ctx, client, in.NodeName, in.ClusterUUID, in.Offset, limit)
	if err != nil {
		return getNodePodsOutput{}, err
	}
	if _, err := resolveNodeCluster(in.ClusterUUID, pods); err != nil {
		return getNodePodsOutput{}, err
	}

	out := getNodePodsOutput{
		Name:       in.NodeName,
		HasMore:    hasMore,
		NextOffset: nextOffset,
	}
	if err := attachNodeMetadata(ctx, client, in.ClusterUUID, in.NodeName, &out, len(pods) == 0 && in.Offset == 0); err != nil {
		return getNodePodsOutput{}, err
	}
	out.Pods = buildNodePodOutputs(ctx, client, pods)
	return out, nil
}

func effectiveNodePodsLimit(limit int32) int32 {
	if limit == 0 {
		return nodePodsDefaultLimit
	}
	return limit
}

// fetchNodePodsPage requests one extra pod so hasMore means that another pod
// really exists, rather than merely echoing the API's full-page cursor hint.
func fetchNodePodsPage(ctx context.Context, client *runaiclient.Client, nodeName, clusterUUID string, offset, limit int32) ([]workloads.Pod, bool, int32, error) {
	if limit <= 0 || limit > common.MaxLimit || offset < 0 || offset > math.MaxInt32-limit {
		return nil, false, 0, fmt.Errorf("invalid pod pagination offset %d and limit %d: %w", offset, limit, runaimcp.ErrValidation)
	}
	filters := []string{common.FilterEq("nodeName", nodeName), common.FilterEq("clusterId", clusterUUID)}
	podsResp, httpResp, err := client.Workloads.PodsAPI.ListPods(ctx).
		FilterBy(filters).
		Completed("false").
		SortBy("name").
		SortOrder("asc").
		Limit(limit + 1).
		Offset(offset).
		Execute()
	if err != nil {
		return nil, false, 0, fmt.Errorf("list pods on node %q: %w", nodeName, runaimcp.MapHTTPError(httpResp, err))
	}
	if podsResp == nil {
		return nil, false, 0, fmt.Errorf("list pods on node %q returned a nil response: %w", nodeName, runaimcp.ErrUpstream)
	}

	hasMore := int32(len(podsResp.Pods)) > limit
	if hasMore {
		podsResp.Pods = podsResp.Pods[:limit]
		return podsResp.Pods, true, offset + limit, nil
	}
	return podsResp.Pods, false, 0, nil
}

// fetchAllNodePods follows the Workloads API's offset cursor in 500-entry
// pages. An explicit cluster UUID scopes the pod query to that cluster; without
// one, all clusters are scanned so ambiguity can be detected. The safety bound
// is deliberate: if the complete set cannot be read, the tool returns an error
// instead of deriving a cluster from partial data.
func fetchAllNodePods(ctx context.Context, client *runaiclient.Client, nodeName, clusterUUID string) ([]workloads.Pod, error) {
	filters := []string{common.FilterEq("nodeName", nodeName)}
	if clusterUUID != "" {
		filters = append(filters, common.FilterEq("clusterId", clusterUUID))
	}
	var all []workloads.Pod
	for offset := int32(0); ; {
		podsResp, httpResp, err := client.Workloads.PodsAPI.ListPods(ctx).
			FilterBy(filters).
			Completed("false").
			SortBy("name").
			SortOrder("asc").
			Limit(nodePodsListLimit).
			Offset(offset).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("list pods on node %q: %w", nodeName, runaimcp.MapHTTPError(httpResp, err))
		}
		if podsResp == nil {
			return nil, fmt.Errorf("list pods on node %q returned a nil response: %w", nodeName, runaimcp.ErrUpstream)
		}
		all = append(all, podsResp.Pods...)
		hasNext := podsResp.Next != nil && *podsResp.Next != 0
		if int32(len(all)) > common.MaxPaginatedResults {
			return nil, fmt.Errorf(
				"node %q has more than the maximum supported %d pods; refusing to make an incomplete cluster decision: %w",
				nodeName,
				common.MaxPaginatedResults,
				runaimcp.ErrUpstream,
			)
		}
		if !hasNext {
			return all, nil
		}
		if *podsResp.Next <= offset {
			return nil, fmt.Errorf("list pods on node %q returned a non-advancing next offset %d: %w", nodeName, *podsResp.Next, runaimcp.ErrUpstream)
		}
		if len(podsResp.Pods) == 0 {
			return nil, fmt.Errorf("list pods on node %q returned an empty page with next offset %d: %w", nodeName, *podsResp.Next, runaimcp.ErrUpstream)
		}
		offset = *podsResp.Next
	}
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
	if resp == nil {
		if requireExists {
			return fmt.Errorf("verify node %q in cluster %q returned a nil response: %w", nodeName, clusterUUID, runaimcp.ErrUpstream)
		}
		logger.Warn().Str("call", "getNodes").Msg("get_node_pods: listNodes returned nil, omitting node metadata")
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
		base := mapPodOutput(p)
		out[i] = nodePodOutput{
			Name:               base.Name,
			ID:                 base.ID,
			NodeName:           base.NodeName,
			CurrentNodePool:    base.CurrentNodePool,
			K8sPhase:           base.K8sPhase,
			CreatedAt:          base.CreatedAt,
			RevisionID:         base.RevisionID,
			ClusterID:          base.ClusterID,
			RequestedResources: base.RequestedResources,
			AllocatedResources: base.AllocatedResources,
			WorkloadID:         p.WorkloadId,
		}
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
	slots := make(chan struct{}, nodePodEnrichmentLimit)
	wg.Add(len(podsToEnrich))
	for id, idxs := range podsToEnrich {
		slots <- struct{}{}
		go func(workloadID string, podIndexes []int) {
			defer wg.Done()
			defer func() { <-slots }()
			applyWorkloadContext(ctx, client, pods, workloadID, podIndexes)
		}(id, idxs)
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
	if wl == nil {
		zerolog.Ctx(ctx).Warn().Str("workload_id", id).Str("call", "getWorkload").
			Msg("get_node_pods: workload enrichment returned nil")
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
