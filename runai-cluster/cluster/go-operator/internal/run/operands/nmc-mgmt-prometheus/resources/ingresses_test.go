/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prom_common "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Ingress Resource", func() {
	var (
		ctx         context.Context
		k8sClient   client.Client
		runaiConfig *runv1.RunaiConfig
		resource    *IngressResource
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = fake.NewClientBuilder().Build()
		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-namespace",
			},
			Spec: runv1.RunaiConfigSpec{
				Global: &global.Global{},
				Internal: &internal.Internal{
					Global: &internal.Global{
						NMC: internal.NMC{
							MgmtPrometheusIngress: &common.Ingress{},
						},
					},
				},
			},
		}
		resource = NewIngressResource("test-operand", k8sClient, runaiConfig)
	})

	Describe("Ingress Class Configuration", func() {
		Context("when no ingress class is specified", func() {
			It("should default to nginx", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("nginx"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("nginx"))
			})
		})

		Context("when NMC mgmt prometheus ingress class is specified", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Class: "nmc-ingress-class",
				}
			})

			It("should use the NMC ingress class", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("nmc-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("nmc-ingress-class"))
			})
		})

		Context("when global ingress class is specified", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Class: "global-ingress-class",
				}
			})

			It("should use the global ingress class", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("global-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("global-ingress-class"))
			})
		})

		Context("when both NMC and global ingress classes are specified", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Class: "nmc-ingress-class",
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Class: "global-ingress-class",
				}
			})

			It("should prefer NMC ingress class over global", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("nmc-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("nmc-ingress-class"))
			})
		})

		Context("when NMC ingress is set but class is empty", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Class: "",
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Class: "global-ingress-class",
				}
			})

			It("should fallback to global ingress class", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("global-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("global-ingress-class"))
			})
		})
	})

	Describe("Ingress Annotations", func() {
		Context("when NMC ingress annotations are specified", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Annotations: map[string]string{
						"nmc-annotation": "nmc-value",
					},
				}
			})

			It("should include NMC annotations", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Annotations["nmc-annotation"]).To(Equal("nmc-value"))
			})
		})

		Context("when global ingress annotations are specified", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Annotations: map[string]string{
						"global-annotation": "global-value",
					},
				}
			})

			It("should include global annotations", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Annotations["global-annotation"]).To(Equal("global-value"))
			})
		})

		Context("when both NMC and global annotations are specified", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Annotations: map[string]string{
						"shared-key":     "nmc-value",
						"nmc-annotation": "nmc-value",
					},
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Annotations: map[string]string{
						"shared-key":        "global-value",
						"global-annotation": "global-value",
					},
				}
			})

			It("should merge annotations with NMC taking precedence for shared keys", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				Expect(ingress.Annotations["nmc-annotation"]).To(Equal("nmc-value"))
				Expect(ingress.Annotations["global-annotation"]).To(Equal("global-value"))
				// NMC annotations are copied after global annotations, so NMC takes precedence
				Expect(ingress.Annotations["shared-key"]).To(Equal("nmc-value"))
			})
		})

		Context("when both NMC and global ingress have classes and annotations", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Class: "nmc-ingress-class",
					Annotations: map[string]string{
						"nmc-annotation": "nmc-value",
						"shared-key":     "nmc-value",
					},
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Class: "global-ingress-class",
					Annotations: map[string]string{
						"global-annotation": "global-value",
						"shared-key":        "global-value",
					},
				}
			})

			It("should use NMC class and merge annotations with NMC precedence", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				// Verify NMC class takes precedence
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("nmc-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("nmc-ingress-class"))

				// Verify annotations are merged with NMC taking precedence
				Expect(ingress.Annotations["nmc-annotation"]).To(Equal("nmc-value"))
				Expect(ingress.Annotations["global-annotation"]).To(Equal("global-value"))
				Expect(ingress.Annotations["shared-key"]).To(Equal("nmc-value"))
			})
		})

		Context("when NMC has only annotations and global has class and annotations", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Annotations: map[string]string{
						"nmc-annotation": "nmc-value",
					},
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Class: "global-ingress-class",
					Annotations: map[string]string{
						"global-annotation": "global-value",
					},
				}
			})

			It("should use global class and merge both annotations", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				// Verify global class is used when NMC class is empty
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("global-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("global-ingress-class"))

				// Verify both annotations are present
				Expect(ingress.Annotations["nmc-annotation"]).To(Equal("nmc-value"))
				Expect(ingress.Annotations["global-annotation"]).To(Equal("global-value"))
			})
		})

		Context("when NMC has class and global has only annotations", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Class: "nmc-ingress-class",
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Annotations: map[string]string{
						"global-annotation": "global-value",
					},
				}
			})

			It("should use NMC class and include global annotations", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				// Verify NMC class is used
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("nmc-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("nmc-ingress-class"))

				// Verify global annotation is included
				Expect(ingress.Annotations["global-annotation"]).To(Equal("global-value"))
			})
		})

		Context("when NMC has empty class but annotations, and global has both", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = &common.Ingress{
					Class: "",
					Annotations: map[string]string{
						"nmc-annotation": "nmc-value",
						"shared-key":     "nmc-value",
					},
				}
				runaiConfig.Spec.Global.Ingress = &global.Ingress{
					Class: "global-ingress-class",
					Annotations: map[string]string{
						"global-annotation": "global-value",
						"shared-key":        "global-value",
					},
				}
			})

			It("should use global class but NMC annotations take precedence in merge", func() {
				obj, err := resource.Object(ctx)
				Expect(err).NotTo(HaveOccurred())

				ingress := obj.(*v1.Ingress)
				// Verify global class is used when NMC class is empty
				Expect(ingress.Spec.IngressClassName).NotTo(BeNil())
				Expect(*ingress.Spec.IngressClassName).To(Equal("global-ingress-class"))
				Expect(ingress.Annotations["kubernetes.io/ingress.class"]).To(Equal("global-ingress-class"))

				// Verify annotations are merged with NMC taking precedence
				Expect(ingress.Annotations["nmc-annotation"]).To(Equal("nmc-value"))
				Expect(ingress.Annotations["global-annotation"]).To(Equal("global-value"))
				Expect(ingress.Annotations["shared-key"]).To(Equal("nmc-value"))
			})
		})
	})

	Describe("Ingress Basic Configuration", func() {
		It("should set correct name and namespace", func() {
			obj, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			ingress := obj.(*v1.Ingress)
			Expect(ingress.Name).To(Equal("nmc-mgmt-prometheus"))
			Expect(ingress.Namespace).To(Equal("test-namespace"))
		})

		It("should configure correct host and path", func() {
			obj, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			ingress := obj.(*v1.Ingress)
			Expect(ingress.Spec.Rules).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].Host).To(Equal(prom_common.MgmtPrometheusDomain))
			Expect(ingress.Spec.Rules[0].HTTP.Paths).To(HaveLen(1))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/"))
		})

		It("should configure TLS", func() {
			obj, err := resource.Object(ctx)
			Expect(err).NotTo(HaveOccurred())

			ingress := obj.(*v1.Ingress)
			Expect(ingress.Spec.TLS).To(HaveLen(1))
			Expect(ingress.Spec.TLS[0].Hosts).To(ContainElement(prom_common.MgmtPrometheusDomain))
			Expect(ingress.Spec.TLS[0].SecretName).To(Equal(tlsSecretName))
		})
	})
})
