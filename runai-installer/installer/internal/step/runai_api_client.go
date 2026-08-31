/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"

	"github.com/run-ai/runai/runai-installer/clients/runai/cluster"
)

type RunAIAPIClient interface {
	GetClusters(ctx context.Context) ([]cluster.DisplayedCluster, int, error)
	CreateCluster(ctx context.Context, req cluster.ClusterCreationRequest) (*cluster.DisplayedCluster, int, error)
	GetClusterByUUID(ctx context.Context, uuid string) (*cluster.DisplayedCluster, int, error)
	GetClusterInstallInfoByUUID(ctx context.Context, uuid, version, remoteClusterURL string) (*cluster.ClusterInstallationInfoResponse, int, error)
}

type RunAIAPIClientProvider func(fqdn, caCertPath string) (RunAIAPIClient, error)
