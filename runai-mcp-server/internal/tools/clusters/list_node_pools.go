package clusters

import (
	"context"
	"fmt"
	"sort"
	"strings"

	clusterservice "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type listNodePoolsInput struct {
	ClusterID string `json:"clusterId,omitempty" jsonschema:"Cluster ID to list node pools for. If omitted, lists node pools across every cluster the caller can access."`
	Limit     int32  `json:"limit,omitempty"     jsonschema:"Maximum number of node pools to return. Defaults to 50, capped at 200."`
	Offset    int32  `json:"offset,omitempty"    jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listNodePoolsInput) Validate() error {
	if err := common.ValidatePagination(in.Limit, in.Offset); err != nil {
		return err
	}
	return common.ValidateNoCommas("clusterId", in.ClusterID)
}

type nodePool struct {
	NodePoolID          string  `json:"nodePoolId"          jsonschema:"Node pool ID."`
	NodePoolName        string  `json:"nodePoolName"        jsonschema:"Node pool display name."`
	ClusterID           string  `json:"clusterId"           jsonschema:"ID of the cluster this node pool belongs to."`
	IsDefault           bool    `json:"isDefault"           jsonschema:"True for the cluster's default node pool."`
	Phase               string  `json:"phase"               jsonschema:"Lifecycle phase. One of: Creating, Updating, Deleting, Empty, Unschedulable, Ready, Deleted."`
	NetworkTopologyName *string `json:"networkTopologyName" jsonschema:"Name of the associated network topology. Null when none is set."`
}

type listNodePoolsOutput struct {
	NodePools  []nodePool `json:"nodePools"            jsonschema:"Node pools matching the request. Always present; empty array when none match."`
	HasMore    bool       `json:"hasMore"              jsonschema:"True when more node pools exist beyond the returned page; pass nextOffset back as offset to fetch them."`
	NextOffset int32      `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch additional node pools. Only set when hasMore is true."`
}

func registerListNodePools(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listNodePoolsInput, listNodePoolsOutput]{
		Name: "list_node_pools",
		Description: "List the node pools in a cluster. Call when the user asks what node pools exist " +
			"(e.g. \"what node pools do we have?\") or to discover pools before drilling into one " +
			"(before list_project_resources or get_cluster_metrics per pool). Returns each pool's " +
			"identity, whether it is the cluster default, its lifecycle phase, and its network topology. " +
			"Does NOT return quota assignments, metrics, or GPU type/capacity.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listNodePoolsInput) (listNodePoolsOutput, error) {
			return listNodePoolsRun(ctx, client, in)
		},
	})
}

func listNodePoolsRun(ctx context.Context, client *runaiclient.Client, in listNodePoolsInput) (listNodePoolsOutput, error) {
	limit := common.EffectiveLimit(in.Limit)

	raw, err := fetchAllNodePools(ctx, client, buildFilters(in))
	if err != nil {
		return listNodePoolsOutput{}, err
	}

	pools := make([]nodePool, 0, len(raw))
	for _, np := range raw {
		pools = append(pools, mapNodePool(np))
	}
	sortNodePools(pools)

	page, hasMore, next := common.Paginate(pools, in.Offset, limit)
	return listNodePoolsOutput{
		NodePools:  page,
		HasMore:    hasMore,
		NextOffset: next,
	}, nil
}

func fetchAllNodePools(ctx context.Context, client *runaiclient.Client, filters []string) ([]clusterservice.Nodepool, error) {
	var all []clusterservice.Nodepool
	offset := int32(0)
	for {
		req := client.Cluster.NodePoolsAPI.GetNodepools(ctx).Offset(offset).Limit(common.UpstreamPageSize)
		if len(filters) > 0 {
			req = req.FilterBy(filters)
		}
		resp, httpResp, err := req.Execute()
		if err != nil {
			return nil, fmt.Errorf("list node pools: %w", runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			break
		}
		all = append(all, resp.Nodepools...)
		if int32(len(all)) > common.MaxPaginatedResults {
			return nil, fmt.Errorf("node pool set exceeds the maximum supported size (%d); narrow with clusterId: %w", common.MaxPaginatedResults, runaimcp.ErrUpstream)
		}
		if int32(len(resp.Nodepools)) < common.UpstreamPageSize {
			break
		}
		offset += common.UpstreamPageSize
	}
	return all, nil
}

func buildFilters(in listNodePoolsInput) []string {
	if strings.TrimSpace(in.ClusterID) == "" {
		return nil
	}
	return []string{common.FilterEq("clusterId", in.ClusterID)}
}

func mapNodePool(np clusterservice.Nodepool) nodePool {
	out := nodePool{
		NodePoolID:          np.Id,
		NodePoolName:        np.Name,
		ClusterID:           np.ClusterId,
		IsDefault:           np.IsDefault,
		NetworkTopologyName: np.NetworkTopologyName.Get(),
	}
	if np.Phase != nil {
		out.Phase = string(*np.Phase)
	}
	return out
}

func sortNodePools(pools []nodePool) {
	sort.SliceStable(pools, func(i, j int) bool {
		if pools[i].ClusterID != pools[j].ClusterID {
			return pools[i].ClusterID < pools[j].ClusterID
		}
		if pools[i].NodePoolName != pools[j].NodePoolName {
			return pools[i].NodePoolName < pools[j].NodePoolName
		}
		return pools[i].NodePoolID < pools[j].NodePoolID
	})
}
