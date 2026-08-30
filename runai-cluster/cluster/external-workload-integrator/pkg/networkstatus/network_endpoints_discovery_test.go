package networkstatus

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// ---------- NodePort ----------

var _ = Describe("findNodePortEndpoints", func() {
	It("creates endpoint for each port with NodePort", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findNodePortEndpoints(svc, "10.0.0.1")
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep.Host).To(Equal("10.0.0.1"))
		Expect(ep.Port).To(Equal("30080"))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.TlsEnabled).To(BeNil())
		Expect(ep.Url).To(Equal("10.0.0.1:30080"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"}))
	})

	It("marks TLS when AppProtocol is https", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np-tls"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{Port: 443, NodePort: 30443, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("https")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findNodePortEndpoints(svc, "10.0.0.2")
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep.Host).To(Equal("10.0.0.2"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("30443"))
		Expect(ep.Url).To(Equal("https://10.0.0.2:30443/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np-tls"}))
	})

	It("does not mark TLS when NodePort is 443 but AppProtocol is not set", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np-443"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{Port: 443, NodePort: 443, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findNodePortEndpoints(svc, "10.0.0.3")
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep.Host).To(Equal("10.0.0.3"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(BeNil())
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("10.0.0.3:443"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np-443"}))
	})

	It("skips ports with NodePort == 0", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np-zero"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 0, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findNodePortEndpoints(svc, "10.0.0.4")
		Expect(eps).To(BeEmpty())
	})

	It("handles multiple ports", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np-multi"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
					{Port: 443, NodePort: 30443, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("https")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findNodePortEndpoints(svc, "10.0.0.5")
		Expect(eps).To(HaveLen(2))

		ep0 := eps[0]
		Expect(ep0.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep0.Host).To(Equal("10.0.0.5"))
		Expect(ep0.Path).To(Equal("/"))
		Expect(ep0.TlsEnabled).To(BeNil())
		Expect(ep0.Protocol).To(Equal("TCP"))
		Expect(ep0.Port).To(Equal("30080"))
		Expect(ep0.Url).To(Equal("10.0.0.5:30080"))
		Expect(ep0.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np-multi"}))

		ep1 := eps[1]
		Expect(ep1.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep1.Host).To(Equal("10.0.0.5"))
		Expect(ep1.Path).To(Equal("/"))
		Expect(ep1.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep1.Protocol).To(Equal("TCP"))
		Expect(ep1.Port).To(Equal("30443"))
		Expect(ep1.Url).To(Equal("https://10.0.0.5:30443/"))
		Expect(ep1.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np-multi"}))
	})
})

// ---------- LoadBalancer ----------

