/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	openshiftv1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	errOwnershipEnforcement     = errors.New("networking ownership enforcement failed")
	errOwnershipRetryNeeded     = errors.New("networking ownership enforcement requires retry")
	errOwnershipEventDrivenWait = errors.New("networking ownership enforcement is waiting for a watched resource")
	errOwnershipOperational     = errors.New("networking ownership enforcement encountered an operational failure")
)

type managedResource struct {
	object      client.Object
	description string
}

type ownershipRetryError struct {
	cause       error
	eventDriven bool
}

func (e ownershipRetryError) Error() string {
	return e.cause.Error()
}

func (e ownershipRetryError) Unwrap() error {
	return e.cause
}

func (e ownershipRetryError) Is(target error) bool {
	return target == errOwnershipRetryNeeded ||
		(e.eventDriven && target == errOwnershipEventDrivenWait) ||
		(!e.eventDriven && target == errOwnershipOperational)
}

func ownershipEnforcementError(action string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, errOwnershipEnforcement) {
		return err
	}
	if !errors.Is(err, errOwnershipViolation) {
		err = ownershipRetryError{
			cause:       err,
			eventDriven: isEventDrivenRetryCause(err),
		}
	}
	return fmt.Errorf("%w: %s: %w", errOwnershipEnforcement, action, err)
}

// isEventDrivenRetryCause identifies races and missing dependencies whose next
// state transition is covered by the networking dependency watches. NotFound
// comes from live Service or exposure checks; owner-walk NotFound is already an
// ownership violation. AlreadyExists and Conflict come from mutations of the
// managed Service, exposure, or auth-proxy Deployment, all of which are watched.
// All parts of a joined error must be event-driven; one operational failure
// still needs controller-runtime's ordinary error retry.
func isEventDrivenRetryCause(err error) bool {
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if !isEventDrivenRetryCause(cause) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok && wrapped.Unwrap() != nil {
		return isEventDrivenRetryCause(wrapped.Unwrap())
	}
	return apierrors.IsNotFound(err) || apierrors.IsAlreadyExists(err) || apierrors.IsConflict(err)
}

func isSettledOwnershipViolation(err error) bool {
	return errors.Is(err, errOwnershipViolation) && !errors.Is(err, errOwnershipRetryNeeded)
}

func (p *Plugin) recordOwnershipEnforcementError(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	state *reconcile.State,
	err error,
	description string,
) error {
	if err == nil || !errors.Is(err, errOwnershipEnforcement) {
		return err
	}
	_ = state.AddPluginIssue(
		ctx, p, runner, err, errors.Is(err, errOwnershipViolation), description, issue.Network,
	)
	if isSettledOwnershipViolation(err) ||
		(errors.Is(err, errOwnershipEventDrivenWait) && !errors.Is(err, errOwnershipOperational)) {
		return nil
	}
	return err
}

func (p *Plugin) ensureOwnedOrAbsent(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	desired client.Object,
	description string,
) error {
	if !p.enforceOwnership {
		return nil
	}
	_, err := p.readOwnedOrAbsent(ctx, runner, desired, description)
	if err != nil {
		return ownershipEnforcementError(
			fmt.Sprintf("refuse to adopt existing %s", description), err,
		)
	}
	return nil
}

func (p *Plugin) readOwnedOrAbsent(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	desired client.Object,
	description string,
) (client.Object, error) {
	if p.ownershipReader == nil {
		return nil, errOwnershipReaderNotConfigured
	}

	// Do not let desired authorization metadata survive into a live read when
	// the Reader decodes into an existing object.
	live := desired.DeepCopyObject().(client.Object)
	live.SetUID("")
	live.SetResourceVersion("")
	live.SetOwnerReferences(nil)
	live.SetManagedFields(nil)
	key := client.ObjectKeyFromObject(desired)
	if err := p.ownershipReader.Get(ctx, key, live); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s %s: %w", description, key, err)
	}
	if err := validateObjectOwner(ctx, p.ownershipReader, runner, live, description); err != nil {
		return live, fmt.Errorf("validate %s %s owner: %w", description, key, err)
	}
	return live, nil
}

