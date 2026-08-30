package common

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	nodepoolutils "github.com/run-ai/runai/runai-cluster/common/node-pool-utils/utils"
)

const (
	PodByPodGroupIndexerName = "podByPodGroupIndexer"
)

const (
	NoSchedulingBackoff     = -1
	SingleSchedulingBackoff = 1

	DefaultMarkUnschedulable = true
	DefaultSchedulingBackoff = NoSchedulingBackoff
)

func GetMarkUnschedulableValue(markUnschedulable *bool) bool {
	if markUnschedulable == nil {
		return DefaultMarkUnschedulable
	}

	return *markUnschedulable
}

func GetSchedulingBackoffValue(schedulingBackoff *int32) int32 {
	if schedulingBackoff == nil {
		return DefaultSchedulingBackoff
	}

	return *schedulingBackoff
}

func IsSingleNodePool(schedulingBackoff *int32) bool {
	return GetSchedulingBackoffValue(schedulingBackoff) == NoSchedulingBackoff
}

func UpdateLabelsWithNodePoolAssignment(labels map[string]string, nodePoolName string, nodePoolLabelKey string) {
	if nodePoolName == config.Config().DefaultNodepoolName {
		delete(labels, nodePoolLabelKey)
	} else {
		labels[nodePoolLabelKey] = nodePoolName
	}
}

// IsPodGroupUpForScheduler is used in runai-scheduler
func IsPodGroupUpForScheduler(podGroup *kaiv2alpha2.PodGroup) bool {
	if GetSchedulingBackoffValue(podGroup.Spec.SchedulingBackoff) == NoSchedulingBackoff {
		return true
	}

	lastSchedulingCondition := GetLastSchedulingCondition(podGroup)
	if lastSchedulingCondition == nil {
		return true
	}

	currentNodePoolName := nodepoolutils.GetNodePoolNameFromLabels(podGroup.Labels, config.Config().NodePoolLabelKey, config.Config().DefaultNodepoolName)

	return lastSchedulingCondition.NodePool != currentNodePoolName
}

func GetLastSchedulingCondition(podGroup *kaiv2alpha2.PodGroup) *kaiv2alpha2.SchedulingCondition {
	schedulingConditions := podGroup.Status.SchedulingConditions
	if len(schedulingConditions) == 0 {
		return nil
	}

	return &podGroup.Status.SchedulingConditions[len(schedulingConditions)-1]
}
