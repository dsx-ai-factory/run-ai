package legacycleanup

import (
	"context"
	"fmt"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateLegacyWebhookState(
	ctx context.Context,
	k8sClient client.Client,
	served []resourceDescriptor,
	ownershipEnabled bool,
) error {
	validating := &admissionv1.ValidatingWebhookConfigurationList{}
	if err := k8sClient.List(ctx, validating); err != nil {
		return fmt.Errorf("list ValidatingWebhookConfigurations: %w", err)
	}
	if err := validateValidatingWebhooksForOwnershipMode(validating.Items, ownershipEnabled); err != nil {
		return err
	}

	mutating := &admissionv1.MutatingWebhookConfigurationList{}
	if err := k8sClient.List(ctx, mutating); err != nil {
		return fmt.Errorf("list MutatingWebhookConfigurations: %w", err)
	}
	if err := validateMutatingWebhooksForOwnershipMode(mutating.Items, ownershipEnabled); err != nil {
		return err
	}
	if !ownershipEnabled {
		return nil
	}
	return validateOwnershipHandoff(served, validating.Items, mutating.Items)
}
