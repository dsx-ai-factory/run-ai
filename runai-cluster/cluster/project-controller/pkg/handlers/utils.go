/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// GetProjectCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetProjectCondition(status *kaiv1alpha1.ProjectStatus, conditionType kaiv1alpha1.ProjectConditionType) (
	int, *kaiv1alpha1.ProjectCondition) {
	if status == nil {
		return -1, nil
	}
	return GetProjectConditionFromList(status.Conditions, conditionType)
}

// GetProjectConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetProjectConditionFromList(conditions []kaiv1alpha1.ProjectCondition, conditionType kaiv1alpha1.ProjectConditionType) (
	int, *kaiv1alpha1.ProjectCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

func UpdateProjectConditions(status *kaiv1alpha1.ProjectStatus, conditions []kaiv1alpha1.ProjectCondition) bool {
	hasChanged := false
	for i := range conditions {
		condition := conditions[i]
		hasChanged = UpdateProjectCondition(status, &condition) || hasChanged
	}
	return hasChanged
}

// UpdateProjectCondition updates existing project condition or creates a new one.
// Sets LastTransitionTime to now if the status has changed.
// Returns true if project condition has changed or has been added.
// -> copied from k8s.io/kubernetes/pkg/api/v1/pod/utils.go:UpdatePodCondition
func UpdateProjectCondition(status *kaiv1alpha1.ProjectStatus, condition *kaiv1alpha1.ProjectCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this project condition.
	conditionIndex, oldCondition := GetProjectCondition(status, condition.Type)

	if oldCondition == nil {
		// We are adding new pod condition.
		status.Conditions = append(status.Conditions, *condition)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastProbeTime.Equal(&oldCondition.LastProbeTime) &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Conditions[conditionIndex] = *condition
	// Return true if one of the fields have changed.
	return !isEqual
}

// GetDepartmentCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetDepartmentCondition(status *kaiv1alpha1.DepartmentStatus, conditionType kaiv1alpha1.DepartmentConditionType) (
	int, *kaiv1alpha1.DepartmentCondition) {
	if status == nil {
		return -1, nil
	}
	return GetDepartmentConditionFromList(status.Conditions, conditionType)
}

// GetDepartmentConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetDepartmentConditionFromList(conditions []kaiv1alpha1.DepartmentCondition, conditionType kaiv1alpha1.DepartmentConditionType) (
	int, *kaiv1alpha1.DepartmentCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

// UpdateDepartmentCondition updates existing department condition or creates a new one.
// Sets LastTransitionTime to now if the status has changed.
// Returns true if department condition has changed or has been added.
// -> copied from k8s.io/kubernetes/pkg/api/v1/pod/utils.go:UpdatePodCondition
func UpdateDepartmentCondition(status *kaiv1alpha1.DepartmentStatus, condition *kaiv1alpha1.DepartmentCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this department condition.
	conditionIndex, oldCondition := GetDepartmentCondition(status, condition.Type)

	if oldCondition == nil {
		// We are adding new department condition.
		status.Conditions = append(status.Conditions, *condition)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastProbeTime.Equal(&oldCondition.LastProbeTime) &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Conditions[conditionIndex] = *condition
	// Return true if one of the fields have changed.
	return !isEqual
}

func getReasonFromError(err error, reason ProjectConditionReason) string {
	if err == nil {
		return ""
	}
	return string(reason)
}

func GetStatusFromError(err error) v1.ConditionStatus {
	if err == nil {
		return v1.ConditionTrue
	}
	return v1.ConditionFalse
}

func GetMessageFromError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func GetObjectNodePoolFromLabels(labels map[string]string) string {
	if nodePoolName, found := labels[config.Get().NodePoolLabelKey]; found {
		return nodePoolName
	}
	return config.Get().DefaultNodepoolName
}

func CreateLabelSelector(key string, operator selection.Operator, values []string) (labels.Selector, error) {
	return AddRequirementToLabelSelector(nil, key, operator, values)
}

func AddRequirementToLabelSelector(selector labels.Selector, key string, operator selection.Operator,
	values []string) (labels.Selector, error) {
	req, err := labels.NewRequirement(key, operator, values)
	if err != nil {
		return nil, err
	}
	if selector == nil {
		selector = labels.NewSelector()
	}
	selector = selector.Add(*req)
	return selector, nil
}

// ShouldPatchAnnotations check in the patch annotations has an update to the annotations
func ShouldPatchAnnotations(annotations map[string]string, patch map[string]interface{}) bool {
	for key, patchValue := range patch {
		curValue, ok := annotations[key]
		// Clearing annotation (patch value is nil) and there is annotation to clear (ok is true)
		if patchValue == nil && ok {
			return true
		}
		// There is a value to patch, and no annotation
		if patchValue != nil && !ok {
			return true
		}
		// If the current annotation is different than the patch
		if patchValue != nil && ok && patchValue != curValue {
			return true
		}
	}
	return false
}
