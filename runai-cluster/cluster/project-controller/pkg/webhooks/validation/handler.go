package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	multierror "github.com/hashicorp/go-multierror"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ProjectWebhookPath is the path the validating handler is served on for
	// Project resources. The webhook configuration that routes admission
	// requests here is provisioned by the deployment's helm chart.
	ProjectWebhookPath = "/validate-project"

	// DepartmentWebhookPath is the path the validating handler is served on for
	// Department resources.
	DepartmentWebhookPath = "/validate-department"

	projectKind    = "Project"
	departmentKind = "Department"
)

// Validator is a validating admission handler for Project and Department
// resources. The same instance is registered on both webhook paths; Handle
// branches on the incoming resource kind.
type Validator struct {
	client client.Client
}

// NewValidator builds a Validator backed by the given client, used to look up
// referenced resources (parent departments, node pools) during validation.
func NewValidator(client client.Client) *Validator {
	return &Validator{client: client}
}

// Handle decodes the incoming object by kind and runs the matching validation.
// A validation failure is returned as a denied (not errored) response so the
// admission message reaches the caller verbatim.
func (v *Validator) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := logf.FromContext(ctx)

	switch req.Kind.Kind {
	case projectKind:
		project := &kaiv1alpha1.Project{}
		if err := json.Unmarshal(req.Object.Raw, project); err != nil {
			logger.Error(err, "failed to unmarshal project")
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal project: %w", err))
		}
		if err := v.validateProject(ctx, project); err != nil {
			logger.Error(err, "project validation failed", "name", project.Name)
			return toAdmissionResponse(err)
		}
	case departmentKind:
		department := &kaiv1alpha1.Department{}
		if err := json.Unmarshal(req.Object.Raw, department); err != nil {
			logger.Error(err, "failed to unmarshal department")
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to unmarshal department: %w", err))
		}
		if err := v.validateDepartment(ctx, department); err != nil {
			logger.Error(err, "department validation failed", "name", department.Name)
			return toAdmissionResponse(err)
		}
	default:
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unexpected resource kind %q", req.Kind.Kind))
	}

	return admission.Allowed("")
}

// validateQueueNodePoolsExist verifies that every queue references a node pool
// that exists in the cluster, and that no node pool is referenced by more than
// one queue. An empty list of queues is allowed, but a queue that names no node
// pool, or names one that does not exist, is rejected. All failures are
// aggregated so the caller sees every problem at once.
func (v *Validator) validateQueueNodePoolsExist(ctx context.Context, queues []kaiv1alpha1.QueueConfig) error {
	if len(queues) == 0 {
		return nil
	}

	existingNodePools, err := v.listNodePoolNames(ctx)
	if err != nil {
		return fmt.Errorf("failed to list node pools: %w", err)
	}

	var errs error
	seenQueues := map[string]bool{}
	for _, queue := range queues {
		if queue.Nodepool == "" {
			errs = multierror.Append(errs, fmt.Errorf("queue %q: node pool must not be empty", queue.Name))
			continue
		}
		if _, ok := seenQueues[queue.Nodepool]; ok {
			errs = multierror.Append(errs, fmt.Errorf("queue %q: node pool %q is already referenced by another queue",
				queue.Name, queue.Nodepool))
			continue
		}
		seenQueues[queue.Nodepool] = true

		if _, ok := existingNodePools[queue.Nodepool]; !ok {
			errs = multierror.Append(errs, fmt.Errorf("queue %q: node pool %q does not exist", queue.Name, queue.Nodepool))
		}
	}

	return errs
}

// listNodePoolNames returns the set of node pool names that currently exist in
// the cluster.
func (v *Validator) listNodePoolNames(ctx context.Context) (map[string]struct{}, error) {
	nodePoolList := &kaiv1alpha1.NodePoolList{}
	if err := v.client.List(ctx, nodePoolList); err != nil {
		return nil, newInternalError(err)
	}

	names := make(map[string]struct{}, len(nodePoolList.Items))
	for _, nodePool := range nodePoolList.Items {
		names[nodePool.Name] = struct{}{}
	}

	// not validating nodepool phases at this point, only existence
	return names, nil
}
