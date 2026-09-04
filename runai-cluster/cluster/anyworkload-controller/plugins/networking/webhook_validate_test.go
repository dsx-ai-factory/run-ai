/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("WebhookValidate", func() {
	var plugin *Plugin

	BeforeEach(func() {
		plugin = NewPlugin(nil)
	})

	DescribeTable("validation errors",
		func(config *dromosv1alpha1.NetworkingConfig, expectedField string, expectedType field.ErrorType) {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runner",
					Namespace: "test-ns",
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Instructions: &dromosv1alpha1.Instructions{
						Networking: config,
					},
				},
			}

			ctx := context.Background()
			_, errs := plugin.WebhookValidate(ctx, nil, runner)

			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring(expectedField))
		},

		Entry("missing authorizedGroups when authorizationType=authorizedGroups",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref: &dromosv1alpha1.NetworkingServiceRef{Name: "svc", Port: 8000},
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthorizedGroups,
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
										}},
									},
								},
							}},
						},
					},
				},
			},
			"authorizedGroups", field.ErrorTypeRequired,
		),

		Entry("authorizedGroups set when authorizationType=authorizedUsers",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref: &dromosv1alpha1.NetworkingServiceRef{Name: "svc", Port: 8000},
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthorizedUsers,
						AuthorizedUsers:   "user1",
						AuthorizedGroups:  "group1",
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
										}},
									},
								},
							}},
						},
					},
				},
			},
			"authorizedGroups", field.ErrorTypeInvalid,
		),

		Entry("invalid authorizedGroups format when authorizationType=authorizedGroups",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref: &dromosv1alpha1.NetworkingServiceRef{Name: "svc", Port: 8000},
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthorizedGroups,
						AuthorizedGroups:  "group with spaces",
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
										}},
									},
								},
							}},
						},
					},
				},
			},
			"authorizedGroups", field.ErrorTypeInvalid,
		),

		Entry("ref without auth",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref: &dromosv1alpha1.NetworkingServiceRef{Name: "svc", Port: 8000},
				}},
			},
			"auth", field.ErrorTypeRequired,
		),

		Entry("both ref and create",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref:    &dromosv1alpha1.NetworkingServiceRef{Name: "svc", Port: 8000},
					Create: &corev1.Service{},
				}},
			},
			"mutually exclusive", field.ErrorTypeInvalid,
		),

		Entry("create without selector",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
			},
			"selector", field.ErrorTypeRequired,
		),

		Entry("ingress without rules",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{},
				},
			},
			"rules", field.ErrorTypeRequired,
		),

		Entry("route without host",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Route: &openshiftv1.Route{},
				},
			},
			"host", field.ErrorTypeRequired,
		),

		Entry("both ingress and route",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{},
					Route:   &openshiftv1.Route{},
				},
			},
			"mutually exclusive", field.ErrorTypeInvalid,
		),

		Entry("auth requires externalAccess",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
					},
				}},
			},
			"externalAccess", field.ErrorTypeRequired,
		),

		Entry("ingress rule without HTTP configuration",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{Host: "test.example.com"}},
						},
					},
				},
			},
			"http configuration is required", field.ErrorTypeRequired,
		),

		Entry("ingress rule with HTTP but no paths",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{},
									},
								},
							}},
						},
					},
				},
			},
			"at least one path is required", field.ErrorTypeRequired,
		),

		Entry("ingress backend name mismatch",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "correct-service"},
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "wrong-service",
													Port: networkingv1.ServiceBackendPort{Number: 8080},
												},
											},
										}},
									},
								},
							}},
						},
					},
				},
			},
			"must match service name", field.ErrorTypeInvalid,
		),

		Entry("ingress backend port mismatch",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "my-service"},
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "my-service",
													Port: networkingv1.ServiceBackendPort{Number: 9000},
												},
											},
										}},
									},
								},
							}},
						},
					},
				},
			},
			"must match service port", field.ErrorTypeInvalid,
		),

		Entry("ingress missing backend",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "my-service"},
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: func() *networkingv1.PathType { pt := networkingv1.PathTypePrefix; return &pt }(),
										}},
									},
								},
							}},
						},
					},
				},
			},
			"backend service must be specified", field.ErrorTypeRequired,
		),

		Entry("httpRoute without rules",
			&dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "test"},
							Ports:    []corev1.ServicePort{{Port: 8080}},
						},
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					HTTPRoute: &gatewayv1.HTTPRoute{},
				},
			},
			"rules", field.ErrorTypeRequired,
		),
	)

	It("passes validation with valid config - create with auth", func() {
		pathType := networkingv1.PathTypePrefix
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Instructions: &dromosv1alpha1.Instructions{
					Networking: &dromosv1alpha1.NetworkingConfig{
						Services: []dromosv1alpha1.NetworkingServiceItem{{
							Create: &corev1.Service{
								Spec: corev1.ServiceSpec{
									Selector: map[string]string{"app": "test"},
									Ports:    []corev1.ServicePort{{Port: 8080}},
								},
							},
							Auth: &dromosv1alpha1.NetworkingAuthConfig{
								AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
							},
						}},
						ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
							Ingress: &networkingv1.Ingress{
								Spec: networkingv1.IngressSpec{
									Rules: []networkingv1.IngressRule{{
										Host: "test.example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{{
													Path:     "/",
													PathType: &pathType,
												}},
											},
										},
									}},
								},
							},
						},
					},
				},
			},
		}

		ctx := context.Background()
		_, errs := plugin.WebhookValidate(ctx, nil, runner)

		Expect(errs).To(BeEmpty())
	})

	It("passes validation with valid config - ref with auth", func() {
		pathType := networkingv1.PathTypePrefix
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Instructions: &dromosv1alpha1.Instructions{
					Networking: &dromosv1alpha1.NetworkingConfig{
						Services: []dromosv1alpha1.NetworkingServiceItem{{
							Ref: &dromosv1alpha1.NetworkingServiceRef{Name: "my-svc", Port: 8080},
							Auth: &dromosv1alpha1.NetworkingAuthConfig{
								AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
							},
						}},
						ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
							Ingress: &networkingv1.Ingress{
								Spec: networkingv1.IngressSpec{
									Rules: []networkingv1.IngressRule{{
										Host: "test.example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{{
													Path:     "/",
													PathType: &pathType,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "my-svc",
															Port: networkingv1.ServiceBackendPort{Number: 8080},
														},
													},
												}},
											},
										},
									}},
								},
							},
						},
					},
				},
			},
		}

		ctx := context.Background()
		_, errs := plugin.WebhookValidate(ctx, nil, runner)

		Expect(errs).To(BeEmpty())
	})

	It("passes validation with valid config - create without auth", func() {
		pathType := networkingv1.PathTypePrefix
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Instructions: &dromosv1alpha1.Instructions{
					Networking: &dromosv1alpha1.NetworkingConfig{
						Services: []dromosv1alpha1.NetworkingServiceItem{{
							Create: &corev1.Service{
								Spec: corev1.ServiceSpec{
									Selector: map[string]string{"app": "test"},
									Ports:    []corev1.ServicePort{{Port: 8080}},
								},
							},
						}},
						ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
							Ingress: &networkingv1.Ingress{
								Spec: networkingv1.IngressSpec{
									Rules: []networkingv1.IngressRule{{
										Host: "test.example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{{
													Path:     "/",
													PathType: &pathType,
												}},
											},
										},
									}},
								},
							},
						},
					},
				},
			},
		}

		ctx := context.Background()
		_, errs := plugin.WebhookValidate(ctx, nil, runner)

		Expect(errs).To(BeEmpty())
	})

	It("passes validation with valid config - httpRoute", func() {
		pathPrefix := gatewayv1.PathMatchPathPrefix
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Instructions: &dromosv1alpha1.Instructions{
					Networking: &dromosv1alpha1.NetworkingConfig{
						Services: []dromosv1alpha1.NetworkingServiceItem{{
							Create: &corev1.Service{
								Spec: corev1.ServiceSpec{
									Selector: map[string]string{"app": "test"},
									Ports:    []corev1.ServicePort{{Port: 8080}},
								},
							},
						}},
						ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
							HTTPRoute: &gatewayv1.HTTPRoute{
								Spec: gatewayv1.HTTPRouteSpec{
									Rules: []gatewayv1.HTTPRouteRule{{
										Matches: []gatewayv1.HTTPRouteMatch{{
											Path: &gatewayv1.HTTPPathMatch{
												Type:  &pathPrefix,
												Value: func() *string { s := "/"; return &s }(),
											},
										}},
									}},
								},
							},
						},
					},
				},
			},
		}

		ctx := context.Background()
		_, errs := plugin.WebhookValidate(ctx, nil, runner)

		Expect(errs).To(BeEmpty())
	})

	It("passes validation with valid config - ingress only without services", func() {
		pathType := networkingv1.PathTypePrefix
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runner",
				Namespace: "test-ns",
			},
			Spec: dromosv1alpha1.WorkloadRunnerSpec{
				Instructions: &dromosv1alpha1.Instructions{
					Networking: &dromosv1alpha1.NetworkingConfig{
						// No services - just creating Ingress for cluster defaults
						ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
							Ingress: &networkingv1.Ingress{
								Spec: networkingv1.IngressSpec{
									Rules: []networkingv1.IngressRule{{
										Host: "test.example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{{
													Path:     "/",
													PathType: &pathType,
													Backend: networkingv1.IngressBackend{
														Service: &networkingv1.IngressServiceBackend{
															Name: "some-external-service",
															Port: networkingv1.ServiceBackendPort{Number: 8080},
														},
													},
												}},
											},
										},
									}},
								},
							},
						},
					},
				},
			},
		}

		ctx := context.Background()
		_, errs := plugin.WebhookValidate(ctx, nil, runner)

		Expect(errs).To(BeEmpty())
	})
})
