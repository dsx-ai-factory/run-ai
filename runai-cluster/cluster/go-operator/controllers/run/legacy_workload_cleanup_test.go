/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"testing"

	legacycleanup "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	cleanupoperand "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-workload-cleanup"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	schemeinit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLegacyCleanupFreshInstallDoesNothing(t *testing.T) {
	config := cleanupTestConfig(true)
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.False(t, changed)
	require.Nil(t, config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupWaitsForReadinessWhenResourcesExist(t *testing.T) {
	config := cleanupTestConfig(false)
	reconciler := cleanupTestReconciler(t, config, legacyDeployment())

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupWaitingForReady))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupReportsRunningJob(t *testing.T) {
	config := cleanupTestConfig(true)
	job := cleanupJob("", "", "")
	job.Status.Conditions = nil
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupRunning))
	require.Len(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources(), 1)
}

func TestLegacyCleanupRetriesFailedJobAfterDeletion(t *testing.T) {
	config := cleanupTestConfig(true)
	job := cleanupJob(batchv1.JobFailed, "BackoffLimitExceeded", "cleanup failed")
	reconciler := cleanupTestReconciler(t, config, job, legacyDeployment())

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupFailed))
	require.Contains(t, cleanupCondition(config).Message, "BackoffLimitExceeded")

	require.NoError(t, reconciler.localClusterClient.Delete(context.Background(), job))
	changed, err = reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupCreating))
}

func TestLegacyCleanupCompletesFromSuccessfulJob(t *testing.T) {
	config := cleanupTestConfig(true)
	job := cleanupJob(batchv1.JobComplete, "Completed", "cleanup complete")
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionTrue, string(runaiv1.LegacyWorkloadCleanupCompleted))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func cleanupTestReconciler(t *testing.T, objects ...client.Object) *RunAIConfigReconciler {
	t.Helper()
	k8sClient := clientfake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(objects...).Build()
	hybrid := kc.New(k8sClient, k8sClient, nil)
	return &RunAIConfigReconciler{
		runaiReconciler:     runaiReconciler{localClusterClient: hybrid, scheme: schemeinit.Scheme()},
		remoteClusterClient: hybrid,
	}
}

func cleanupTestConfig(ready bool) *runaiv1.RunaiConfig {
	config := &runaiv1.RunaiConfig{ObjectMeta: metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName, Generation: 7,
	}}
	if ready {
		config.Status.ObservedGeneration = config.Generation
		config.Status.Conditions = []metav1.Condition{
			{Type: string(runaiv1.ConditionTypeReconciled), Status: metav1.ConditionTrue, ObservedGeneration: config.Generation},
			{Type: string(runaiv1.ConditionTypeAvailable), Status: metav1.ConditionTrue, ObservedGeneration: config.Generation},
		}
	}
	return config
}

func legacyDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace, Name: "workload-controller",
	}}
}

func cleanupJob(conditionType batchv1.JobConditionType, reason, message string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: runaiv1.DefaultNamespace, Name: legacycleanup.CleanupJobName},
		Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
			Type: conditionType, Status: corev1.ConditionTrue, Reason: reason, Message: message,
		}}},
	}
}

func cleanupCondition(config *runaiv1.RunaiConfig) *metav1.Condition {
	return config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete)
}
