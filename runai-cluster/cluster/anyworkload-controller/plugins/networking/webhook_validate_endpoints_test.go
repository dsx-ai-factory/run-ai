/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Bypasses endpoints.Set so callers can produce structurally invalid payloads to exercise admission.
func runnerWithEndpoints(cfg *endpoints.Config) *dromosv1alpha1.WorkloadRunner {
	cfg.SchemaVersion = endpoints.SchemaVersionV1
	raw, err := json.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:        testWorkloadName,
			Namespace:   testNamespace,
			Annotations: map[string]string{endpoints.Annotation: string(raw)},
		},
	}
}

func minimalEndpointService(name string, port int32) endpoints.Service {
	return endpoints.Service{
		Create: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": name},
				Ports:    []corev1.ServicePort{{Port: port, TargetPort: intstr.FromInt32(port)}},
			},
		},
	}
}

func metricsEndpointService(name string) endpoints.Service {
	return endpoints.Service{
		Create: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					constants.LabelWorkloadMetricsService:    "true",
					constants.LabelNIMWorkload:               "true",
					constants.LabelNIMWorkloadMetricsService: "true",
					constants.LabelWorkloadKind:              constants.KindInferenceWorkload,
				},
			},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{constants.LabelWorkloadName: testWorkloadName},
				Ports: []corev1.ServicePort{{
					Name:       "metrics",
					Port:       8080,
					TargetPort: intstr.FromInt32(8080),
					Protocol:   corev1.ProtocolTCP,
				}},
			},
		},
	}
}

func useVLLMEngineLabels(endpoint *endpoints.Endpoint) {
	labels := endpoint.Services[0].Create.Labels
	delete(labels, constants.LabelNIMWorkload)
	delete(labels, constants.LabelNIMWorkloadMetricsService)
	labels[constants.LabelVLLMWorkload] = "true"
	labels[constants.LabelVLLMWorkloadMetricsService] = "true"
}

func minimalEndpointIngress(svcName string, svcPort int32) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: svcName + "-ing"},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: svcName + ".example.com",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: svcName,
									Port: networkingv1.ServiceBackendPort{Number: svcPort},
								},
							},
						}},
					},
				},
			}},
		},
	}
}

func minimalEndpointRoute(name string) *openshiftv1.Route {
	return &openshiftv1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-route"},
		Spec: openshiftv1.RouteSpec{
			Host: name + ".apps.example.com",
		},
	}
}

func minimalEndpointHTTPRoute(name string) *gatewayv1.HTTPRoute {
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: name + "-httproute"},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(name + ".gateway.example.com")},
			Rules:     []gatewayv1.HTTPRouteRule{{}},
		},
	}
}