// ensureNoOrphanEndpointSlices prevents a settled ownership denial from
// becoming a Service create/delete loop. Kubernetes removes ordinary slices
// with their Service; any same-name slice that remains while the Service is
// absent must disappear before a new Service UID can safely claim the name.
func (p *Plugin) ensureNoOrphanEndpointSlices(
	ctx context.Context,
	service *corev1.Service,
	description string,
) error {
	if !p.enforceOwnership {
		return nil
	}
	if p.ownershipReader == nil {
		return ownershipEnforcementError(
			"inspect EndpointSlices before creating "+description,
			errOwnershipReaderNotConfigured,
		)
	}

	slices := &discoveryv1.EndpointSliceList{}
	if err := p.ownershipReader.List(
		ctx,
		slices,
		client.InNamespace(service.Namespace),
		client.MatchingLabels{discoveryv1.LabelServiceName: service.Name},
	); err != nil {
		return ownershipEnforcementError("inspect EndpointSlices before creating "+description, err)
	}
	if len(slices.Items) == 0 {
		return nil
	}

	return ownershipEnforcementError(
		"refuse to create "+description,
		ownershipViolationf(
			"%d EndpointSlice(s) still reference absent Service %s/%s",
			len(slices.Items), service.Namespace, service.Name,
		),
	)
}

func (p *Plugin) validateDesiredService(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	desired *corev1.Service,
	description string,
) error {
	if !p.enforceOwnership {
		return nil
	}
	if err := validateDesiredServiceOwnership(ctx, p.ownershipReader, runner, desired); err != nil {
		return ownershipEnforcementError("validate desired "+description, err)
	}
	return nil
}

func (p *Plugin) reconcileCreatedService(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	owner client.Object,
	desired *corev1.Service,
	k8sClient client.Client,
	description string,
) error {
	deny := func(err error) error {
		if !isSettledOwnershipViolation(err) {
			return err
		}
		return errors.Join(err, p.deleteIfOwned(ctx, runner, desired, k8sClient, description))
	}
	if err := p.validateDesiredService(ctx, runner, desired, description); err != nil {
		return deny(err)
	}
	if err := p.setOwnerAndApply(ctx, runner, k8sClient, owner, desired, description); err != nil {
		return err
	}
	if !p.enforceOwnership {
		return nil
	}
	if err := validateCreatedServiceOwnership(
		ctx,
		p.ownershipReader,
		runner,
		types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name},
	); err != nil {
		return deny(ownershipEnforcementError("validate live "+description, err))
	}
	return nil
}

func (p *Plugin) validateReferencedService(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
) error {
	if !p.enforceOwnership {
		return nil
	}
	err := validateReferencedServiceOwnership(
		ctx,
		p.ownershipReader,
		runner,
		types.NamespacedName{Namespace: runner.Namespace, Name: name},
	)
	return ownershipEnforcementError("validate referenced Service", err)
}

func (p *Plugin) deleteIfOwned(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	object client.Object,
	k8sClient client.Client,
	description string,
) error {
	if !p.enforceOwnership {
		return nil
	}
	live, err := p.readOwnedOrAbsent(ctx, runner, object, description)
	if err != nil {
		if errors.Is(err, errOwnershipViolation) {
			log.FromContext(ctx).Info(
				"Refusing to delete resource whose ownership cannot be proven",
				"resourceType", description,
				"resource", client.ObjectKeyFromObject(live),
			)
			return nil
		}
		return ownershipEnforcementError("inspect "+description+" before deletion", err)
	}
	if live == nil {
		return nil
	}
	uid := live.GetUID()
	if err := k8sClient.Delete(
		ctx, live, client.Preconditions{UID: &uid},
	); err != nil && !apierrors.IsNotFound(err) {
		return ownershipEnforcementError("delete "+description, err)
	}
	return nil
}

func (p *Plugin) buildExternalAccessResources(
	runner *dromosv1alpha1.WorkloadRunner,
	ingressSpec *networkingv1.Ingress,
	routeSpec *openshiftv1.Route,
	httpRouteSpec *gatewayv1.HTTPRoute,
	backendServiceName string,
	backendServicePort int32,
) []managedResource {
	resources := make([]managedResource, 0, 3)
	if ingressSpec != nil {
		resources = append(resources, managedResource{
			object:      p.buildIngress(runner, ingressSpec, backendServiceName, backendServicePort),
			description: "ingress",
		})
	}
	if routeSpec != nil {
		resources = append(resources, managedResource{
			object:      p.buildRoute(runner, routeSpec, backendServiceName),
			description: "OpenShift route",
		})
	}
	if httpRouteSpec != nil {
		resources = append(resources, managedResource{
			object:      p.buildHTTPRoute(runner, httpRouteSpec, backendServiceName, backendServicePort),
			description: "HTTPRoute",
		})
	}
	return resources
}

