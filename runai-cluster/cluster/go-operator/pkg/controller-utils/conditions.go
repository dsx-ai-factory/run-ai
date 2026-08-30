package controller_utils

import (
	"fmt"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func SetCondition(runaiCustomType runaiv1.RunaiCustomTypeI, conditionType runaiv1.ConditionType,
	status metav1.ConditionStatus,
	reason runaiv1.ConditionReason, message string, generation int64) {

	newCondition := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: generation,
	}

	key := func(condition metav1.Condition) string {
		return fmt.Sprintf("%s-%s-%s-%s-%d",
			string(condition.Type),
			string(condition.Status),
			string(condition.Reason),
			condition.Message,
			condition.ObservedGeneration)
	}

	for i := range runaiCustomType.GetRunaiCustomStatus().Conditions {
		if runaiCustomType.GetRunaiCustomStatus().Conditions[i].Type == string(conditionType) {
			if key(runaiCustomType.GetRunaiCustomStatus().Conditions[i]) != key(newCondition) {
				runaiCustomType.GetRunaiCustomStatus().Conditions[i] = newCondition
				runaiCustomType.GetRunaiCustomStatus().Conditions[i].LastTransitionTime = metav1.NewTime(time.Now())
			}

			return
		}
	}

	newConditions := append(runaiCustomType.GetRunaiCustomStatus().Conditions, metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		LastTransitionTime: metav1.NewTime(time.Now()),
		ObservedGeneration: generation,
	})

	runaiCustomType.GetRunaiCustomStatus().Conditions = newConditions
}
