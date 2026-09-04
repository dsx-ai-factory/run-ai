/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"fmt"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// WebhookValidate dispatches admission validation on the configured input shape; see README "Input shapes".
func (p *Plugin) WebhookValidate(
	ctx context.Context,
	existingRunner *dromosv1alpha1.WorkloadRunner,
	desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	shape, cfg, err := classifyInput(desiredRunner)
	if err != nil {
		return nil, field.ErrorList{field.Forbidden(
			field.NewPath("metadata", "annotations").Key(endpoints.Annotation),
			err.Error(),
		)}
	}

	switch shape {
	case inputLegacy:
		return p.webhookValidateLegacy(ctx, existingRunner, desiredRunner)
	case inputAnnotation:
		return nil, validateEndpointsConfig(desiredRunner, cfg)
	default:
		return nil, nil
	}
}

// Preserved verbatim so spec.instructions.networking admission semantics are unchanged.
func (p *Plugin) webhookValidateLegacy(
	_ context.Context,
	_ *dromosv1alpha1.WorkloadRunner,
	desiredRunner *dromosv1alpha1.WorkloadRunner,
) (admission.Warnings, field.ErrorList) {
	config := extractNetworkingConfig(desiredRunner)
	if config == nil {
		return nil, nil
	}

	var errs field.ErrorList
	basePath := field.NewPath("spec", "instructions", "networking")

	errs = append(errs, validateServices(config, basePath)...)
	errs = append(errs, validateExternalAccess(config, basePath)...)

	return nil, errs
}

func validateServices(config *dromosv1alpha1.NetworkingConfig, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList

	servicesCount := len(config.Services)
	if servicesCount == 0 {
		return errs
	}

	if servicesCount > 1 {
		errs = append(errs, field.TooMany(
			basePath.Child("services"),
			servicesCount,
			1))
	}

	svc := config.Services[0]
	svcPath := basePath.Child("services").Index(0)

	// Validate ref/create mutual exclusivity
	if svc.Ref == nil && svc.Create == nil {
		errs = append(errs, field.Required(
			svcPath,
			"either ref or create must be set"))
	}
	if svc.Ref != nil && svc.Create != nil {
		errs = append(errs, field.Invalid(
			svcPath,
			"both ref and create",
			"ref and create are mutually exclusive"))
	}

	// Validate ref requires auth
	if svc.Ref != nil && svc.Auth == nil {
		errs = append(errs, field.Required(
			svcPath.Child("auth"),
			"auth is required when using service reference; if you don't need authentication, create a plain Ingress resource instead of using this plugin"))
	}

	// Validate ref fields
	if svc.Ref != nil {
		if svc.Ref.Name == "" {
			errs = append(errs, field.Required(
				svcPath.Child("ref", "name"),
				"name is required"))
		}
		if svc.Ref.Port == 0 {
			errs = append(errs, field.Required(
				svcPath.Child("ref", "port"),
				"port is required"))
		}
	}

	// Validate create fields
	if svc.Create != nil {
		if len(svc.Create.Spec.Selector) == 0 {
			errs = append(errs, field.Required(
				svcPath.Child("create", "spec", "selector"),
				"spec.selector is required"))
		}
		if len(svc.Create.Spec.Ports) == 0 {
			errs = append(errs, field.Required(
				svcPath.Child("create", "spec", "ports"),
				"spec.ports is required"))
		}
	}

	// Validate auth configuration if present
	if svc.Auth != nil {
		errs = append(errs, validateAuthConfig(svc.Auth, svcPath.Child("auth"))...)

		// Validate auth requires externalAccess
		if config.ExternalAccess == nil {
			errs = append(errs, field.Required(
				basePath.Child("externalAccess"),
				"externalAccess configuration is required when auth is enabled"))
		}
	}

	return errs
}

