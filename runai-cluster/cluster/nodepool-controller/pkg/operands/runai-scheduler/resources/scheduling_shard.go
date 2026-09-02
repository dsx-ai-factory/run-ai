/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/operands"
	"maps"
	"time"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	usagedbapi "github.com/kai-scheduler/KAI-scheduler/pkg/scheduler/cache/usagedb/api"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/rs/zerolog/log"
	"github.com/xhit/go-str2duration/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/kai/schedulingshardargs"
)

const (
	defaultTimeBasedFairShareHistoricalUsageWeight = float32(1.0)
	defaultTimeBasedFairShareDecayHalfLife         = 0 * time.Second
	defaultTimeBasedFairShareWindowDuration        = 7 * 24 * time.Hour
	defaultTimeBasedFairShareWindowType            = usagedbapi.SlidingWindow

	cpuWorkerNodeLabelKey = "node-role.kubernetes.io/runai-cpu-worker"
	gpuWorkerNodeLabelKey = "node-role.kubernetes.io/runai-gpu-worker"
	migWorkerNodeLabelKey = "node-role.kubernetes.io/runai-mig-enabled"
)

func SchedulerBaseOperandName() string { return config.Get().SchedulerName }

func SchedulingShardForNodePool(
	ctx context.Context, k8sReader client.Reader, nodePool *v1alpha1.NodePool,
	params *common.NodePoolControllerParams, operandName string,
) (client.Object, error) {
	shard := &kaiv1.SchedulingShard{}
	if err := k8sReader.Get(ctx, types.NamespacedName{Name: operandName}, shard); err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	shard.Name = operandName
	if shard.Labels == nil {
		shard.Labels = map[string]string{}
	}
	shard.Labels["app"] = SchedulerBaseOperandName()
	cfg := nodePool.Spec.SchedulingShardConfig
	shard.Spec = kaiv1.SchedulingShardSpec{
		Args:                buildShardArgs(cfg, params),
		PlacementStrategy:   getPlacementStrategy(cfg),
		PartitionLabelValue: getNodePoolNameLabelValueForScheduler(nodePool.Name),
		QueueDepthPerAction: getQueueDepthPerAction(cfg),
		MinRuntime:          getMinRuntime(cfg),
		KValue:              getKValue(nodePool),
		UsageDBConfig:       getTimeBasedFairShareFields(nodePool),
		Plugins:             getPlugins(cfg),
		Actions:             getActions(cfg),
	}
	return shard, nil
}

func buildShardArgs(cfg *v1alpha1.SchedulingShardConfig, params *common.NodePoolControllerParams) map[string]string {
	args := map[string]string{}
	if cfg != nil {
		maps.Copy(args, cfg.Args)
	}
	maps.Copy(args, params.SchedulingShardArgs)
	args[schedulingshardargs.CPUWorkerNodeLabelKey] = cpuWorkerNodeLabelKey
	args[schedulingshardargs.GPUWorkerNodeLabelKey] = gpuWorkerNodeLabelKey
	args[schedulingshardargs.MIGWorkerNodeLabelKey] = migWorkerNodeLabelKey
	return args
}

func getPlacementStrategy(cfg *v1alpha1.SchedulingShardConfig) *kaiv1.PlacementStrategy {
	if cfg == nil {
		return nil
	}
	return cfg.PlacementStrategy
}

func getMinRuntime(cfg *v1alpha1.SchedulingShardConfig) *kaiv1.MinRuntime {
	if cfg == nil {
		return nil
	}
	return cfg.MinRuntime
}

func getPlugins(cfg *v1alpha1.SchedulingShardConfig) map[string]kaiv1.PluginConfig {
	if cfg == nil {
		return nil
	}
	return cfg.Plugins
}

func getActions(cfg *v1alpha1.SchedulingShardConfig) map[string]kaiv1.ActionConfig {
	if cfg == nil {
		return nil
	}
	return cfg.Actions
}

func getQueueDepthPerAction(cfg *v1alpha1.SchedulingShardConfig) map[string]int {
	if cfg == nil {
		return nil
	}
	return cfg.QueueDepthPerAction
}

func SchedulingShardStatus(
	ctx context.Context, k8sReader client.Reader, _ *v1alpha1.NodePool,
	_ *common.NodePoolControllerParams, operandName string,
) (operands.Status, error) {
	shard := &kaiv1.SchedulingShard{}
	err := k8sReader.Get(ctx, types.NamespacedName{Name: operandName}, shard)
	if err != nil && !errors.IsNotFound(err) {
		return operands.NotReadyStatus("scheduling shard is not deployed"), err
	}

	var deployed, available bool
	message := ""
	for _, condition := range shard.Status.Conditions {
		switch condition.Type {
		case string(kaiv1.ConditionTypeDeployed):
			deployed = condition.Status == metav1.ConditionTrue
			if !deployed {
				message = condition.Message
			}
		case string(kaiv1.ConditionTypeAvailable):
			available = condition.Status == metav1.ConditionTrue
			if !available && message == "" {
				message = condition.Message
			}
		}
	}
	if deployed && available {
		return operands.ReadyStatus(), nil
	}
	if message == "" {
		message = "no status message available"
	}
	return operands.NotReadyStatus(fmt.Sprintf("scheduler [%s] is not running yet: %s", operandName, message)), nil
}

func getNodePoolNameLabelValueForScheduler(nodePoolName string) string {
	if nodePoolName == config.Get().DefaultNodepoolName {
		return ""
	}
	return nodePoolName
}

