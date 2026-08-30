package restart

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

// eventSource is the event-recorder source
const eventSource = "anyworkload-controller"

// MetadataUIDField is the cache index key for object UID (metadata.uid).
const MetadataUIDField = "metadata.uid"

// NewSetup returns the dromos manager setup func that registers all restart components
func NewSetup(maxConcurrentReconciles int, childDeletionTimeout time.Duration) func(ctx context.Context, mgr ctrl.Manager) error {
	return func(ctx context.Context, mgr ctrl.Manager) error {
		// the restart reconciler resolves the run.ai/request-ref label by listing RestartRequests with
		// client.MatchingFields, which fails against the cached client without this index
		err := mgr.GetFieldIndexer().IndexField(
			ctx,
			&runv1alpha1.RestartRequest{},
			MetadataUIDField,
			func(obj client.Object) []string {
				return []string{string(obj.GetUID())}
			})
		if err != nil {
			return fmt.Errorf("index restart requests by %s: %w", MetadataUIDField, err)
		}

		//nolint:staticcheck // SA1019: repo-wide convention is record.EventRecorder; migrate all controllers together
		recorder := mgr.GetEventRecorderFor(eventSource)
		if err := NewHandoffReconciler(mgr.GetClient(), recorder).
			SetupWithManager(mgr, maxConcurrentReconciles); err != nil {
			return fmt.Errorf("setup restart handoff reconciler: %w", err)
		}

		if err := NewRestartReconciler(mgr.GetClient(), childDeletionTimeout).
			SetupWithManager(mgr, maxConcurrentReconciles); err != nil {
			return fmt.Errorf("setup restart reconciler: %w", err)
		}

		// registered under the restart gate on purpose: with restart off the path is unserved and the
		// unconditional fail-closed runai-restartrequest-validation webhook rejects every RestartRequest write
		mgr.GetWebhookServer().Register(RestartRequestValidatorPath,
			&webhook.Admission{Handler: NewRestartRequestValidator(mgr.GetClient())})
		return nil
	}
}
