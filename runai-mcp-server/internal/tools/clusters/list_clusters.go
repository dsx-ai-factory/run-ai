/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type listClustersInput struct {
	Name   string `json:"name,omitempty"   jsonschema:"Exact cluster display name, sent to the API as name==value. Omit to list visible clusters."`
	Limit  int32  `json:"limit,omitempty"  jsonschema:"Maximum matching clusters to return. Defaults to 50; API maximum 500."`
	Offset int32  `json:"offset,omitempty" jsonschema:"API pagination offset. Defaults to 0; advance by the page size after a full page."`
}

func (in listClustersInput) Validate() error {
	return common.ValidateFilterValue("name", in.Name)
}

type clusterListItem struct {
	ClusterID string `json:"clusterId" jsonschema:"Cluster UUID accepted by cluster-scoped tool inputs."`
	Name      string `json:"name"      jsonschema:"Cluster display name. Names are not guaranteed to be unique."`
}

type listClustersOutput struct {
	Clusters []clusterListItem `json:"clusters" jsonschema:"Cluster IDs and names in the API page, preserving API order. Empty when no clusters match."`
}

func registerListClusters(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listClustersInput, listClustersOutput]{
		Name: "list_clusters",
		Description: "List visible clusters and their UUIDs using API filtering and pagination. " +
			"After a full page, advance offset by the page size; stop after a shorter page. " +
			"When resolving a name, check every page and ask the user if multiple clusters match.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listClustersInput) (listClustersOutput, error) {
			return listClustersRun(ctx, client, in)
		},
	})
}

func listClustersRun(ctx context.Context, client *runaiclient.Client, in listClustersInput) (listClustersOutput, error) {
	request := client.Cluster.ClustersAPI.GetClustersMinimalData(ctx).Limit(common.EffectiveLimit(in.Limit)).Offset(in.Offset)
	if in.Name != "" {
		request = request.FilterBy([]string{common.FilterEq("name", in.Name)})
	}
	response, resp, err := request.Execute()
	if err != nil {
		return listClustersOutput{}, fmt.Errorf("list clusters: %w", runaimcp.MapHTTPError(resp, err))
	}

	clusters := response.GetClusters()
	items := make([]clusterListItem, 0, len(clusters))
	for _, cluster := range clusters {
		items = append(items, clusterListItem{ClusterID: cluster.Uuid, Name: cluster.Name})
	}
	return listClustersOutput{Clusters: items}, nil
}
