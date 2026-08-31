/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamiealquiza/tachymeter"
	"github.com/run-ai/runai/scale-tests/sdk/reports"
	workloadscale "github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale/aggregated_report/common/stages"
	"github.com/samber/lo"
)

type WorkloadAnalisysAggregator struct{}

func NewWorkloadAnalisysAggregator() *WorkloadAnalisysAggregator {
	return &WorkloadAnalisysAggregator{}
}

func (a *WorkloadAnalisysAggregator) Aggregate(analyzedWorkloads []workloadscale.WorkloadFinalRawDataResponse) (map[string]reports.PerformanceMetric, error) {
	workloadSyncToClusterMetricDetails, err := calculateStageMetricDetails(analyzedWorkloads, stages.WorkloadStageCreatedOnCluster)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate workload sync to cluster metric details: %w", err)
	}
	workloadFirstPodCreationMetricDetails, err := calculateStageMetricDetails(analyzedWorkloads, stages.WorkloadStageFirstPodCreated)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate workload first pod creation metric details: %w", err)
	}
	workloadPodSchedulingMetricDetails, err := calculateStageMetricDetails(analyzedWorkloads, stages.WorkloadStageFirstPodBound)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate workload pod scheduling metric details: %w", err)
	}
	workloadSyncToCpMetricDetails, err := calculateStageMetricDetails(analyzedWorkloads, stages.WorkloadStagePhaseChangeSynced)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate workload sync to cp metric details: %w", err)
	}

	workloadLifecycleMetricDetails, err := calculateStageMetricDetails(analyzedWorkloads, stages.WorkloadStageWholeLifecycle)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate workload lifecycle metric details: %w", err)
	}

	successFailureAnalysis := calculateSuccessFailureAnalysis(analyzedWorkloads)

	return map[string]reports.PerformanceMetric{
		"workload_sync_to_cluster":       workloadSyncToClusterMetricDetails,
		"workload_first_pod_creation":    workloadFirstPodCreationMetricDetails,
		"workload_pod_scheduling":        workloadPodSchedulingMetricDetails,
		"workload_sync_to_control_plane": workloadSyncToCpMetricDetails,
		"workload_whole_lifecycle":       workloadLifecycleMetricDetails,
		"success_failure_analysis":       successFailureAnalysis,
	}, nil
}

func calculateStageMetricDetails(analyzedWorkloads []workloadscale.WorkloadFinalRawDataResponse, stage stages.WorkloadStage) (*StagePerformanceMetric, error) {
	workloadsStageDetails := lo.FilterMap(analyzedWorkloads, func(workload workloadscale.WorkloadFinalRawDataResponse, _ int) (workloadscale.StageDetails, bool) {
		stageDetails, ok := workload.Stages[stage]
		return stageDetails, ok
	})

	if len(workloadsStageDetails) == 0 {
		return nil, fmt.Errorf("no details found for stage %s across all workloads", stage)
	}

	pendingCount := lo.Reduce(workloadsStageDetails, func(acc int, stageDetails workloadscale.StageDetails, _ int) int {
		if !stageDetails.Completed {
			return acc + 1
		}
		return acc
	}, 0)

	t := tachymeter.New(&tachymeter.Config{Size: len(workloadsStageDetails)})
	for _, stageDetails := range workloadsStageDetails {
		t.AddTime(stageDetails.Duration)
	}
	latencyMetrics := t.Calc()

	return &StagePerformanceMetric{
		PendingRate: float64(pendingCount) / float64(len(workloadsStageDetails)),
		LatencyMS: Latency{
			Mean:   float64(latencyMetrics.Time.Avg.Milliseconds()),
			Max:    float64(latencyMetrics.Time.Max.Milliseconds()),
			Min:    float64(latencyMetrics.Time.Min.Milliseconds()),
			P95:    float64(latencyMetrics.Time.P95.Milliseconds()),
			StdDev: float64(latencyMetrics.Time.StdDev.Milliseconds()),
			P50:    float64(latencyMetrics.Time.P50.Milliseconds()),
			P75:    float64(latencyMetrics.Time.P75.Milliseconds()),
		},
		SampleSize: len(workloadsStageDetails),
	}, nil
}

