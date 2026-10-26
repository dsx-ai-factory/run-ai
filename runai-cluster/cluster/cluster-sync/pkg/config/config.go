/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"math"

	"go.uber.org/zap"

	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

type ClusterSyncConfig struct {
	ClusterUuid   string `mapstructure:"CLUSTER_ID" config:"secret"`
	ClusterDomain string `mapstructure:"CLUSTER_DOMAIN"`

	// Client details to get Application token from RUN:AI backend
	ClientName   string `mapstructure:"CLIENT_NAME" config:"secret"`
	ClientSecret string `mapstructure:"CLIENT_SECRET" config:"secret"`

	MaxNumOfEventsPerRequest      int `mapstructure:"MAX_NUM_OF_EVENTS_PER_REQUEST"`
	MaxNumOfPodsPerRequest        int `mapstructure:"MAX_NUM_OF_PODS_PER_REQUEST"`
	MaxNumOfNodesPerRequest       int `mapstructure:"MAX_NUM_OF_NODES_PER_REQUEST"`
	MaxNumOfWorkloadsPerRequest   int `mapstructure:"MAX_NUM_OF_WORKLOADS_PER_REQUEST"`
	MaxNumOfDatavolumesPerRequest int `mapstructure:"MAX_NUM_OF_DATAVOLUMES_PER_REQUEST"`

	SyncEventsIntervalSeconds int `mapstructure:"SYNC_EVENTS_INTERVAL_SECONDS"`

	// Base endpoint  for requests to run:ai backend
	RunaiApiEndPoint string `mapstructure:"ENDPOINT"`

	LivenessIntervalSeconds        int `mapstructure:"LIVENESS_INTERVAL_SECONDS"`
	ClusterInfoSyncIntervalSeconds int `mapstructure:"CLUSTER_INFO_SYNC_INTERVAL_SECONDS"`

	// Retries information for token retrieval and for calculating max failed liveness attempts
	MaxNumberOfRetries     int `mapstructure:"MAX_NUMBER_OF_RETRIES"`
	RetriesIntervalSeconds int `mapstructure:"RETRIES_INTERVAL_SECONDS"`
	SyncIntervalMinutes    int `mapstructure:"SYNC_INTERVAL_MINUTES"`

	ExistingSecretToSyncLabel string `mapstructure:"EXISTING_SECRET_TO_SYNC_LABEL"`
	AssetsSecretToSyncLabel   string `mapstructure:"ASSETS_SECRET_TO_SYNC_LABEL"`

	ConfigMapToSyncLabel string `mapstructure:"CONFIG_MAP_TO_SYNC_LABEL"`

	SSLCertificateVerification bool `mapstructure:"SSL_CERTIFICATE_VERIFICATION"`

	MetricsPort string `mapstructure:"METRICS_PORT"`
	MetricsPath string `mapstructure:"METRICS_PATH"`

	EnableProfiler       bool   `mapstructure:"ENABLE_PROFILER"`
	ProfilerApiPort      string `mapstructure:"PROFILER_API_PORT"`
	K8sClientConfigQPS   int    `mapstructure:"K8S_CLIENT_CONFIG_QPS"`
	K8sClientConfigBurst int    `mapstructure:"K8S_CLIENT_CONFIG_BURST"`

	NvidiaGPULabel                      string `mapstructure:"NVIDIA_GPU_LABEL"`
	NodeLabelNvidiaContainerToolkit     string `mapstructure:"NODE_LABEL_NVIDIA_CONTAINER_TOOLKIT"`
	NodeLabelNvidiaDcgmExporter         string `mapstructure:"NODE_LABEL_NVIDIA_DCGM_EXPORTER"`
	EnableNodeNvidiaComponentsReadiness bool   `mapstructure:"ENABLE_NODE_NVIDIA_COMPONENTS_READINESS"`

	SyncPodsWithWorkloads bool `mapstructure:"SYNC_PODS_WITH_WORKLOADS"`

	// for workloads, pods, etc...
	RateLimiterWorkloadsDelaySeconds int `mapstructure:"RATE_LIMITER_WORKLOADS_DELAY_SECONDS"`
	RateLimiterWorkloadsNumOfWorkers int `mapstructure:"RATE_LIMITER_WORKLOADS_NUM_OF_WORKERS"`

	// for projects, nodepools, etc...
	RateLimiterRegularDelaySeconds int `mapstructure:"RATE_LIMITER_REGULAR_DELAY_SECONDS"`
	RateLimiterRegularNumOfWorkers int `mapstructure:"RATE_LIMITER_REGULAR_NUM_OF_WORKERS"`

	EnableLeaderElection bool `mapstructure:"LEADER_ELECT"`
}

