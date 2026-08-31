/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package seed_cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	workload_scale "github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale/analyze"
	workloadclient "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkloadSamplesReportsGenerator struct {
	sampleGenerationCompletedWithSuccess bool
	analyzer                             *analyze.WorkloadAnalyzer
	workloadSampler                      *WorkloadSampler
}

func NewWorkloadSamplesReportsGenerator(sampler *WorkloadSampler,
	sampleGenerationCompletedWithSuccess bool) *WorkloadSamplesReportsGenerator {
	return &WorkloadSamplesReportsGenerator{
		workloadSampler:                      sampler,
		sampleGenerationCompletedWithSuccess: sampleGenerationCompletedWithSuccess,
		analyzer:                             analyze.NewWorkloadAnalyzer(),
	}
}
func (w *WorkloadSamplesReportsGenerator) GenerateRawDataReport() (*scale.SamplesReport, error) {
	data, err := w.collectWorkloadResults()
	if err != nil {
		return nil, fmt.Errorf("failed to collect workload results:[%v]", err)
	}
	rawReport := scale.NewSamplesReport(w.sampleGenerationCompletedWithSuccess, data)
	
	// Add workload creation timing metrics to metadata
	if rawReport.MetaData == nil {
		rawReport.MetaData = make(map[string]string)
	}
	rawReport.MetaData["workload_creation_duration_seconds"] = fmt.Sprintf("%.2f", w.workloadSampler.creationDuration.Seconds())
	rawReport.MetaData["workload_creation_throughput_per_sec"] = fmt.Sprintf("%.2f", w.workloadSampler.workloadsPerSecond)
	
	return rawReport, nil
}