var _ = Describe("WebhookValidate (annotation path)", func() {
	var plugin *Plugin
	ctx := context.Background()

	BeforeEach(func() {
		plugin = NewPlugin(nil)
	})

	Describe("mutual exclusivity", func() {
		It("rejects a runner that sets both spec.instructions.networking and the annotation", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "svc"},
						Spec: corev1.ServiceSpec{
							Selector: map[string]string{"app": "x"},
							Ports:    []corev1.ServicePort{{Port: 80}},
						},
					},
				}},
			})
			Expect(endpoints.Set(runner, &endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{minimalEndpointService("a", 80)}}},
			})).To(Succeed())

			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Error()).To(ContainSubstring("mutually exclusive"))
		})
	})

	Describe("payload parse errors", func() {
		It("rejects a malformed JSON annotation", func() {
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:        testWorkloadName,
					Namespace:   testNamespace,
					Annotations: map[string]string{endpoints.Annotation: "{not json"},
				},
			}
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Error()).To(ContainSubstring("malformed"))
		})

		It("rejects an unsupported schemaVersion", func() {
			payload, _ := json.Marshal(map[string]any{"schemaVersion": "v9", "endpoints": []any{}})
			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:        testWorkloadName,
					Namespace:   testNamespace,
					Annotations: map[string]string{endpoints.Annotation: string(payload)},
				},
			}
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(HaveLen(1))
			Expect(errs[0].Error()).To(ContainSubstring("unsupported"))
		})
	})

	Describe("structural validation", func() {
		It("accepts a single-endpoint internal-only payload", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("accepts a single-endpoint payload with Ingress and matching backend", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)},
					ExternalAccess: &endpoints.ExternalAccess{
						Ingress: minimalEndpointIngress("jupyter", 8888),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("accepts Ingress and HTTPRoute together for gateway migration parity", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)},
					ExternalAccess: &endpoints.ExternalAccess{
						Ingress:   minimalEndpointIngress("jupyter", 8888),
						HTTPRoute: minimalEndpointHTTPRoute("jupyter"),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("accepts Route and HTTPRoute together for gateway migration parity", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)},
					ExternalAccess: &endpoints.ExternalAccess{
						Route:     minimalEndpointRoute("jupyter"),
						HTTPRoute: minimalEndpointHTTPRoute("jupyter"),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("rejects Ingress and Route together", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)},
					ExternalAccess: &endpoints.ExternalAccess{
						Ingress: minimalEndpointIngress("jupyter", 8888),
						Route:   minimalEndpointRoute("jupyter"),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("externalAccess.ingress and externalAccess.route are mutually exclusive"))
		})

		It("accepts an external-access-only endpoint", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					ExternalAccess: &endpoints.ExternalAccess{
						Ingress: minimalEndpointIngress("external-service", 9000),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("rejects a payload with zero endpoints", func() {
			runner := runnerWithEndpoints(&endpoints.Config{Endpoints: nil})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("at least one endpoint"))
		})

		It("rejects an endpoint with neither services nor externalAccess", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("at least one service or externalAccess"))
		})

		It("rejects more than one service per endpoint (v1 cap)", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{
						minimalEndpointService("a", 80),
						minimalEndpointService("b", 81),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("Too many"))
		})

		It("rejects a service with neither create nor ref", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{{}},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("either ref or create must be set"))
		})

		It("rejects a service with both create and ref", func() {
			svc := minimalEndpointService("dup", 80)
			svc.Ref = &endpoints.ServiceRef{Name: "x", Port: 80}
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{svc}}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("mutually exclusive"))
		})

		It("rejects an unsupported authMode value", func() {
			svc := minimalEndpointService("jupyter", 8888)
			svc.AuthMode = "InPod"
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{svc}}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("supported values"))
		})

		It("rejects auth without externalAccess", func() {
			svc := minimalEndpointService("jupyter", 8888)
			svc.Auth = &endpoints.AuthConfig{
				AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
			}
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{svc}}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("externalAccess is required when service auth is enabled"))
		})

		It("rejects an ingress backend that doesn't match the endpoint's service", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)},
					ExternalAccess: &endpoints.ExternalAccess{
						Ingress: minimalEndpointIngress("OTHER", 9999),
					},
				}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("must match endpoint service name"))
		})

		It("accepts two endpoints with distinct service names", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{
					{Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)}},
					{Services: []endpoints.Service{minimalEndpointService("tensorboard", 6006)}},
				},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		DescribeTable("accepts exactly one complete metrics engine-label pair",
			func(vllm bool) {
				metricsEndpoint := endpoints.Endpoint{
					Services: []endpoints.Service{metricsEndpointService(testWorkloadName + "-metrics")},
				}
				if vllm {
					useVLLMEngineLabels(&metricsEndpoint)
				}
				runner := runnerWithEndpoints(&endpoints.Config{
					Endpoints: []endpoints.Endpoint{
						{Services: []endpoints.Service{minimalEndpointService("serving", 80)}},
						metricsEndpoint,
					},
				})
				_, errs := plugin.WebhookValidate(ctx, nil, runner)
				Expect(errs).To(BeEmpty())
			},
			Entry("NIM", false),
			Entry("vLLM", true),
		)

		DescribeTable("rejects invalid label-identified metrics Services",
			func(mutate func(*endpoints.Endpoint), expected string) {
				endpoint := endpoints.Endpoint{Services: []endpoints.Service{metricsEndpointService(testWorkloadName + "-metrics")}}
				mutate(&endpoint)
				runner := runnerWithEndpoints(&endpoints.Config{Endpoints: []endpoints.Endpoint{endpoint}})
				_, errs := plugin.WebhookValidate(ctx, nil, runner)
				Expect(errs).NotTo(BeEmpty())
				Expect(errs.ToAggregate().Error()).To(ContainSubstring(expected))
			},
			Entry("implicit Service type", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Type = ""
			}, "explicitly use ClusterIP"),
			Entry("NodePort Service type", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Type = corev1.ServiceTypeNodePort
			}, "explicitly use ClusterIP"),
			Entry("more than one port", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Ports = append(
					e.Services[0].Create.Spec.Ports,
					corev1.ServicePort{Name: "other", Port: 9090, TargetPort: intstr.FromInt32(9090), Protocol: corev1.ProtocolTCP},
				)
			}, "exactly one metrics port"),
			Entry("wrong port name", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Ports[0].Name = "http"
			}, `port name must be "metrics"`),
			Entry("non-TCP port", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Ports[0].Protocol = corev1.ProtocolUDP
			}, "protocol must be TCP"),
			Entry("missing target port", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Ports[0].TargetPort = intstr.IntOrString{}
			}, "targetPort is required"),
			Entry("authentication", func(e *endpoints.Endpoint) {
				e.Services[0].Auth = &endpoints.AuthConfig{
					AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
				}
			}, "must not configure authentication"),
			Entry("external access", func(e *endpoints.Endpoint) {
				e.ExternalAccess = &endpoints.ExternalAccess{
					Ingress: minimalEndpointIngress("metrics", 8080),
				}
			}, "must not configure externalAccess"),
			Entry("arbitrary service name", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Name = "other-metrics"
			}, "metrics Service name must be"),
			Entry("empty selector", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Selector = map[string]string{}
			}, "metrics Service must have at least one selector"),
			Entry("selector missing workloadName key", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Selector = map[string]string{"app": "other"}
			}, "missing Service selector "+constants.LabelWorkloadName),
			Entry("selector with wrong workloadName value", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Spec.Selector = map[string]string{constants.LabelWorkloadName: "other-workload"}
			}, "metrics Service selector must be "+constants.LabelWorkloadName+"="+testWorkloadName),
			Entry("missing inference workload kind", func(e *endpoints.Endpoint) {
				delete(e.Services[0].Create.Labels, constants.LabelWorkloadKind)
			}, "metrics Service must target"),
			Entry("missing engine metrics label", func(e *endpoints.Endpoint) {
				delete(e.Services[0].Create.Labels, constants.LabelNIMWorkloadMetricsService)
			}, "exactly one consistent NIM or vLLM engine-label pair"),
			Entry("conflicting engine labels", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Labels[constants.LabelVLLMWorkload] = "true"
				e.Services[0].Create.Labels[constants.LabelVLLMWorkloadMetricsService] = "true"
			}, "exactly one consistent NIM or vLLM engine-label pair"),
			Entry("NIM pair with vLLM workload label", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Labels[constants.LabelVLLMWorkload] = "true"
			}, "exactly one consistent NIM or vLLM engine-label pair"),
			Entry("NIM pair with vLLM metrics Service label", func(e *endpoints.Endpoint) {
				e.Services[0].Create.Labels[constants.LabelVLLMWorkloadMetricsService] = "true"
			}, "exactly one consistent NIM or vLLM engine-label pair"),
			Entry("vLLM pair with NIM workload label", func(e *endpoints.Endpoint) {
				useVLLMEngineLabels(e)
				e.Services[0].Create.Labels[constants.LabelNIMWorkload] = "true"
			}, "exactly one consistent NIM or vLLM engine-label pair"),
			Entry("vLLM pair with NIM metrics Service label", func(e *endpoints.Endpoint) {
				useVLLMEngineLabels(e)
				e.Services[0].Create.Labels[constants.LabelNIMWorkloadMetricsService] = "true"
			}, "exactly one consistent NIM or vLLM engine-label pair"),
		)

		It("does not apply metrics validation for engine-specific labels alone", func() {
			svc := minimalEndpointService("nim-metrics-looking", 8080)
			svc.Create.Labels = map[string]string{constants.LabelNIMWorkloadMetricsService: "true"}
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{svc}}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		// Regression: the CP workloads-manager appends run.ai/networking-workload-id
		// to every created Service selector as part of the ownership contract
		// (bindCreatedServiceSelectors). Admission must accept the resulting
		// multi-key selector as long as workloadName=<name> is present.
		It("accepts a metrics Service selector with additional keys alongside workloadName", func() {
			svc := metricsEndpointService(testWorkloadName + "-metrics")
			svc.Create.Spec.Selector = map[string]string{
				constants.LabelWorkloadName:     testWorkloadName,
				"run.ai/networking-workload-id": "cp-workload-id",
			}
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{svc}}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).To(BeEmpty())
		})

		It("rejects duplicate service names across endpoints", func() {
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{
					{Services: []endpoints.Service{minimalEndpointService("jupyter", 8888)}},
					{Services: []endpoints.Service{minimalEndpointService("jupyter", 8889)}},
				},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("service names must be unique"))
		})

		It("rejects a Create service with no metadata.name", func() {
			svc := minimalEndpointService("ignored", 8888)
			svc.Create.Name = ""
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{{Services: []endpoints.Service{svc}}},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("metadata.name is required"))
		})

		It("rejects when one endpoint's declared service collides with another's generated auth-proxy name", func() {
			authSvc := minimalEndpointService("api", 8888)
			authSvc.Auth = &endpoints.AuthConfig{
				AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
			}
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{
					{
						Services: []endpoints.Service{authSvc},
						ExternalAccess: &endpoints.ExternalAccess{
							Ingress: minimalEndpointIngress("api", 8888),
						},
					},
					{Services: []endpoints.Service{minimalEndpointService("api-auth-proxy", 9000)}},
				},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("service names must be unique"))
		})

		It("rejects two endpoints whose Ingress names default to the same runner name", func() {
			ing := minimalEndpointIngress("svcA", 80)
			ing.Name = ""
			ing2 := minimalEndpointIngress("svcB", 81)
			ing2.Name = ""
			runner := runnerWithEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{
					{
						Services:       []endpoints.Service{minimalEndpointService("svcA", 80)},
						ExternalAccess: &endpoints.ExternalAccess{Ingress: ing},
					},
					{
						Services:       []endpoints.Service{minimalEndpointService("svcB", 81)},
						ExternalAccess: &endpoints.ExternalAccess{Ingress: ing2},
					},
				},
			})
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs).NotTo(BeEmpty())
			Expect(errs.ToAggregate().Error()).To(ContainSubstring("ingress names must be unique"))
		})
	})
})
