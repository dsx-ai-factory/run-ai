package run

import (
	"context"
	"testing"

	legacycleanup "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	cleanupoperand "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-workload-cleanup"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLegacyCleanupCreatesJobFromUpgradeAndLiveState(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupCreating))
	require.Len(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources(), 1)
}

func TestLegacyCleanupRestoresCompletionFromSuccessfulJob(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	setCleanupControllerFlags(config, false)
	reconciler := cleanupTestReconciler(t, config, cleanupJob(batchv1.JobComplete, "Completed", "cleanup complete"))

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionTrue, string(runaiv1.LegacyWorkloadCleanupCompleted))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupDoesNotRecreateDeletedSuccessfulJob(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	setCleanupControllerFlags(config, false)
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
