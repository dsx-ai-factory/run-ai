/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"fmt"
	"reflect"

	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type serviceBackend struct {
	key    types.NamespacedName
	source string
}

// validateExposureServiceOwnership enumerates every standard traffic
// destination in rendered exposure objects and applies the shared referenced
// Service validation. Plugin-created Services are additionally checked through
// validateCreatedServiceOwnership before this exposure check is invoked.
// The reader used here must perform live, uncached API reads.
func validateExposureServiceOwnership(
	ctx context.Context,
	reader client.Reader,
	runner *dromosv1alpha1.WorkloadRunner,
	exposures ...client.Object,
) error {
	ctx, cancel := context.WithTimeout(ctx, ownershipValidationTimeout)
	defer cancel()

	if reader == nil {
		return errOwnershipReaderNotConfigured
	}
	validationReader := newOwnershipValidationReader(reader)
	if err := validateRunnerIdentityIsComplete(runner); err != nil {
		return err
	}

	errs := make([]error, 0)
	backends := make([]serviceBackend, 0)
	for _, exposure := range exposures {
		if isNilClientObject(exposure) {
			errs = append(errs, ownershipViolationf("exposure object is nil"))
			continue
		}
		extracted, err := exposureServiceBackends(exposure)
		errs = append(errs, err)
		backends = append(backends, extracted...)

		namespace, supported := exposureObjectNamespace(exposure)
		if !supported {
			continue
		}
		if namespace != runner.Namespace {
			errs = append(errs, ownershipViolationf(
				"exposure namespace %q does not match WorkloadRunner namespace %q",
				namespace, runner.Namespace,
			))
		}
	}

	validated := datastructs.NewSet[types.NamespacedName]()
	for _, backend := range backends {
		if validated.Has(backend.key) {
			continue
		}
		if err := validateServiceOwnership(ctx, validationReader, runner, backend.key, false); err != nil {
			errs = append(errs, fmt.Errorf("%s references Service %s: %w", backend.source, backend.key, err))
		}
		validated.Add(backend.key)
	}
	return prioritizeOwnershipErrors(errs...)
}

func exposureObjectNamespace(exposure client.Object) (string, bool) {
	if isNilClientObject(exposure) {
		return "", false
	}
	switch exposure.(type) {
	case *networkingv1.Ingress, *openshiftv1.Route, *gatewayv1.HTTPRoute:
		return exposure.GetNamespace(), true
	}
	return "", false
}

func exposureServiceBackends(exposure client.Object) ([]serviceBackend, error) {
	if isNilClientObject(exposure) {
		return nil, ownershipViolationf("exposure is nil")
	}
	switch object := exposure.(type) {
	case *networkingv1.Ingress:
		return ingressServiceBackends(object)
	case *openshiftv1.Route:
		return routeServiceBackends(object)
	case *gatewayv1.HTTPRoute:
		return httpRouteServiceBackends(object)
	default:
		return nil, ownershipViolationf("unsupported exposure type %T", exposure)
	}
}

func isNilClientObject(object client.Object) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	return value.Kind() == reflect.Ptr && value.IsNil()
}

func ingressServiceBackends(ingress *networkingv1.Ingress) ([]serviceBackend, error) {
	backends := make([]serviceBackend, 0)
	errs := make([]error, 0)
	if ingress.Spec.DefaultBackend != nil {
		backend, err := ingressServiceBackend(
			ingress.Namespace, *ingress.Spec.DefaultBackend,
			fmt.Sprintf("Ingress %s/%s default backend", ingress.Namespace, ingress.Name),
		)
		if err != nil {
			errs = append(errs, err)
		} else {
			backends = append(backends, backend)
		}
	}
	for ruleIndex := range ingress.Spec.Rules {
		rule := &ingress.Spec.Rules[ruleIndex]
		if rule.HTTP == nil {
			continue
		}
		for pathIndex := range rule.HTTP.Paths {
			backend, err := ingressServiceBackend(
				ingress.Namespace,
				rule.HTTP.Paths[pathIndex].Backend,
				fmt.Sprintf("Ingress %s/%s rule[%d].path[%d] backend", ingress.Namespace, ingress.Name, ruleIndex, pathIndex),
			)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			backends = append(backends, backend)
		}
	}
	return backends, prioritizeOwnershipErrors(errs...)
}

func ingressServiceBackend(
	namespace string,
	backend networkingv1.IngressBackend,
	source string,
) (serviceBackend, error) {
	if backend.Service == nil {
		return serviceBackend{}, ownershipViolationf("%s is not a Service backend", source)
	}
	if backend.Resource != nil {
		return serviceBackend{}, ownershipViolationf("%s contains both Service and resource backends", source)
	}
	return serviceBackendForName(namespace, backend.Service.Name, source)
}

