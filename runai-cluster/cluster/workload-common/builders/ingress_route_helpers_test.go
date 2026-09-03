/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnrichIngressWithDefaults(t *testing.T) {
	t.Run("sets ingressClassName from config when not specified", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarIngressClass, "nginx")
		defer os.Unsetenv(workloadconfig.EnvVarIngressClass)

		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		assert.NotNil(t, enriched.Spec.IngressClassName)
		assert.Equal(t, "nginx", *enriched.Spec.IngressClassName)
	})

	t.Run("preserves user-specified ingressClassName", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarIngressClass, "nginx")
		defer os.Unsetenv(workloadconfig.EnvVarIngressClass)

		userClass := "traefik"
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &userClass,
			},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		assert.Equal(t, "traefik", *enriched.Spec.IngressClassName)
	})

	t.Run("adds body size annotation if not present", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		assert.Contains(t, enriched.Annotations, constants.IngressBodySizeAnnotationKey)
		assert.Equal(t, constants.IngressBodySizeAnnotationDefaultValue, enriched.Annotations[constants.IngressBodySizeAnnotationKey])
	})

	t.Run("preserves user-specified body size annotation", func(t *testing.T) {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					constants.IngressBodySizeAnnotationKey: "100m",
				},
			},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		assert.Equal(t, "100m", enriched.Annotations[constants.IngressBodySizeAnnotationKey])
	})

	t.Run("merges default annotations while preserving user annotations", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarIngressAnnotations, `{"default-key":"default-value","shared-key":"default-shared"}`)
		defer os.Unsetenv(workloadconfig.EnvVarIngressAnnotations)

		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					"user-key":   "user-value",
					"shared-key": "user-shared",
				},
			},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		// User annotations take precedence
		assert.Equal(t, "user-value", enriched.Annotations["user-key"])
		assert.Equal(t, "user-shared", enriched.Annotations["shared-key"])
		// Default annotations are added if not present
		assert.Equal(t, "default-value", enriched.Annotations["default-key"])
	})

	t.Run("returns nil for nil input", func(t *testing.T) {
		enriched := EnrichIngressWithDefaults(nil)
		assert.Nil(t, enriched)
	})
}

func TestEnrichRouteWithDefaults(t *testing.T) {
	t.Run("sets default TLS when not specified", func(t *testing.T) {
		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		assert.NotNil(t, enriched.Spec.TLS)
		assert.Equal(t, openshiftv1.TLSTerminationEdge, enriched.Spec.TLS.Termination)
		assert.Equal(t, openshiftv1.InsecureEdgeTerminationPolicyRedirect, enriched.Spec.TLS.InsecureEdgeTerminationPolicy)
	})

	t.Run("preserves user-specified TLS", func(t *testing.T) {
		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
				TLS: &openshiftv1.TLSConfig{
					Termination: openshiftv1.TLSTerminationPassthrough,
				},
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		assert.Equal(t, openshiftv1.TLSTerminationPassthrough, enriched.Spec.TLS.Termination)
	})

	t.Run("uses passthrough termination when configured by env", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, workloadconfig.OpenshiftRouteTerminationPolicyPassthrough)
		defer os.Unsetenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy)

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		assert.NotNil(t, enriched.Spec.TLS)
		assert.Equal(t, openshiftv1.TLSTerminationPassthrough, enriched.Spec.TLS.Termination)
		assert.Equal(t, openshiftv1.InsecureEdgeTerminationPolicyRedirect, enriched.Spec.TLS.InsecureEdgeTerminationPolicy)
	})

	t.Run("falls back to edge for unknown termination policy", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, "reencrypt")
		defer os.Unsetenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy)

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		assert.Equal(t, openshiftv1.TLSTerminationEdge, enriched.Spec.TLS.Termination)
	})

	t.Run("trims subdomain when enabled", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarSubdomainSupport, "true")
		defer os.Unsetenv(workloadconfig.EnvVarSubdomainSupport)

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "runai-test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		assert.NotContains(t, enriched.Spec.Host, workloadconfig.RunaiTrimStr)
	})

	t.Run("merges default annotations while preserving user annotations", func(t *testing.T) {
		os.Setenv(workloadconfig.EnvVarIngressAnnotations, `{"default-key":"default-value"}`)
		defer os.Unsetenv(workloadconfig.EnvVarIngressAnnotations)

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					"user-key": "user-value",
				},
			},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		assert.Equal(t, "user-value", enriched.Annotations["user-key"])
		assert.Equal(t, "default-value", enriched.Annotations["default-key"])
	})

	t.Run("returns nil for nil input", func(t *testing.T) {
		enriched := EnrichRouteWithDefaults(nil)
		assert.Nil(t, enriched)
	})
}