func validateAuthConfig(auth *dromosv1alpha1.NetworkingAuthConfig, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList

	if auth.AuthorizationType == "" {
		errs = append(errs, field.Required(
			basePath.Child("authorizationType"),
			"authorizationType is required when auth is enabled"))
	}

	if auth.AuthorizationType == dromosv1alpha1.AuthorizationTypeAuthorizedUsers {
		if auth.AuthorizedUsers == "" {
			errs = append(errs, field.Required(
				basePath.Child("authorizedUsers"),
				"authorizedUsers is required when authorizationType is authorizedUsers"))
		}
		if auth.AuthorizedGroups != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedGroups"),
				auth.AuthorizedGroups,
				"authorizedGroups must not be set when authorizationType is authorizedUsers"))
		}
	}

	if auth.AuthorizationType == dromosv1alpha1.AuthorizationTypeAuthorizedGroups {
		if auth.AuthorizedGroups == "" {
			errs = append(errs, field.Required(
				basePath.Child("authorizedGroups"),
				"authorizedGroups is required when authorizationType is authorizedGroups"))
		} else if _, err := workloadbuilders.SanitizeAuthorizedGroups(auth.AuthorizedGroups); err != nil {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedGroups"),
				auth.AuthorizedGroups,
				err.Error()))
		}
		if auth.AuthorizedUsers != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedUsers"),
				auth.AuthorizedUsers,
				"authorizedUsers must not be set when authorizationType is authorizedGroups"))
		}
	}

	if auth.AuthorizationType == dromosv1alpha1.AuthorizationTypeAuthenticatedUsers {
		if auth.AuthorizedUsers != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedUsers"),
				auth.AuthorizedUsers,
				"authorizedUsers must not be set when authorizationType is authenticatedUsers"))
		}
		if auth.AuthorizedGroups != "" {
			errs = append(errs, field.Invalid(
				basePath.Child("authorizedGroups"),
				auth.AuthorizedGroups,
				"authorizedGroups must not be set when authorizationType is authenticatedUsers"))
		}
	}

	return errs
}

func validateExternalAccess(config *dromosv1alpha1.NetworkingConfig, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList

	if config.ExternalAccess == nil {
		return errs
	}

	ea := config.ExternalAccess

	if ea.Ingress == nil && ea.Route == nil && ea.HTTPRoute == nil {
		errs = append(errs, field.Required(
			basePath.Child("externalAccess"),
			"one of externalAccess.ingress, externalAccess.route, or externalAccess.httpRoute must be set"))
	}
	if ea.Ingress != nil && ea.Route != nil {
		errs = append(errs, field.Invalid(
			basePath.Child("externalAccess"),
			"both ingress and route",
			"externalAccess.ingress and externalAccess.route are mutually exclusive"))
	}

	if ea.Ingress != nil {
		if len(ea.Ingress.Spec.Rules) == 0 {
			errs = append(errs, field.Required(
				basePath.Child("externalAccess", "ingress", "spec", "rules"),
				"at least one ingress rule is required"))
		}

		for i, rule := range ea.Ingress.Spec.Rules {
			rulePath := basePath.Child("externalAccess", "ingress", "spec", "rules").Index(i)

			if rule.HTTP == nil {
				errs = append(errs, field.Required(
					rulePath.Child("http"),
					"http configuration is required for ingress rules"))
			} else if len(rule.HTTP.Paths) == 0 {
				errs = append(errs, field.Required(
					rulePath.Child("http", "paths"),
					"at least one path is required"))
			}
		}

		errs = append(errs, validateIngressBackends(config, basePath)...)
	}

	if ea.Route != nil && ea.Route.Spec.Host == "" {
		errs = append(errs, field.Required(
			basePath.Child("externalAccess", "route", "spec", "host"),
			"host is required"))
	}

	if ea.HTTPRoute != nil && len(ea.HTTPRoute.Spec.Rules) == 0 {
		errs = append(errs, field.Required(
			basePath.Child("externalAccess", "httpRoute", "spec", "rules"),
			"at least one HTTPRoute rule is required"))
	}

	return errs
}

func validateIngressBackends(config *dromosv1alpha1.NetworkingConfig, basePath *field.Path) field.ErrorList {
	var errs field.ErrorList

	if config.ExternalAccess == nil || config.ExternalAccess.Ingress == nil {
		return errs
	}

	if !hasService(config) {
		return errs
	}

	info := getWorkloadServiceInfo(config, "")
	if info.Name == "" {
		return errs // Service not properly configured, will be caught elsewhere
	}

	ingress := config.ExternalAccess.Ingress

	// Validate ALL paths in ALL rules have backends matching the service
	for i, rule := range ingress.Spec.Rules {
		if rule.HTTP != nil {
			for j, path := range rule.HTTP.Paths {
				pathBasePath := basePath.Child("externalAccess", "ingress", "spec", "rules").Index(i).
					Child("http", "paths").Index(j).Child("backend", "service")

				if path.Backend.Service == nil {
					errs = append(errs, field.Required(
						pathBasePath,
						"backend service must be specified and must match the service"))
					continue
				}

				if path.Backend.Service.Name != info.Name {
					errs = append(errs, field.Invalid(
						pathBasePath.Child("name"),
						path.Backend.Service.Name,
						fmt.Sprintf("must match service name '%s'; all ingress paths must route to the same service", info.Name)))
				}

				if path.Backend.Service.Port.Number != info.Port {
					errs = append(errs, field.Invalid(
						pathBasePath.Child("port", "number"),
						path.Backend.Service.Port.Number,
						fmt.Sprintf("must match service port %d; all ingress paths must route to the same service", info.Port)))
				}
			}
		}
	}

	return errs
}