var _ = Describe("findLoadBalancerEndpoints", func() {
	It("creates endpoint from ingress IP, defaults to TLS for non-80 port", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 8080, Protocol: corev1.ProtocolTCP},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(ep.Host).To(Equal("1.2.3.4"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Url).To(Equal("1.2.3.4:8080"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb"}))
	})

	It("falls back to hostname when IP is empty", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-hostname"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("http")},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{Hostname: "my-lb.example.com"}},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(ep.Host).To(Equal("my-lb.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Url).To(Equal("http://my-lb.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-hostname"}))
	})

	It("marks TLS when AppProtocol is https", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-tls-proto"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 443, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("https")},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: "5.6.7.8"}},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(ep.Host).To(Equal("5.6.7.8"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://5.6.7.8/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-tls-proto"}))
	})

	It("marks TLS when port is 443", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-443"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 443, Protocol: corev1.ProtocolTCP},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: "1.1.1.1"}},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(ep.Host).To(Equal("1.1.1.1"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("1.1.1.1:443"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-443"}))
	})

	It("marks TLS when port is 80 and no AppProtocol (default mode assumes TLS)", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-port80"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: corev1.ProtocolTCP},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: "2.2.2.2"}},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(ep.Host).To(Equal("2.2.2.2"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Url).To(Equal("2.2.2.2:80"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-port80"}))
	})

	It("marks non-TLS when appProtocol is http", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-http-proto"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 8080, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("http")},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: "2.2.2.2"}},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(ep.Host).To(Equal("2.2.2.2"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Url).To(Equal("http://2.2.2.2:8080/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-http-proto"}))
	})

	It("handles multiple ingresses and ports", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-multi"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("http")},
					{Port: 443, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("https")},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{IP: "1.1.1.1"},
						{IP: "2.2.2.2"},
					},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		// 2 ingresses * 2 ports = 4 endpoints
		Expect(eps).To(HaveLen(4))

		// ingress 1.1.1.1, port 80
		Expect(eps[0].SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(eps[0].Host).To(Equal("1.1.1.1"))
		Expect(eps[0].Path).To(Equal("/"))
		Expect(eps[0].TlsEnabled).To(HaveValue(BeFalse()))
		Expect(eps[0].Protocol).To(Equal("TCP"))
		Expect(eps[0].Port).To(Equal("80"))
		Expect(eps[0].Url).To(Equal("http://1.1.1.1/"))
		Expect(eps[0].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-multi"}))

		// ingress 1.1.1.1, port 443
		Expect(eps[1].SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(eps[1].Host).To(Equal("1.1.1.1"))
		Expect(eps[1].Path).To(Equal("/"))
		Expect(eps[1].TlsEnabled).To(HaveValue(BeTrue()))
		Expect(eps[1].Protocol).To(Equal("TCP"))
		Expect(eps[1].Port).To(Equal("443"))
		Expect(eps[1].Url).To(Equal("https://1.1.1.1/"))
		Expect(eps[1].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-multi"}))

		// ingress 2.2.2.2, port 80
		Expect(eps[2].SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(eps[2].Host).To(Equal("2.2.2.2"))
		Expect(eps[2].Path).To(Equal("/"))
		Expect(eps[2].TlsEnabled).To(HaveValue(BeFalse()))
		Expect(eps[2].Protocol).To(Equal("TCP"))
		Expect(eps[2].Port).To(Equal("80"))
		Expect(eps[2].Url).To(Equal("http://2.2.2.2/"))
		Expect(eps[2].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-multi"}))

		// ingress 2.2.2.2, port 443
		Expect(eps[3].SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(eps[3].Host).To(Equal("2.2.2.2"))
		Expect(eps[3].Path).To(Equal("/"))
		Expect(eps[3].TlsEnabled).To(HaveValue(BeTrue()))
		Expect(eps[3].Protocol).To(Equal("TCP"))
		Expect(eps[3].Port).To(Equal("443"))
		Expect(eps[3].Url).To(Equal("https://2.2.2.2/"))
		Expect(eps[3].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb-multi"}))
	})

	It("returns nothing when no ingress status", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb-empty"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findLoadBalancerEndpoints(svc)
		Expect(eps).To(BeEmpty())
	})
})

// ---------- Ingress (ClusterIP) ----------

var _ = Describe("findIngressEndpointsForService", func() {
	BeforeEach(func() {
		os.Setenv(config.EnvVarClusterDomain, "https://app.example.com:443")
	})
	AfterEach(func() {
		os.Unsetenv(config.EnvVarClusterDomain)
	})

	It("returns nil when ingressList is nil", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		eps := NewNetworkStatusUpdater(nil).findIngressEndpointsForService(svc, nil)
		Expect(eps).To(BeNil())
	})

	It("creates endpoint for matching ingress rule", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		ingressList := &networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "my-ingress", Namespace: "ns"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "app.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{
											{
												Path: "/api",
												Backend: networkingv1.IngressBackend{
													Service: &networkingv1.IngressServiceBackend{
														Name: "my-svc",
													},
												},
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
		eps := NewNetworkStatusUpdater(nil).findIngressEndpointsForService(svc, ingressList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(ep.Host).To(Equal("app.example.com"))
		Expect(ep.Path).To(Equal("/api"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Url).To(Equal("http://app.example.com:443/api"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "my-ingress"}))
	})

	It("marks TLS when ingress has TLS for the host", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc-tls"}}
		ingressList := &networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ing-tls"},
					Spec: networkingv1.IngressSpec{
						TLS: []networkingv1.IngressTLS{
							{Hosts: []string{"secure.example.com"}},
						},
						Rules: []networkingv1.IngressRule{
							{
								Host: "secure.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{
											{
												Path: "/",
												Backend: networkingv1.IngressBackend{
													Service: &networkingv1.IngressServiceBackend{Name: "svc-tls"},
												},
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
		eps := NewNetworkStatusUpdater(nil).findIngressEndpointsForService(svc, ingressList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(ep.Host).To(Equal("secure.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://secure.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "ing-tls"}))
	})

	It("skips rules with no HTTP section", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		ingressList := &networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ing-no-http"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{Host: "example.com"},
						},
					},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findIngressEndpointsForService(svc, ingressList)
		Expect(eps).To(BeEmpty())
	})

	It("skips paths pointing to a different service", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc-a"}}
		ingressList := &networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ing"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{
											{
												Path: "/other",
												Backend: networkingv1.IngressBackend{
													Service: &networkingv1.IngressServiceBackend{Name: "svc-b"},
												},
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
		eps := NewNetworkStatusUpdater(nil).findIngressEndpointsForService(svc, ingressList)
		Expect(eps).To(BeEmpty())
	})

	It("defaults empty path to /", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		ingressList := &networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ing"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{
											{
												Path: "",
												Backend: networkingv1.IngressBackend{
													Service: &networkingv1.IngressServiceBackend{Name: "svc"},
												},
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
		eps := NewNetworkStatusUpdater(nil).findIngressEndpointsForService(svc, ingressList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(ep.Host).To(Equal("example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("http://example.com:443/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "ing"}))
	})
})

// ---------- OpenShift Route ----------

func admittedRouteIngress(host string) []openshiftV1.RouteIngress {
	return []openshiftV1.RouteIngress{
		{
			Host: host,
			Conditions: []openshiftV1.RouteIngressCondition{
				{Type: openshiftV1.RouteAdmitted, Status: corev1.ConditionTrue},
			},
		},
	}
}

var _ = Describe("findRouteEndpointsForService", func() {
	It("returns nil when routeList is nil", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, nil)
		Expect(eps).To(BeNil())
	})

	It("creates HTTP endpoint for admitted Route without TLS", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "my-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "app.example.com",
						Path: "/api",
						To:   openshiftV1.RouteTargetReference{Name: "my-svc"},
					},
					Status: openshiftV1.RouteStatus{Ingress: admittedRouteIngress("app.example.com")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(ep.Host).To(Equal("app.example.com"))
		Expect(ep.Path).To(Equal("/api"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Url).To(Equal("http://app.example.com/api"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Route", Name: "my-route"}))
	})

	It("creates HTTPS endpoint for admitted Route with TLS", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "secure.example.com",
						To:   openshiftV1.RouteTargetReference{Name: "my-svc"},
						TLS:  &openshiftV1.TLSConfig{Termination: openshiftV1.TLSTerminationEdge},
					},
					Status: openshiftV1.RouteStatus{Ingress: admittedRouteIngress("secure.example.com")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(ep.Host).To(Equal("secure.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.Url).To(Equal("https://secure.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Route", Name: "tls-route"}))
	})

	It("uses admitted host from status over spec host", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "status-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "spec.example.com",
						To:   openshiftV1.RouteTargetReference{Name: "my-svc"},
					},
					Status: openshiftV1.RouteStatus{Ingress: admittedRouteIngress("actual.apps.cluster.com")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(HaveLen(1))
		Expect(eps[0].Host).To(Equal("actual.apps.cluster.com"))
	})

	It("falls back to spec host when admitted host is empty", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "fallback-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "spec.example.com",
						To:   openshiftV1.RouteTargetReference{Name: "my-svc"},
					},
					Status: openshiftV1.RouteStatus{Ingress: admittedRouteIngress("")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(HaveLen(1))
		Expect(eps[0].Host).To(Equal("spec.example.com"))
	})

	It("skips Route that is not admitted", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pending-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "pending.example.com",
						To:   openshiftV1.RouteTargetReference{Name: "my-svc"},
					},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(BeEmpty())
	})

	It("skips Route targeting a different service", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "other-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "other.example.com",
						To:   openshiftV1.RouteTargetReference{Name: "other-svc"},
					},
					Status: openshiftV1.RouteStatus{Ingress: admittedRouteIngress("other.example.com")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(BeEmpty())
	})

	It("matches Route via alternate backend", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "alt-svc"}}
		routeList := &openshiftV1.RouteList{
			Items: []openshiftV1.Route{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "alt-route"},
					Spec: openshiftV1.RouteSpec{
						Host: "alt.example.com",
						To:   openshiftV1.RouteTargetReference{Name: "primary-svc"},
						AlternateBackends: []openshiftV1.RouteTargetReference{
							{Name: "alt-svc"},
						},
					},
					Status: openshiftV1.RouteStatus{Ingress: admittedRouteIngress("alt.example.com")},
				},
			},
		}
		eps := NewNetworkStatusUpdater(nil).findRouteEndpointsForService(svc, routeList)
		Expect(eps).To(HaveLen(1))
		Expect(eps[0].Host).To(Equal("alt.example.com"))
		Expect(eps[0].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Route", Name: "alt-route"}))
	})
})

// ---------- HTTPRoute (GatewayAPI, ClusterIP) ----------

var _ = Describe("findHTTPRouteEndpointsForService", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns nil when httpRouteList is nil", func() {
		fc := newFakeClient(scheme)
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, nil)
		Expect(eps).To(BeNil())
	})

	It("creates endpoints for matching HTTPRoute with route hostnames and path matches", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "http",
						Port:     8080,
						Protocol: gwv1.HTTPProtocolType,
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route1", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"app.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{
								{Name: "gw"},
							},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								Matches: []gwv1.HTTPRouteMatch{
									{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/api")}},
								},
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("app.example.com"))
		Expect(ep.Path).To(Equal("/api"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Url).To(Equal("http://app.example.com:8080/api"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route1"}))
	})

	It("creates endpoints for HTTPS listener (TLS enabled)", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-https", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "https",
						Port:     443,
						Protocol: gwv1.HTTPSProtocolType,
						TLS:      &gwv1.ListenerTLSConfig{},
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-https", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"secure.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-https"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-https"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("secure.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://secure.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-https"}))
	})

	It("falls back to listener hostname when route has no hostnames", func() {
		listenerHost := gwv1.Hostname("listener.example.com")
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-lh", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "http",
						Port:     80,
						Protocol: gwv1.HTTPProtocolType,
						Hostname: &listenerHost,
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-no-host", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-lh"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-lh"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("listener.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Url).To(Equal("http://listener.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-no-host"}))
	})

	It("preserves wildcard hostnames as-is", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-wc", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-wc", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"*.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-wc"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-wc"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("*.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Url).To(Equal("http://*.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-wc"}))
	})

	It("produces cartesian product of multiple hostnames and paths", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-multi", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-multi", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"a.example.com", "b.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-multi"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								Matches: []gwv1.HTTPRouteMatch{
									{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/api")}},
									{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/v2")}},
								},
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-multi"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		// 2 hostnames * 2 paths = 4 endpoints
		Expect(eps).To(HaveLen(4))

		for _, ep := range eps {
			Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
			Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
			Expect(ep.Protocol).To(Equal("http"))
			Expect(ep.Port).To(Equal("80"))
			Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-multi"}))
		}

		urls := make(map[string]bool)
		for _, ep := range eps {
			urls[ep.Url] = true
		}
		Expect(urls).To(HaveKey("http://a.example.com/api"))
		Expect(urls).To(HaveKey("http://a.example.com/v2"))
		Expect(urls).To(HaveKey("http://b.example.com/api"))
		Expect(urls).To(HaveKey("http://b.example.com/v2"))
	})

	It("defaults path to / when matches list is empty", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-nopath", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-nopath", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-nopath"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								// no matches => defaults to "/"
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-nopath"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("host.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Url).To(Equal("http://host.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-nopath"}))
	})

	It("skips backends referencing a different service", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-other", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "other-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(BeEmpty())
	})

	It("filters listeners by parentRef SectionName", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-section", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
					{Name: "https", Port: 443, Protocol: gwv1.HTTPSProtocolType, TLS: &gwv1.ListenerTLSConfig{}},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		sectionName := gwv1.SectionName("https")
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-section", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-section", SectionName: &sectionName}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-section", SectionName: &sectionName}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("host.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://host.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-section"}))
	})

	It("filters listeners by parentRef Port", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-port", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
					{Name: "http-alt", Port: 8080, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		refPort := gwv1.PortNumber(8080)
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-port", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-port", Port: &refPort}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-port", Port: &refPort}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("host.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Url).To(Equal("http://host.example.com:8080/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-port"}))
	})

	It("resolves gateway from different namespace via parentRef.Namespace", func() {
		gwNs := gwv1.Namespace("gw-ns")
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-cross", Namespace: "gw-ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-cross", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"cross.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-cross", Namespace: &gwNs}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-cross", Namespace: &gwNs}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("cross.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Url).To(Equal("http://cross.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-cross"}))
	})

	It("returns no endpoints when gateway is not found", func() {
		fc := newFakeClient(scheme) // no gateway

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-no-gw", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "nonexistent-gw"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "nonexistent-gw"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(BeEmpty())
	})

	It("returns no endpoints when no hostnames are available from route or listener", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-nohost", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-nohost", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						// no hostnames on route, no hostname on listener
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-nohost"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-nohost"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)
		Expect(eps).To(BeEmpty())
	})
})

