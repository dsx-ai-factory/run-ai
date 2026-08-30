package analyze

import (
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale/aggregated_report/common/stages"
	workloadclient "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"sort"
	"time"
)

type WorkloadAnalyzer struct {
}

func NewWorkloadAnalyzer() *WorkloadAnalyzer {
	return &WorkloadAnalyzer{}
}

func (a *WorkloadAnalyzer) Analyze(wd *workload_scale.WorkloadFinalRawDataResponse) *workload_scale.WorkloadFinalRawDataResponse {
	firstPod := getFirstPod(wd.WorkloadK8sData.Pods)
	stages := analyzeStages(wd, firstPod)
	wd.Stages = stages
	return wd

}

func analyzeStages(wd *workload_scale.WorkloadFinalRawDataResponse, firstPod *workload_scale.PodK8sData) map[stages.WorkloadStage]workload_scale.StageDetails {
	ret := map[stages.WorkloadStage]workload_scale.StageDetails{}

	createdOnClusterReadinessCondition := wd.ReachedExpectedPhase && wd.CPCreatedAt != nil && isAfterOrEqual(wd.CPCreatedAt, wd.WorkloadK8sData.WorkloadCreationTime) && wd.WorkloadK8sData.WorkloadCreationTime != nil
	ret[stages.WorkloadStageCreatedOnCluster] = analyzeStage(wd.CPCreatedAt, wd.WorkloadK8sData.WorkloadCreationTime, createdOnClusterReadinessCondition)

	firstPodCreatedReadinessCondition := wd.ReachedExpectedPhase && wd.WorkloadK8sData.WorkloadCreationTime != nil && isAfterOrEqual(wd.WorkloadK8sData.WorkloadCreationTime, firstPod.CreationTime) && firstPod.CreationTime != nil
	ret[stages.WorkloadStageFirstPodCreated] = analyzeStage(wd.WorkloadK8sData.WorkloadCreationTime, firstPod.CreationTime, firstPodCreatedReadinessCondition)

	firstPodCreatedBoundCondition := wd.ReachedExpectedPhase && firstPod.CreationTime != nil && isAfterOrEqual(firstPod.CreationTime, firstPod.ScheduledTime) && firstPod.ScheduledTime != nil
	ret[stages.WorkloadStageFirstPodBound] = analyzeStage(firstPod.CreationTime, firstPod.ScheduledTime, firstPodCreatedBoundCondition)

	syncedToClusterReadinessCondition := wd.ReachedExpectedPhase && wd.WorkloadK8sData.ClusterPhaseTime != nil && wd.CPPhaseUpdatedAt != nil && isAfterOrEqual(wd.WorkloadK8sData.ClusterPhaseTime, wd.CPPhaseUpdatedAt) && wd.Phase == wd.WorkloadK8sData.ClusterPhase
	ret[stages.WorkloadStagePhaseChangeSynced] = analyzeStage(wd.WorkloadK8sData.ClusterPhaseTime, wd.CPPhaseUpdatedAt, syncedToClusterReadinessCondition)

	lifeCycleReadinessCondition := wd.ReachedExpectedPhase && isPhaseTerminal(wd.Phase) && wd.CPPhaseUpdatedAt != nil && wd.CPCreatedAt != nil && isAfterOrEqual(wd.CPCreatedAt, wd.CPPhaseUpdatedAt)
	ret[stages.WorkloadStageWholeLifecycle] = analyzeStage(wd.CPCreatedAt, wd.CPPhaseUpdatedAt, lifeCycleReadinessCondition)

	return ret
}

func isAfterOrEqual(from, to *int64) bool {
	return from != nil && to != nil && *to >= *from
}

func analyzeStage(from, to *int64, completed bool) workload_scale.StageDetails {
	if !completed {
		return workload_scale.StageDetails{
			Duration:  0,
			Completed: false,
		}
	}
	fromTime := time.Unix(0, *from).UTC()
	toTime := time.Unix(0, *to).UTC()
	return workload_scale.StageDetails{
		Duration:  toTime.Sub(fromTime),
		Completed: true,
	}
}

func getFirstPod(pods []workload_scale.PodK8sData) *workload_scale.PodK8sData {
	if len(pods) == 0 {
		// Returning a pod with no creation or scheduled time, representing a non-existent pod
		return &workload_scale.PodK8sData{}
	}

	sortedByCreationTimePods := sortPodsByCreateTime(pods)
	sortedBySchedulingTimePods := sortPodsBySchedulingTime(pods)
	firstPodStatistics := &workload_scale.PodK8sData{}
	if sortedByCreationTimePods[0].CreationTime != nil {
		firstPodStatistics.CreationTime = sortedByCreationTimePods[0].CreationTime
	}
	if sortedBySchedulingTimePods[0].ScheduledTime != nil {
		firstPodStatistics.ScheduledTime = sortedBySchedulingTimePods[0].ScheduledTime
	}
	return firstPodStatistics
}

// Sorts the pods by creation time
func sortPodsByCreateTime(pods []workload_scale.PodK8sData) []workload_scale.PodK8sData {
	sort.SliceStable(pods, func(i, j int) bool {
		if pods[i].CreationTime == nil {
			return false
		}
		if pods[j].CreationTime == nil {
			return true
		}
		return (*pods[i].CreationTime) < (*pods[j].CreationTime)
	})
	return pods
}

// Sorts the pods by creation time
func sortPodsBySchedulingTime(pods []workload_scale.PodK8sData) []workload_scale.PodK8sData {
	sort.SliceStable(pods, func(i, j int) bool {
		if pods[i].ScheduledTime == nil {
			return false
		}
		if pods[j].ScheduledTime == nil {
			return true
		}
		return (*pods[i].ScheduledTime) < (*pods[j].ScheduledTime)
	})
	return pods
}

func isPhaseTerminal(phase workloadclient.Phase) bool {
	terminalPhases := []workloadclient.Phase{workloadclient.PHASE_COMPLETED, workloadclient.PHASE_FAILED}

	for _, terminalPhase := range terminalPhases {
		if phase == terminalPhase {
			return true
		}
	}
	return false
}
