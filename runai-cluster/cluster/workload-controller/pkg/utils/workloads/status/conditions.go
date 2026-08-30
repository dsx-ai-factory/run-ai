package status

import (
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Path: workload-controller/pkg/utils/workloads/conditions.go

func UpsertCondition(workloadConditions *[]common.WorkloadCondition, workloadConditionType common.WorkloadConditionType,
	status v1.ConditionStatus, reason string, message string) {
	if *workloadConditions == nil {
		*workloadConditions = []common.WorkloadCondition{}
	}
	found := false
	var newCondition common.WorkloadCondition
	newConditions := make([]common.WorkloadCondition, len(*workloadConditions))
	copy(newConditions, *workloadConditions)

	time := metav1.Now()
	for index := range newConditions {
		if newConditions[index].Type == workloadConditionType {
			newConditions[index].Message = message
			newConditions[index].Reason = reason
			if newConditions[index].Status != status {
				newConditions[index].LastTransitionTime = time
				newConditions[index].Status = status
			}
			found = true
			break
		}
	}

	if !found {
		// the condition does not exist yet
		newCondition = common.WorkloadCondition{
			Type:               workloadConditionType,
			Status:             status,
			LastTransitionTime: time,
			Message:            message,
			Reason:             reason,
		}
		newConditions = append(newConditions, newCondition)
	}
	*workloadConditions = newConditions
}

func UnsetCondition(workloadConditions *[]common.WorkloadCondition, workloadConditionType common.WorkloadConditionType) {
	if *workloadConditions == nil {
		return
	}
	newConditions := make([]common.WorkloadCondition, len(*workloadConditions))
	copy(newConditions, *workloadConditions)
	found := false
	for index := range newConditions {
		if newConditions[index].Type == workloadConditionType {
			newConditions[index] = newConditions[len(newConditions)-1]
			found = true
			break
		}
	}
	if found {
		*workloadConditions = newConditions[:len(newConditions)-1]
	}
}

func GetWorkloadCondition(workloadConditions *[]common.WorkloadCondition, conditionType common.WorkloadConditionType) *common.WorkloadCondition {
	for _, condition := range *workloadConditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func GetWorkloadConditionsAsMap(workloadConditions *[]common.WorkloadCondition) map[common.WorkloadConditionType]*common.WorkloadCondition {
	cm := make(map[common.WorkloadConditionType]*common.WorkloadCondition)
	for _, condition := range *workloadConditions {
		cm[condition.Type] = &condition
	}

	return cm
}

func GetRunaiCondition(runaiJobConditions []runaiJobV1.JobCondition, conditionType runaiJobV1.JobConditionType) *runaiJobV1.JobCondition {
	for _, condition := range runaiJobConditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil

}
