/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"fmt"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// validateHTTPRouteFilters bounds the extension surface before the ownership
// walk. RequestMirror destinations remain checked by exposureServiceBackends.
func validateHTTPRouteFilters(route *gatewayv1.HTTPRoute) error {
	if route == nil {
		return ownershipViolationf("HTTPRoute is nil")
	}
	var errs []error
	for ruleIndex := range route.Spec.Rules {
		rule := &route.Spec.Rules[ruleIndex]
		source := fmt.Sprintf("HTTPRoute %s/%s rule[%d]", route.Namespace, route.Name, ruleIndex)
		errs = append(errs, validateHTTPRouteFilterList(rule.Filters, source))
		for backendIndex := range rule.BackendRefs {
			errs = append(errs, validateHTTPRouteFilterList(
				rule.BackendRefs[backendIndex].Filters,
				fmt.Sprintf("%s.backend[%d]", source, backendIndex),
			))
		}
	}
	return prioritizeOwnershipErrors(errs...)
}

func validateHTTPRouteFilterList(filters []gatewayv1.HTTPRouteFilter, source string) error {
	var errs []error
	for i := range filters {
		if err := validateHTTPRouteFilter(&filters[i]); err != nil {
			errs = append(errs, fmt.Errorf("%s filter[%d]: %w", source, i, err))
		}
	}
	return prioritizeOwnershipErrors(errs...)
}

func validateHTTPRouteFilter(filter *gatewayv1.HTTPRouteFilter) error {
	switch filter.Type {
	case gatewayv1.HTTPRouteFilterRequestHeaderModifier,
		gatewayv1.HTTPRouteFilterResponseHeaderModifier,
		gatewayv1.HTTPRouteFilterURLRewrite,
		gatewayv1.HTTPRouteFilterRequestRedirect,
		gatewayv1.HTTPRouteFilterCORS,
		gatewayv1.HTTPRouteFilterRequestMirror:
	default:
		return ownershipViolationf("filter type %q is not supported by networking ownership validation", filter.Type)
	}

	// Endpoint annotations bypass the HTTPRoute CRD schema. Check every union
	// member so a permitted type cannot carry an unchecked extension alongside it.
	configured := map[gatewayv1.HTTPRouteFilterType]bool{
		gatewayv1.HTTPRouteFilterRequestHeaderModifier:  filter.RequestHeaderModifier != nil,
		gatewayv1.HTTPRouteFilterResponseHeaderModifier: filter.ResponseHeaderModifier != nil,
		gatewayv1.HTTPRouteFilterURLRewrite:             filter.URLRewrite != nil,
		gatewayv1.HTTPRouteFilterRequestRedirect:        filter.RequestRedirect != nil,
		gatewayv1.HTTPRouteFilterCORS:                   filter.CORS != nil,
		gatewayv1.HTTPRouteFilterRequestMirror:          filter.RequestMirror != nil,
		gatewayv1.HTTPRouteFilterExtensionRef:           filter.ExtensionRef != nil,
		gatewayv1.HTTPRouteFilterExternalAuth:           filter.ExternalAuth != nil,
	}
	count := 0
	for _, present := range configured {
		if present {
			count++
		}
	}
	if count != 1 || !configured[filter.Type] {
		return ownershipViolationf("filter type %q must set exactly its matching configuration", filter.Type)
	}
	return nil
}
