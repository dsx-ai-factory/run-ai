/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"fmt"
	"strings"

	legacycleanup "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	cu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups="",resources=services;serviceaccounts;secrets,verbs=get
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;clusterroles;clusterrolebindings,verbs=get
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list
//+kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get

func (r *RunAIConfigReconciler) reconcileLegacyWorkloadCleanup(
	ctx context.Context,
	config *runaiv1.RunaiConfig,
) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("RunaiConfig is missing")
	}
	if condition := config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete); condition != nil &&
		condition.Status == metav1.ConditionTrue {
		return false, nil
	}

	job := &batchv1.Job{}
	jobKey := client.ObjectKey{Namespace: runaiv1.DefaultNamespace, Name: legacycleanup.CleanupJobName}
	err := r.localClusterClient.Get(ctx, jobKey, job)
	if apierrors.IsNotFound(err) {
		resourcesPresent, presenceErr := legacycleanup.LegacyControllerResourcesPresent(ctx, r.remoteClusterClient)
		if presenceErr != nil {
			return false, presenceErr
		}
		if !resourcesPresent {
			return clearCleanupCondition(config), nil
		}
		if readinessErr := legacycleanup.ValidateReadinessMarker(config); readinessErr != nil {
			return setCleanupCondition(
				config, metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupWaitingForReady, readinessErr.Error(),
			), nil
		}
		return setCleanupCondition(
			config, metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupCreating,
			"Creating legacy workload cleanup Job",
		), nil
	}
	if err != nil {
		return false, fmt.Errorf("get cleanup Job %s: %w", jobKey, err)
	}

	if condition := trueJobCondition(job, batchv1.JobFailed); condition != nil {
		return setCleanupCondition(
			config, metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupFailed, jobConditionError(condition),
		), nil
	}
	if condition := trueJobCondition(job, batchv1.JobComplete); condition != nil {
		return setCleanupCondition(
			config, metav1.ConditionTrue, runaiv1.LegacyWorkloadCleanupCompleted,
			"Legacy workload controller cleanup is complete",
		), nil
	}

	return setCleanupCondition(
		config, metav1.ConditionFalse, runaiv1.LegacyWorkloadCleanupRunning,
		"Legacy workload cleanup Job is running",
	), nil
}

func setCleanupCondition(
	config *runaiv1.RunaiConfig,
	status metav1.ConditionStatus,
	reason runaiv1.ConditionReason,
	message string,
) bool {
	existing := config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete)
	if existing != nil && existing.Status == status && existing.Reason == string(reason) &&
		existing.Message == message && existing.ObservedGeneration == config.Generation {
		return false
	}
	cu.SetCondition(
		config, runaiv1.ConditionTypeLegacyWorkloadCleanupComplete, status, reason, message, config.Generation,
	)
	return true
}

func clearCleanupCondition(config *runaiv1.RunaiConfig) bool {
	conditions := config.Status.Conditions
	for i := range conditions {
		if conditions[i].Type != string(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete) {
			continue
		}
		config.Status.Conditions = append(conditions[:i], conditions[i+1:]...)
		return true
	}
	return false
}

func trueJobCondition(job *batchv1.Job, conditionType batchv1.JobConditionType) *batchv1.JobCondition {
	for i := range job.Status.Conditions {
		condition := &job.Status.Conditions[i]
		if condition.Type == conditionType && condition.Status == corev1.ConditionTrue {
			return condition
		}
	}
	return nil
}

func jobConditionError(condition *batchv1.JobCondition) string {
	parts := make([]string, 0, 2)
	if condition.Reason != "" {
		parts = append(parts, condition.Reason)
	}
	if condition.Message != "" {
		parts = append(parts, condition.Message)
	}
	if len(parts) == 0 {
		return "cleanup Job failed"
	}
	return strings.Join(parts, ": ")
}
