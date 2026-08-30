package networking

import (
	"context"
	"errors"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var ErrServiceMissingCreateAndRef = errors.New("endpoint service must set either create or ref")

func (p *Plugin) handleEndpointsAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	cfg *endpoints.Config,
	resource *unstructured.Unstructured,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx)
	logger.V(1).Info("HandleAdditionalResources: entry (endpoints path)",
		"workloadRunner", runner.Name,
		"endpointsCount", len(cfg.Endpoints))

	owner, done, err := p.resolveOwnerOrDone(ctx, runner, resource, k8sClient)
	if done {
		return err
	}

	// Continue-on-error so siblings reconcile; return joined err so controller-runtime requeues failed ones.
	var sibErrs error
	desiredMetricsServices := map[string]struct{}{}
	for i := range cfg.Endpoints {
		endpoint := &cfg.Endpoints[i]
		if svc := firstService(endpoint); svc != nil && svc.Create != nil && isMetricsService(svc.Create) {
			desiredMetricsServices[svc.Create.Name] = struct{}{}
		}
		if err := p.reconcileEndpoint(ctx, runner, endpoint, i, owner, k8sClient, state); err != nil {
			_ = state.AddPluginIssue(ctx, p, runner, err, true,
				fmt.Sprintf("endpoint[%d]", i), issue.Network)
			logger.Error(err, "endpoint reconcile failed; continuing with siblings",
				"workloadRunner", runner.Name, "endpointIndex", i)
			sibErrs = errors.Join(sibErrs, fmt.Errorf("endpoint[%d]: %w", i, err))
		}
	}
	if sibErrs != nil {
		return sibErrs
	}

	if err := p.pruneStaleMetricsServices(ctx, runner, owner, desiredMetricsServices, k8sClient); err != nil {
		_ = state.AddPluginIssue(ctx, p, runner, err, true, "prune stale services", issue.Network)
		return err
	}
	return nil
}

func (p *Plugin) pruneStaleMetricsServices(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	owner client.Object,
	desired map[string]struct{},
	k8sClient client.Client,
) error {
	services := &corev1.ServiceList{}
	if err := k8sClient.List(
		ctx,
		services,
		client.InNamespace(runner.Namespace),
		client.MatchingLabels{
			constants.LabelWorkloadMetricsService: "true",
			LabelWorkloadRunner:                   runner.Name,
		},
	); err != nil {
		return fmt.Errorf("failed to list owned services for pruning: %w", err)
	}

	var pruneErr error
	for i := range services.Items {
		svc := &services.Items[i]
		if _, keep := desired[svc.Name]; keep || !isOwnedBy(svc, owner) {
			continue
		}
		if err := k8sClient.Delete(ctx, svc); err != nil {
			pruneErr = errors.Join(pruneErr, fmt.Errorf("failed to delete stale service %s: %w", svc.Name, err))
		}
	}
	return pruneErr
}

// pruneAllMetricsServices removes metrics Services previously created for the
// WorkloadRunner. It is used when endpoint networking is removed or replaced
// by the legacy shape so obsolete metrics Services are not left behind.
func (p *Plugin) pruneAllMetricsServices(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	owner, done, err := p.resolveOwnerOrDone(ctx, runner, resource, k8sClient)
	if done {
		return err
	}
	if err := p.pruneStaleMetricsServices(ctx, runner, owner, map[string]struct{}{}, k8sClient); err != nil {
		_ = state.AddPluginIssue(ctx, p, runner, err, true, "prune stale services", issue.Network)
		return err
	}
	return nil
}

func isOwnedBy(obj client.Object, owner client.Object) bool {
	if obj == nil || owner == nil {
		return false
	}
	for _, ref := range obj.GetOwnerReferences() {
		if ref.Name != owner.GetName() {
			continue
		}
		if owner.GetUID() == "" || ref.UID == owner.GetUID() {
			return true
		}
	}
	return false
}

