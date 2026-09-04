/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	networkingv1 "k8s.io/api/networking/v1"
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
			Expect(ingress.Spec.Rules[0].HTTP.Paths).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/cluster-api"))
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
			runaiConfig.Spec.ClusterApi.Ingress.TlsSecret = "runai-cluster-tls-secret"
			initIngress(runaiConfig, ingress)
			Expect(ingress.Spec.TLS[0].Hosts[0]).To(Equal("www.example.com"))
			Expect(ingress.Spec.TLS[0].SecretName).To(Equal("runai-cluster-tls-secret"))
		})
	})

	Context("Custom annotations", func() {
		It("should have default nginx annotations", func() {
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations).To(Not(BeEmpty()))
		})

		It("should have custom annotations from ClusterApi config", func() {
			runaiConfig.Spec.ClusterApi.Ingress.Annotations = map[string]string{
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

		It("should merge annotations from both ClusterApi and Global configs", func() {
			runaiConfig.Spec.ClusterApi.Ingress.Annotations = map[string]string{
				"cluster-api-key": "cluster-api-value",
			}
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Annotations: map[string]string{
					"global-key": "global-value",
				},
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["cluster-api-key"]).To(Equal("cluster-api-value"))
			Expect(ingress.ObjectMeta.Annotations["global-key"]).To(Equal("global-value"))
		})

		It("should give precedence to ClusterApi annotations over Global annotations", func() {
			runaiConfig.Spec.ClusterApi.Ingress.Annotations = map[string]string{
				"shared-key": "from-cluster-api",
			}
			runaiConfig.Spec.Global.Ingress = &global.Ingress{
				Annotations: map[string]string{
					"shared-key": "from-global",
				},
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["shared-key"]).To(Equal("from-cluster-api"))
		})

		It("should set CORS annotations by default", func() {
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/enable-cors"]).To(Equal("true"))
			Expect(ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/cors-allow-headers"]).To(ContainSubstring("Authorization"))
			Expect(ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/cors-allow-credentials"]).To(Equal("true"))
		})

		It("should scope cors-allow-origin to the control-plane origin (scheme+host, no path/port)", func() {
			runaiConfig.Spec.Internal.Global.ControlPlane.URL = "https://acme.run.ai:443/some/path"
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/cors-allow-origin"]).To(Equal("https://acme.run.ai"))
		})

		It("should omit cors-allow-origin when the control-plane URL is unknown", func() {
			runaiConfig.Spec.Internal.Global.ControlPlane.URL = ""
			initIngress(runaiConfig, ingress)
			_, ok := ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/cors-allow-origin"]
			Expect(ok).To(BeFalse())
		})

		It("should let customer annotations override the CORS defaults", func() {
			runaiConfig.Spec.ClusterApi.Ingress.Annotations = map[string]string{
				"nginx.ingress.kubernetes.io/enable-cors": "false",
			}
			initIngress(runaiConfig, ingress)
			Expect(ingress.ObjectMeta.Annotations["nginx.ingress.kubernetes.io/enable-cors"]).To(Equal("false"))
		})

		It("should not share the annotations map reference (deep copy verification)", func() {
			originalAnnotations := map[string]string{
				"original-key": "original-value",
			}
			runaiConfig.Spec.ClusterApi.Ingress.Annotations = originalAnnotations

			initIngress(runaiConfig, ingress)

			// Modify the ingress annotations
			ingress.ObjectMeta.Annotations["new-key"] = "new-value"

			// Original map should not be affected
			Expect(originalAnnotations).To(HaveLen(1))
			Expect(originalAnnotations).ToNot(HaveKey("new-key"))
			Expect(originalAnnotations["original-key"]).To(Equal("original-value"))
		})
	})
})
