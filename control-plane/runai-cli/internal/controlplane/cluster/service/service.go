/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
)

type ClusterClient interface {
	GetClusterList(ctx context.Context) ([]openapi.ClustersMinimalData, error)
	GetClusterByID(ctx context.Context, uuid string) (*openapi.ClustersMinimalData, error)
	GetClusterByName(ctx context.Context, name string) (*openapi.ClustersMinimalData, error)
}

type BackendClusterService struct {
	clusterClient ClusterClient
}

func NewBackendClusterService(clusterClient ClusterClient) *BackendClusterService {
	return &BackendClusterService{clusterClient: clusterClient}
}

func (bcs *BackendClusterService) GetClusterList(ctx context.Context) (*list.Output, error) {
	clusterList, err := bcs.clusterClient.GetClusterList(ctx)
	if err != nil {
		return nil, err
	}
	return list.NewOutput(clusterList, nil)
}

func (bcs *BackendClusterService) GetClusterByUuid(ctx context.Context, uuid string) (*manipulate.BriefCluster, error) {
	cluster, err := bcs.clusterClient.GetClusterByID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	briefCluster := manipulate.ReduceStruct(*cluster)
	return &briefCluster, nil
}

func (bcs *BackendClusterService) GetClusterByName(ctx context.Context, name string) (*manipulate.BriefCluster, error) {
	cluster, err := bcs.clusterClient.GetClusterByName(ctx, name)
	if err != nil {
		return nil, err
	}
	briefCluster := manipulate.ReduceStruct(*cluster)
	return &briefCluster, nil
}
