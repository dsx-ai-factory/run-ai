/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KartaConditionStatuses struct {
	RBACReady   metav1.ConditionStatus
	SyncEnabled metav1.ConditionStatus
}

type KartaDeletionConditionStatuses struct {
	SyncDisabled        metav1.ConditionStatus
	NoWorkloadsUseKarta metav1.ConditionStatus
	RBACCleanupSuccess  metav1.ConditionStatus
}

// upsertKartaConditions updates/appends the given conditions and preserves any not in the
// map, so EWI does not delete the conditions the OSS controller owns (KartaValidated,
// CRDExists, Ready) while writing its own.
func upsertKartaConditions(conditions *[]metav1.Condition, newConditions map[runv1alpha1.KartaConditionType]metav1.Condition) {
	if conditions == nil {
		return
	}

	var result []metav1.Condition

	now := metav1.Now()
	// forloop on existing conditions - to preserve the order
	for _, condition := range *conditions {
		newCondition, exists := newConditions[runv1alpha1.KartaConditionType(condition.Type)]
		if !exists {
			result = append(result, condition)
			continue
		}

		if condition.Status != newCondition.Status {
			newCondition.LastTransitionTime = now
		} else {
			newCondition.LastTransitionTime = condition.LastTransitionTime
		}

		result = append(result, newCondition)
		delete(newConditions, runv1alpha1.KartaConditionType(condition.Type))
	}

	// If there are any new conditions that did not exist, append them
	for _, newCondition := range newConditions {
		newCondition.LastTransitionTime = now
		result = append(result, newCondition)
	}

	*conditions = result
}

func SetConditions(status *kartav1alpha1.KartaStatus, conditionStatuses KartaConditionStatuses) {
	// Ready is owned by the OSS controller; read it to derive RunaiReady.
	ossReady := metav1.ConditionUnknown
	for _, c := range status.Conditions {
		if c.Type == string(runv1alpha1.Ready) {
			ossReady = c.Status
			break
		}
	}

	newConditions := map[runv1alpha1.KartaConditionType]metav1.Condition{
		runv1alpha1.RBACReady:   buildCondition(runv1alpha1.RBACReady, conditionStatuses.RBACReady, "The RBAC permissions for the group/kind are not ready"),
		runv1alpha1.SyncEnabled: buildCondition(runv1alpha1.SyncEnabled, conditionStatuses.SyncEnabled, "The sync is not enabled"),
		runv1alpha1.RunaiReady:  buildRunaiReadyCondition(ossReady, conditionStatuses.RBACReady),
	}

	upsertKartaConditions(&status.Conditions, newConditions)
}

// buildRunaiReadyCondition is True only when both Ready and RBACReady are True.
func buildRunaiReadyCondition(ready, rbacReady metav1.ConditionStatus) metav1.Condition {
	condition := metav1.Condition{
		Type:   string(runv1alpha1.RunaiReady),
		Status: metav1.ConditionFalse,
		Reason: string(runv1alpha1.RunaiReady),
	}

	switch {
	case ready != metav1.ConditionTrue:
		condition.Reason = "ReadyNotMet"
		condition.Message = "The Karta is not runai-ready because the Ready condition is not True"
	case rbacReady != metav1.ConditionTrue:
		condition.Reason = "RBACNotReady"
		condition.Message = "The Karta is not runai-ready because the RBACReady condition is not True"
	default:
		condition.Status = metav1.ConditionTrue
	}

	return condition
}

func setDeletionConditions(status *kartav1alpha1.KartaStatus, conditionStatuses KartaDeletionConditionStatuses) {
	newDeletionConditions := map[runv1alpha1.KartaConditionType]metav1.Condition{
		runv1alpha1.SyncDisabled:        buildCondition(runv1alpha1.SyncDisabled, conditionStatuses.SyncDisabled, "Sync is still enabled"),
		runv1alpha1.NoWorkloadsUseKarta: buildCondition(runv1alpha1.NoWorkloadsUseKarta, conditionStatuses.NoWorkloadsUseKarta, "There are still ExternalWorkloads using this Karta"),
		runv1alpha1.RBACCleanupSuccess:  buildCondition(runv1alpha1.RBACCleanupSuccess, conditionStatuses.RBACCleanupSuccess, "RBAC cleanup failed"),
		runv1alpha1.RunaiReady:          buildCondition(runv1alpha1.RunaiReady, metav1.ConditionFalse, "The Karta is being deleted"),
	}

	upsertKartaConditions(&status.Conditions, newDeletionConditions)
}

func WasStatusUpdated(originalStatus, newStatus *kartav1alpha1.KartaStatus) bool {
	if len(originalStatus.Conditions) != len(newStatus.Conditions) {
		return true
	}

	for i := range originalStatus.Conditions {
		if originalStatus.Conditions[i].Type != newStatus.Conditions[i].Type ||
			originalStatus.Conditions[i].Status != newStatus.Conditions[i].Status ||
			originalStatus.Conditions[i].Reason != newStatus.Conditions[i].Reason ||
			originalStatus.Conditions[i].Message != newStatus.Conditions[i].Message {
			return true
		}
	}

	return false
}

func buildCondition(condType runv1alpha1.KartaConditionType, condStatus metav1.ConditionStatus, errorMsg string) metav1.Condition {
	condition := metav1.Condition{
		Type:   string(condType),
		Status: condStatus,
		Reason: string(condType),
	}

	if condStatus != metav1.ConditionTrue {
		condition.Message = errorMsg
	}

	return condition
}
