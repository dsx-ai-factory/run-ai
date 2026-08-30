package run

import (
	"context"
	"testing"

	legacycleanup "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/legacy-workload-cleanup"
	cleanupoperand "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-workload-cleanup"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	schemeinit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	inferencecontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	jobcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	internalv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	workloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLegacyCleanupFreshInstallDoesNothing(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	setCleanupControllerFlags(config, false)
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.False(t, changed)
	require.Nil(t, config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupBlocksMissingUpgradeProvenance(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupBlocked))
	require.Contains(t, cleanupCondition(config).Message, "cleanup source version annotation is missing while legacy controllers are enabled")
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupDoesNotRunForPost3Source(t *testing.T) {
	for _, test := range []struct {
		name   string
		target string
		source string
	}{
		{name: "three to three", target: "3.1.0", source: "3.0.0"},
		{name: "branch build after three", target: "0.0.0-12345.v301-main", source: "3.0.0"},
		{name: "branch source is not pre-three evidence", target: "0.0.0-12346.v300-main", source: "0.0.0-12345.v300-main"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := cleanupTestConfig(test.target, true)
			config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: test.source}
			reconciler := cleanupTestReconciler(t, config)

			changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

			require.NoError(t, err)
			require.False(t, changed)
			require.Nil(t, config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete))
		})
	}
}

func TestLegacyCleanupReportsInvalidVersionAsBlocked(t *testing.T) {
	for _, test := range []struct {
		name      string
		target    string
		source    string
		wantError string
	}{
		{name: "source", target: "3.0.0", source: "not-a-version", wantError: "parse source version"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := cleanupTestConfig(test.target, true)
			config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: test.source}
			reconciler := cleanupTestReconciler(t, config)

			changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

			require.NoError(t, err)
			require.True(t, changed)
			requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupBlocked))
			require.Contains(t, cleanupCondition(config).Message, test.wantError)

			changed, err = reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
			require.NoError(t, err)
			require.False(t, changed)
		})
	}
}

func TestLegacyCleanupWaitsForTargetReadiness(t *testing.T) {
	config := cleanupTestConfig("3.0.0", false)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupWaitingForReady))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupRunsForBranchBuildTargetVersion(t *testing.T) {
	config := cleanupTestConfig("0.0.0-12345.v300-main", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupCreating))
	require.Len(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources(), 1)
}

func TestLegacyCleanupReportsRunningJob(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	job := cleanupJob("", "", "")
	job.Status.Conditions = nil
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupRunning))
	require.Len(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources(), 1)
}

func TestLegacyCleanupRetriesFailedJobOnlyAfterDeletion(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	job := cleanupJob(batchv1.JobFailed, "BackoffLimitExceeded", "cleanup failed")
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupFailed))
	require.Contains(t, cleanupCondition(config).Message, "BackoffLimitExceeded")

	changed, err = reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.False(t, changed)

	require.NoError(t, reconciler.localClusterClient.Delete(context.Background(), job))
	changed, err = reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)
	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupCreating))
}

func TestLegacyCleanupBlocksMixedControllerFlags(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	config.Spec.WorkloadController.Enabled = ptr.To(false)
	reconciler := cleanupTestReconciler(t, config)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupBlocked))
	require.Contains(t, cleanupCondition(config).Message, "not atomic")
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupCompletesFromSuccessfulJobAndDisabledControllers(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	setCleanupControllerFlags(config, false)
	job := cleanupJob(batchv1.JobComplete, "Completed", "cleanup complete")
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionTrue, string(runaiv1.LegacyWorkloadCleanupCompleted))
	require.Empty(t, cleanupoperand.Operand(reconciler.localClusterClient, config).Resources())
}

func TestLegacyCleanupRejectsCompletedJobWithEnabledControllers(t *testing.T) {
	config := cleanupTestConfig("3.0.0", true)
	config.Annotations = map[string]string{legacycleanup.SourceVersionAnnotation: "2.26.0"}
	job := cleanupJob(batchv1.JobComplete, "Completed", "cleanup complete")
	reconciler := cleanupTestReconciler(t, config, job)

	changed, err := reconciler.reconcileLegacyWorkloadCleanup(context.Background(), config)

	require.NoError(t, err)
	require.True(t, changed)
	requireCleanupCondition(t, config, metav1.ConditionFalse, string(runaiv1.LegacyWorkloadCleanupFailed))
	require.Contains(t, cleanupCondition(config).Message, "controllers remained enabled")
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

func cleanupTestConfig(target string, ready bool) *runaiv1.RunaiConfig {
	config := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  runaiv1.DefaultNamespace,
			Name:       runaiv1.DefaultName,
			Generation: 7,
		},
		Spec: runaiv1.RunaiConfigSpec{
			Internal:                    &internalv1.Internal{Global: &internalv1.Global{Version: target}},
			WorkloadController:          &workloadcontroller.WorkloadController{Enabled: ptr.To(true)},
			InferenceWorkloadController: &inferencecontroller.InferenceWorkloadController{Enabled: ptr.To(true)},
			JobController:               &jobcontroller.JobController{Enabled: ptr.To(true)},
		},
	}
	if ready {
		config.Status.ObservedGeneration = config.Generation
		config.Status.Conditions = []metav1.Condition{
			{Type: string(runaiv1.ConditionTypeReconciled), Status: metav1.ConditionTrue, ObservedGeneration: config.Generation},
			{Type: string(runaiv1.ConditionTypeAvailable), Status: metav1.ConditionTrue, ObservedGeneration: config.Generation},
		}
	}
	return config
}

func cleanupJob(conditionType batchv1.JobConditionType, reason, message string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: runaiv1.DefaultNamespace, Name: legacycleanup.CleanupJobName},
		Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
			Type: conditionType, Status: corev1.ConditionTrue, Reason: reason, Message: message,
		}}},
	}
}

func setCleanupControllerFlags(config *runaiv1.RunaiConfig, enabled bool) {
	config.Spec.WorkloadController.Enabled = ptr.To(enabled)
	config.Spec.InferenceWorkloadController.Enabled = ptr.To(enabled)
	config.Spec.JobController.Enabled = ptr.To(enabled)
}

func cleanupCondition(config *runaiv1.RunaiConfig) *metav1.Condition {
	return config.Status.GetCondition(runaiv1.ConditionTypeLegacyWorkloadCleanupComplete)
}