func timeBasedFairShare(nodePool *v1alpha1.NodePool) *v1alpha1.TimeBasedFairShare {
	cfg := nodePool.Spec.SchedulingShardConfig
	if cfg == nil || cfg.TimeBasedFairShare == nil || !ptr.Deref(cfg.TimeBasedFairShare.Enabled, false) {
		return nil
	}
	return cfg.TimeBasedFairShare
}

func getKValue(nodePool *v1alpha1.NodePool) *float64 {
	tbfs := timeBasedFairShare(nodePool)
	if tbfs == nil {
		return nil
	}
	value := float64(ptr.Deref(tbfs.HistoricalUsageWeight, defaultTimeBasedFairShareHistoricalUsageWeight))
	return &value
}

func getTimeBasedFairShareFields(nodePool *v1alpha1.NodePool) *usagedbapi.UsageDBConfig {
	tbfs := timeBasedFairShare(nodePool)
	if tbfs == nil {
		return nil
	}

	var windowSize, windowType *string
	var tumblingStart *metav1.Time
	var cronString string
	if window := tbfs.Window; window != nil {
		windowSize = window.Size
		windowType = window.Type
		tumblingStart = window.TumblingStartTime
		cronString = ptr.Deref(window.CronString, "")
	}

	usageParams := &usagedbapi.UsageParams{
		HalfLifePeriod: parseDurationParam(tbfs.HalfLifePeriod, defaultTimeBasedFairShareDecayHalfLife, "HalfLifePeriod", nodePool.Name),
		WindowSize:     parseMonitoringDurationParam(windowSize, defaultTimeBasedFairShareWindowDuration, "Window.Size", nodePool.Name),
		WindowType:     ptr.To(defaultTimeBasedFairShareWindowType),
		CronString:     cronString,
	}
	if windowType != nil {
		if wt := usagedbapi.WindowType(*windowType); wt.IsValid() {
			usageParams.WindowType = ptr.To(wt)
		}
	}
	if tumblingStart != nil {
		usageParams.TumblingWindowStartTime = ptr.To(*tumblingStart)
	}
	if sampling := tbfs.Sampling; sampling != nil {
		usageParams.FetchInterval = parseOptionalDurationParam(sampling.FetchInterval, "Sampling.FetchInterval", nodePool.Name)
		usageParams.StalenessPeriod = parseOptionalDurationParam(sampling.StalenessPeriod, "Sampling.StalenessPeriod", nodePool.Name)
		usageParams.WaitTimeout = parseOptionalDurationParam(sampling.WaitTimeout, "Sampling.WaitTimeout", nodePool.Name)
	}

	result := &usagedbapi.UsageDBConfig{ClientType: "prometheus", UsageParams: usageParams}
	overrideCapacityMetricsIfNeeded(result.UsageParams, nodePool.Name)
	return result
}

func parseOptionalDurationParam(value *string, field, nodePool string) *metav1.Duration {
	if value == nil {
		return nil
	}
	if duration, err := str2duration.ParseDuration(*value); err == nil {
		return &metav1.Duration{Duration: duration}
	}
	log.Warn().Msgf("failed to parse TimeBasedFairShare.%s '%s' for NodePool '%s'", field, *value, nodePool)
	return nil
}

func parseDurationParam(value *string, fallback time.Duration, field, nodePool string) *metav1.Duration {
	if parsed := parseOptionalDurationParam(value, field, nodePool); parsed != nil {
		return parsed
	}
	return &metav1.Duration{Duration: fallback}
}

func parseMonitoringDurationParam(value *string, fallback time.Duration, field, nodePool string) *monitoringv1.Duration {
	if value != nil {
		if _, err := str2duration.ParseDuration(*value); err == nil {
			return ptr.To(monitoringv1.Duration(*value))
		}
		log.Warn().Msgf("failed to parse TimeBasedFairShare.%s '%s' for NodePool '%s'", field, *value, nodePool)
	}
	return ptr.To(monitoringv1.Duration(fallback.String()))
}

func overrideCapacityMetricsIfNeeded(params *usagedbapi.UsageParams, nodePoolName string) {
	if params.ExtraParams == nil {
		params.ExtraParams = map[string]string{}
	}
	prefix := config.Get().MetricsNamespace
	setDefaultExtraParam(params, "gpuAllocationMetric", fmt.Sprintf("%s_queue_allocated_gpus", prefix))
	setDefaultExtraParam(params, "cpuAllocationMetric", fmt.Sprintf("%s_queue_allocated_cpu_cores", prefix))
	setDefaultExtraParam(params, "memoryAllocationMetric", fmt.Sprintf("%s_queue_allocated_memory_bytes", prefix))
	setDefaultExtraParam(params, "gpuCapacityMetric", fmt.Sprintf("resources_per_nodepool{resource=\"nvidia_com_gpu\", nodepool=\"%s\"}", nodePoolName))
	setDefaultExtraParam(params, "cpuCapacityMetric", fmt.Sprintf("resources_per_nodepool{resource=\"cpu\", nodepool=\"%s\"}", nodePoolName))
	setDefaultExtraParam(params, "memoryCapacityMetric", fmt.Sprintf("resources_per_nodepool{resource=\"memory\", nodepool=\"%s\"}", nodePoolName))
}

func setDefaultExtraParam(params *usagedbapi.UsageParams, key, value string) {
	if params.ExtraParams[key] == "" {
		params.ExtraParams[key] = value
	}
}
