package ownership

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	// ValidationPath is the workload ownership validation endpoint.
	ValidationPath = "/validate-workload-ownership-protection"
	// MutationPath is the workload ownership update mutation endpoint.
	MutationPath = "/mutate-workload-ownership"
)

// NewSetup returns manager setup that registers the ownership endpoints.
func NewSetup() func(context.Context, ctrl.Manager) error {
	return func(_ context.Context, manager ctrl.Manager) error {
		server := manager.GetWebhookServer()
		server.Register(ValidationPath, &admission.Webhook{Handler: NewValidator(manager.GetClient(), manager.GetAPIReader())})
		server.Register(MutationPath, &admission.Webhook{Handler: NewUpdateMutator()})
		return nil
	}
}
