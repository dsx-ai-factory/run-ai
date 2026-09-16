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

// RunAIAPIClientProvider returns a control-plane client authenticated with the
// admin credentials the installer reads from the control plane's own
// Kubernetes secrets. It only works when the control plane runs in the
// cluster the installer is connected to.
type RunAIAPIClientProvider func(fqdn, caCertPath string) (RunAIAPIClient, error)

// ControlPlaneCredentials are user-supplied credentials for an external
// control plane: either Username+Password or a pre-issued bearer Token.
type ControlPlaneCredentials struct {
	Username string
	Password string
	Token    string
}

// RunAIAPICredentialedClientProvider returns a control-plane client for an
// external control plane, authenticated with the supplied credentials rather
// than with secrets read from the local cluster.
type RunAIAPICredentialedClientProvider func(url, caCertPath string, creds ControlPlaneCredentials) (RunAIAPIClient, error)
