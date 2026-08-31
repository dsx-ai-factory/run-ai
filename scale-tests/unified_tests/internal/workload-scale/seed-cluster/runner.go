/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package seed_cluster

import (
	"fmt"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	seeder "github.com/run-ai/runai/scale-tests/unified_tests/internal/cluster-seeder"
	projects "github.com/run-ai/runai/scale-tests/unified_tests/internal/projects/seed_cluster"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

type WorkloadSamplesRunner struct {
	samplerConfig *WorkloadSamplerConfig
	clusterConfig *scale.ClusterConfig
}

func NewWorkloadSamplesRunner(clusterConfig *scale.ClusterConfig) seeder.SamplerRunner {
	return &WorkloadSamplesRunner{
		samplerConfig: ConstructConfig(),
		clusterConfig: clusterConfig,
	}
}

func (wstr *WorkloadSamplesRunner) InitSampledEntities(clients *sampler_actions.Clients) ([]scale.Entity, error) {
	sampledEntities := []scale.Entity{}
	projectPrefix, err := generateUniqueProjectIdentifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique project identifier: %v", err)
	}
	projectsConfig := toProjectConfig(wstr.clusterConfig, wstr.samplerConfig)
	workloadConfig := toWorkloadConfig(wstr.clusterConfig, wstr.samplerConfig)
	projectsTestActions := projects.NewProjectSampler(wstr.clusterConfig, projectsConfig, clients, projectPrefix)
	workloadsTestActions := NewWorkloadSampler(wstr.clusterConfig, workloadConfig, clients, projectsTestActions.(*projects.ProjectSampler))
	sampledEntities = append(sampledEntities, projectsTestActions)
	sampledEntities = append(sampledEntities, workloadsTestActions)
	return sampledEntities, nil
}

func (wstr *WorkloadSamplesRunner) UpdateRawResultReportMetaData(report *scale.SamplesReport) {
	wstr.samplerConfig.updateSamplerReportMetaData(report)
}

func (wstr *WorkloadSamplesRunner) GetMetricLabels() map[string]string {
	return map[string]string{
		"workload_count":   fmt.Sprintf("%d", wstr.samplerConfig.WorkloadCount),
		"projects_count":   fmt.Sprintf("%d", wstr.samplerConfig.ProjectCount),
		"nodepool_count":   fmt.Sprintf("%d", wstr.clusterConfig.NodepoolCount),
		"nodes_per_pool":   fmt.Sprintf("%d", wstr.clusterConfig.NodesPerPool),
		"gpu_per_workload": fmt.Sprintf("%d", int(wstr.samplerConfig.GpuPerWorkload)),
		"cluster_id":       wstr.clusterConfig.ClusterID,
		"operator_version": wstr.clusterConfig.OperatorVersion,
		"cp_version":       wstr.clusterConfig.CpVersion,
		"scale_group":      wstr.clusterConfig.ScaleGroup,
	}
}

func generateUniqueProjectIdentifier() (string, error) {
	projectPrefix, err := gonanoid.New()
	if err != nil {
		return "", err
	}
	projectPrefix = strings.ToLower(projectPrefix)
	projectPrefix = strings.Replace(projectPrefix, "_", "", -1)
	return projectPrefix, nil
}

func toProjectConfig(clusterConfig *scale.ClusterConfig, samplerConfig *WorkloadSamplerConfig) *projects.ProjectConfig {
	durationForProjectWait, _ := time.ParseDuration(samplerConfig.ProjectCreateTimeOut)
	return &projects.ProjectConfig{
		ProjectCount:           samplerConfig.ProjectCount,
		ControlPlaneQPS:        clusterConfig.ControlPlaneQPS,
		ClusterID:              clusterConfig.ClusterID,
		ProjectCreationTimeout: durationForProjectWait,
		MonitorWaitTime:        scale.WaitTime(1 * time.Minute),
		BatchSize:              clusterConfig.BatchSize,
	}
}

func toWorkloadConfig(clusterConfig *scale.ClusterConfig, samplerConfig *WorkloadSamplerConfig) *WorkloadConfig {
	durationForWorkloadsWait, _ := time.ParseDuration(samplerConfig.WorkloadCreateTimeOut)

	workloadConfig := &WorkloadConfig{
		WorkloadCount:               samplerConfig.WorkloadCount,
		WorkloadType:                samplerConfig.WorkloadType,
		DistributedFramework:        samplerConfig.DistributedFramework,
		WorkloadWorkerCount:         samplerConfig.WorkloadWorkerCount,
		ProjectsCount:               samplerConfig.ProjectCount,
		ControlPlaneQPS:             clusterConfig.ControlPlaneQPS,
		WorkloadCreateTimeOut:       durationForWorkloadsWait,
		ClusterID:                   clusterConfig.ClusterID,
		NodepoolCount:               clusterConfig.NodepoolCount,
		GpuPerWorkload:              samplerConfig.GpuPerWorkload,
		ProjectLoadMethod:           samplerConfig.ProjectLoadMethod,
		NumberOfProjectsToHeavyLoad: samplerConfig.NumberOfProjectsToHeavyLoad,
		MonitorWaitTime:             scale.WaitTime(durationForWorkloadsWait),
		BatchSize:                   clusterConfig.BatchSize,
		KwokEnabled:                 clusterConfig.KwokEnabled,
		GpuProduct:                  clusterConfig.GpuProduct,
	}
	workloadCommands := []string{}
	if samplerConfig.RunInfiniteWorkloads {
		workloadCommands = append(workloadCommands, "sleep infinity")
	}
	for _, sleepTime := range samplerConfig.WorkloadSleepTimesInSeconds {
		workloadCommands = append(workloadCommands, fmt.Sprintf("sleep %d", sleepTime))
	}
	workloadConfig.WorkloadsCommands = workloadCommands

	if samplerConfig.NodeChaosEnabled {
		workloadConfig.NodeChaosEnabled = true
		workloadConfig.NodeChaosCount = samplerConfig.NodeChaosCount
		workloadConfig.NodeChaosDownDuration, _ = time.ParseDuration(samplerConfig.NodeChaosDownDuration)
		workloadConfig.NodeChaosNodeSelector = samplerConfig.NodeChaosNodeSelector
		workloadConfig.NodeChaosStartDelay, _ = time.ParseDuration(samplerConfig.NodeChaosStartDelay)
		workloadConfig.NodeChaosDelayBetween, _ = time.ParseDuration(samplerConfig.NodeChaosDelayBetween)
	}

	return workloadConfig
}