// ---------- GRPCRoute (GatewayAPI, ClusterIP) ----------

var _ = Describe("findGRPCRouteEndpointsForService", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns nil when grpcRouteList is nil", func() {
		fc := newFakeClient(scheme)
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, nil)
		Expect(eps).To(BeNil())
	})

	It("creates endpoints for matching GRPCRoute with route hostnames", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "http",
						Port:     8080,
						Protocol: gwv1.HTTPProtocolType,
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "grpc-route1", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						Hostnames: []gwv1.Hostname{"grpc.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{
								{Name: "gw"},
							},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("grpc.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Url).To(Equal("http://grpc.example.com:8080/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "GRPCRoute", Name: "grpc-route1"}))
	})

	It("creates endpoints for HTTPS listener (TLS enabled)", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-https", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "https",
						Port:     443,
						Protocol: gwv1.HTTPSProtocolType,
						TLS:      &gwv1.ListenerTLSConfig{},
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "grpc-route-https", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						Hostnames: []gwv1.Hostname{"secure-grpc.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-https"}},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-https"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("secure-grpc.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://secure-grpc.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "GRPCRoute", Name: "grpc-route-https"}))
	})

	It("falls back to listener hostname when route has no hostnames", func() {
		listenerHost := gwv1.Hostname("listener.example.com")
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-lh", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "http",
						Port:     80,
						Protocol: gwv1.HTTPProtocolType,
						Hostname: &listenerHost,
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "grpc-route-no-host", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-lh"}},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-lh"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.Host).To(Equal("listener.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.Url).To(Equal("http://listener.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "GRPCRoute", Name: "grpc-route-no-host"}))
	})

	It("skips backends referencing a different service", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "grpc-route-other", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "other-svc"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)
		Expect(eps).To(BeEmpty())
	})

	It("returns no endpoints when gateway is not found", func() {
		fc := newFakeClient(scheme) // no gateway

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "grpc-route-no-gw", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "nonexistent-gw"}},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "nonexistent-gw"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)
		Expect(eps).To(BeEmpty())
	})

	It("produces endpoints for multiple hostnames", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-multi", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "grpc-route-multi", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						Hostnames: []gwv1.Hostname{"a.example.com", "b.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-multi"}},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-multi"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)
		Expect(eps).To(HaveLen(2))

		urls := make(map[string]bool)
		for _, ep := range eps {
			Expect(ep.Path).To(Equal("/"))
			Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "GRPCRoute", Name: "grpc-route-multi"}))
			urls[ep.Url] = true
		}
		Expect(urls).To(HaveKey("http://a.example.com/"))
		Expect(urls).To(HaveKey("http://b.example.com/"))
	})
})

// ---------- TLSRoute (GatewayAPI, ClusterIP) ----------

var _ = Describe("findTLSRouteEndpointsForService", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns nil when tlsRouteList is nil", func() {
		fc := newFakeClient(scheme)
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, nil)
		Expect(eps).To(BeNil())
	})

	It("creates endpoints for matching TLSRoute with route hostnames", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "tls",
						Port:     8443,
						Protocol: gwv1.TLSProtocolType,
						TLS:      &gwv1.ListenerTLSConfig{},
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route1", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						Hostnames: []gwv1.Hostname{"secure.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("secure.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.Port).To(Equal("8443"))
		Expect(ep.Protocol).To(Equal("tls"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Url).To(Equal("https://secure.example.com:8443/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "tls-route1"}))
	})

	It("falls back to listener hostname when route has no hostnames", func() {
		listenerHost := gwv1.Hostname("listener-tls.example.com")
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls-lh", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "tls",
						Port:     443,
						Protocol: gwv1.TLSProtocolType,
						Hostname: &listenerHost,
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route-nohost", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls-lh"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls-lh"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("listener-tls.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("tls"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://listener-tls.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "tls-route-nohost"}))
	})

	It("preserves wildcard hostnames as-is", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls-wc", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls", Port: 443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route-wc", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						Hostnames: []gwv1.Hostname{"*.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls-wc"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls-wc"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("*.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("tls"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Url).To(Equal("https://*.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "tls-route-wc"}))
	})

	It("creates one endpoint per hostname (no path permutations for TLS)", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls-multi", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls", Port: 443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route-multi", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						Hostnames: []gwv1.Hostname{"a.example.com", "b.example.com", "c.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls-multi"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls-multi"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(HaveLen(3))
		hosts := map[string]bool{}
		for _, ep := range eps {
			Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
			Expect(ep.Path).To(Equal("/"))
			Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
			Expect(ep.Protocol).To(Equal("tls"))
			Expect(ep.Port).To(Equal("443"))
			Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "tls-route-multi"}))
			hosts[ep.Host] = true
		}
		Expect(hosts).To(HaveKey("a.example.com"))
		Expect(hosts).To(HaveKey("b.example.com"))
		Expect(hosts).To(HaveKey("c.example.com"))
	})

	It("returns no endpoints when no hostnames are available from route or listener", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls-nohost", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls", Port: 443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route-nohost", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls-nohost"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls-nohost"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(BeEmpty())
	})

	It("filters by parentRef SectionName", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls-sec", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls-a", Port: 8443, Protocol: gwv1.TLSProtocolType},
					{Name: "tls-b", Port: 9443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		sectionName := gwv1.SectionName("tls-b")
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route-sec", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						Hostnames: []gwv1.Hostname{"sec.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls-sec", SectionName: &sectionName}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls-sec", SectionName: &sectionName}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("sec.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Protocol).To(Equal("tls"))
		Expect(ep.Port).To(Equal("9443"))
		Expect(ep.Url).To(Equal("https://sec.example.com:9443/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "tls-route-sec"}))
	})

	It("skips backends referencing a different service", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls-other", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls", Port: 443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "my-svc", Namespace: "ns"}}
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tls-route-other", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						Hostnames: []gwv1.Hostname{"other.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls-other"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "other-svc"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls-other"}}),
						},
					},
				},
			},
		}

		eps := NewNetworkStatusUpdater(fc).findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)
		Expect(eps).To(BeEmpty())
	})
})

// ---------- extractEndpointsFromService ----------

