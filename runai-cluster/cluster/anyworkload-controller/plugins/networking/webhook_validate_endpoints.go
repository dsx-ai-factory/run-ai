/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	templateutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// endpointNameTracker reserves K8s object names per kind so admission rejects collisions instead of silently overwriting on apply.
type endpointNameTracker struct {
	runner               *dromosv1alpha1.WorkloadRunner
	attachedWorkloadName string
	services             map[string]int
	ingresses            map[string]int
	routes               map[string]int
	httpRoutes           map[string]int
}

func newEndpointNameTracker(runner *dromosv1alpha1.WorkloadRunner) *endpointNameTracker {
	attachedWorkloadName := runner.Name
	if template, err := templateutils.NewParser().Parse(runner.Spec.Template); err == nil && template.GetName() != "" {
		attachedWorkloadName = template.GetName()
	}
	return &endpointNameTracker{
		runner:               runner,
		attachedWorkloadName: attachedWorkloadName,
		services:             map[string]int{},
		ingresses:            map[string]int{},
		routes:               map[string]int{},
		httpRoutes:           map[string]int{},
	}
}

func (t *endpointNameTracker) recordService(
	name string, endpointIdx int, path *field.Path,
) field.ErrorList {
	return recordName(t.services, "service", name, endpointIdx, path)
}

func (t *endpointNameTracker) recordIngress(
	name string, endpointIdx int, path *field.Path,
) field.ErrorList {
	return recordName(t.ingresses, "ingress", name, endpointIdx, path)
}

func (t *endpointNameTracker) recordRoute(
	name string, endpointIdx int, path *field.Path,
) field.ErrorList {
	return recordName(t.routes, "route", name, endpointIdx, path)
}

func (t *endpointNameTracker) recordHTTPRoute(
	name string, endpointIdx int, path *field.Path,
) field.ErrorList {
	return recordName(t.httpRoutes, "httpRoute", name, endpointIdx, path)
}

func recordName(
	m map[string]int, kind, name string, endpointIdx int, path *field.Path,
) field.ErrorList {
	if name == "" {
		return nil
	}
	if prev, ok := m[name]; ok {
		return field.ErrorList{field.Duplicate(
			path, fmt.Sprintf(
				"%s name %q is already used by endpoints[%d]; %s names must be unique across endpoints",
				kind, name, prev, kind,
			),
		)}
	}
	m[name] = endpointIdx
	return nil
}

func validateEndpointsConfig(runner *dromosv1alpha1.WorkloadRunner, cfg *endpoints.Config) field.ErrorList {
	if cfg == nil {
		return nil
	}

	var errs field.ErrorList
	basePath := field.NewPath("metadata", "annotations").Key(endpoints.Annotation)

	if len(cfg.Endpoints) == 0 {
		errs = append(errs, field.Required(
			basePath.Child("endpoints"),
			"at least one endpoint is required",
		))
		return errs
	}

	tracker := newEndpointNameTracker(runner)
	for i := range cfg.Endpoints {
		endpoint := &cfg.Endpoints[i]
		endpointPath := basePath.Child("endpoints").Index(i)
		errs = append(errs, validateEndpoint(endpoint, endpointPath, tracker, i)...)
	}

	return errs
}

func validateEndpoint(
	endpoint *endpoints.Endpoint,
	endpointPath *field.Path,
	tracker *endpointNameTracker,
	endpointIdx int,
) field.ErrorList {
	var errs field.ErrorList

	errs = append(errs, validateEndpointServices(endpoint, endpointPath, tracker, endpointIdx)...)
	errs = append(errs, validateEndpointExternalAccess(endpoint, endpointPath, tracker, endpointIdx)...)

	return errs
}

