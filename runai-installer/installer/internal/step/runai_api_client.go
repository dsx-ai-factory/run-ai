/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"

	cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type RunAIAPIClient interface {
	GetClusters(ctx context.Context) ([]cluster.DisplayedCluster, int, error)
	CreateCluster(ctx context.Context, req cluster.ClusterCreationRequest) (*cluster.DisplayedCluster, int, error)
	GetClusterByUUID(ctx context.Context, uuid string) (*cluster.DisplayedCluster, int, error)
	GetClusterInstallInfoByUUID(ctx context.Context, uuid, version, remoteClusterURL string) (*cluster.ClusterInstallationInfoResponse, int, error)
}

type RunAIAPIClientProvider func(fqdn, caCertPath string) (RunAIAPIClient, error)

type ControlPlaneCredentials struct {
	Username string
	Password string
	Token    string
}

type RunAIAPICredentialedClientProvider func(url, caCertPath string, creds ControlPlaneCredentials) (RunAIAPIClient, error)