var _ = Describe("extractEndpointsFromService", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
		os.Setenv(config.EnvVarClusterDomain, "https://app.run.ai:443")
	})

	AfterEach(func() {
		os.Unsetenv(config.EnvVarClusterDomain)
	})

	It("returns correct endpoints for all service types", func() {
		// Create gateways needed by HTTPRoute and TLSRoute
		gwHTTP := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		gwTLS := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-tls", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls", Port: 443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gwHTTP, gwTLS)

		var allEndpoints []v1alpha1.ExternalEndpoint

		// --- NodePort service ---
		svcNP := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		allEndpoints = append(allEndpoints, NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svcNP, nil, nil, nil, nil, nil, "10.0.0.1", false)...)

		// --- LoadBalancer service ---
		svcLB := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-lb", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: corev1.ProtocolTCP},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}},
				},
			},
		}
		allEndpoints = append(allEndpoints, NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svcLB, nil, nil, nil, nil, nil, "", false)...)

		// --- ClusterIP service with Ingress ---
		svcCI := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-ci", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		ingressList := &networkingv1.IngressList{
			Items: []networkingv1.Ingress{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "ing", Namespace: "ns"},
					Spec: networkingv1.IngressSpec{
						Rules: []networkingv1.IngressRule{
							{
								Host: "app.run.ai",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{
											{
												Path: "/",
												Backend: networkingv1.IngressBackend{
													Service: &networkingv1.IngressServiceBackend{Name: "svc-ci"},
												},
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
		allEndpoints = append(allEndpoints, NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svcCI, ingressList, nil, nil, nil, nil, "", false)...)

		// --- ClusterIP service with HTTPRoute ---
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "hr", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"host.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "svc-ci"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}
		allEndpoints = append(allEndpoints, NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svcCI, nil, httpRouteList, nil, nil, nil, "", false)...)

		// --- ClusterIP service with GRPCRoute ---
		grpcRouteList := &gwv1.GRPCRouteList{
			Items: []gwv1.GRPCRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "gr", Namespace: "ns"},
					Spec: gwv1.GRPCRouteSpec{
						Hostnames: []gwv1.Hostname{"grpc.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
						},
						Rules: []gwv1.GRPCRouteRule{
							{
								BackendRefs: []gwv1.GRPCBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "svc-ci"}}},
								},
							},
						},
					},
					Status: gwv1.GRPCRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}
		allEndpoints = append(allEndpoints, NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svcCI, nil, nil, grpcRouteList, nil, nil, "", false)...)

		// --- ClusterIP service with TLSRoute ---
		tlsRouteList := &gwv1alpha2.TLSRouteList{
			Items: []gwv1alpha2.TLSRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "tr", Namespace: "ns"},
					Spec: gwv1alpha2.TLSRouteSpec{
						Hostnames: []gwv1.Hostname{"tls.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-tls"}},
						},
						Rules: []gwv1alpha2.TLSRouteRule{
							{
								BackendRefs: []gwv1.BackendRef{
									{BackendObjectReference: gwv1.BackendObjectReference{Name: "svc-ci"}},
								},
							},
						},
					},
					Status: gwv1alpha2.TLSRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-tls"}}),
						},
					},
				},
			},
		}
		allEndpoints = append(allEndpoints, NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svcCI, nil, nil, nil, tlsRouteList, nil, "", false)...)

		// --- Validate all 6 endpoints ---
		Expect(allEndpoints).To(HaveLen(6))

		// Endpoint 0: NodePort
		Expect(allEndpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(allEndpoints[0].Host).To(Equal("10.0.0.1"))
		Expect(allEndpoints[0].Path).To(Equal("/"))
		Expect(allEndpoints[0].TlsEnabled).To(BeNil())
		Expect(allEndpoints[0].Protocol).To(Equal("TCP"))
		Expect(allEndpoints[0].Port).To(Equal("30080"))
		Expect(allEndpoints[0].Url).To(Equal("10.0.0.1:30080"))
		Expect(allEndpoints[0].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"}))

		// Endpoint 1: LoadBalancer
		Expect(allEndpoints[1].SourceType).To(Equal(v1alpha1.ExternalEndpointLoadBalanceSource))
		Expect(allEndpoints[1].Host).To(Equal("1.2.3.4"))
		Expect(allEndpoints[1].Path).To(Equal("/"))
		Expect(allEndpoints[1].TlsEnabled).To(HaveValue(BeTrue()))
		Expect(allEndpoints[1].Protocol).To(Equal("TCP"))
		Expect(allEndpoints[1].Port).To(Equal("80"))
		Expect(allEndpoints[1].Url).To(Equal("1.2.3.4:80"))
		Expect(allEndpoints[1].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-lb"}))

		// Endpoint 2: Ingress
		Expect(allEndpoints[2].SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(allEndpoints[2].Host).To(Equal("app.run.ai"))
		Expect(allEndpoints[2].Path).To(Equal("/"))
		Expect(allEndpoints[2].TlsEnabled).To(HaveValue(BeFalse()))
		Expect(allEndpoints[2].Protocol).To(Equal("http"))
		Expect(allEndpoints[2].Port).To(Equal("443"))
		Expect(allEndpoints[2].Url).To(Equal("http://app.run.ai:443/"))
		Expect(allEndpoints[2].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "ing"}))

		// Endpoint 3: HTTPRoute
		Expect(allEndpoints[3].SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(allEndpoints[3].Host).To(Equal("host.example.com"))
		Expect(allEndpoints[3].Path).To(Equal("/"))
		Expect(allEndpoints[3].TlsEnabled).To(HaveValue(BeFalse()))
		Expect(allEndpoints[3].Protocol).To(Equal("http"))
		Expect(allEndpoints[3].Port).To(Equal("80"))
		Expect(allEndpoints[3].Url).To(Equal("http://host.example.com/"))
		Expect(allEndpoints[3].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "hr"}))

		// Endpoint 4: GRPCRoute
		Expect(allEndpoints[4].SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(allEndpoints[4].Host).To(Equal("grpc.example.com"))
		Expect(allEndpoints[4].Path).To(Equal("/"))
		Expect(allEndpoints[4].TlsEnabled).To(HaveValue(BeFalse()))
		Expect(allEndpoints[4].Protocol).To(Equal("http"))
		Expect(allEndpoints[4].Port).To(Equal("80"))
		Expect(allEndpoints[4].Url).To(Equal("http://grpc.example.com/"))
		Expect(allEndpoints[4].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "GRPCRoute", Name: "gr"}))

		// Endpoint 5: TLSRoute
		Expect(allEndpoints[5].SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(allEndpoints[5].Host).To(Equal("tls.example.com"))
		Expect(allEndpoints[5].Path).To(Equal("/"))
		Expect(allEndpoints[5].TlsEnabled).To(HaveValue(BeTrue()))
		Expect(allEndpoints[5].Protocol).To(Equal("tls"))
		Expect(allEndpoints[5].Port).To(Equal("443"))
		Expect(allEndpoints[5].Url).To(Equal("https://tls.example.com/"))
		Expect(allEndpoints[5].SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "tr"}))
	})

	It("returns empty for unsupported service type (ExternalName)", func() {
		fc := newFakeClient(scheme)
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-ext"},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeExternalName,
			},
		}
		eps := NewNetworkStatusUpdater(fc).extractEndpointsFromService(ctx, svc, nil, nil, nil, nil, nil, "", false)
		Expect(eps).To(BeEmpty())
	})

})

// ---------- helper functions ----------

var _ = Describe("getNodeAddress", func() {
	var scheme *runtime.Scheme

	BeforeEach(func() {
		scheme = testScheme()
	})

	It("returns empty for empty nodeName", func() {
		fc := newFakeClient(scheme)
		addr := getNodeAddress(context.Background(), fc, "")
		Expect(addr).To(BeEmpty())
	})

	It("returns external IP when available", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.1"},
					{Type: corev1.NodeExternalIP, Address: "203.0.113.1"},
				},
			},
		}
		fc := newFakeClient(scheme, node)
		addr := getNodeAddress(context.Background(), fc, "node1")
		Expect(addr).To(Equal("203.0.113.1"))
	})

	It("falls back to internal IP when no external IP", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node2"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.2"},
				},
			},
		}
		fc := newFakeClient(scheme, node)
		addr := getNodeAddress(context.Background(), fc, "node2")
		Expect(addr).To(Equal("192.168.1.2"))
	})

	It("uses the annotated external IP before the internal IP", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "node-with-annotated-external-ip",
				Annotations: map[string]string{externalNodeIPAnnotation: "203.0.113.2"},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.2"},
				},
			},
		}
		fc := newFakeClient(scheme, node)
		addr := getNodeAddress(context.Background(), fc, node.Name)
		Expect(addr).To(Equal("203.0.113.2"))
	})

	It("prefers a Kubernetes external IP over the annotation", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "node-with-external-ip-and-annotation",
				Annotations: map[string]string{externalNodeIPAnnotation: "203.0.113.3"},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.3"},
					{Type: corev1.NodeExternalIP, Address: "198.51.100.3"},
				},
			},
		}
		fc := newFakeClient(scheme, node)
		addr := getNodeAddress(context.Background(), fc, node.Name)
		Expect(addr).To(Equal("198.51.100.3"))
	})

	It("ignores an empty external IP annotation", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "node-with-empty-external-ip-annotation",
				Annotations: map[string]string{externalNodeIPAnnotation: ""},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.4"},
				},
			},
		}
		fc := newFakeClient(scheme, node)
		addr := getNodeAddress(context.Background(), fc, node.Name)
		Expect(addr).To(Equal("192.168.1.4"))
	})

	It("returns empty when node not found", func() {
		fc := newFakeClient(scheme)
		addr := getNodeAddress(context.Background(), fc, "nonexistent")
		Expect(addr).To(BeEmpty())
	})
})