func validateEndpointServices(
	endpoint *endpoints.Endpoint,
	endpointPath *field.Path,
	tracker *endpointNameTracker,
	endpointIdx int,
) field.ErrorList {
	var errs field.ErrorList

	servicesPath := endpointPath.Child("services")
	if len(endpoint.Services) == 0 {
		if endpoint.ExternalAccess == nil {
			errs = append(errs, field.Required(
				endpointPath,
				"at least one service or externalAccess is required",
			))
		}
		return errs
	}
	if len(endpoint.Services) > 1 {
		errs = append(errs, field.TooMany(servicesPath, len(endpoint.Services), 1))
	}

	for j := range endpoint.Services {
		svc := &endpoint.Services[j]
		svcPath := servicesPath.Index(j)

		if svc.Ref == nil && svc.Create == nil {
			errs = append(errs, field.Required(svcPath, "either ref or create must be set"))
			continue
		}
		if svc.Ref != nil && svc.Create != nil {
			errs = append(errs, field.Invalid(
				svcPath, "both ref and create",
				"ref and create are mutually exclusive",
			))
		}

		if svc.Ref != nil {
			errs = append(errs, validateEndpointServiceRef(svc, svcPath)...)
		}
		if svc.Create != nil {
			errs = append(errs, validateEndpointServiceCreate(svc, svcPath)...)
			if isMetricsService(svc.Create) {
				errs = append(errs, validateMetricsService(
					endpoint, svc, tracker.attachedWorkloadName, endpointPath, svcPath,
				)...)
			}
		}

		if svc.AuthMode != "" && svc.AuthMode != endpoints.AuthModeStandalone {
			errs = append(errs, field.NotSupported(
				svcPath.Child("authMode"),
				string(svc.AuthMode),
				[]string{string(endpoints.AuthModeStandalone)},
			))
		}

		if svc.Auth != nil {
			errs = append(errs, validateEndpointAuthConfig(svc.Auth, svcPath.Child("auth"))...)
			if endpoint.ExternalAccess == nil {
				errs = append(errs, field.Required(
					endpointPath.Child("externalAccess"),
					"externalAccess is required when service auth is enabled",
				))
			}
		}

		name := endpointServiceName(svc)
		errs = append(errs, tracker.recordService(name, endpointIdx, svcPath)...)
		// Reserve the generated auth-proxy Service name so it cannot collide with another endpoint's declared Service.
		if name != "" && svc.Auth != nil {
			errs = append(errs, tracker.recordService(
				getEndpointAuthProxyName(name), endpointIdx,
				svcPath.Child("auth"),
			)...)
		}
	}

	return errs
}

