package networking

import (
	"context"
	"fmt"

	openshiftv1 "github.com/openshift/api/route/v1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const upstreamURLFormat = "http://%s.%s.svc.cluster.local:%d"

// HandleAdditionalResources dispatches on the configured input shape; see README "Input shapes".
func (p *Plugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	shape, cfg, err := classifyInput(runner)
	if err != nil {
		_ = state.AddPluginIssue(ctx, p, runner, err, false, "", issue.Network)
		return nil
	}

	switch shape {
	case inputLegacy:
		if err := p.pruneAllMetricsServices(ctx, runner, resource, k8sClient, state); err != nil {
			return err
		}
		return p.handleLegacyAdditionalResources(ctx, runner, resource, karta, k8sClient, state)
	case inputAnnotation:
		return p.handleEndpointsAdditionalResources(ctx, runner, cfg, resource, k8sClient, state)
	default:
		return p.pruneAllMetricsServices(ctx, runner, resource, k8sClient, state)
	}
}

func (p *Plugin) handleLegacyAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	_ *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx)
	config := extractNetworkingConfig(runner)
	if config == nil {
		return nil
	}

	resourceGVK := ""
	resourceName := ""
	if resource != nil {
		resourceGVK = resource.GroupVersionKind().String()
		resourceName = resource.GetName()
	}
	logger.V(1).Info("HandleAdditionalResources: entry",
		"workloadRunner", runner.Name,
		"resourceGVK", resourceGVK,
		"resourceName", resourceName)

	owner, done, err := p.resolveOwnerOrDone(ctx, runner, resource, k8sClient)
	if done {
		return err
	}

	if !hasService(config) {
		return p.createExternalAccess(ctx, runner, config, owner, k8sClient, "", 0)
	}

	workloadServiceName, workloadServicePort, err := p.getOrCreateWorkloadService(ctx, runner, config, owner, k8sClient)
	if err != nil {
		return err
	}

	var backendServiceName string
	var backendServicePort int32
	authEnabled := isAuthEnabled(config)

	if authEnabled {
		if err := p.createOrUpdateAuthProxy(ctx, runner, config, workloadServiceName, workloadServicePort, owner, k8sClient, state); err != nil {
			return err
		}
		backendServiceName = getAuthProxyName(runner)
		backendServicePort = AuthProxyServicePort
	} else {
		backendServiceName = workloadServiceName
		backendServicePort = workloadServicePort
	}

	return p.createExternalAccess(ctx, runner, config, owner, k8sClient, backendServiceName, backendServicePort)
}

// resolveOwner determines the owner for networking resources.
// If the primary resource (e.g. NIMService) exists in the cluster, it is used as the owner.
// If it does not exist yet (first reconcile), we skip creating networking resources
// entirely — the primary resource creation will trigger a new reconcile where we'll retry.
// If resource is nil (e.g. in tests), we fall back to the WorkloadRunner as owner.
func (p *Plugin) resolveOwner(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	k8sClient client.Client,
) (owner client.Object, skip bool, err error) {
	logger := log.FromContext(ctx)

	if resource == nil || resource.GetName() == "" || resource.GroupVersionKind().Empty() {
		logger.V(1).Info("resolveOwner: no primary resource template, using WorkloadRunner as owner",
			"workloadRunner", runner.Name)
		return runner, false, nil
	}

	liveResource := &unstructured.Unstructured{}
	liveResource.SetGroupVersionKind(resource.GroupVersionKind())
	key := client.ObjectKey{Namespace: runner.Namespace, Name: resource.GetName()}
	if err := k8sClient.Get(ctx, key, liveResource); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("resolveOwner: primary resource not yet created, skipping networking resources",
				"resourceGVK", resource.GroupVersionKind().String(),
				"resourceName", resource.GetName())
			return nil, true, nil
		}
		logger.Error(err, "resolveOwner: failed to fetch primary resource",
			"resourceGVK", resource.GroupVersionKind().String(),
			"resourceName", resource.GetName())
		return nil, false, fmt.Errorf("failed to fetch primary resource for owner reference: %w", err)
	}

	logger.Info("resolveOwner: using primary resource as owner",
		"ownerGVK", liveResource.GroupVersionKind().String(),
		"ownerName", liveResource.GetName(),
		"ownerUID", liveResource.GetUID())
	return liveResource, false, nil
}

func (p *Plugin) resolveOwnerOrDone(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	k8sClient client.Client,
) (owner client.Object, done bool, err error) {
	owner, skip, err := p.resolveOwner(ctx, runner, resource, k8sClient)
	if err != nil || skip {
		return nil, true, err
	}
	return owner, false, nil
}

func (p *Plugin) getOrCreateWorkloadService(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	config *dromosv1alpha1.NetworkingConfig,
	owner client.Object,
	k8sClient client.Client,
) (name string, port int32, err error) {
	info := getWorkloadServiceInfo(config, runner.Name)

	if info.Action == WorkloadServiceActionCreate {
		svc := p.buildWorkloadService(runner, config.Services[0].Create)
		if err := p.setOwnerAndApply(ctx, k8sClient, owner, svc, "workload service"); err != nil {
			return "", 0, err
		}
		info.Name = svc.Name
	}

	return info.Name, info.Port, nil
}

