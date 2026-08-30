package networking

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("Builders", func() {
	Describe("getEndpointAuthProxyName", func() {
		It("preserves existing names that fit the Kubernetes limit", func() {
			serviceName := strings.Repeat("a", 63-len(AuthProxySuffix))

			Expect(getEndpointAuthProxyName(serviceName)).To(Equal(serviceName + AuthProxySuffix))
		})

		It("shortens long derived names deterministically and retains the auth-proxy suffix", func() {
			serviceName := strings.Repeat("a", 63)
			name := getEndpointAuthProxyName(serviceName)

			Expect(name).To(HaveLen(63))
			Expect(name).To(HaveSuffix(AuthProxySuffix))
			Expect(validation.IsDNS1123Label(name)).To(BeEmpty())
			Expect(name).To(Equal(getEndpointAuthProxyName(serviceName)))
		})

		It("does not collide when long names have the same truncated prefix", func() {
			first := getEndpointAuthProxyName(strings.Repeat("a", 62) + "b")
			second := getEndpointAuthProxyName(strings.Repeat("a", 62) + "c")

			Expect(first).NotTo(Equal(second))
		})
	})

	Describe("buildAuthProxyService", func() {
		It("creates correct service", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)

			svc := plugin.buildAuthProxyService(runner, getAuthProxyName(runner))

			Expect(svc.Name).To(Equal(testWorkloadName + "-auth-proxy"))
			Expect(svc.Namespace).To(Equal(testNamespace))
			Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(svc.Spec.Ports).To(HaveLen(1))
			Expect(svc.Spec.Ports[0].Port).To(Equal(int32(AuthProxyServicePort)))
		})
	})

	Describe("buildWorkloadService", func() {
		It("copies and sets namespace", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			inputSvc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "my-svc"},
				Spec: corev1.ServiceSpec{
					Type:     corev1.ServiceTypeNodePort,
					Ports:    []corev1.ServicePort{{Port: 8080}},
					Selector: map[string]string{"app": "test"},
				},
			}

			svc := plugin.buildWorkloadService(runner, inputSvc)

			Expect(svc.Name).To(Equal("my-svc"))
			Expect(svc.Namespace).To(Equal(testNamespace))
			Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
		})

		It("defaults name to runner name", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			inputSvc := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{Port: 8080}},
				},
			}

			svc := plugin.buildWorkloadService(runner, inputSvc)

			Expect(svc.Name).To(Equal(testWorkloadName))
		})
	})

	Describe("buildEndpointWorkloadService", func() {
		It("preserves service type and port details", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			inputSvc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "nodeport-svc"},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeNodePort,
					Ports: []corev1.ServicePort{{
						Name:       "http",
						Protocol:   corev1.ProtocolTCP,
						Port:       8080,
						TargetPort: intstr.FromInt32(8888),
						NodePort:   32080,
					}},
					Selector: map[string]string{"app": "test"},
				},
			}

			svc := plugin.buildEndpointWorkloadService(runner, inputSvc)

			Expect(svc.Name).To(Equal("nodeport-svc"))
			Expect(svc.Namespace).To(Equal(testNamespace))
			Expect(svc.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
			Expect(svc.Spec.Ports[0].Name).To(Equal("http"))
			Expect(svc.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(svc.Spec.Ports[0].Port).To(Equal(int32(8080)))
			Expect(svc.Spec.Ports[0].TargetPort).To(Equal(intstr.FromInt32(8888)))
			Expect(svc.Spec.Ports[0].NodePort).To(Equal(int32(32080)))
		})
	})

	Describe("buildIngress", func() {
		It("overrides backend to auth proxy", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			inputIngress := &networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "test.example.com",
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Path: "/",
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "original-svc",
													Port: networkingv1.ServiceBackendPort{Number: 8080},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			ingress := plugin.buildIngress(runner, inputIngress, "auth-proxy-svc", 80)

			Expect(ingress.Namespace).To(Equal(testNamespace))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal("auth-proxy-svc"))
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number).To(Equal(int32(80)))
		})
	})

	Describe("buildRoute", func() {
		It("overrides backend to auth proxy", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			inputRoute := &openshiftv1.Route{
				Spec: openshiftv1.RouteSpec{
					Host: "route.openshift.com",
					To: openshiftv1.RouteTargetReference{
						Kind: "Service",
						Name: "original-svc",
					},
				},
			}

			route := plugin.buildRoute(runner, inputRoute, "auth-proxy-svc")

			Expect(route.Namespace).To(Equal(testNamespace))
			Expect(route.Spec.To.Name).To(Equal("auth-proxy-svc"))
			Expect(route.Spec.To.Kind).To(Equal("Service"))
		})
	})

	Describe("buildHTTPRoute", func() {
		It("overrides backend name and port", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			origPort := gatewayv1.PortNumber(8080)
			inputHTTPRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: "my-httproute"},
				Spec: gatewayv1.HTTPRouteSpec{
					Rules: []gatewayv1.HTTPRouteRule{{
						BackendRefs: []gatewayv1.HTTPBackendRef{{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: "original-svc",
									Port: &origPort,
								},
							},
						}},
					}},
				},
			}

			httpRoute := plugin.buildHTTPRoute(runner, inputHTTPRoute, "auth-proxy-svc", 80)

			Expect(httpRoute.Namespace).To(Equal(testNamespace))
			Expect(httpRoute.Name).To(Equal("my-httproute"))
			Expect(httpRoute.Spec.Rules).To(HaveLen(1))
			Expect(httpRoute.Spec.Rules[0].BackendRefs).To(HaveLen(1))
			Expect(string(httpRoute.Spec.Rules[0].BackendRefs[0].Name)).To(Equal("auth-proxy-svc"))
			Expect(int32(*httpRoute.Spec.Rules[0].BackendRefs[0].Port)).To(Equal(int32(80)))
		})

		It("defaults name to runner name when spec name is empty", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			inputHTTPRoute := &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Rules: []gatewayv1.HTTPRouteRule{{}},
				},
			}

			httpRoute := plugin.buildHTTPRoute(runner, inputHTTPRoute, "", 0)

			Expect(httpRoute.Name).To(Equal(testWorkloadName))
		})

		It("preserves parentRefs and hostnames", func() {
			plugin := NewPlugin(nil)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)
			gwNs := gatewayv1.Namespace("gateway-ns")
			inputHTTPRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: "my-httproute"},
				Spec: gatewayv1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name:      "my-gateway",
							Namespace: &gwNs,
						}},
					},
					Hostnames: []gatewayv1.Hostname{"test.example.com"},
					Rules:     []gatewayv1.HTTPRouteRule{{}},
				},
			}

			httpRoute := plugin.buildHTTPRoute(runner, inputHTTPRoute, "", 0)

			Expect(httpRoute.Spec.ParentRefs).To(HaveLen(1))
			Expect(string(httpRoute.Spec.ParentRefs[0].Name)).To(Equal("my-gateway"))
			Expect(httpRoute.Spec.Hostnames).To(HaveLen(1))
			Expect(string(httpRoute.Spec.Hostnames[0])).To(Equal("test.example.com"))
		})
	})
})
