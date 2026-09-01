/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"testing"

	cleanupoperand "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-workload-cleanup"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLegacyCleanupCreatesJobWhenLegacyResourcesExist(t *testing.T) {
	config := cleanupTestConfig(true)
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace,
		Name:      "workload-controller",
	}}
	reconciler := cleanupTestReconciler(t, config, deployment)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupCreating))
	require.Len(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources(), 1)
}

func TestLegacyCleanupCreatesJobForRemainingLegacyService(t *testing.T) {
	config := cleanupTestConfig(true)
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Namespace: runaiv1.DefaultNamespace,
		Name:      "workload-controller",
	}}
	reconciler := cleanupTestReconciler(t, config, service)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupCreating))
}

func TestLegacyCleanupRestoresCompletionFromSuccessfulJob(t *testing.T) {
	config := cleanupTestConfig(true)
	reconciler := cleanupTestReconciler(t, config, cleanupJob(batchv1.JobComplete, "Completed", "cleanup complete"))

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionTrue, string(runaiv1.LegacyWorkloadCleanupCompleted))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupDoesNotRecreateDeletedSuccessfulJob(t *testing.T) {
	config := cleanupTestConfig(true)
	job := cleanupJob(batchv1.JobComplete, "Completed", "cleanup complete")
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionTrue, string(runaiv1.LegacyWorkloadCleanupCompleted))

	require.NoError(t, reconciler.localClusterClient.Delete(context.Background(), job))
	changed, err = reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.False(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionTrue, string(runaiv1.LegacyWorkloadCleanupCompleted))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func requireCleanupCondition(
	t *testing.T,
	config *runaiv1.RunaiConfig,
	status metav1.ConditionStatus,
	reason string,
) {
	t.Helper()
	condition := config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete)
	require.NotNil(t, condition)
	require.Equal(t, status, condition.Status)
	require.Equal(t, reason, condition.Reason)
}