func (p *Plugin) reconcileEndpoint(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	endpoint *endpoints.Endpoint,
	endpointIdx int,
	owner client.Object,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	svc := firstService(endpoint)
	if svc == nil {
		if endpoint.ExternalAccess == nil {
			return nil
		}
		return p.createEndpointExternalAccess(ctx, runner, endpoint, owner, k8sClient, "", 0)
	}

	workloadServiceName, workloadServicePort, err := p.reconcileEndpointService(ctx, runner, svc, owner, k8sClient)
	if err != nil {
		return fmt.Errorf("reconcile workload service: %w", err)
	}

	backendServiceName := workloadServiceName
	backendServicePort := workloadServicePort

	if svc.Auth != nil {
		authProxyName := getEndpointAuthProxyName(workloadServiceName)
		if err := p.createOrUpdateEndpointAuthProxy(
			ctx, runner, endpoint, svc, authProxyName,
			workloadServiceName, workloadServicePort,
			owner, k8sClient,
		); err != nil {
			return fmt.Errorf("reconcile auth proxy: %w", err)
		}
		backendServiceName = authProxyName
		backendServicePort = AuthProxyServicePort
	}

	if endpoint.ExternalAccess != nil {
		if err := p.createEndpointExternalAccess(
			ctx, runner, endpoint, owner, k8sClient,
			backendServiceName, backendServicePort,
		); err != nil {
			return fmt.Errorf("reconcile external access: %w", err)
		}
	}
	return nil
}

func (p *Plugin) reconcileEndpointService(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	svc *endpoints.Service,
	owner client.Object,
	k8sClient client.Client,
) (string, int32, error) {
	switch {
	case svc.Create != nil:
		built := p.buildEndpointWorkloadService(runner, svc.Create)
		if err := p.setOwnerAndApply(ctx, k8sClient, owner, built, "workload service"); err != nil {
			return "", 0, err
		}
		port := int32(0)
		if len(built.Spec.Ports) > 0 {
			port = built.Spec.Ports[0].Port
		}
		return built.Name, port, nil
	case svc.Ref != nil:
		return svc.Ref.Name, svc.Ref.Port, nil
	default:
		return "", 0, ErrServiceMissingCreateAndRef
	}
}

func (p *Plugin) createOrUpdateEndpointAuthProxy(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	endpoint *endpoints.Endpoint,
	svc *endpoints.Service,
	authProxyName string,
	workloadServiceName string,
	workloadServicePort int32,
	owner client.Object,
	k8sClient client.Client,
) error {
	upstreamURL := fmt.Sprintf(upstreamURLFormat,
		workloadServiceName, runner.Namespace, workloadServicePort)
	externalHost, externalPath := getEndpointExternalHostAndPath(endpoint)

	deployment, err := p.buildEndpointAuthProxyDeployment(
		runner, svc, authProxyName, upstreamURL, externalHost, externalPath,
	)
	if err != nil {
		return fmt.Errorf("build auth proxy deployment: %w", err)
	}
	if err := p.setOwnerAndApply(ctx, k8sClient, owner, deployment, "auth proxy deployment"); err != nil {
		return err
	}

	authSvc := p.buildAuthProxyService(runner, authProxyName)
	if err := p.setOwnerAndApply(ctx, k8sClient, owner, authSvc, "auth proxy service"); err != nil {
		return err
	}
	return nil
}

func (p *Plugin) createEndpointExternalAccess(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	endpoint *endpoints.Endpoint,
	owner client.Object,
	k8sClient client.Client,
	backendServiceName string,
	backendServicePort int32,
) error {
	ea := endpoint.ExternalAccess
	if ea == nil {
		return nil
	}
	return p.applyExternalAccess(ctx, k8sClient, runner, owner,
		ea.Ingress, ea.Route, ea.HTTPRoute, backendServiceName, backendServicePort)
}