func validateMetricsService(
	endpoint *endpoints.Endpoint,
	svc *endpoints.Service,
	workloadName string,
	endpointPath *field.Path,
	svcPath *field.Path,
) field.ErrorList {
	var errs field.ErrorList
	createPath := svcPath.Child("create")
	labelsPath := createPath.Child("metadata", "labels")

	expectedName := workloadName + "-metrics"
	if svc.Create.Name != expectedName {
		errs = append(errs, field.Invalid(
			createPath.Child("metadata", "name"),
			svc.Create.Name,
			fmt.Sprintf("metrics Service name must be %q", expectedName),
		))
	}

	labels := svc.Create.Labels
	if labels[constants.LabelWorkloadKind] != constants.KindInferenceWorkload {
		errs = append(errs, field.Invalid(
			labelsPath.Key(constants.LabelWorkloadKind),
			labels[constants.LabelWorkloadKind],
			fmt.Sprintf("metrics Service must target %s", constants.KindInferenceWorkload),
		))
	}
	nimWorkload := labels[constants.LabelNIMWorkload] == "true"
	nimMetricsService := labels[constants.LabelNIMWorkloadMetricsService] == "true"
	vllmWorkload := labels[constants.LabelVLLMWorkload] == "true"
	vllmMetricsService := labels[constants.LabelVLLMWorkloadMetricsService] == "true"
	validNIMPair := nimWorkload && nimMetricsService && !vllmWorkload && !vllmMetricsService
	validVLLMPair := vllmWorkload && vllmMetricsService && !nimWorkload && !nimMetricsService
	if !validNIMPair && !validVLLMPair {
		errs = append(errs, field.Invalid(
			labelsPath,
			labels,
			"metrics Service must contain exactly one consistent NIM or vLLM engine-label pair",
		))
	}

	selector := svc.Create.Spec.Selector
	selectorPath := createPath.Child("spec", "selector")
	if len(selector) == 0 {
		errs = append(errs, field.Invalid(
			selectorPath,
			selector,
			"metrics Service must have at least one selector",
		))
	}

	byWorkload, found := selector[constants.LabelWorkloadName]
	if !found {
		errs = append(errs, field.Invalid(
			selectorPath,
			selector,
			fmt.Sprintf("missing Service selector %s=%s", constants.LabelWorkloadName, workloadName),
		))
	} else if byWorkload != workloadName {
		errs = append(errs, field.Invalid(
			selectorPath,
			selector,
			fmt.Sprintf("metrics Service selector must be %s=%s", constants.LabelWorkloadName, workloadName),
		))
	}

	if svc.Ref != nil {
		errs = append(errs, field.Forbidden(
			svcPath.Child("ref"),
			fmt.Sprintf("%s=true services must use create, not ref", constants.LabelWorkloadMetricsService),
		))
	}
	if svc.Create.Spec.Type != corev1.ServiceTypeClusterIP {
		errs = append(errs, field.Invalid(
			createPath.Child("spec", "type"),
			svc.Create.Spec.Type,
			fmt.Sprintf("%s=true services must explicitly use ClusterIP", constants.LabelWorkloadMetricsService),
		))
	}

	portsPath := createPath.Child("spec", "ports")
	if len(svc.Create.Spec.Ports) != 1 {
		errs = append(errs, field.Invalid(
			portsPath,
			len(svc.Create.Spec.Ports),
			fmt.Sprintf("%s=true services must define exactly one metrics port", constants.LabelWorkloadMetricsService),
		))
	} else {
		port := svc.Create.Spec.Ports[0]
		portPath := portsPath.Index(0)
		if port.Name != "metrics" {
			errs = append(errs, field.Invalid(
				portPath.Child("name"), port.Name, `metrics Service port name must be "metrics"`,
			))
		}
		if port.Protocol != corev1.ProtocolTCP {
			errs = append(errs, field.Invalid(
				portPath.Child("protocol"), port.Protocol, "metrics Service port protocol must be TCP",
			))
		}
		if port.Port < 1 || port.Port > 65535 {
			errs = append(errs, field.Invalid(
				portPath.Child("port"), port.Port, "metrics Service port must be between 1 and 65535",
			))
		}
		if port.TargetPort.IntVal == 0 && port.TargetPort.StrVal == "" {
			errs = append(errs, field.Required(
				portPath.Child("targetPort"), "metrics Service targetPort is required",
			))
		}
	}

	if svc.Auth != nil || svc.AuthMode != "" {
		errs = append(errs, field.Forbidden(
			svcPath.Child("auth"),
			fmt.Sprintf("%s=true services must not configure authentication", constants.LabelWorkloadMetricsService),
		))
	}
	if endpoint.ExternalAccess != nil {
		errs = append(errs, field.Forbidden(
			endpointPath.Child("externalAccess"),
			fmt.Sprintf("%s=true services must not configure externalAccess", constants.LabelWorkloadMetricsService),
		))
	}
	return errs
}

func validateEndpointServiceRef(svc *endpoints.Service, svcPath *field.Path) field.ErrorList {
	var errs field.ErrorList
	if svc.Auth == nil {
		errs = append(errs, field.Required(
			svcPath.Child("auth"),
			"auth is required when using service reference; "+
				"if you don't need authentication, create a plain Ingress resource instead of using this plugin",
		))
	}
	if svc.Ref.Name == "" {
		errs = append(errs, field.Required(svcPath.Child("ref", "name"), "name is required"))
	}
	if svc.Ref.Port == 0 {
		errs = append(errs, field.Required(svcPath.Child("ref", "port"), "port is required"))
	}
	return errs
}

func validateEndpointServiceCreate(svc *endpoints.Service, svcPath *field.Path) field.ErrorList {
	var errs field.ErrorList
	// Name is required: the plugin preserves the spec name verbatim so producers can reference it from other endpoints' service refs.
	if svc.Create.Name == "" {
		errs = append(errs, field.Required(
			svcPath.Child("create", "metadata", "name"),
			"metadata.name is required",
		))
	}
	if len(svc.Create.Spec.Selector) == 0 {
		errs = append(errs, field.Required(
			svcPath.Child("create", "spec", "selector"),
			"spec.selector is required",
		))
	}
	if len(svc.Create.Spec.Ports) == 0 {
		errs = append(errs, field.Required(
			svcPath.Child("create", "spec", "ports"),
			"spec.ports is required",
		))
	}
	return errs
}

