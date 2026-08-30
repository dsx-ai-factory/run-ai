package validation

import (
	"context"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

// validateProject runs all create/update validations for a Project:
//   - an empty parent is allowed; a non-empty parent must reference an existing
//     department,
//   - every node pool referenced by a queue must exist in the cluster (empty
//     queues are allowed),
//   - every default node pool must be backed by a queue in the same spec.
//
// All failures are aggregated so the caller sees every problem at once.
func (v *Validator) validateProject(ctx context.Context, project *kaiv1alpha1.Project) error {
	var errs error

	if err := v.validateParent(ctx, project.Spec.Parent); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := v.validateQueueNodePoolsExist(ctx, project.Spec.Queues); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := validateDefaultNodePoolsHaveQueues(project.Spec.DefaultNodePools, project.Spec.Queues); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs
}

// validateParent allows a project with no parent; otherwise the referenced
// parent department must exist.
func (v *Validator) validateParent(ctx context.Context, parent string) error {
	if parent == "" {
		return nil
	}

	department := &kaiv1alpha1.Department{}
	err := v.client.Get(ctx, types.NamespacedName{Name: parent}, department)
	if apierrors.IsNotFound(err) {
		return fmt.Errorf("parent department %q does not exist", parent)
	}
	if err != nil {
		return newInternalError(fmt.Errorf("failed to get parent department %q: %w", parent, err))
	}
	return nil
}

// validateDefaultNodePoolsHaveQueues ensures every node pool listed in the
// project's default node pools is referenced by at least one queue in the same
// spec. This is a spec-consistency check only; it performs no cluster lookup.
func validateDefaultNodePoolsHaveQueues(defaultNodePools []string, queues []kaiv1alpha1.QueueConfig) error {
	queueNodePools := make(map[string]struct{}, len(queues))
	for _, queue := range queues {
		queueNodePools[queue.Nodepool] = struct{}{}
	}

	var errs error
	for _, nodePool := range defaultNodePools {
		if _, ok := queueNodePools[nodePool]; !ok {
			errs = multierror.Append(errs, fmt.Errorf("the nodepool %q from the DefaultNodePools list has no queue defined in the spec", nodePool))
		}
	}
	return errs
}