func (p *Plugin) reconcileExternalAccess(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	owner client.Object,
	resources []managedResource,
	k8sClient client.Client,
) error {
	if len(resources) == 0 {
		return nil
	}
	fail := func(err error) error {
		if !p.enforceOwnership || !isSettledOwnershipViolation(err) {
			return err
		}
		return errors.Join(err, p.deleteResourcesIfOwned(ctx, runner, resources, k8sClient))
	}

	if p.enforceOwnership {
		exposures := make([]client.Object, 0, len(resources))
		for _, resource := range resources {
			exposures = append(exposures, resource.object)
		}
		if err := validateExposureServiceOwnership(ctx, p.ownershipReader, runner, exposures...); err != nil {
			return fail(ownershipEnforcementError("validate external access", err))
		}

		// Check every name before applying any object so a collision cannot leave
		// a partially exposed graph.
		for _, resource := range resources {
			if err := p.ensureOwnedOrAbsent(ctx, runner, resource.object, resource.description); err != nil {
				return fail(err)
			}
		}
	}

	for _, resource := range resources {
		if err := p.setOwnerAndApply(
			ctx, runner, k8sClient, owner, resource.object, resource.description,
		); err != nil {
			return fail(err)
		}
	}
	if p.enforceOwnership {
		if err := p.validateLiveExternalAccess(ctx, runner, resources); err != nil {
			return fail(err)
		}
	}
	return nil
}

func (p *Plugin) validateLiveExternalAccess(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resources []managedResource,
) error {
	liveExposures := make([]client.Object, 0, len(resources))
	for _, resource := range resources {
		live := resource.object.DeepCopyObject().(client.Object)
		if err := p.ownershipReader.Get(ctx, client.ObjectKeyFromObject(resource.object), live); err != nil {
			return ownershipEnforcementError("read live "+resource.description, err)
		}
		if err := validateObjectOwner(ctx, p.ownershipReader, runner, live, resource.description); err != nil {
			return ownershipEnforcementError("validate live "+resource.description+" owner", err)
		}
		liveExposures = append(liveExposures, live)
	}
	if err := validateExposureServiceOwnership(
		ctx, p.ownershipReader, runner, liveExposures...,
	); err != nil {
		return ownershipEnforcementError("validate live external access", err)
	}
	return nil
}

func (p *Plugin) deleteResourcesIfOwned(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resources []managedResource,
	k8sClient client.Client,
) error {
	var cleanupErr error
	for _, resource := range resources {
		cleanupErr = errors.Join(
			cleanupErr,
			p.deleteIfOwned(ctx, runner, resource.object, k8sClient, resource.description),
		)
	}
	return cleanupErr
}

func (p *Plugin) cleanupAfterOwnershipError(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	err error,
	resources []managedResource,
	authProxyName string,
	removeAuthProxy bool,
	k8sClient client.Client,
) error {
	if !isSettledOwnershipViolation(err) {
		return err
	}
	cleanupErr := p.deleteResourcesIfOwned(ctx, runner, resources, k8sClient)
	if removeAuthProxy && authProxyName != "" {
		cleanupErr = errors.Join(cleanupErr, p.deleteAuthProxy(ctx, runner, authProxyName, k8sClient))
	}
	return errors.Join(err, cleanupErr)
}

func (p *Plugin) deleteAuthProxy(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	k8sClient client.Client,
) error {
	return p.deleteResourcesIfOwned(ctx, runner, []managedResource{
		{
			object: &corev1.Service{
				ObjectMeta: objectMeta(runner.Namespace, name),
			},
			description: "auth proxy service",
		},
		{
			object: &appsv1.Deployment{
				ObjectMeta: objectMeta(runner.Namespace, name),
			},
			description: "auth proxy deployment",
		},
	}, k8sClient)
}

func objectMeta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Namespace: namespace, Name: name}
}

func shouldDeleteAuthProxyFor(err error) bool {
	return isSettledOwnershipViolation(err)
}
