package migration

import (
	"context"
	"errors"

	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/spec_checksum"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ChecksumSHA256Migration rewrites any InferenceWorkload whose spec-checksum annotation
// is in the legacy MD5 format (32 hex chars) to the new SHA-256 format (64 hex chars).
// Running this at startup ensures the reconcile loop never sees a cross-algorithm mismatch,
// which would otherwise trigger a spurious Knative service rebuild for every existing workload.
type ChecksumSHA256Migration struct {
	k8sClient client.Client
}

func NewChecksumSHA256Migration(k8sClient client.Client) *ChecksumSHA256Migration {
	return &ChecksumSHA256Migration{k8sClient: k8sClient}
}

func (m *ChecksumSHA256Migration) String() string { return ChecksumSHA256MigrationName }

func (m *ChecksumSHA256Migration) RequiresWebhooks() bool { return false }

func (m *ChecksumSHA256Migration) Migrate(ctx context.Context) error {
	log.Info("Starting checksum SHA-256 migration")

	workloads := &infw.InferenceWorkloadList{}
	if err := m.k8sClient.List(ctx, workloads); err != nil {
		if meta.IsNoMatchError(err) {
			log.Info("Skipping checksum SHA-256 migration because the InferenceWorkload API is not served")
			return nil
		}
		return err
	}

	migratedCount := 0
	skippedCount := 0
	var err error
	for i := range workloads.Items {
		wl := &workloads.Items[i]

		current := spec_checksum.GetSpecChecksumAnnotation(wl)
		if !spec_checksum.IsLegacyMD5Checksum(current) {
			skippedCount++
			continue
		}

		newChecksum, er := spec_checksum.CalculateSpecChecksum(wl)
		if er != nil {
			err = errors.Join(err, er)
			log.Error(er, "failed to calculate SHA-256 checksum", "workload", wl.Name, "namespace", wl.Namespace)
			continue
		}

		patch := client.MergeFrom(wl.DeepCopy())
		if wl.Annotations == nil {
			wl.Annotations = make(map[string]string)
		}
		wl.Annotations[spec_checksum.SpecChecksumAnnotation] = newChecksum

		if patchErr := m.k8sClient.Patch(ctx, wl, patch); patchErr != nil {
			err = errors.Join(err, patchErr)
			log.Error(patchErr, "failed to patch checksum annotation", "workload", wl.Name, "namespace", wl.Namespace)
			continue
		}

		migratedCount++
		log.V(4).Info("Migrated checksum to SHA-256", "workload", wl.Name, "namespace", wl.Namespace)
	}

	log.Info("Checksum SHA-256 migration completed",
		"migrated", migratedCount,
		"skipped", skippedCount,
		"failed", len(workloads.Items)-migratedCount-skippedCount)

	if err != nil {
		return err
	}
	return nil
}
