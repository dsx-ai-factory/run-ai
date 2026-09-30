/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"encoding/json"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	openshiftv1 "github.com/openshift/api/route/v1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
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
	if p.configMapName != "" {
		if err := p.ensureOAuth2Templates(ctx, runner, k8sClient); err != nil {
			return err
		}
	}
	if p.routingDisabled {
		return nil
	}
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
		err := p.handleLegacyAdditionalResources(ctx, runner, resource, karta, k8sClient, state)
		return p.recordOwnershipEnforcementError(ctx, runner, state, err, "networking ownership validation")
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
		return p.cleanupAfterOwnershipError(
			ctx, runner, err, p.buildLegacyExternalAccessResources(runner, config),
			getAuthProxyName(runner), isAuthEnabled(config) && shouldDeleteAuthProxyFor(err), k8sClient,
		)
	}

	var backendServiceName string
	var backendServicePort int32
	authEnabled := isAuthEnabled(config)

	if authEnabled {
		if err := p.createOrUpdateAuthProxy(ctx, runner, config, workloadServiceName, workloadServicePort, owner, k8sClient, state); err != nil {
			return p.cleanupAfterOwnershipError(
				ctx, runner, err, p.buildLegacyExternalAccessResources(runner, config),
				getAuthProxyName(runner), true, k8sClient,
			)
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
	reader := client.Reader(k8sClient)
	if p.enforceOwnership {
		reader = p.ownershipReader
	}
	if reader == nil {
		return nil, false, ownershipEnforcementError(
			"resolve primary resource owner", errOwnershipReaderNotConfigured,
		)
	}
	if err := reader.Get(ctx, key, liveResource); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("resolveOwner: primary resource not yet created, skipping networking resources",
				"resourceGVK", resource.GroupVersionKind().String(),
				"resourceName", resource.GetName())
			return nil, true, nil
		}
		logger.Error(err, "resolveOwner: failed to fetch primary resource",
			"resourceGVK", resource.GroupVersionKind().String(),
			"resourceName", resource.GetName())
		return nil, false, ownershipEnforcementError(
			"fetch primary resource for owner reference", err,
		)
	}
	if p.enforceOwnership {
		if err := validateObjectOwner(
			ctx, p.ownershipReader, runner, liveResource, "primary resource",
		); err != nil {
			return nil, false, ownershipEnforcementError("validate primary resource owner", err)
		}
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
		if err := p.reconcileCreatedService(ctx, runner, owner, svc, k8sClient, "workload service"); err != nil {
			return "", 0, err
		}
		info.Name = svc.Name
	} else if err := p.validateReferencedService(ctx, runner, info.Name); err != nil {
		return "", 0, err
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
	authSvc := p.buildAuthProxyService(runner, getAuthProxyName(runner))
	if err := p.reconcileCreatedService(ctx, runner, owner, authSvc, k8sClient, "auth proxy service"); err != nil {
		return err
	}
	if err := p.setOwnerAndApply(ctx, runner, k8sClient, owner, deployment, "auth proxy deployment"); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) setOwnerAndApply(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	k8sClient client.Client,
	owner client.Object,
	resource client.Object,
	resourceType string,
) error {
	logger := log.FromContext(ctx)
	var live client.Object
	if p.enforceOwnership {
		var err error
		live, err = p.readOwnedOrAbsent(ctx, runner, resource, resourceType)
		if err != nil {
			return ownershipEnforcementError(
				fmt.Sprintf("refuse to adopt existing %s", resourceType), err,
			)
		}
		if live == nil {
			if service, ok := resource.(*corev1.Service); ok {
				if err := p.ensureNoOrphanEndpointSlices(ctx, service, resourceType); err != nil {
					return err
				}
			}
		}
	}

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
	if p.enforceOwnership && live == nil {
		// Consume the observed absence with Create so a same-name object that
		// appears after validation cannot be adopted by a forced apply.
		resource.SetResourceVersion("")
		if err := k8sClient.Create(ctx, resource, client.FieldOwner(PluginName)); err != nil {
			if !apierrors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create %s: %w", resourceType, err)
			}
			if _, validationErr := p.readOwnedOrAbsent(ctx, runner, resource, resourceType); validationErr != nil {
				return ownershipEnforcementError(
					"refuse to adopt concurrently created "+resourceType, validationErr,
				)
			}
			return ownershipEnforcementError("retry after concurrent creation of "+resourceType, err)
		}
	} else {
		if live == nil {
			resource.SetResourceVersion("")
		} else {
			// Bind the apply to the exact identity and version authorized above.
			resource.SetUID(live.GetUID())
			resource.SetResourceVersion(live.GetResourceVersion())
		}
		applyPatch, err := json.Marshal(resource)
		if err != nil {
			return fmt.Errorf("failed to marshal %s for server-side apply: %w", resourceType, err)
		}
		if err := k8sClient.Patch(
			ctx, resource, client.RawPatch(types.ApplyPatchType, applyPatch),
			client.ForceOwnership, client.FieldOwner(PluginName),
		); err != nil {
			if p.enforceOwnership && apierrors.IsConflict(err) {
				return ownershipEnforcementError("apply "+resourceType, err)
			}
			return fmt.Errorf("failed to apply %s: %w", resourceType, err)
		}
	}

	logger.Info("setOwnerAndApply: reconciled",
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

func (p *Plugin) buildLegacyExternalAccessResources(
	runner *dromosv1alpha1.WorkloadRunner,
	config *dromosv1alpha1.NetworkingConfig,
) []managedResource {
	if config == nil || config.ExternalAccess == nil {
		return nil
	}
	ea := config.ExternalAccess
	return p.buildExternalAccessResources(runner, ea.Ingress, ea.Route, ea.HTTPRoute, "", 0)
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
	resources := p.buildExternalAccessResources(
		runner, ingressSpec, routeSpec, httpRouteSpec, backendServiceName, backendServicePort,
	)
	return p.reconcileExternalAccess(ctx, runner, owner, resources, k8sClient)
}