var _ = Describe("routeOrListenerHostnames", func() {
	It("returns route hostnames when present", func() {
		h := routeOrListenerHostnames([]gwv1.Hostname{"a.com", "b.com"}, nil)
		Expect(h).To(Equal([]string{"a.com", "b.com"}))
	})

	It("falls back to listener hostname", func() {
		lh := gwv1.Hostname("listener.com")
		h := routeOrListenerHostnames(nil, &lh)
		Expect(h).To(Equal([]string{"listener.com"}))
	})

	It("returns nil when both are empty", func() {
		h := routeOrListenerHostnames(nil, nil)
		Expect(h).To(BeNil())
	})
})

var _ = Describe("routeMatchPaths", func() {
	It("returns / for empty matches", func() {
		p := routeMatchPaths(nil)
		Expect(p).To(Equal([]string{"/"}))
	})

	It("extracts paths from matches", func() {
		p := routeMatchPaths([]gwv1.HTTPRouteMatch{
			{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/api")}},
			{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/v2")}},
		})
		Expect(p).To(Equal([]string{"/api", "/v2"}))
	})

	It("defaults to / when path is nil in match", func() {
		p := routeMatchPaths([]gwv1.HTTPRouteMatch{
			{},
		})
		Expect(p).To(Equal([]string{"/"}))
	})

	It("defaults to / when path value is empty", func() {
		p := routeMatchPaths([]gwv1.HTTPRouteMatch{
			{Path: &gwv1.HTTPPathMatch{Value: ptr.To("")}},
		})
		Expect(p).To(Equal([]string{"/"}))
	})
})

var _ = Describe("hasTLSForHost", func() {
	It("returns true when host is in TLS list", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				TLS: []networkingv1.IngressTLS{
					{Hosts: []string{"a.com", "b.com"}},
				},
			},
		}
		Expect(hasTLSForHost(ing, "b.com")).To(BeTrue())
	})

	It("returns false when host is not in TLS list", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				TLS: []networkingv1.IngressTLS{
					{Hosts: []string{"a.com"}},
				},
			},
		}
		Expect(hasTLSForHost(ing, "c.com")).To(BeFalse())
	})

	It("returns false when no TLS configured", func() {
		ing := &networkingv1.Ingress{}
		Expect(hasTLSForHost(ing, "a.com")).To(BeFalse())
	})
})

var _ = Describe("hasTlsAnnotation", func() {
	It("returns true for AWS SSL cert annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/aws-load-balancer-ssl-cert": "arn:aws:acm:us-east-1:123:cert/abc",
		})).To(BeTrue())
	})

	It("returns true for DigitalOcean cert annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/do-loadbalancer-certificate-id": "cert-id",
		})).To(BeTrue())
	})

	It("returns true for Azure cert annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/azure-load-balancer-ssl-cert-managed": "true",
		})).To(BeTrue())
	})

	It("returns true for Alibaba cert annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cert-id": "cert-id",
		})).To(BeTrue())
	})

	It("returns true for AWS SSL ports annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/aws-load-balancer-ssl-ports": "443",
		})).To(BeTrue())
	})

	It("returns true for DigitalOcean TLS ports annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/do-loadbalancer-tls-ports": "443",
		})).To(BeTrue())
	})

	It("returns false for empty annotations", func() {
		Expect(hasTlsAnnotation(map[string]string{})).To(BeFalse())
	})

	It("returns false for nil annotations", func() {
		Expect(hasTlsAnnotation(nil)).To(BeFalse())
	})

	It("returns false for empty value in TLS annotation", func() {
		Expect(hasTlsAnnotation(map[string]string{
			"service.beta.kubernetes.io/aws-load-balancer-ssl-cert": "",
		})).To(BeFalse())
	})
})

var _ = Describe("labelsMatchSelector", func() {
	It("returns true when all selector labels match", func() {
		labels := map[string]string{"app": "web", "env": "prod", "extra": "val"}
		selector := map[string]string{"app": "web", "env": "prod"}
		Expect(labelsMatchSelector(labels, selector)).To(BeTrue())
	})

	It("returns false when a selector label is missing", func() {
		labels := map[string]string{"app": "web"}
		selector := map[string]string{"app": "web", "env": "prod"}
		Expect(labelsMatchSelector(labels, selector)).To(BeFalse())
	})

	It("returns false when a selector value does not match", func() {
		labels := map[string]string{"app": "api"}
		selector := map[string]string{"app": "web"}
		Expect(labelsMatchSelector(labels, selector)).To(BeFalse())
	})

	It("returns true for empty selector", func() {
		labels := map[string]string{"app": "web"}
		Expect(labelsMatchSelector(labels, map[string]string{})).To(BeTrue())
	})

	It("returns true when both nil", func() {
		Expect(labelsMatchSelector(nil, nil)).To(BeTrue())
	})

	It("returns false when labels nil but selector non-empty", func() {
		Expect(labelsMatchSelector(nil, map[string]string{"app": "web"})).To(BeFalse())
	})

	It("returns true when both empty maps", func() {
		Expect(labelsMatchSelector(map[string]string{}, map[string]string{})).To(BeTrue())
	})
})

var _ = Describe("findGwListeners", func() {
	It("returns all listeners matching compatible protocols", func() {
		gw := &gwv1.Gateway{
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
					{Name: "https", Port: 443, Protocol: gwv1.HTTPSProtocolType},
					{Name: "tls", Port: 8443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		ref := gwv1.RouteParentStatus{}
		matched := findGwListeners(gw, ref, []gwv1.ProtocolType{gwv1.HTTPProtocolType, gwv1.HTTPSProtocolType})
		Expect(matched).To(HaveLen(2))
		Expect(matched[0].Name).To(Equal(gwv1.SectionName("http")))
		Expect(matched[1].Name).To(Equal(gwv1.SectionName("https")))
	})

	It("filters by SectionName", func() {
		gw := &gwv1.Gateway{
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
					{Name: "https", Port: 443, Protocol: gwv1.HTTPSProtocolType},
				},
			},
		}
		sn := gwv1.SectionName("https")
		ref := gwv1.RouteParentStatus{ParentRef: gwv1.ParentReference{SectionName: &sn}}
		matched := findGwListeners(gw, ref, []gwv1.ProtocolType{gwv1.HTTPProtocolType, gwv1.HTTPSProtocolType})
		Expect(matched).To(HaveLen(1))
		Expect(matched[0].Name).To(Equal(gwv1.SectionName("https")))
	})

	It("filters by Port", func() {
		gw := &gwv1.Gateway{
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
					{Name: "http-alt", Port: 8080, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		p := gwv1.PortNumber(8080)
		ref := gwv1.RouteParentStatus{ParentRef: gwv1.ParentReference{Port: &p}}
		matched := findGwListeners(gw, ref, []gwv1.ProtocolType{gwv1.HTTPProtocolType})
		Expect(matched).To(HaveLen(1))
		Expect(matched[0].Port).To(Equal(gwv1.PortNumber(8080)))
	})

	It("returns empty when no protocols match", func() {
		gw := &gwv1.Gateway{
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 80, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		ref := gwv1.RouteParentStatus{ParentRef: gwv1.ParentReference{}}
		matched := findGwListeners(gw, ref, []gwv1.ProtocolType{gwv1.TLSProtocolType})
		Expect(matched).To(BeEmpty())
	})
})

// ---------- findServicesForPod ----------

var _ = Describe("findServicesForPod", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns services whose selector matches pod labels", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-match", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web", "extra": "val"}},
		}
		fc := newFakeClient(scheme, svc)
		h := NewNetworkStatusUpdater(fc)

		services, err := h.findServicesForPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(HaveLen(1))
		Expect(services[0].Name).To(Equal("svc-match"))
	})

	It("skips services with nil selector", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-nil-sel", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{Port: 80}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, svc)
		h := NewNetworkStatusUpdater(fc)

		services, err := h.findServicesForPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(BeEmpty())
	})

	It("returns empty when no services match", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-no-match", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "api"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, svc)
		h := NewNetworkStatusUpdater(fc)

		services, err := h.findServicesForPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(BeEmpty())
	})

	It("returns empty when namespace has no services", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme)
		h := NewNetworkStatusUpdater(fc)

		services, err := h.findServicesForPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(BeEmpty())
	})

	It("does not return services from a different namespace", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-other-ns", Namespace: "other-ns"},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, svc)
		h := NewNetworkStatusUpdater(fc)

		services, err := h.findServicesForPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(BeEmpty())
	})

	It("returns multiple matching services", func() {
		svc1 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		svc2 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 8080}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, svc1, svc2)
		h := NewNetworkStatusUpdater(fc)

		services, err := h.findServicesForPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(services).To(HaveLen(2))
		names := []string{services[0].Name, services[1].Name}
		Expect(names).To(ContainElements("svc1", "svc2"))
	})
})

