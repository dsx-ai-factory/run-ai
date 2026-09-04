/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Ingress", func() {

	var runaiConfig *runaiv1.RunaiConfig
	var ingress *networkingv1.Ingress

	BeforeEach(func() {
		runaiConfig = utils.GetDefaultRunaiConfig()
		ingress = &networkingv1.Ingress{}
	})

	Context("UseCertManager", func() {
		It("Parse url path", func() {
			runaiConfig.Spec.Internal.Global.ClusterURL = "https://automationtenant-35-223-190-115.webredirect.org"
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.Rules[0].HTTP.Paths).To(HaveLen(3))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/researcher"))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[1].Path).To(Equal("/cli"))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[2].Path).To(Equal("/oauth2"))

		})

		It("Parse url host", func() {
			runaiConfig.Spec.Internal.Global.ClusterURL = "https://automationtenant-35-223-190-115.webredirect.org"
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.Rules[0].Host).To(Equal("automationtenant-35-223-190-115.webredirect.org"))
		})
	})

	Context("TlsSecret", func() {
		It("Use correct tls secret", func() {
			runaiConfig.Spec.Internal.Global.ClusterURL = "https://www.example.com"
			runaiConfig.Spec.ResearcherService.Ingress.TlsSecret = "runai-cluster-tls-secret"
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.TLS[0].Hosts[0]).To(Equal("www.example.com"))
			Expect(ingress.Spec.TLS[0].SecretName).To(Equal("runai-cluster-tls-secret"))
		})

		It("should not use a secret if TLS is disabled", func() {
			runaiConfig.Spec.ResearcherService.Ingress.TLSEnabled = pointer.Bool(false)
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.TLS).To(BeEmpty())
		})
	})

	Context("Ingress Class Configuration", func() {
		It("should default to nginx when no ingress class is specified", func() {
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
			Expect(*ingress.Spec.IngressClassName).To(Equal("nginx"))
			Expect(ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"]).To(Equal("nginx"))
		})

		It("should use ResearcherService ingress class when specified", func() {
			runaiConfig.Spec.ResearcherService.Ingress.Class = "researcher-ingress-class"
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
			Expect(*ingress.Spec.IngressClassName).To(Equal("researcher-ingress-class"))
			Expect(ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"]).To(Equal("researcher-ingress-class"))
		})

		It("should use global ingress class when specified", func() {
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Class: "global-ingress-class",
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
			Expect(*ingress.Spec.IngressClassName).To(Equal("global-ingress-class"))
			Expect(ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"]).To(Equal("global-ingress-class"))
		})

		It("should prefer ResearcherService ingress class over global", func() {
			runaiConfig.Spec.ResearcherService.Ingress.Class = "researcher-ingress-class"
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Class: "global-ingress-class",
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
			Expect(*ingress.Spec.IngressClassName).To(Equal("researcher-ingress-class"))
			Expect(ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"]).To(Equal("researcher-ingress-class"))
		})

		It("should fallback to global ingress class when ResearcherService class is empty", func() {
			runaiConfig.Spec.ResearcherService.Ingress.Class = ""
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Class: "global-ingress-class",
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
			Expect(*ingress.Spec.IngressClassName).To(Equal("global-ingress-class"))
			Expect(ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"]).To(Equal("global-ingress-class"))
		})
	})

	Context("Custom annotations", func() {
		It("should have default nginx annotations", func() {
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations).To(Not(BeEmpty()))
		})

		It("should have custom annotations from ResearcherService config", func() {
			runaiConfig.Spec.ResearcherService.Ingress.Annotations = map[string]string{
				"test": "test",
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations).To(Not(BeEmpty()))
			Expect(ingress.ObjectMeta.Annotations["test"]).To(Equal("test"))
		})

		It("should have custom annotations from Global config", func() {
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Annotations: map[string]string{
					"global-key": "global-value",
				},
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations).To(Not(BeEmpty()))
			Expect(ingress.ObjectMeta.Annotations["global-key"]).To(Equal("global-value"))
		})

		It("should merge annotations from both ResearcherService and Global configs", func() {
			runaiConfig.Spec.ResearcherService.Ingress.Annotations = map[string]string{
				"researcher-key": "researcher-value",
			}
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Annotations: map[string]string{
					"global-key": "global-value",
				},
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["researcher-key"]).To(Equal("researcher-value"))
			Expect(ingress.ObjectMeta.Annotations["global-key"]).To(Equal("global-value"))
		})

		It("should give precedence to ResearcherService annotations over Global annotations", func() {
			runaiConfig.Spec.ResearcherService.Ingress.Annotations = map[string]string{
				"shared-key": "from-researcher",
			}
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Annotations: map[string]string{
					"shared-key": "from-global",
				},
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["shared-key"]).To(Equal("from-researcher"))
		})

		It("should not share the annotations map reference (deep copy verification)", func() {
			originalAnnotations := map[string]string{
				"original-key": "original-value",
			}
			runaiConfig.Spec.ResearcherService.Ingress.Annotations = originalAnnotations

			initIngress(runaiConfig, ingress)

			// Modify the ingress annotations
			ingress.ObjectMeta.Annotations["new-key"] = "new-value"

			// Original map should not be affected
			Expect(originalAnnotations).To(HaveLen(1))
			Expect(originalAnnotations).ToNot(HaveKey("new-key"))
			Expect(originalAnnotations["original-key"]).To(Equal("original-value"))
		})
	})

	Context("Enabled gating on ResearcherService.Enabled", func() {
		// Helper that returns a runaiConfig where every condition in
		// IngressResource.Enabled() is satisfied EXCEPT spec.ResearcherService.Enabled
		// (which is the single knob each test toggles).
		makeRunaiConfigWithIngressEligible := func() *runaiv1.RunaiConfig {
			c := utils.GetDefaultRunaiConfig()
			c.Spec.Internal.Global.ClusterURL = "https://example.runailabs.com"
			c.Spec.Global.Ingress = &global.Ingress{Enabled: pointer.Bool(true)}
			c.Spec.Internal.ResearcherService = &runv1_internal.ResearcherService{
				Ingress: &runv1_internal.ResearcherServiceIngress{
					Enabled: pointer.Bool(true),
				},
			}
			return c
		}

		It("returns false when ResearcherService.Enabled is false", func() {
			runaiConfig := makeRunaiConfigWithIngressEligible()
			runaiConfig.Spec.ResearcherService.Enabled = pointer.Bool(false)

			r := NewIngressResource("researcher-service", fake.NewClientBuilder().Build(), runaiConfig)
			Expect(r.Enabled()).To(BeFalse())
		})

		It("returns true when ResearcherService.Enabled is true and other conditions allow", func() {
			runaiConfig := makeRunaiConfigWithIngressEligible()
			runaiConfig.Spec.ResearcherService.Enabled = pointer.Bool(true)

			r := NewIngressResource("researcher-service", fake.NewClientBuilder().Build(), runaiConfig)
			Expect(r.Enabled()).To(BeTrue())
		})

		It("IngressStatus reports Ready when ResearcherService is disabled even with no ingress on cluster", func() {
			runaiConfig := makeRunaiConfigWithIngressEligible()
			runaiConfig.Spec.ResearcherService.Enabled = pointer.Bool(false)

			status, err := IngressStatus(context.Background(), fake.NewClientBuilder().Build(), runaiConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeTrue())
		})
	})
})
