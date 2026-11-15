/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
)

// ClusterSnapshot is the in-memory aggregator of cluster-reported
// configuration values consumed by the anyworkload-path init container
// constructors. Composers
// read it at workload-composition time so they don't have to traverse the
// wire-format ClusterReportedConfig shape. Nil fields mean the cluster
// hasn't reported the value yet; consumers should surface a typed
// errClusterValueMissing for fields they treat as required.
type ClusterSnapshot struct {
	GitSyncImage                           *string
	S3FileSystemImage                      *string
	S3FileSystemMountCheckImage            *string
	FipsMode                               *string
	RunaiOauth2ProxyImage                  *string
	RunaiOauth2ProxyUsernameLoaderImage    *string
	RunaiOauth2ProxyRedisFQDN              *string
	RunaiOauth2ProxyWorkloadAuthSecretName *string
	RunaiOauth2ProxySessionStoreType       *string
	ExternalAuthUrlEnabled                 *bool
	OpenshiftRouteTerminationPolicy        *string
	CustomCAGit                            *ClusterSnapshotCustomCA
	CustomCAS3                             *ClusterSnapshotCustomCA
	CustomCAOauth2Proxy                    *ClusterSnapshotCustomCA
}

// ClusterSnapshotCustomCA mirrors api.ClusterReportedCustomCa for a single
// consumer (git-sync, S3, or oauth2-proxy). It exists so callers don't take
// a dependency on the generated package and can be evolved independently.
type ClusterSnapshotCustomCA struct {
	Enabled    *bool
	SecretName *string
	SecretKey  *string
	Optional   *bool
}

// NewClusterSnapshotFromCluster composes a ClusterSnapshot from a cluster's
// stored DisplayedCluster.Status.Config (the JSONB written by cluster-sync's
// ship-up). Returns a non-nil snapshot even when the cluster has no reported
// config yet, so callers don't need to nil-check the result.
func NewClusterSnapshotFromCluster(cluster *api.DisplayedCluster) *ClusterSnapshot {
	snapshot := &ClusterSnapshot{}
	if cluster == nil || cluster.Status == nil || cluster.Status.Config == nil {
		return snapshot
	}
	cfg := cluster.Status.Config

	snapshot.GitSyncImage = cfg.GitSyncImage
	snapshot.S3FileSystemImage = cfg.S3FileSystemImage
	snapshot.S3FileSystemMountCheckImage = cfg.S3FileSystemMountCheckImage
	if cfg.FipsMode != nil {
		fipsMode := string(*cfg.FipsMode)
		snapshot.FipsMode = &fipsMode
	}
	snapshot.RunaiOauth2ProxyImage = cfg.RunaiOauth2ProxyImage
	snapshot.RunaiOauth2ProxyUsernameLoaderImage = cfg.RunaiOauth2ProxyUsernameLoaderImage
	snapshot.RunaiOauth2ProxyRedisFQDN = cfg.RunaiOauth2ProxyRedisFQDN
	snapshot.RunaiOauth2ProxyWorkloadAuthSecretName = cfg.RunaiOauth2ProxyWorkloadAuthSecretName
	snapshot.RunaiOauth2ProxySessionStoreType = cfg.RunaiOauth2ProxySessionStoreType
	snapshot.ExternalAuthUrlEnabled = cfg.ExternalAuthUrlEnabled
	if cfg.OpenshiftRouteTerminationPolicy != nil {
		terminationPolicy := string(*cfg.OpenshiftRouteTerminationPolicy)
		snapshot.OpenshiftRouteTerminationPolicy = &terminationPolicy
	}
	snapshot.CustomCAGit = customCAToSnapshot(cfg.CustomCAGit)
	snapshot.CustomCAS3 = customCAToSnapshot(cfg.CustomCAS3)
	snapshot.CustomCAOauth2Proxy = customCAToSnapshot(cfg.CustomCAOauth2Proxy)

	return snapshot
}

func customCAToSnapshot(ca *api.ClusterReportedCustomCa) *ClusterSnapshotCustomCA {
	if ca == nil {
		return nil
	}
	return &ClusterSnapshotCustomCA{
		Enabled:    ca.Enabled,
		SecretName: ca.SecretName,
		SecretKey:  ca.SecretKey,
		Optional:   ca.Optional,
	}
}