func calculateSuccessFailureAnalysis(analyzedWorkloads []workloadscale.WorkloadFinalRawDataResponse) *SuccessFailureAnalysisMetric {
	totalWorkloads := len(analyzedWorkloads)
	if totalWorkloads == 0 {
		return &SuccessFailureAnalysisMetric{
			TotalWorkloads:    0,
			SuccessCount:      0,
			FailureCount:      0,
			SuccessRate:       0,
			FailureBreakdown:  make(map[string]int),
			TopFailureReasons: []FailureReasonCount{},
			StageFailures:     make(map[string]StageFailureInfo),
		}
	}

	successCount, failureBreakdown, failureReasonsMap, stageFailures := analyzeWorkloadResults(analyzedWorkloads)

	failureCount := totalWorkloads - successCount
	successRate := (float64(successCount) / float64(totalWorkloads)) * 100

	topFailureReasons := extractTopFailureReasons(failureReasonsMap, 10)

	return &SuccessFailureAnalysisMetric{
		TotalWorkloads:    totalWorkloads,
		SuccessCount:      successCount,
		FailureCount:      failureCount,
		SuccessRate:       successRate,
		FailureBreakdown:  failureBreakdown,
		TopFailureReasons: topFailureReasons,
		StageFailures:     stageFailures,
	}
}

func analyzeWorkloadResults(analyzedWorkloads []workloadscale.WorkloadFinalRawDataResponse) (
	successCount int,
	failureBreakdown map[string]int,
	failureReasonsMap map[string]int,
	stageFailures map[string]StageFailureInfo,
) {
	failureBreakdown = make(map[string]int)
	failureReasonsMap = make(map[string]int)
	stageFailures = initializeStageFailures()

	for _, workload := range analyzedWorkloads {
		if workload.ReachedExpectedPhase {
			successCount++
		} else {
			if workload.FailureCategory != workloadscale.FailureCategoryNone {
				failureBreakdown[string(workload.FailureCategory)]++
			}

			if workload.FailureReason != "" {
				normalizedReason := strings.TrimSpace(strings.ToLower(workload.FailureReason))
				if normalizedReason != "" {
					failureReasonsMap[normalizedReason]++
				}
			}
		}

		updateStageFailures(workload, stageFailures)
	}

	return successCount, failureBreakdown, failureReasonsMap, stageFailures
}

func initializeStageFailures() map[string]StageFailureInfo {
	allStages := []stages.WorkloadStage{
		stages.WorkloadStageCreatedOnCluster,
		stages.WorkloadStageFirstPodCreated,
		stages.WorkloadStageFirstPodBound,
		stages.WorkloadStagePhaseChangeSynced,
		stages.WorkloadStageWholeLifecycle,
	}

	stageFailures := make(map[string]StageFailureInfo)
	for _, stage := range allStages {
		stageFailures[string(stage)] = StageFailureInfo{
			SuccessCount: 0,
			FailureCount: 0,
		}
	}
	return stageFailures
}

func updateStageFailures(workload workloadscale.WorkloadFinalRawDataResponse, stageFailures map[string]StageFailureInfo) {
	allStages := []stages.WorkloadStage{
		stages.WorkloadStageCreatedOnCluster,
		stages.WorkloadStageFirstPodCreated,
		stages.WorkloadStageFirstPodBound,
		stages.WorkloadStagePhaseChangeSynced,
		stages.WorkloadStageWholeLifecycle,
	}

	for _, stage := range allStages {
		stageKey := string(stage)
		stageInfo := stageFailures[stageKey]

		if stageDetails, ok := workload.Stages[stage]; ok {
			if stageDetails.Completed {
				stageInfo.SuccessCount++
			} else {
				stageInfo.FailureCount++
			}
		} else {
			stageInfo.FailureCount++
		}
		stageFailures[stageKey] = stageInfo
	}
}

func extractTopFailureReasons(failureReasonsMap map[string]int, topN int) []FailureReasonCount {
	topFailureReasons := make([]FailureReasonCount, 0, len(failureReasonsMap))
	for reason, count := range failureReasonsMap {
		topFailureReasons = append(topFailureReasons, FailureReasonCount{
			Reason: reason,
			Count:  count,
		})
	}

	sort.Slice(topFailureReasons, func(i, j int) bool {
		return topFailureReasons[i].Count > topFailureReasons[j].Count
	})

	if len(topFailureReasons) > topN {
		topFailureReasons = topFailureReasons[:topN]
	}

	return topFailureReasons
}