func validateEndpointAuthConfig(auth *endpoints.AuthConfig, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList

	if auth.AuthorizationType == "" {
		errs = append(errs, field.Required(
			basePath.Child("authorizationType"),
			"authorizationType is required when auth is enabled",
		))
	}

	switch auth.AuthorizationType {
	case endpoints.AuthorizationTypeAuthorizedUsers:
		if auth.AuthorizedUsers == "" {
			errs = append(errs, field.Required(
				basePath.Child("authorizedUsers"),
				"authorizedUsers is required when authorizationType is authorizedUsers",
			))
		}
		if auth.AuthorizedGroups != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedGroups"), auth.AuthorizedGroups,
				"authorizedGroups must not be set when authorizationType is authorizedUsers",
			))
		}
	case endpoints.AuthorizationTypeAuthorizedGroups:
		if auth.AuthorizedGroups == "" {
			errs = append(errs, field.Required(
				basePath.Child("authorizedGroups"),
				"authorizedGroups is required when authorizationType is authorizedGroups",
			))
		} else if _, err := workloadbuilders.SanitizeAuthorizedGroups(auth.AuthorizedGroups); err != nil {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedGroups"), auth.AuthorizedGroups, err.Error(),
			))
		}
		if auth.AuthorizedUsers != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedUsers"), auth.AuthorizedUsers,
				"authorizedUsers must not be set when authorizationType is authorizedGroups",
			))
		}
	case endpoints.AuthorizationTypeAuthenticatedUsers:
		if auth.AuthorizedUsers != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedUsers"), auth.AuthorizedUsers,
				"authorizedUsers must not be set when authorizationType is authenticatedUsers",
			))
		}
		if auth.AuthorizedGroups != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedGroups"), auth.AuthorizedGroups,
				"authorizedGroups must not be set when authorizationType is authenticatedUsers",
			))
		}
	default:
		// Empty handled above; reject non-empty unknown values so admission catches typos.
		if auth.AuthorizationType != "" {
			errs = append(errs, field.NotSupported(
				basePath.Child("authorizationType"),
				string(auth.AuthorizationType),
				[]string{
					string(endpoints.AuthorizationTypeAuthorizedUsers),
					string(endpoints.AuthorizationTypeAuthorizedGroups),
					string(endpoints.AuthorizationTypeAuthenticatedUsers),
				},
			))
		}
	}

	return errs
}

func validateEndpointExternalAccess(
	endpoint *endpoints.Endpoint,
	endpointPath *field.Path,
	tracker *endpointNameTracker,
	endpointIdx int,
) field.ErrorList {
	var errs field.ErrorList

	if endpoint.ExternalAccess == nil {
		return errs
	}
	ea := endpoint.ExternalAccess
	eaPath := endpointPath.Child("externalAccess")

	if ea.Ingress == nil && ea.Route == nil && ea.HTTPRoute == nil {
		errs = append(errs, field.Required(
			eaPath,
			"one of externalAccess.ingress, externalAccess.route, or externalAccess.httpRoute must be set",
		))
		return errs
	}
	if ea.Ingress != nil && ea.Route != nil {
		errs = append(errs, field.Invalid(
			eaPath, "both ingress and route",
			"externalAccess.ingress and externalAccess.route are mutually exclusive",
		))
	}

	if ea.Ingress != nil {
		errs = append(errs, validateEndpointIngress(endpoint, eaPath.Child("ingress"))...)
		errs = append(errs, tracker.recordIngress(
			getResourceName(ea.Ingress.Name, tracker.runner), endpointIdx,
			eaPath.Child("ingress", "metadata", "name"),
		)...)
	}
	if ea.Route != nil {
		if ea.Route.Spec.Host == "" {
			errs = append(errs, field.Required(
				eaPath.Child("route", "spec", "host"), "host is required",
			))
		}
		errs = append(errs, tracker.recordRoute(
			getResourceName(ea.Route.Name, tracker.runner), endpointIdx,
			eaPath.Child("route", "metadata", "name"),
		)...)
	}
	if ea.HTTPRoute != nil {
		if len(ea.HTTPRoute.Spec.Rules) == 0 {
			errs = append(errs, field.Required(
				eaPath.Child("httpRoute", "spec", "rules"),
				"at least one HTTPRoute rule is required",
			))
		}
		errs = append(errs, tracker.recordHTTPRoute(
			getResourceName(ea.HTTPRoute.Name, tracker.runner), endpointIdx,
			eaPath.Child("httpRoute", "metadata", "name"),
		)...)
	}

	return errs
}