func (p *Plugin) createOrUpdateAuthProxy(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	config *dromosv1alpha1.NetworkingConfig,
	workloadServiceName string,
	workloadServicePort int32,
	owner client.Object,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	upstreamURL := fmt.Sprintf(upstreamURLFormat,
		workloadServiceName, runner.Namespace, workloadServicePort)
	externalHost, externalPath := getExternalHostAndPath(config)

	deployment, err := p.buildAuthProxyDeployment(runner, config, upstreamURL, externalHost, externalPath)
	if err != nil {
		buildErr := fmt.Errorf("failed to build auth proxy deployment: %w", err)
		_ = state.AddPluginIssue(ctx, p, runner, buildErr, true, "", issue.Internal) // allows Status to detect build failure in the same reconcile loop
		return buildErr
	}
	if err := p.setOwnerAndApply(ctx, k8sClient, owner, deployment, "auth proxy deployment"); err != nil {
		return err
	}

	authSvc := p.buildAuthProxyService(runner, getAuthProxyName(runner))
	if err := p.setOwnerAndApply(ctx, k8sClient, owner, authSvc, "auth proxy service"); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) setOwnerAndApply(
	ctx context.Context,
	k8sClient client.Client,
	owner client.Object,
	resource client.Object,
	resourceType string,
) error {
	logger := log.FromContext(ctx)

	logger.V(1).Info("setOwnerAndApply: setting owner reference",
		"resourceType", resourceType,
		"resourceName", resource.GetName(),
		"ownerKind", owner.GetObjectKind().GroupVersionKind().Kind,
		"ownerName", owner.GetName(),
		"ownerUID", owner.GetUID())

	if err := controllerutil.SetOwnerReference(owner, resource, k8sClient.Scheme()); err != nil {
		return fmt.Errorf("failed to set owner reference on %s: %w", resourceType, err)
	}

	logger.V(1).Info("setOwnerAndApply: ownerReferences set",
		"resourceName", resource.GetName(),
		"ownerRefs", fmt.Sprintf("%v", resource.GetOwnerReferences()))

	resource.SetManagedFields(nil)
	resource.SetResourceVersion("")
	if err := k8sClient.Patch(ctx, resource, client.Apply, client.ForceOwnership, client.FieldOwner(PluginName)); err != nil {
		return fmt.Errorf("failed to apply %s: %w", resourceType, err)
	}

	logger.Info("setOwnerAndApply: applied",
		"resourceType", resourceType,
		"resourceName", resource.GetName(),
		"resourceUID", resource.GetUID())

	verified := resource.DeepCopyObject().(client.Object)
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(resource), verified); err != nil {
		logger.Error(err, "setOwnerAndApply: POST-APPLY VERIFICATION FAILED",
			"resourceType", resourceType,
			"resourceName", resource.GetName())
	} else {
		logger.V(1).Info("setOwnerAndApply: post-apply verification OK",
			"resourceType", resourceType,
			"resourceName", verified.GetName(),
			"resourceUID", verified.GetUID(),
			"ownerRefs", fmt.Sprintf("%v", verified.GetOwnerReferences()))
	}

	return nil
}

func (p *Plugin) createExternalAccess(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	config *dromosv1alpha1.NetworkingConfig,
	owner client.Object,
	k8sClient client.Client,
	backendServiceName string,
	backendServicePort int32,
) error {
	if config.ExternalAccess == nil {
		return nil
	}
	ea := config.ExternalAccess
	return p.applyExternalAccess(ctx, k8sClient, runner, owner,
		ea.Ingress, ea.Route, ea.HTTPRoute, backendServiceName, backendServicePort)
}

func (p *Plugin) applyExternalAccess(
	ctx context.Context,
	k8sClient client.Client,
	runner *dromosv1alpha1.WorkloadRunner,
	owner client.Object,
	ingressSpec *networkingv1.Ingress,
	routeSpec *openshiftv1.Route,
	httpRouteSpec *gatewayv1.HTTPRoute,
	backendServiceName string,
	backendServicePort int32,
) error {
	if ingressSpec != nil {
		ingress := p.buildIngress(runner, ingressSpec, backendServiceName, backendServicePort)
		if err := p.setOwnerAndApply(ctx, k8sClient, owner, ingress, "ingress"); err != nil {
			return err
		}
	}
	if routeSpec != nil {
		route := p.buildRoute(runner, routeSpec, backendServiceName)
		if err := p.setOwnerAndApply(ctx, k8sClient, owner, route, "OpenShift route"); err != nil {
			return err
		}
	}
	if httpRouteSpec != nil {
		httpRoute := p.buildHTTPRoute(runner, httpRouteSpec, backendServiceName, backendServicePort)
		if err := p.setOwnerAndApply(ctx, k8sClient, owner, httpRoute, "HTTPRoute"); err != nil {
			return err
		}
	}
	return nil
}
