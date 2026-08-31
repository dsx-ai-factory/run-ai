/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package seed_cluster

import (
	"fmt"
	"reflect"
	"time"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	wclient "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/pkg/errors"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type LoadMethod string

var Uniform LoadMethod = "uniform"
var Heavy LoadMethod = "heavy"

type WorkloadSamplerConfig struct {
	ProjectCreateTimeOut        string               `mapstructure:"project_create_timeout"`
	WorkloadCreateTimeOut       string               `mapstructure:"workload_create_timeout"`
	ProjectCount                int                  `mapstructure:"project_count"`
	WorkloadType                wclient.WorkloadType `mapstructure:"workload_type"`
	WorkloadCount               int                  `mapstructure:"workload_count"`
	GpuPerWorkload              float64              `mapstructure:"gpu_per_workload"`
	ProjectLoadMethod           LoadMethod           `mapstructure:"project_load_method"`
	NumberOfProjectsToHeavyLoad int                  `mapstructure:"number_of_projects_to_heavy_load"`
	RunInfiniteWorkloads        bool                 `mapstructure:"run_infinite_workloads"`
	WorkloadSleepTimesInSeconds []int                `mapstructure:"workload_sleep_times"`
	GpuProduct                  []string             `mapstructure:"gpu_product"`

	// The following fields are relevant for distributed workloads only
	WorkloadWorkerCount  int                          `mapstructure:"workload_worker_count"`
	DistributedFramework wclient.DistributedFramework `mapstructure:"distributed_framework"`

	// Node chaos configuration - simulates node failures during workload execution
	NodeChaosEnabled      bool   `mapstructure:"node_chaos_enabled"`
	NodeChaosCount        int    `mapstructure:"node_chaos_count"`
	NodeChaosDownDuration string `mapstructure:"node_chaos_down_duration"`
	NodeChaosNodeSelector string `mapstructure:"node_chaos_node_selector"`
	NodeChaosStartDelay   string `mapstructure:"node_chaos_start_delay"`
	NodeChaosDelayBetween string `mapstructure:"node_chaos_delay_between"`
}

func (wtc *WorkloadSamplerConfig) SetDefaults() {
	wtc.WorkloadCreateTimeOut = "20m"
	wtc.ProjectCreateTimeOut = "20m"
	wtc.ProjectCount = 100
	wtc.WorkloadType = wclient.WORKLOADTYPE_DISTRIBUTED
	wtc.WorkloadCount = 400
	wtc.WorkloadWorkerCount = 4
	wtc.DistributedFramework = wclient.DISTRIBUTEDFRAMEWORK_PY_TORCH
	wtc.GpuPerWorkload = 1
	wtc.ProjectLoadMethod = Uniform
	wtc.NumberOfProjectsToHeavyLoad = 1
	wtc.RunInfiniteWorkloads = false
	wtc.WorkloadSleepTimesInSeconds = []int{180}

	wtc.NodeChaosEnabled = false
	wtc.NodeChaosCount = 0
	wtc.NodeChaosDownDuration = "5m"
	wtc.NodeChaosNodeSelector = "nvidia.com/gpu.present=true"
	wtc.NodeChaosStartDelay = "2m"
	wtc.NodeChaosDelayBetween = "10s"
}

func ConstructConfig() *WorkloadSamplerConfig {
	workloadConfig, err := config.GetR[*WorkloadSamplerConfig]()
	if err != nil {
		logging.Fatalf("failed to register config, error: %v. Service is terminating", err)
	}
	err = validateConfig(workloadConfig)
	if err != nil {
		logging.Fatalf("failed to validate config, error: %v. Service is terminating", err)
	}
	return workloadConfig
}

func validateConfig(testConfig *WorkloadSamplerConfig) error {
	allErrors := errors.New("config validation errors")
	hadErrors := false
	_, err := time.ParseDuration(testConfig.WorkloadCreateTimeOut)
	if err != nil {
		allErrors = errors.Wrapf(allErrors, "error parsing: [%s], error: [%v]", "WorkloadCreateTimeOut", err)
		hadErrors = true
	}
	_, err = time.ParseDuration(testConfig.ProjectCreateTimeOut)
	if err != nil {
		allErrors = errors.Wrapf(allErrors, "error parsing: [%s], error: [%v]", "ProjectCreateTimeOut", err)
		hadErrors = true
	}
	if testConfig.NodeChaosEnabled {
		_, err = time.ParseDuration(testConfig.NodeChaosDownDuration)
		if err != nil {
			allErrors = errors.Wrapf(allErrors, "error parsing: [%s], error: [%v]", "NodeChaosDownDuration", err)
			hadErrors = true
		}
		_, err = time.ParseDuration(testConfig.NodeChaosStartDelay)
		if err != nil {
			allErrors = errors.Wrapf(allErrors, "error parsing: [%s], error: [%v]", "NodeChaosStartDelay", err)
			hadErrors = true
		}
		_, err = time.ParseDuration(testConfig.NodeChaosDelayBetween)
		if err != nil {
			allErrors = errors.Wrapf(allErrors, "error parsing: [%s], error: [%v]", "NodeChaosDelayBetween", err)
			hadErrors = true
		}
		if testConfig.NodeChaosCount <= 0 {
			allErrors = errors.Wrapf(allErrors, "node_chaos_count must be > 0 when node_chaos_enabled is true")
			hadErrors = true
		}
	}
	if hadErrors {
		return allErrors
	}
	return nil
}

func (w *WorkloadSamplerConfig) updateSamplerReportMetaData(samplerReport *scale.SamplesReport) {
	if samplerReport.MetaData == nil {
		samplerReport.MetaData = make(map[string]string)
	}
	val := reflect.ValueOf(*w)
	typ := reflect.TypeOf(*w)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		samplerReport.MetaData[field.Name] = fmt.Sprintf("%v", value)
	}
}