func validateEndpointIngress(endpoint *endpoints.Endpoint, ingressPath *field.Path) field.ErrorList {
	var errs field.ErrorList

	ingress := endpoint.ExternalAccess.Ingress
	if len(ingress.Spec.Rules) == 0 {
		errs = append(errs, field.Required(
			ingressPath.Child("spec", "rules"),
			"at least one ingress rule is required",
		))
		return errs
	}

	for i, rule := range ingress.Spec.Rules {
		rulePath := ingressPath.Child("spec", "rules").Index(i)
		if rule.HTTP == nil {
			errs = append(errs, field.Required(
				rulePath.Child("http"),
				"http configuration is required for ingress rules",
			))
		} else if len(rule.HTTP.Paths) == 0 {
			errs = append(errs, field.Required(
				rulePath.Child("http", "paths"),
				"at least one path is required",
			))
		}
	}

	errs = append(errs, validateEndpointIngressBackends(endpoint, ingressPath)...)
	return errs
}

func validateEndpointIngressBackends(endpoint *endpoints.Endpoint, ingressPath *field.Path) field.ErrorList {
	var errs field.ErrorList

	if endpoint.ExternalAccess == nil || endpoint.ExternalAccess.Ingress == nil {
		return errs
	}
	svc := firstService(endpoint)
	if svc == nil {
		return errs
	}
	expectedName := endpointServiceName(svc)
	expectedPort := endpointServicePort(svc)
	if expectedName == "" {
		return errs
	}

	ingress := endpoint.ExternalAccess.Ingress
	for i, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for j, path := range rule.HTTP.Paths {
			pathBase := ingressPath.Child("spec", "rules").Index(i).
				Child("http", "paths").Index(j).Child("backend", "service")
			errs = append(errs, validateIngressBackendPath(path, pathBase, expectedName, expectedPort)...)
		}
	}
	return errs
}

func validateIngressBackendPath(
	path networkingv1.HTTPIngressPath,
	pathBase *field.Path,
	expectedName string,
	expectedPort int32,
) field.ErrorList {
	var errs field.ErrorList
	if path.Backend.Service == nil {
		errs = append(errs, field.Required(
			pathBase,
			"backend service must be specified and must match the endpoint's service",
		))
		return errs
	}
	if path.Backend.Service.Name != expectedName {
		errs = append(errs, field.Invalid(
			pathBase.Child("name"), path.Backend.Service.Name,
			fmt.Sprintf("must match endpoint service name %q", expectedName),
		))
	}
	if expectedPort != 0 && path.Backend.Service.Port.Number != expectedPort {
		errs = append(errs, field.Invalid(
			pathBase.Child("port", "number"), path.Backend.Service.Port.Number,
			fmt.Sprintf("must match endpoint service port %d", expectedPort),
		))
	}
	return errs
}

func endpointServiceName(svc *endpoints.Service) string {
	if svc == nil {
		return ""
	}
	if svc.Create != nil {
		return svc.Create.Name
	}
	if svc.Ref != nil {
		return svc.Ref.Name
	}
	return ""
}

func endpointServicePort(svc *endpoints.Service) int32 {
	if svc == nil {
		return 0
	}
	if svc.Create != nil && len(svc.Create.Spec.Ports) > 0 {
		return svc.Create.Spec.Ports[0].Port
	}
	if svc.Ref != nil {
		return svc.Ref.Port
	}
	return 0
}