// ---------- extractServiceNamesFromIngress ----------

var _ = Describe("extractServiceNamesFromIngress", func() {
	It("extracts service names from rules", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "svc-a"}}},
								{Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "svc-b"}}},
							},
						},
					}},
				},
			},
		}
		names := extractServiceNamesFromIngress(ing)
		Expect(names).To(ConsistOf("svc-a", "svc-b"))
	})

	It("includes default backend service", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				DefaultBackend: &networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{Name: "default-svc"},
				},
			},
		}
		names := extractServiceNamesFromIngress(ing)
		Expect(names).To(ConsistOf("default-svc"))
	})

	It("deduplicates service names", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "svc-a"}}},
								{Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "svc-a"}}},
							},
						},
					}},
				},
			},
		}
		names := extractServiceNamesFromIngress(ing)
		Expect(names).To(HaveLen(1))
		Expect(names).To(ConsistOf("svc-a"))
	})

	It("skips rules without HTTP", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{IngressRuleValue: networkingv1.IngressRuleValue{HTTP: nil}},
				},
			},
		}
		names := extractServiceNamesFromIngress(ing)
		Expect(names).To(BeEmpty())
	})
})

// ---------- extractServiceNamesFromHTTPRoute ----------

var _ = Describe("extractServiceNamesFromHTTPRoute", func() {
	It("extracts service names from backend refs", func() {
		route := &gwv1.HTTPRoute{
			Spec: gwv1.HTTPRouteSpec{
				Rules: []gwv1.HTTPRouteRule{
					{BackendRefs: []gwv1.HTTPBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "svc-x"}}},
					}},
				},
			},
		}
		names := extractServiceNamesFromHTTPRoute(route)
		Expect(names).To(ConsistOf("svc-x"))
	})

	It("skips non-Service kind refs", func() {
		nonSvcKind := gwv1.Kind("Bucket")
		route := &gwv1.HTTPRoute{
			Spec: gwv1.HTTPRouteSpec{
				Rules: []gwv1.HTTPRouteRule{
					{BackendRefs: []gwv1.HTTPBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
							Kind: &nonSvcKind,
							Name: "my-bucket",
						}}},
					}},
				},
			},
		}
		names := extractServiceNamesFromHTTPRoute(route)
		Expect(names).To(BeEmpty())
	})
})

// ---------- extractServiceNamesFromGRPCRoute ----------

var _ = Describe("extractServiceNamesFromGRPCRoute", func() {
	It("extracts service names from backend refs", func() {
		route := &gwv1.GRPCRoute{
			Spec: gwv1.GRPCRouteSpec{
				Rules: []gwv1.GRPCRouteRule{
					{BackendRefs: []gwv1.GRPCBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "grpc-svc"}}},
					}},
				},
			},
		}
		names := extractServiceNamesFromGRPCRoute(route)
		Expect(names).To(ConsistOf("grpc-svc"))
	})
})

// ---------- extractServiceNamesFromTLSRoute ----------

var _ = Describe("extractServiceNamesFromTLSRoute", func() {
	It("extracts service names from backend refs", func() {
		route := &gwv1alpha2.TLSRoute{
			Spec: gwv1alpha2.TLSRouteSpec{
				Rules: []gwv1alpha2.TLSRouteRule{
					{BackendRefs: []gwv1.BackendRef{
						{BackendObjectReference: gwv1.BackendObjectReference{Name: "tls-svc"}},
					}},
				},
			},
		}
		names := extractServiceNamesFromTLSRoute(route)
		Expect(names).To(ConsistOf("tls-svc"))
	})
})

// ---------- FindPodsForService ----------

var _ = Describe("FindPodsForService", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns nil when service has no selector", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{},
		}
		fc := newFakeClient(scheme)
		updater := NewNetworkStatusUpdater(fc)

		pods, err := updater.FindPodsForService(ctx, svc)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods).To(BeNil())
	})

	It("returns matching Pods", func() {
		pod1 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		pod2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns", Labels: map[string]string{"app": "api"}},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, pod1, pod2)
		updater := NewNetworkStatusUpdater(fc)

		pods, err := updater.FindPodsForService(ctx, svc)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods).To(HaveLen(1))
		Expect(pods[0].Name).To(Equal("pod1"))
	})
})

// ---------- DiscoverEndpointsForEwAndPods ----------

var _ = Describe("DiscoverEndpointsForEwAndPods", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
		os.Setenv(config.EnvVarClusterDomain, "https://app.example.com:443")
	})

	AfterEach(func() {
		os.Unsetenv(config.EnvVarClusterDomain)
	})

	It("returns empty when no services match the pod", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(BeEmpty())
	})

	It("returns NodePort endpoints for matching NodePort service", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
			Spec:       corev1.PodSpec{NodeName: "node1"},
		}
		fc := newFakeClient(scheme, node, svc)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(1))
		ep := endpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep.Host).To(Equal("10.0.0.1"))
		Expect(ep.Port).To(Equal("30080"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.TlsEnabled).To(BeNil())
		Expect(ep.Url).To(Equal("10.0.0.1:30080"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"}))
	})

	It("returns endpoints from a ClusterIP service matched via HTTPRoute", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 8080, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-cluster", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		httpRoute := &gwv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "route1", Namespace: "ns"},
			Spec: gwv1.HTTPRouteSpec{
				Hostnames: []gwv1.Hostname{"app.example.com"},
				CommonRouteSpec: gwv1.CommonRouteSpec{
					ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
				},
				Rules: []gwv1.HTTPRouteRule{
					{
						Matches: []gwv1.HTTPRouteMatch{
							{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/api")}},
						},
						BackendRefs: []gwv1.HTTPBackendRef{
							{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "svc-cluster"}}},
						},
					},
				},
			},
			Status: gwv1.HTTPRouteStatus{
				RouteStatus: gwv1.RouteStatus{
					Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, gw, svc, httpRoute)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(1))
		ep := endpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("app.example.com"))
		Expect(ep.Path).To(Equal("/api"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Url).To(Equal("http://app.example.com:8080/api"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route1"}))
	})

	It("aggregates endpoints from multiple matching services", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		}
		svc1 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		svc2 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports: []corev1.ServicePort{
					{Port: 8080, NodePort: 30880, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
			Spec:       corev1.PodSpec{NodeName: "node1"},
		}
		fc := newFakeClient(scheme, node, svc1, svc2)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(2))
	})

	It("includes spec.urls as endpoints alongside service endpoints", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec: v1alpha1.ExternalWorkloadSpec{
				Urls: []string{"https://inference.example.com/v1/models", "http://my-service.ns.svc.cluster.local/predict"},
			},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
			Spec:       corev1.PodSpec{NodeName: "node1"},
		}
		fc := newFakeClient(scheme, node, svc)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(3)) // 1 NodePort + 2 from spec.urls

		// First endpoint is from the NodePort service
		Expect(endpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(endpoints[0].Host).To(Equal("10.0.0.1"))
		Expect(endpoints[0].Port).To(Equal("30080"))

		// Second endpoint is from the first spec.url
		Expect(endpoints[1].SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(endpoints[1].Host).To(Equal("inference.example.com"))
		Expect(endpoints[1].Port).To(Equal("443"))
		Expect(endpoints[1].Path).To(Equal("/v1/models"))
		Expect(endpoints[1].Protocol).To(Equal("https"))
		Expect(endpoints[1].TlsEnabled).To(HaveValue(BeTrue()))
		Expect(endpoints[1].Url).To(Equal("https://inference.example.com/v1/models"))

		// Third endpoint is from the second spec.url
		Expect(endpoints[2].SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(endpoints[2].Host).To(Equal("my-service.ns.svc.cluster.local"))
		Expect(endpoints[2].Port).To(Equal("80"))
		Expect(endpoints[2].Path).To(Equal("/predict"))
		Expect(endpoints[2].Protocol).To(Equal("http"))
		Expect(endpoints[2].TlsEnabled).To(HaveValue(BeFalse()))
		Expect(endpoints[2].Url).To(Equal("http://my-service.ns.svc.cluster.local/predict"))
	})

	It("includes spec.urls as endpoints even with no matching services", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec: v1alpha1.ExternalWorkloadSpec{
				Urls: []string{"https://inference.example.com/v1/models"},
			},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(1))
		Expect(endpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(endpoints[0].Host).To(Equal("inference.example.com"))
		Expect(endpoints[0].Path).To(Equal("/v1/models"))
		Expect(endpoints[0].Url).To(Equal("https://inference.example.com/v1/models"))
	})

	It("includes spec.urls as endpoints even with no pods", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec: v1alpha1.ExternalWorkloadSpec{
				Urls: []string{"https://inference.example.com/v1/models"},
			},
		}
		fc := newFakeClient(scheme)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(1))
		Expect(endpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(endpoints[0].Host).To(Equal("inference.example.com"))
		Expect(endpoints[0].Url).To(Equal("https://inference.example.com/v1/models"))
	})

	It("deduplicates spec.urls that match already discovered service endpoints", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports: []corev1.ServicePort{
					{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP},
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec: v1alpha1.ExternalWorkloadSpec{
				Urls: []string{
					"10.0.0.1:30080", // duplicate of the NodePort endpoint
					"https://inference.example.com/v1/models", // new, should be included
				},
			},
		}
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
			Spec:       corev1.PodSpec{NodeName: "node1"},
		}
		fc := newFakeClient(scheme, node, svc)
		h := NewNetworkStatusUpdater(fc)

		endpoints, err := h.DiscoverEndpointsForEwAndPods(ctx, ew, []corev1.Pod{pod})
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(2)) // 1 NodePort + 1 new from spec.urls (duplicate skipped)
		Expect(endpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(endpoints[1].SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(endpoints[1].Host).To(Equal("inference.example.com"))
	})
})