func (a *ClusterSyncConfig) SetDefaults() {
	a.MaxNumOfEventsPerRequest = 1000
	a.MaxNumOfPodsPerRequest = 1800
	a.MaxNumOfNodesPerRequest = 1000
	a.MaxNumOfWorkloadsPerRequest = 1000
	a.MaxNumOfDatavolumesPerRequest = 2000
	a.SyncEventsIntervalSeconds = 10
	a.LivenessIntervalSeconds = 10
	a.ClusterInfoSyncIntervalSeconds = 10
	a.MaxNumberOfRetries = 10
	a.RetriesIntervalSeconds = 10
	a.SyncIntervalMinutes = 180
	a.ExistingSecretToSyncLabel = constants.RunaiResourceLabel
	a.AssetsSecretToSyncLabel = constants.RunaiAssetLabel
	a.ConfigMapToSyncLabel = constants.RunaiResourceLabel
	a.SSLCertificateVerification = true
	a.MetricsPort = "9400"
	a.MetricsPath = "/metrics"
	a.EnableProfiler = false
	a.ProfilerApiPort = "8182"
	a.K8sClientConfigQPS = 50
	a.K8sClientConfigBurst = 300
	a.NvidiaGPULabel = "feature.node.kubernetes.io/pci-10de.present"
	a.SyncPodsWithWorkloads = false
	a.NodeLabelNvidiaContainerToolkit = "nvidia.com/gpu.deploy.container-toolkit"
	a.NodeLabelNvidiaDcgmExporter = "nvidia.com/gpu.deploy.dcgm-exporter"
	a.EnableNodeNvidiaComponentsReadiness = false
	a.RateLimiterWorkloadsDelaySeconds = 10
	a.RateLimiterWorkloadsNumOfWorkers = 3
	a.RateLimiterRegularDelaySeconds = 5
	a.RateLimiterRegularNumOfWorkers = 1
	a.EnableLeaderElection = false
}

func RetriesIntervalTime() int {
	return int(math.Max(float64(Config().RetriesIntervalSeconds), 1))
}

func MaxAllowedWaitTimeInSec() int {
	var configuredMaxAllowedTime = Config().MaxNumberOfRetries * RetriesIntervalTime()
	const lowestTimeAllowedInSec = 60 * 2 //2 minutes
	return int(math.Max(float64(configuredMaxAllowedTime), lowestTimeAllowedInSec))
}

func Init() *ClusterSyncConfig {
	zap.S().Debug("initializing config")

	err := rcpconfig.SetSecretConfigFile(constants.SecretMountPath)
	if err != nil {
		zap.S().Fatal("Unable to load secret: ", err.Error())
	}

	err = rcpconfig.Register[*ClusterSyncConfig]()
	if err != nil {
		zap.S().Fatal("Unable to register configuration: ", err.Error())
	}

	config := Config()
	if config == nil {
		zap.S().Fatal("Failed to load agent config, terminating...")
	}

	return config
}

func Config() *ClusterSyncConfig {
	return rcpconfig.Get[*ClusterSyncConfig]()
}
