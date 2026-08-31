/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	testNamespace    = "runai-test"
	testWorkloadName = "test-workload"
)

func init() {
	_ = dromosv1alpha1.AddToScheme(scheme.Scheme)
	_ = networkingv1.AddToScheme(scheme.Scheme)
	_ = openshiftv1.Install(scheme.Scheme)
}

func newTestRunner(name, namespace string, networkingConfig *dromosv1alpha1.NetworkingConfig) *dromosv1alpha1.WorkloadRunner {
	runner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       "test-uid-12345",
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{},
	}
	if networkingConfig != nil {
		runner.Spec.Instructions = &dromosv1alpha1.Instructions{
			Networking: networkingConfig,
		}
	}
	return runner
}

var _ = Describe("Helpers", func() {
	Describe("extractNetworkingConfig", func() {
		It("returns nil when no instructions", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				Spec: dromosv1alpha1.WorkloadRunnerSpec{},
			}
			config := extractNetworkingConfig(runner)
			Expect(config).To(BeNil())
		})

		It("returns config when present", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
					},
				}},
			})
			config := extractNetworkingConfig(runner)
			Expect(config).NotTo(BeNil())
			Expect(config.Services).To(HaveLen(1))
			Expect(config.Services[0].Auth).NotTo(BeNil())
		})
	})

	Describe("isAuthEnabled", func() {
		It("returns false when config is nil", func() {
			Expect(isAuthEnabled(nil)).To(BeFalse())
		})

		It("returns false when auth is nil", func() {
			Expect(isAuthEnabled(&dromosv1alpha1.NetworkingConfig{})).To(BeFalse())
		})

		It("returns true when auth section is present", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthorizedGroups,
						AuthorizedGroups:  "data-scientists",
					},
				}},
			}
			Expect(isAuthEnabled(config)).To(BeTrue())
		})
	})

	Describe("getWorkloadServiceInfo", func() {
		It("returns defaults when service is nil", func() {
			config := &dromosv1alpha1.NetworkingConfig{}
			info := getWorkloadServiceInfo(config, "default-name")
			Expect(info.Name).To(Equal("default-name"))
			Expect(info.Port).To(Equal(int32(0)))
			Expect(info.Action).To(Equal(WorkloadServiceActionNone))
		})

		It("returns create service info", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "my-svc"},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
			}
			info := getWorkloadServiceInfo(config, "default-name")
			Expect(info.Name).To(Equal("my-svc"))
			Expect(info.Port).To(Equal(int32(8080)))
			Expect(info.Action).To(Equal(WorkloadServiceActionCreate))
		})

		It("returns ref service info", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref: &dromosv1alpha1.NetworkingServiceRef{
						Name: "existing-svc",
						Port: 9090,
					},
				}},
			}
			info := getWorkloadServiceInfo(config, "default-name")
			Expect(info.Name).To(Equal("existing-svc"))
			Expect(info.Port).To(Equal(int32(9090)))
			Expect(info.Action).To(Equal(WorkloadServiceActionRef))
		})
	})

	Describe("getExternalHostAndPath", func() {
		It("extracts from ingress", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{
									Host: "test.example.com",
									IngressRuleValue: networkingv1.IngressRuleValue{
										HTTP: &networkingv1.HTTPIngressRuleValue{
											Paths: []networkingv1.HTTPIngressPath{
												{Path: "/api"},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			host, path := getExternalHostAndPath(config)
			Expect(host).To(Equal("test.example.com"))
			Expect(path).To(Equal("/api"))
		})

		It("extracts from route", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Route: &openshiftv1.Route{
						Spec: openshiftv1.RouteSpec{
							Host: "route.openshift.com",
							Path: "/app",
						},
					},
				},
			}
			host, path := getExternalHostAndPath(config)
			Expect(host).To(Equal("route.openshift.com"))
			Expect(path).To(Equal("/app"))
		})

		It("strips nginx path rewrite regex suffix from ingress path", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{
											{Path: "/api/v1/my-nim(/|$)(.*)"},
										},
									},
								},
							}},
						},
					},
				},
			}
			host, path := getExternalHostAndPath(config)
			Expect(host).To(Equal("test.example.com"))
			Expect(path).To(Equal("/api/v1/my-nim"))
		})

		It("defaults path to /", func() {
			config := &dromosv1alpha1.NetworkingConfig{
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{
								{Host: "test.example.com"},
							},
						},
					},
				},
			}
			host, path := getExternalHostAndPath(config)
			Expect(host).To(Equal("test.example.com"))
			Expect(path).To(Equal("/"))
		})
	})
})