// ---------- isInferencePoolRef ----------

var _ = Describe("isInferencePoolRef", func() {
	It("matches when Kind, Group, and Name all match", func() {
		ref := gwv1.BackendObjectReference{
			Kind:      ptr.To(gwv1.Kind(kindInferencePool)),
			Group:     ptr.To(gwv1.Group(infv1.GroupName)),
			Name:      "my-pool",
			Namespace: ptr.To(gwv1.Namespace("ns")),
		}
		Expect(isInferencePoolRef(ref, "my-pool", "ns")).To(BeTrue())
	})

	It("does not match when name differs", func() {
		ref := gwv1.BackendObjectReference{
			Kind:      ptr.To(gwv1.Kind(kindInferencePool)),
			Group:     ptr.To(gwv1.Group(infv1.GroupName)),
			Name:      "other-pool",
			Namespace: ptr.To(gwv1.Namespace("ns")),
		}
		Expect(isInferencePoolRef(ref, "my-pool", "ns")).To(BeFalse())
	})

	It("does not match when Kind is nil", func() {
		ref := gwv1.BackendObjectReference{
			Group:     ptr.To(gwv1.Group(infv1.GroupName)),
			Name:      "my-pool",
			Namespace: ptr.To(gwv1.Namespace("ns")),
		}
		Expect(isInferencePoolRef(ref, "my-pool", "ns")).To(BeFalse())
	})

	It("does not match when Group is nil", func() {
		ref := gwv1.BackendObjectReference{
			Kind:      ptr.To(gwv1.Kind(kindInferencePool)),
			Name:      "my-pool",
			Namespace: ptr.To(gwv1.Namespace("ns")),
		}
		Expect(isInferencePoolRef(ref, "my-pool", "ns")).To(BeFalse())
	})

	It("does not match a Service ref with the same name", func() {
		ref := gwv1.BackendObjectReference{
			Kind:      ptr.To(gwv1.Kind("Service")),
			Name:      "my-pool",
			Namespace: ptr.To(gwv1.Namespace("ns")),
		}
		Expect(isInferencePoolRef(ref, "my-pool", "ns")).To(BeFalse())
	})
})

// ---------- httpRoutesToGatewayRoutes + findGatewayRouteEndpoints for InferencePool ----------

var _ = Describe("httpRoutesToGatewayRoutes for InferencePool", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	findEndpointsForPool := func(n *NetworkStatusUpdater, ctx context.Context, pool *infv1.InferencePool, httpRouteList *gwv1.HTTPRouteList) []v1alpha1.ExternalEndpoint {
		if httpRouteList == nil || len(httpRouteList.Items) == 0 {
			return nil
		}
		routes := httpRoutesToGatewayRoutes(httpRouteList)
		return n.findGatewayRouteEndpoints(ctx, pool.Name, func(ref gwv1.BackendObjectReference) bool {
			return isInferencePoolRef(ref, pool.Name, pool.Namespace)
		}, routes)
	}

	It("returns nil when httpRouteList is nil", func() {
		fc := newFakeClient(scheme)
		pool := &infv1.InferencePool{ObjectMeta: metav1.ObjectMeta{Name: "pool1"}}
		eps := findEndpointsForPool(NewNetworkStatusUpdater(fc), ctx, pool, nil)
		Expect(eps).To(BeNil())
	})

	It("returns nil when httpRouteList is empty", func() {
		fc := newFakeClient(scheme)
		pool := &infv1.InferencePool{ObjectMeta: metav1.ObjectMeta{Name: "pool1"}}
		eps := findEndpointsForPool(NewNetworkStatusUpdater(fc), ctx, pool, &gwv1.HTTPRouteList{})
		Expect(eps).To(BeNil())
	})

	It("creates endpoints for matching HTTPRoute referencing an InferencePool backend", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "http",
						Port:     8080,
						Protocol: gwv1.HTTPProtocolType,
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		pool := &infv1.InferencePool{ObjectMeta: metav1.ObjectMeta{Name: "my-pool", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-pool", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"inference.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								Matches: []gwv1.HTTPRouteMatch{
									{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/predict")}},
								},
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
										Kind:      ptr.To(gwv1.Kind(kindInferencePool)),
										Group:     ptr.To(gwv1.Group(infv1.GroupName)),
										Name:      "my-pool",
										Namespace: ptr.To(gwv1.Namespace("ns")),
									}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}

		eps := findEndpointsForPool(NewNetworkStatusUpdater(fc), ctx, pool, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("inference.example.com"))
		Expect(ep.Path).To(Equal("/predict"))
		Expect(ep.Port).To(Equal("8080"))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Url).To(Equal("http://inference.example.com:8080/predict"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-pool"}))
	})

	It("ignores HTTPRoutes that reference a Service, not InferencePool", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 8080, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		pool := &infv1.InferencePool{ObjectMeta: metav1.ObjectMeta{Name: "my-pool", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-svc", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"app.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "my-pool"}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
						},
					},
				},
			},
		}

		eps := findEndpointsForPool(NewNetworkStatusUpdater(fc), ctx, pool, httpRouteList)
		Expect(eps).To(BeEmpty())
	})

	It("creates HTTPS endpoints for InferencePool behind HTTPS gateway", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw-https", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{
						Name:     "https",
						Port:     443,
						Protocol: gwv1.HTTPSProtocolType,
						TLS:      &gwv1.ListenerTLSConfig{},
					},
				},
			},
		}
		fc := newFakeClient(scheme, gw)

		pool := &infv1.InferencePool{ObjectMeta: metav1.ObjectMeta{Name: "my-pool", Namespace: "ns"}}
		httpRouteList := &gwv1.HTTPRouteList{
			Items: []gwv1.HTTPRoute{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "route-pool-https", Namespace: "ns"},
					Spec: gwv1.HTTPRouteSpec{
						Hostnames: []gwv1.Hostname{"secure-inference.example.com"},
						CommonRouteSpec: gwv1.CommonRouteSpec{
							ParentRefs: []gwv1.ParentReference{{Name: "gw-https"}},
						},
						Rules: []gwv1.HTTPRouteRule{
							{
								BackendRefs: []gwv1.HTTPBackendRef{
									{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
										Kind:      ptr.To(gwv1.Kind(kindInferencePool)),
										Group:     ptr.To(gwv1.Group(infv1.GroupName)),
										Name:      "my-pool",
										Namespace: ptr.To(gwv1.Namespace("ns")),
									}}},
								},
							},
						},
					},
					Status: gwv1.HTTPRouteStatus{
						RouteStatus: gwv1.RouteStatus{
							Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw-https"}}),
						},
					},
				},
			},
		}

		eps := findEndpointsForPool(NewNetworkStatusUpdater(fc), ctx, pool, httpRouteList)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("secure-inference.example.com"))
		Expect(ep.Path).To(Equal("/"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Url).To(Equal("https://secure-inference.example.com/"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-pool-https"}))
	})
})

