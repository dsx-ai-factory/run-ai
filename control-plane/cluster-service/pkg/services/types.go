/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	clustercpconfigprovider "github.com/run-ai/runai/control-plane/cluster-service/pkg/business/cluster-cp-config/provider"
	cluster_delete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"
	cluster_install "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-installation"
)

type Verbosity int

const (
	VerbosityBrief Verbosity = iota
	VerbosityVerbose
)

type ClusterServiceServices struct {
	HealthApiService         api.HealthAPIServicer
	NodesApiService          api.NodesAPIServicer
	NodepoolsApiService      api.NodePoolsAPIServicer
	ClusterDeleter           cluster_delete.ClusterDeleter
	ClusterInstaller         cluster_install.ClusterInstallService
	ClustersApiService       api.ClustersAPIServicer
	InternalApiService       api.InternalAPIServicer
	CpConfigProvider         *clustercpconfigprovider.ClusterControlPlaneConfigProvider
	CpConfigApiService       api.ControlPlaneConfigAPIServicer
	ClustersApiConfigService api.ClusterConfigAPIServicer
	NetworkTopologiesService api.NetworkTopologiesAPIServicer
	SettingsApiService       api.SettingsAPIServicer
}
