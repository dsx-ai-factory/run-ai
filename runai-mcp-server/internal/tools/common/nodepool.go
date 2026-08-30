package common

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

// ResolveNodePoolIDs resolves names with one list call and matches in memory — names aren't
// pushed into the filter, so the per-name validation is an early reject, not an injection
// guard. Errors on an unknown or ambiguous name.
func ResolveNodePoolIDs(ctx context.Context, client *runaiclient.Client, clusterID string, names []string) (map[string]string, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("clusterId is required to resolve node pools: %w", runaimcp.ErrValidation)
	}
	if err := ValidateFilterValue("clusterId", clusterID); err != nil {
		return nil, fmt.Errorf("%v: %w", err, runaimcp.ErrValidation)
	}
	for _, n := range names {
		if err := ValidateFilterValue("nodePoolName", n); err != nil {
			return nil, fmt.Errorf("%v: %w", err, runaimcp.ErrValidation)
		}
	}
	resp, httpResp, err := client.Cluster.NodePoolsAPI.GetNodepools(ctx).
		FilterBy([]string{FilterEq("clusterId", clusterID)}).Limit(200).Execute()
	if err != nil {
		return nil, fmt.Errorf("list node pools on cluster %q: %w", clusterID, runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return nil, fmt.Errorf("list node pools on cluster %q: nil response: %w", clusterID, runaimcp.ErrUpstream)
	}
	byName := map[string]string{}
	ambiguous := map[string]bool{}
	for _, np := range resp.Nodepools {
		if _, seen := byName[np.Name]; seen {
			ambiguous[np.Name] = true
		}
		byName[np.Name] = np.Id
	}
	out := make(map[string]string, len(names))
	for _, n := range names {
		if ambiguous[n] {
			return nil, fmt.Errorf("ambiguous_node_pool: name %q matches multiple pools on cluster %q; pass an explicit ID: %w", n, clusterID, runaimcp.ErrValidation)
		}
		id, ok := byName[n]
		if !ok {
			return nil, fmt.Errorf("node pool %q on cluster %q: %w", n, clusterID, runaimcp.ErrNotFound)
		}
		out[n] = id
	}
	return out, nil
}

func ResolveNodePoolID(ctx context.Context, client *runaiclient.Client, clusterID, name string) (string, error) {
	ids, err := ResolveNodePoolIDs(ctx, client, clusterID, []string{name})
	if err != nil {
		return "", err
	}
	return ids[name], nil
}
