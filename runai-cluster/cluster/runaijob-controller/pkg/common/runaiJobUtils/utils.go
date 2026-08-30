package runaiJobUtils

import (
	runajob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	JobConditionMap map[runajob.JobConditionType]runajob.JobCondition
)

var (
	PodConditions = []runajob.JobConditionType{runajob.JobPending, runajob.JobInitializing, runajob.JobRunning}
)

func GenerateConditionMap(conditions *[]runajob.JobCondition) JobConditionMap {
	result := make(JobConditionMap)
	for _, condition := range *conditions {
		result[condition.Type] = condition
	}
	return result
}

func NewCondition(conditionType runajob.JobConditionType, status v1.ConditionStatus, reason, message string) runajob.JobCondition {
	return runajob.JobCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// EnsureJobConditionStatus appends or updates an existing job condition of the
// given type with the given status value. Note that this function will not
// append to the conditions list if the new condition's status is false
// (because going from nothing to false is meaningless); it can, however,
// update the status condition to false. The function returns a bool to let the
// caller know if the list was changed (either appended or updated).
func EnsureJobConditionStatus(list []runajob.JobCondition, cType runajob.JobConditionType, status v1.ConditionStatus, reason, message string) ([]runajob.JobCondition, bool) {
	for i := range list {
		if list[i].Type == cType {
			if list[i].Status != status || list[i].Reason != reason || list[i].Message != message {
				list[i].Status = status
				list[i].LastTransitionTime = metav1.Now()
				list[i].Reason = reason
				list[i].Message = message
				return list, true
			}
			return list, false
		}
	}
	// A condition with that type doesn't exist in the list.
	if status != v1.ConditionFalse {
		return append(list, NewCondition(cType, status, reason, message)), true
	}
	return list, false
}
