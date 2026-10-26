/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

var RequiredNodeTelemetryTypes = []openapi.NodeTelemetryType{
	openapi.NODETELEMETRYTYPE_FREE_GPUS,
}

type Client interface {
	GetNodepoolList(ctx context.Context, filters []string) (*openapi.GetNodepools200Response, error)
	GetNodepoolMinimalList(ctx context.Context, filters []string) (*openapi.GetNodepoolsMinimal200Response, error)
}

type BackendNodepoolService struct {
	nodepoolClient Client
	nodeService    nodes.Client
}

func NewBackendNodepoolService(nodepoolClient Client, nodeService nodes.Client) *BackendNodepoolService {
	return &BackendNodepoolService{nodepoolClient: nodepoolClient, nodeService: nodeService}
}

func (bns *BackendNodepoolService) GetNodepoolList(ctx context.Context,
	clusterUuid string) (*openapi.GetNodepools200Response, *openapi.Nodes, api.NodeTelemetriesByType, error) {
	nodesList, err := nodes.FetchAllNodes(ctx, bns.nodeService, clusterUuid)
	if err != nil {
		return nil, nil, nil, err
	}

	nodesTelemetry, err := bns.nodeService.GetNodesTelemetry(ctx, clusterUuid, RequiredNodeTelemetryTypes)
	if err != nil {
		return nil, nil, nil, err
	}

	nodepoolList, err := bns.nodepoolClient.GetNodepoolList(ctx, filter.NewBuilder().WithClusterID(clusterUuid).StringList())
	if err != nil {
		return nil, nil, nil, err
	}

	return nodepoolList, &openapi.Nodes{Nodes: nodesList}, nodesTelemetry, nil
}

func (bns *BackendNodepoolService) GetNodepoolMinimalList(ctx context.Context,
	clusterUuid string) (*openapi.GetNodepoolsMinimal200Response, error) {
	nodepoolMinimalList, err := bns.nodepoolClient.GetNodepoolMinimalList(ctx, filter.NewBuilder().WithClusterID(clusterUuid).StringList())
	if err != nil {
		return nil, err
	}
	return nodepoolMinimalList, nil
}