func (w *WorkloadSamplesReportsGenerator) collectWorkloadResults() ([]*workload_scale.WorkloadFinalRawDataResponse, error) {
	res := []*workload_scale.WorkloadFinalRawDataResponse{}
	k8sWorkloadsData, err := w.getWorkloadClusterDataForAllProjects(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to pull k8s workload data:[%v]", err)
	}

	// Get CP workloads and create a map for efficient lookup
	cpWorkloadsMap := make(map[string]*workloadclient.Workload)
	filterString := fmt.Sprintf("projectName=@%s", w.workloadSampler.projects.GetProjectsPrefix())
	countWorkloadsRequest := w.workloadSampler.clients.WorkloadServiceClient.WorkloadsAPI.CountWorkloads(context.Background())
	countWorkloadsRequest = countWorkloadsRequest.FilterBy([]string{filterString})
	count, _, _ := countWorkloadsRequest.Execute()
	for i := 0; i < int(count.Count); i += w.workloadSampler.config.BatchSize {
		getWorkloadsRequest := w.workloadSampler.clients.WorkloadServiceClient.WorkloadsAPI.GetWorkloads(context.Background())
		getWorkloadsRequest = getWorkloadsRequest.FilterBy([]string{filterString})
		if i < int(count.Count) {
			getWorkloadsRequest = getWorkloadsRequest.Limit(int32(w.workloadSampler.config.BatchSize))
		}
		getWorkloadsRequest = getWorkloadsRequest.Offset(int32(i))
		workloads, _, _ := getWorkloadsRequest.Execute()
		for _, cpWorkload := range workloads.Workloads {
			cpWorkloadsMap[cpWorkload.Name] = &cpWorkload
		}
	}

	// Loop through all workloads in the sync.Map (includes creation failures)
	w.workloadSampler.workload.Range(func(key, value interface{}) bool {
		workloadSamplerEntity := value.(*workload_scale.WorkloadCreateSampleResponse)
		workloadIdentifier := w.getWorkloadUniqueIdentifier(workloadSamplerEntity.ProjectName, workloadSamplerEntity.Name)

		workloadFinalRawDataResponse := &workload_scale.WorkloadFinalRawDataResponse{
			WorkloadName: workloadSamplerEntity.Name,
			ProjectName:  workloadSamplerEntity.ProjectName,
			Phase:        workloadclient.PHASE_UNKNOWN,
		}

		// Check if workload was created successfully
		if !workloadSamplerEntity.SuccessAtCreation {
			// Workload failed at creation
			workloadFinalRawDataResponse.CreatedSuccessfully = false
			workloadFinalRawDataResponse.ReachedExpectedPhase = false
			workloadFinalRawDataResponse.FailureReason = workloadSamplerEntity.CreateFailMessage
			workloadFinalRawDataResponse.FailureCategory = workload_scale.FailureCategoryCreation
			workloadFinalRawDataResponse.WorkloadK8sData = workload_scale.WorkloadK8sData{
				ClusterPhase: workloadclient.PHASE_UNKNOWN,
			}
		} else {
			// Workload created successfully, check CP and K8s data
			workloadFinalRawDataResponse.CreatedSuccessfully = true
			workloadFinalRawDataResponse.CPCreatedAt = workloadSamplerEntity.CreateTime

			// Check if workload exists in CP
			if cpWorkload, ok := cpWorkloadsMap[workloadSamplerEntity.Name]; ok {
				phaseUpdatedAt := cpWorkload.PhaseUpdatedAt
				workloadFinalRawDataResponse.Phase = cpWorkload.Phase
				workloadFinalRawDataResponse.CPPhaseUpdatedAt = scale.GetTimeInUtc(&phaseUpdatedAt)
				workloadFinalRawDataResponse.ReachedExpectedPhase = cpWorkload.Phase == workloadSamplerEntity.ExpectedEndPhase

				// Categorize failure if not reached expected phase
				if !workloadFinalRawDataResponse.ReachedExpectedPhase {
					if cpWorkload.Phase == workloadclient.PHASE_FAILED {
						workloadFinalRawDataResponse.FailureCategory = workload_scale.FailureCategoryPhaseMismatch
						workloadFinalRawDataResponse.FailureReason = fmt.Sprintf("workload reached phase %s instead of expected phase %s", cpWorkload.Phase, workloadSamplerEntity.ExpectedEndPhase)
					} else {
						workloadFinalRawDataResponse.FailureCategory = workload_scale.FailureCategoryTimeout
						workloadFinalRawDataResponse.FailureReason = fmt.Sprintf("workload did not reach expected phase %s within timeout (current phase: %s)", workloadSamplerEntity.ExpectedEndPhase, cpWorkload.Phase)
					}
				}
			} else {
				// Workload created successfully but not found in CP
				workloadFinalRawDataResponse.ReachedExpectedPhase = false
				workloadFinalRawDataResponse.FailureCategory = workload_scale.FailureCategoryMissingData
				workloadFinalRawDataResponse.FailureReason = "workload created successfully but not found in control plane"
			}

			// Check K8s data
			if workloadK8sData, ok := k8sWorkloadsData[workloadIdentifier]; ok {
				workloadFinalRawDataResponse.WorkloadK8sData = *workloadK8sData
			} else {
				workloadFinalRawDataResponse.WorkloadK8sData = workload_scale.WorkloadK8sData{
					ClusterPhase: workloadclient.PHASE_UNKNOWN,
				}
				// If not in K8s and not already categorized as missing data
				if workloadFinalRawDataResponse.FailureCategory == workload_scale.FailureCategoryNone && !workloadFinalRawDataResponse.ReachedExpectedPhase {
					workloadFinalRawDataResponse.FailureCategory = workload_scale.FailureCategoryMissingData
					workloadFinalRawDataResponse.FailureReason = "workload not found in kubernetes cluster"
				}
			}
		}

		// Analyze workload stages (only for successfully created workloads)
		if workloadFinalRawDataResponse.CreatedSuccessfully {
			workloadFinalRawDataResponse = w.analyzer.Analyze(workloadFinalRawDataResponse)
		}

		res = append(res, workloadFinalRawDataResponse)
		return true
	})

	return res, nil
}

func (w *WorkloadSamplesReportsGenerator) getWorkloadClusterDataForAllProjects(ctx context.Context) (map[string]*workload_scale.WorkloadK8sData, error) {
	allWorkloadsData := map[string]*workload_scale.WorkloadK8sData{}
	pods := w.getPodMetrics()
	for _, project := range w.workloadSampler.projects.GetActiveProjects() {
		namespace := sampler_actions.ProjectToNs(project.Name)
		workloadClusterData, err := w.analyzeWorkloadK8sData(ctx, namespace)
		if err != nil {
			continue
		}
		for workloadName, workloadData := range workloadClusterData {
			workloadIdentifier := w.getWorkloadUniqueIdentifier(project.Name, workloadName)
			workloadData.Pods = pods[workloadName]
			allWorkloadsData[workloadIdentifier] = workloadData
		}
	}
	return allWorkloadsData, nil
}

