/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("EnrichIngressWithDefaults", func() {
	It("sets ingressClassName from config when not specified", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarIngressClass, "nginx")

		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		Expect(enriched.Spec.IngressClassName).NotTo(BeNil())
		Expect(*enriched.Spec.IngressClassName).To(Equal("nginx"))
	})

	It("preserves user-specified ingressClassName", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarIngressClass, "nginx")

		userClass := "traefik"
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &userClass,
			},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		Expect(*enriched.Spec.IngressClassName).To(Equal("traefik"))
	})

	It("adds body size annotation if not present", func() {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		Expect(enriched.Annotations).To(HaveKey(constants.IngressBodySizeAnnotationKey))
		Expect(enriched.Annotations[constants.IngressBodySizeAnnotationKey]).To(Equal(constants.IngressBodySizeAnnotationDefaultValue))
	})

	It("preserves user-specified body size annotation", func() {
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					constants.IngressBodySizeAnnotationKey: "100m",
				},
			},
		}

		enriched := EnrichIngressWithDefaults(ingress)

		Expect(enriched.Annotations[constants.IngressBodySizeAnnotationKey]).To(Equal("100m"))
	})

	It("merges default annotations while preserving user annotations", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarIngressAnnotations, `{"default-key":"default-value","shared-key":"default-shared"}`)

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
		Expect(enriched.Annotations["user-key"]).To(Equal("user-value"))
		Expect(enriched.Annotations["shared-key"]).To(Equal("user-shared"))
		// Default annotations are added if not present
		Expect(enriched.Annotations["default-key"]).To(Equal("default-value"))
	})

	It("returns nil for nil input", func() {
		enriched := EnrichIngressWithDefaults(nil)
		Expect(enriched).To(BeNil())
	})
})

var _ = Describe("EnrichRouteWithDefaults", func() {
	BeforeEach(func() {
		GinkgoT().Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, "")
	})

	It("sets default TLS when not specified", func() {
		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		Expect(enriched.Spec.TLS).NotTo(BeNil())
		Expect(enriched.Spec.TLS.Termination).To(Equal(openshiftv1.TLSTerminationEdge))
		Expect(enriched.Spec.TLS.InsecureEdgeTerminationPolicy).To(Equal(openshiftv1.InsecureEdgeTerminationPolicyRedirect))
	})

	It("preserves user-specified TLS", func() {
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

		Expect(enriched.Spec.TLS.Termination).To(Equal(openshiftv1.TLSTerminationPassthrough))
	})

	It("uses passthrough termination when configured by env", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, workloadconfig.OpenshiftRouteTerminationPolicyPassthrough)

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		Expect(enriched.Spec.TLS).NotTo(BeNil())
		Expect(enriched.Spec.TLS.Termination).To(Equal(openshiftv1.TLSTerminationPassthrough))
		Expect(enriched.Spec.TLS.InsecureEdgeTerminationPolicy).To(Equal(openshiftv1.InsecureEdgeTerminationPolicyRedirect))
	})

	It("falls back to edge for unknown termination policy", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, "reencrypt")

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		Expect(enriched.Spec.TLS.Termination).To(Equal(openshiftv1.TLSTerminationEdge))
	})

	It("trims subdomain when enabled", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarSubdomainSupport, "true")

		route := &openshiftv1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Spec: openshiftv1.RouteSpec{
				Host: "runai-test.example.com",
			},
		}

		enriched := EnrichRouteWithDefaults(route)

		Expect(enriched.Spec.Host).NotTo(ContainSubstring(workloadconfig.RunaiTrimStr))
	})

	It("merges default annotations while preserving user annotations", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarIngressAnnotations, `{"default-key":"default-value"}`)

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

		Expect(enriched.Annotations["user-key"]).To(Equal("user-value"))
		Expect(enriched.Annotations["default-key"]).To(Equal("default-value"))
	})

	It("returns nil for nil input", func() {
		enriched := EnrichRouteWithDefaults(nil)
		Expect(enriched).To(BeNil())
	})
})
