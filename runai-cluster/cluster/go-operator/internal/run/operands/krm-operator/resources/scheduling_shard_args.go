/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"encoding/json"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/kai/schedulingshardargs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runai_scheduler "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-scheduler"
)

// marshalShardArgs builds the cluster-wide SchedulingShard args from RunaiConfig
// - both the RunAIScheduler settings and the global NodeAffinity policy - and
// JSON-encodes them in KAI vocabulary for the --scheduling-shard-args flag.
//
// Only SchedulingShard-facing fields are translated here; the
// scheduler-deployment fields (image, resources, vpa, ...) are consumed
// separately by the kai-operator operand. Keeping this translation on the
// operator side lets nodepool-controller receive already-KAI args and stay
// free of run.ai types. The controller owns the base worker-node-label args,
// so they are intentionally not produced here.
func marshalShardArgs(schedulerConfig *runai_scheduler.RunAIScheduler, nodeAffinity *global.NodeAffinity) (string, error) {
	args := buildShardArgs(schedulerConfig, nodeAffinity)
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	return string(argsBytes), nil
}

func buildShardArgs(config *runai_scheduler.RunAIScheduler, nodeAffinity *global.NodeAffinity) map[string]string {
	args := map[string]string{}
	if nodeAffinity != nil {
		args[schedulingshardargs.RestrictNodeScheduling] = strconv.FormatBool(nodeAffinity.RestrictScheduling)
	}
	if config == nil {
		return args
	}
	setBoolArg(args, schedulingshardargs.DetailedFitErrors, config.DetailedFitErrors)
	setBoolArg(args, schedulingshardargs.ScheduleCSIStorage, config.AdvancedCSIScheduling)
	if config.Profiling != nil {
		setBoolArg(args, schedulingshardargs.EnableProfiler, &config.Profiling.Enabled)
	}
	if config.Args == nil {
		return args
	}
	a := config.Args
	setIntArg(args, schedulingshardargs.Verbosity, a.Verbosity)
	setStringArg(args, schedulingshardargs.SchedulerName, a.DefaultSchedulerName)
	setStringArg(args, schedulingshardargs.SchedulePeriod, a.DefaultSchedulerPeriod)
	setIntArg(args, schedulingshardargs.MaxConsolidationPreemptees, a.MaxNumberConsolidationPreemptees)
	setIntArg(args, schedulingshardargs.QPS, a.ClientQPS)
	setIntArg(args, schedulingshardargs.Burst, a.ClientBurst)
	setBoolArg(args, schedulingshardargs.UseSchedulingSignatures, a.UseSchedulingSignatures)
	setBoolArg(args, schedulingshardargs.FullHierarchyFairness, a.FullHierarchyFairness)
	setBoolArg(args, schedulingshardargs.AllowConsolidatingReclaim, a.AllowConsolidatingReclaim)
	setStringArg(args, schedulingshardargs.DefaultStalenessGracePeriod, a.DefaultStalenessGracePeriod)
	setIntArg(args, schedulingshardargs.NumOfStatusRecordingWorkers, a.NumOfStatusRecordingWorkers)
	return args
}

func setStringArg(args map[string]string, key string, value *string) {
	if value != nil {
		args[key] = *value
	}
}

func setIntArg(args map[string]string, key string, value *int) {
	if value != nil {
		args[key] = strconv.Itoa(*value)
	}
}

func setBoolArg(args map[string]string, key string, value *bool) {
	if value != nil {
		args[key] = strconv.FormatBool(*value)
	}
}
