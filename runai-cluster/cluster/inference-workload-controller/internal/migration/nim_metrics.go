package migration

import (
	"context"

	constants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NIMMetricsServiceMigration struct {
	k8sClient client.Client
}

func NewNIMMetricsServiceMigration(k8sClient client.Client) *NIMMetricsServiceMigration {
	return &NIMMetricsServiceMigration{
		k8sClient: k8sClient,
	}
}

func (m *NIMMetricsServiceMigration) String() string {
	return NIMMetricsServiceMigrationName
}

func (m *NIMMetricsServiceMigration) RequiresWebhooks() bool {
	return false
}

func (m *NIMMetricsServiceMigration) Migrate(ctx context.Context) error {
	log.Info("Starting NIM metrics service migration")

	workloads := &infv2.InferenceWorkloadList{}
	if err := m.k8sClient.List(ctx, workloads); err != nil {
		if meta.IsNoMatchError(err) {
			log.Info("Skipping NIM metrics service migration because InferenceWorkload API is not served")
			return nil
		}
		return err
	}

	patchedCount := 0
	skippedCount := 0

	for i := range workloads.Items {
		wl := &workloads.Items[i]

		if !isNIMWorkload(wl) {
			skippedCount++
			continue
		}

		if hasMetricsMigrationAnnotation(wl) {
			skippedCount++
			continue
		}

		if err := m.triggerReconciliation(ctx, wl); err != nil {
			log.Error(err, "Failed to trigger reconciliation", "workload", wl.Name, "namespace", wl.Namespace)
			continue
		}

		patchedCount++
		log.V(4).Info("Triggered reconciliation", "workload", wl.Name, "namespace", wl.Namespace)
	}

	log.Info("NIM metrics service migration completed",
		"patched", patchedCount,
		"skipped", skippedCount,
		"total", len(workloads.Items))

	return nil
}

func (m *NIMMetricsServiceMigration) triggerReconciliation(ctx context.Context, wl *infv2.InferenceWorkload) error {
	patch := client.MergeFrom(wl.DeepCopy())

	if wl.Spec.Annotations == nil {
		wl.Spec.Annotations = &field_types.DictionaryField{}
	}

	wl.Spec.Annotations.Put(NIMMetricsMigratedAnnotation, "true")

	return m.k8sClient.Patch(ctx, wl, patch)
}

func isNIMWorkload(wl *infv2.InferenceWorkload) bool {
	labels := wl.Spec.Labels.Deref()
	if val, exists := labels[constants.LabelNIMWorkload]; exists && val.Value != nil && *val.Value == constants.BoolTrueAsString {
		return true
	}
	return false
}

func hasMetricsMigrationAnnotation(wl *infv2.InferenceWorkload) bool {
	annotations := wl.Spec.Annotations.Deref()
	_, exists := annotations[NIMMetricsMigratedAnnotation]
	return exists
}