// ---------- discoverEndpointsFromInferencePools ----------

var _ = Describe("discoverEndpointsFromInferencePools", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns empty endpoints and the HTTPRouteList when no routes reference the pool", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 8080, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		httpRoute := &gwv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "route-svc", Namespace: "ns"},
			Spec: gwv1.HTTPRouteSpec{
				Hostnames: []gwv1.Hostname{"app.example.com"},
				CommonRouteSpec: gwv1.CommonRouteSpec{
					ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
				},
				Rules: []gwv1.HTTPRouteRule{
					{
						BackendRefs: []gwv1.HTTPBackendRef{
							{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "some-svc"}}},
						},
					},
				},
			},
			Status: gwv1.HTTPRouteStatus{
				RouteStatus: gwv1.RouteStatus{
					Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
				},
			},
		}
		fc := newFakeClient(scheme, gw, httpRoute)
		h := NewNetworkStatusUpdater(fc)

		pools := []infv1.InferencePool{
			{ObjectMeta: metav1.ObjectMeta{Name: "my-pool", Namespace: "ns"}},
		}
		endpoints, routeList, err := h.discoverEndpointsFromInferencePools(ctx, "ns", pools)
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(BeEmpty())
		Expect(routeList).NotTo(BeNil())
		Expect(routeList.Items).To(HaveLen(1))
	})

	It("discovers endpoints from InferencePools referenced by HTTPRoutes", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 9090, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		httpRoute := &gwv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "route-pool", Namespace: "ns"},
			Spec: gwv1.HTTPRouteSpec{
				Hostnames: []gwv1.Hostname{"inference.example.com"},
				CommonRouteSpec: gwv1.CommonRouteSpec{
					ParentRefs: []gwv1.ParentReference{{Name: "gw"}},
				},
				Rules: []gwv1.HTTPRouteRule{
					{
						Matches: []gwv1.HTTPRouteMatch{
							{Path: &gwv1.HTTPPathMatch{Value: ptr.To("/v1/predict")}},
						},
						BackendRefs: []gwv1.HTTPBackendRef{
							{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{
								Kind:      ptr.To(gwv1.Kind(kindInferencePool)),
								Group:     ptr.To(gwv1.Group(infv1.GroupName)),
								Name:      "my-pool",
								Namespace: ptr.To(gwv1.Namespace("ns")),
							}}},
						},
					},
				},
			},
			Status: gwv1.HTTPRouteStatus{
				RouteStatus: gwv1.RouteStatus{
					Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
				},
			},
		}
		fc := newFakeClient(scheme, gw, httpRoute)
		h := NewNetworkStatusUpdater(fc)

		pools := []infv1.InferencePool{
			{ObjectMeta: metav1.ObjectMeta{Name: "my-pool", Namespace: "ns"}},
		}
		endpoints, routeList, err := h.discoverEndpointsFromInferencePools(ctx, "ns", pools)
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints).To(HaveLen(1))
		ep := endpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.Host).To(Equal("inference.example.com"))
		Expect(ep.Path).To(Equal("/v1/predict"))
		Expect(ep.Port).To(Equal("9090"))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
		Expect(ep.Url).To(Equal("http://inference.example.com:9090/v1/predict"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "route-pool"}))
		Expect(routeList).NotTo(BeNil())
		Expect(routeList.Items).To(HaveLen(1))
	})
})

// ---------- parseUrlsToEndpoints ----------

var _ = Describe("parseUrlsToEndpoints", func() {
	updater := NewNetworkStatusUpdater(nil)

	It("parses an https URL with default port", func() {
		eps := updater.parseUrlsToEndpoints([]string{"https://inference.example.com/v1/models"}, nil)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(ep.Host).To(Equal("inference.example.com"))
		Expect(ep.Path).To(Equal("/v1/models"))
		Expect(ep.Port).To(Equal("443"))
		Expect(ep.Protocol).To(Equal("https"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Url).To(Equal("https://inference.example.com/v1/models"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{}))
	})

	It("parses an http URL with default port", func() {
		eps := updater.parseUrlsToEndpoints([]string{"http://my-service.ns.svc.cluster.local/predict"}, nil)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.Host).To(Equal("my-service.ns.svc.cluster.local"))
		Expect(ep.Path).To(Equal("/predict"))
		Expect(ep.Port).To(Equal("80"))
		Expect(ep.Protocol).To(Equal("http"))
		Expect(ep.TlsEnabled).To(HaveValue(BeFalse()))
	})

	It("parses a URL with an explicit port", func() {
		eps := updater.parseUrlsToEndpoints([]string{"https://host.example.com:8443/api"}, nil)
		Expect(eps).To(HaveLen(1))
		ep := eps[0]
		Expect(ep.Host).To(Equal("host.example.com"))
		Expect(ep.Port).To(Equal("8443"))
		Expect(ep.Path).To(Equal("/api"))
		Expect(ep.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(ep.Url).To(Equal("https://host.example.com:8443/api"))
	})

	It("defaults path to / when URL has no path", func() {
		eps := updater.parseUrlsToEndpoints([]string{"https://host.example.com"}, nil)
		Expect(eps).To(HaveLen(1))
		Expect(eps[0].Path).To(Equal("/"))
	})

	It("parses multiple URLs", func() {
		eps := updater.parseUrlsToEndpoints([]string{
			"https://a.example.com/foo",
			"http://b.example.com:9090/bar",
		}, nil)
		Expect(eps).To(HaveLen(2))
		Expect(eps[0].Host).To(Equal("a.example.com"))
		Expect(eps[0].TlsEnabled).To(HaveValue(BeTrue()))
		Expect(eps[1].Host).To(Equal("b.example.com"))
		Expect(eps[1].Port).To(Equal("9090"))
		Expect(eps[1].TlsEnabled).To(HaveValue(BeFalse()))
	})

	It("skips invalid URLs", func() {
		eps := updater.parseUrlsToEndpoints([]string{"://bad", "https://good.example.com/ok"}, nil)
		Expect(eps).To(HaveLen(1))
		Expect(eps[0].Host).To(Equal("good.example.com"))
	})

	It("skips URLs with no host", func() {
		eps := updater.parseUrlsToEndpoints([]string{"/just-a-path"}, nil)
		Expect(eps).To(BeEmpty())
	})

	It("returns empty for nil input", func() {
		eps := updater.parseUrlsToEndpoints(nil, nil)
		Expect(eps).To(BeEmpty())
	})

	It("returns empty for empty slice", func() {
		eps := updater.parseUrlsToEndpoints([]string{}, nil)
		Expect(eps).To(BeEmpty())
	})

	It("skips URLs that already exist in existing endpoints", func() {
		existing := []v1alpha1.ExternalEndpoint{
			{Url: "https://inference.example.com/v1/models"},
		}
		eps := updater.parseUrlsToEndpoints([]string{
			"https://inference.example.com/v1/models",
			"https://new.example.com/api",
		}, existing)
		Expect(eps).To(HaveLen(1))
		Expect(eps[0].Host).To(Equal("new.example.com"))
	})

	It("deduplicates URLs with trailing slash differences", func() {
		existing := []v1alpha1.ExternalEndpoint{
			{Url: "https://inference.example.com/v1/models/"},
		}
		eps := updater.parseUrlsToEndpoints([]string{
			"https://inference.example.com/v1/models",
		}, existing)
		Expect(eps).To(BeEmpty())
	})
})

// Suppress unused import warnings
var _ = fmt.Sprint
var _ = os.Getenv
var _ = ptr.To("")
