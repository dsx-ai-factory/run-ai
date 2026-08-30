package pre_hooks

import (
	"context"
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// migrateManagedNodesConfig migrates the legacy managed-nodes-config
// (run.ai/v1alpha1) to its kai twin (kai.resources/v1alpha1), keeping the same name.
// It is a singleton, so there is nothing to page through and nothing owns it - unlike
// the nodepool migration there are no children to re-parent and so no rollback chain.
//
// It is idempotent: an existing kai twin short-circuits the copy.
func migrateManagedNodesConfig(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)

	oldCRDInstalled, err := hooks.OldManagedNodesConfigCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldCRDInstalled {
		logger.Info("No legacy managed-nodes-config CRD installed, nothing to migrate")
		return nil
	}

	old, found, err := hooks.GetLegacyManagedNodesConfig(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !found {
		logger.Info("No legacy managed-nodes-config found, nothing to migrate")
		return nil
	}

	var existing kaiv1alpha1.ManagedNodesConfig
	err = k8sClient.Get(ctx, types.NamespacedName{Name: old.Name}, &existing)
	if err == nil {
		logger.Info(fmt.Sprintf("kai managed-nodes-config %s already exists, skipping", old.Name))
		return clearLegacyManagedNodesConfigFinalizers(ctx, k8sClient, old)
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("get kai managed-nodes-config %s: %w", old.Name, err)
	}

	if err = k8sClient.Create(ctx, buildNewManagedNodesConfig(old)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create kai managed-nodes-config: %w", err)
	}
	logger.Info(fmt.Sprintf("migrated managed-nodes-config %s", old.Name))

	return clearLegacyManagedNodesConfigFinalizers(ctx, k8sClient, old)
}

// buildNewManagedNodesConfig maps the legacy managed-nodes-config onto its kai twin.
// The name is carried over unchanged - go-operator pins it for the deployed controller
// via --managed-nodes-config-name, so only the group may differ. The spec is
// field-identical between the two groups, so the inclusion criteria are copied straight
// across. Labels and annotations are preserved; the legacy finalizer is not (the kai
// object gets its own from its controller; copying it would wedge deletion while that
// controller is down).
func buildNewManagedNodesConfig(old *runv1alpha1.ManagedNodesConfig) *kaiv1alpha1.ManagedNodesConfig {
	return &kaiv1alpha1.ManagedNodesConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:        old.Name,
			Labels:      copyStringMap(old.Labels),
			Annotations: copyStringMap(old.Annotations),
		},
		Spec: kaiv1alpha1.ManagedNodesConfigSpec{
			InclusionCriteria: old.Spec.InclusionCriteria,
		},
	}
}

// clearLegacyManagedNodesConfigFinalizers strips the legacy object's finalizers so the
// post-upgrade cleanup can delete it: its controller is down for the migration window
// and would never clear them itself.
func clearLegacyManagedNodesConfigFinalizers(ctx context.Context, k8sClient client.Client,
	old *runv1alpha1.ManagedNodesConfig) error {
	if len(old.Finalizers) == 0 {
		return nil
	}

	log.FromContext(ctx).Info(fmt.Sprintf("clearing finalizers on legacy managed-nodes-config %s", old.Name),
		"finalizers", old.Finalizers)
	patch := client.MergeFrom(old.DeepCopy())
	old.Finalizers = nil
	if err := k8sClient.Patch(ctx, old, patch); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("clear finalizers on legacy managed-nodes-config %s: %w", old.Name, err)
	}
	return nil
}