func (w *WorkloadSamplesReportsGenerator) analyzeWorkloadK8sData(ctx context.Context, namespace string) (map[string]*workload_scale.WorkloadK8sData, error) {
	data := make(map[string]*workload_scale.WorkloadK8sData)
	trainingWorkloads, err := sampler_actions.GetK8sTrainingWorkloadFromCluster(ctx, namespace, w.workloadSampler.clients.KubeClient)
	if err != nil {
		return nil, err
	}
	distributedWorkloads, err := sampler_actions.GetK8sDistributedWorkloadFromCluster(ctx, namespace, w.workloadSampler.clients.KubeClient)
	if err != nil {
		return nil, err
	}

	type commonWorkloadData struct {
		CreationTimestamp metav1.Time
		Phase             workloadclient.Phase
		Conditions        []workloadcommon.WorkloadCondition
	}
	workloads := make(map[string]commonWorkloadData)
	for wName, wData := range trainingWorkloads {
		workloads[wName] = commonWorkloadData{
			CreationTimestamp: wData.CreationTimestamp,
			Phase:             workloadclient.Phase(wData.Status.Phase),
			Conditions:        wData.Status.Conditions,
		}
	}
	for wName, wData := range distributedWorkloads {
		workloads[wName] = commonWorkloadData{
			CreationTimestamp: wData.CreationTimestamp,
			Phase:             workloadclient.Phase(wData.Status.Phase),
			Conditions:        wData.Status.Conditions,
		}
	}

	for workloadName, workloadData := range workloads {
		clusterPhaseTime := extractWorkloadPhaseTransitionTime(workloadData.Phase, workloadData.Conditions)
		workloadCreationTimeInCluster := workloadData.CreationTimestamp
		workloadK8sData := &workload_scale.WorkloadK8sData{
			WorkloadCreationTime: scale.GetTimeInUtc(&workloadCreationTimeInCluster.Time),
			ClusterPhase:         workloadData.Phase,
			ClusterPhaseTime:     scale.GetTimeInUtc(clusterPhaseTime),
		}
		if workloadK8sData.ClusterPhase == "" {
			workloadK8sData.ClusterPhase = workloadclient.PHASE_UNKNOWN
		}
		data[workloadName] = workloadK8sData
	}
	return data, nil
}

func extractWorkloadPhaseTransitionTime(phase workloadclient.Phase, conditions []workloadcommon.WorkloadCondition) *time.Time {
	conditionMap := map[workloadcommon.WorkloadConditionType]*workloadcommon.WorkloadCondition{}
	for _, condition := range conditions {
		if condition.Status == "True" {
			conditionMap[condition.Type] = &condition
		}
	}

	if phase == workloadclient.PHASE_PENDING {
		if pending, ok := conditionMap[workloadcommon.Pending]; ok {
			return &pending.LastTransitionTime.Time
		}
	}

	if phase == workloadclient.PHASE_RUNNING {
		if running, ok := conditionMap[workloadcommon.Ready]; ok {
			if initialized, ok := conditionMap[workloadcommon.Initialized]; ok {
				return sampler_actions.MaxTime(&running.LastTransitionTime.Time, &initialized.LastTransitionTime.Time)

			}
		}
	}
	if phase == workloadclient.PHASE_COMPLETED {
		if completed, ok := conditionMap[workloadcommon.Succeeded]; ok {
			return &completed.LastTransitionTime.Time
		}
	}
	if phase == workloadclient.PHASE_FAILED {
		if failed, ok := conditionMap[workloadcommon.Failed]; ok {
			return &failed.LastTransitionTime.Time
		}
	}
	return nil
}

func (w *WorkloadSamplesReportsGenerator) getWorkloadUniqueIdentifier(projectName, workloadName string) string {
	return projectName + "_" + workloadName
}

func (w *WorkloadSamplesReportsGenerator) getPodMetrics() map[string][]workload_scale.PodK8sData {
	data := make(map[string][]workload_scale.PodK8sData)
	for _, pod := range w.workloadSampler.Pods {
		workloadName := pod.WorkloadName
		podSlice, ok := data[workloadName]
		if !ok {
			podSlice = []workload_scale.PodK8sData{}

		}
		podSlice = append(podSlice, *pod)
		data[workloadName] = podSlice
	}
	return data
}