func routeServiceBackends(route *openshiftv1.Route) ([]serviceBackend, error) {
	backends := make([]serviceBackend, 0, 1+len(route.Spec.AlternateBackends))
	errs := make([]error, 0)
	primary, err := routeServiceBackend(
		route.Namespace, route.Spec.To,
		fmt.Sprintf("Route %s/%s primary backend", route.Namespace, route.Name),
	)
	if err != nil {
		errs = append(errs, err)
	} else {
		backends = append(backends, primary)
	}
	for i := range route.Spec.AlternateBackends {
		backend, err := routeServiceBackend(
			route.Namespace, route.Spec.AlternateBackends[i],
			fmt.Sprintf("Route %s/%s alternate backend[%d]", route.Namespace, route.Name, i),
		)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		backends = append(backends, backend)
	}
	return backends, prioritizeOwnershipErrors(errs...)
}

func routeServiceBackend(
	namespace string,
	backend openshiftv1.RouteTargetReference,
	source string,
) (serviceBackend, error) {
	if backend.Kind != "" && backend.Kind != serviceKind {
		return serviceBackend{}, ownershipViolationf("%s has unsupported kind %q", source, backend.Kind)
	}
	return serviceBackendForName(namespace, backend.Name, source)
}

func httpRouteServiceBackends(route *gatewayv1.HTTPRoute) ([]serviceBackend, error) {
	backends := make([]serviceBackend, 0)
	errs := make([]error, 0)
	for ruleIndex := range route.Spec.Rules {
		rule := &route.Spec.Rules[ruleIndex]
		for backendIndex := range rule.BackendRefs {
			backend := &rule.BackendRefs[backendIndex]
			ref, err := gatewayServiceBackend(
				route.Namespace,
				backend.BackendObjectReference,
				fmt.Sprintf("HTTPRoute %s/%s rule[%d].backend[%d]", route.Namespace, route.Name, ruleIndex, backendIndex),
			)
			if err != nil {
				errs = append(errs, err)
			} else {
				backends = append(backends, ref)
			}
			mirrors, err := httpMirrorServiceBackends(
				route.Namespace,
				backend.Filters,
				fmt.Sprintf("HTTPRoute %s/%s rule[%d].backend[%d]", route.Namespace, route.Name, ruleIndex, backendIndex),
			)
			errs = append(errs, err)
			backends = append(backends, mirrors...)
		}
		mirrors, err := httpMirrorServiceBackends(
			route.Namespace,
			rule.Filters,
			fmt.Sprintf("HTTPRoute %s/%s rule[%d]", route.Namespace, route.Name, ruleIndex),
		)
		errs = append(errs, err)
		backends = append(backends, mirrors...)
	}
	return backends, prioritizeOwnershipErrors(errs...)
}

func httpMirrorServiceBackends(
	namespace string,
	filters []gatewayv1.HTTPRouteFilter,
	source string,
) ([]serviceBackend, error) {
	backends := make([]serviceBackend, 0)
	errs := make([]error, 0)
	for i := range filters {
		filter := &filters[i]
		if filter.RequestMirror == nil {
			if filter.Type == gatewayv1.HTTPRouteFilterRequestMirror {
				errs = append(errs, ownershipViolationf(
					"%s filter[%d] has no request-mirror backend", source, i,
				))
			}
			continue
		}
		backend, err := gatewayServiceBackend(
			namespace,
			filter.RequestMirror.BackendRef,
			fmt.Sprintf("%s filter[%d] request-mirror backend", source, i),
		)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		backends = append(backends, backend)
	}
	return backends, prioritizeOwnershipErrors(errs...)
}

func gatewayServiceBackend(
	namespace string,
	backend gatewayv1.BackendObjectReference,
	source string,
) (serviceBackend, error) {
	if backend.Group != nil && *backend.Group != "" {
		return serviceBackend{}, ownershipViolationf("%s has non-core group %q", source, *backend.Group)
	}
	if backend.Kind != nil && *backend.Kind != serviceKind {
		return serviceBackend{}, ownershipViolationf("%s has non-Service kind %q", source, *backend.Kind)
	}
	backendNamespace := namespace
	if backend.Namespace != nil {
		backendNamespace = string(*backend.Namespace)
	}
	if backendNamespace != namespace {
		return serviceBackend{}, ownershipViolationf(
			"%s targets namespace %q instead of %q", source, backendNamespace, namespace,
		)
	}
	return serviceBackendForName(namespace, string(backend.Name), source)
}

func serviceBackendForName(namespace, name, source string) (serviceBackend, error) {
	if name == "" {
		return serviceBackend{}, ownershipViolationf("%s has an empty Service name", source)
	}
	return serviceBackend{
		key:    types.NamespacedName{Namespace: namespace, Name: name},
		source: source,
	}, nil
}
