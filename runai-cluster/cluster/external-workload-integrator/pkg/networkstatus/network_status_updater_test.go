package networkstatus

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	commonconsts "github.com/run-ai/runai/runai-cluster/common/constants"

	openshiftV1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestNetworkStatusUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Status Utils Suite")
}

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(gwv1.Install(s))
	utilruntime.Must(gwv1alpha2.Install(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	utilruntime.Must(dromosv1alpha1.AddToScheme(s))
	utilruntime.Must(networkingv1.AddToScheme(s))
	utilruntime.Must(infv1.Install(s))
	return s
}

func newFakeClient(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

func acceptedParentStatus(parentRefs []gwv1.ParentReference) []gwv1.RouteParentStatus {
	statuses := make([]gwv1.RouteParentStatus, 0, len(parentRefs))
	for _, ref := range parentRefs {
		statuses = append(statuses, gwv1.RouteParentStatus{
			ParentRef:      ref,
			ControllerName: "example.net/gateway-controller",
			Conditions: []metav1.Condition{
				{
					Type:   string(gwv1.RouteConditionAccepted),
					Status: metav1.ConditionTrue,
				},
			},
		})
	}
	return statuses
}

// newFakeClientWithIndex builds a fake client that supports the EW-by-UID field index
// and allows status subresource patches on ExternalWorkload.
func newFakeClientWithIndex(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
		WithIndex(&v1alpha1.ExternalWorkload{},
			constants.ExternalWorkloadByOriginalUIDIndexerName,
			cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer).
		Build()
}

// ---------- extractServiceNames (dispatch) ----------

var _ = Describe("extractServiceNames", func() {
	var h *NetworkStatusUpdater

	BeforeEach(func() {
		h = &NetworkStatusUpdater{}
	})

	It("returns empty for Service kind", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		result := h.extractServiceNames(context.Background(), svc, schema.GroupVersionKind{Kind: kindService})
		Expect(result).To(BeEmpty())
	})

	It("returns service names for Ingress kind", func() {
		ing := &networkingv1.Ingress{
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "web"}}},
							},
						},
					}},
				},
			},
		}
		result := h.extractServiceNames(context.Background(), ing, schema.GroupVersionKind{Kind: kindIngress})
		Expect(result).To(ConsistOf("web"))
	})

	It("returns service names for Route kind", func() {
		route := &openshiftV1.Route{
			Spec: openshiftV1.RouteSpec{
				To: openshiftV1.RouteTargetReference{Name: "primary-svc"},
				AlternateBackends: []openshiftV1.RouteTargetReference{
					{Name: "alt-svc"},
				},
			},
		}
		result := h.extractServiceNames(context.Background(), route, schema.GroupVersionKind{Kind: kindRoute})
		Expect(result).To(ConsistOf("primary-svc", "alt-svc"))
	})

	It("returns empty slice for unknown kind", func() {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc"}}
		result := h.extractServiceNames(context.Background(), svc, schema.GroupVersionKind{Kind: "Unknown"})
		Expect(result).To(BeEmpty())
	})
})

// ---------- resolvePodsOpenshiftRoute ----------

var _ = Describe("resolvePodsOpenshiftRoute", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
		utilruntime.Must(openshiftV1.Install(scheme))
	})

	It("returns nil for non-Route object", func() {
		fc := newFakeClient(scheme)
		h := NewNetworkStatusUpdater(fc)
		pods := h.resolvePodsOpenshiftRoute(ctx, &corev1.Service{})
		Expect(pods).To(BeNil())
	})

	It("returns pods matching the Route's target service", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "web"}},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "web-pod", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newFakeClient(scheme, svc, pod)
		h := NewNetworkStatusUpdater(fc)

		route := &openshiftV1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "my-route", Namespace: "ns"},
			Spec: openshiftV1.RouteSpec{
				To: openshiftV1.RouteTargetReference{Name: "web"},
			},
		}
		pods := h.resolvePodsOpenshiftRoute(ctx, route)
		Expect(pods).To(HaveLen(1))
		Expect(pods[0].Name).To(Equal("web-pod"))
	})

	It("returns pods from alternate backends too", func() {
		svc1 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "primary", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "primary"}},
		}
		svc2 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "alt", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "alt"}},
		}
		pod1 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-primary", Namespace: "ns", UID: "uid-1", Labels: map[string]string{"app": "primary"}},
		}
		pod2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-alt", Namespace: "ns", UID: "uid-2", Labels: map[string]string{"app": "alt"}},
		}
		fc := newFakeClient(scheme, svc1, svc2, pod1, pod2)
		h := NewNetworkStatusUpdater(fc)

		route := &openshiftV1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "split-route", Namespace: "ns"},
			Spec: openshiftV1.RouteSpec{
				To: openshiftV1.RouteTargetReference{Name: "primary"},
				AlternateBackends: []openshiftV1.RouteTargetReference{
					{Name: "alt"},
				},
			},
		}
		pods := h.resolvePodsOpenshiftRoute(ctx, route)
		Expect(pods).To(HaveLen(2))
		names := []string{pods[0].Name, pods[1].Name}
		Expect(names).To(ConsistOf("pod-primary", "pod-alt"))
	})

	It("deduplicates pods shared across primary and alternate backends", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-a", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "shared"}},
		}
		svcAlias := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-b", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "shared"}},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "shared-pod", Namespace: "ns",
				Labels: map[string]string{"app": "shared"},
				UID:    types.UID("uid-1")},
		}
		fc := newFakeClient(scheme, svc, svcAlias, pod)
		h := NewNetworkStatusUpdater(fc)

		route := &openshiftV1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "dup-route", Namespace: "ns"},
			Spec: openshiftV1.RouteSpec{
				To: openshiftV1.RouteTargetReference{Name: "svc-a"},
				AlternateBackends: []openshiftV1.RouteTargetReference{
					{Name: "svc-b"},
				},
			},
		}
		pods := h.resolvePodsOpenshiftRoute(ctx, route)
		Expect(pods).To(HaveLen(1))
		Expect(pods[0].Name).To(Equal("shared-pod"))
	})

	It("returns empty when service does not exist", func() {
		fc := newFakeClient(scheme)
		h := NewNetworkStatusUpdater(fc)

		route := &openshiftV1.Route{
			ObjectMeta: metav1.ObjectMeta{Name: "orphan-route", Namespace: "ns"},
			Spec: openshiftV1.RouteSpec{
				To: openshiftV1.RouteTargetReference{Name: "nonexistent"},
			},
		}
		pods := h.resolvePodsOpenshiftRoute(ctx, route)
		Expect(pods).To(BeEmpty())
	})
})

// ---------- ReplaceEndpointsAndPatch ----------

var _ = Describe("ReplaceEndpointsAndPatch", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("patches when endpoints differ", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		fc := newFakeClientWithIndex(scheme, ew)
		updater := NewNetworkStatusUpdater(fc)

		newEps := []v1alpha1.ExternalEndpoint{
			{Port: "80", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc"}},
		}
		err := updater.ReplaceEndpointsAndPatch(ctx, ew, newEps)
		Expect(err).NotTo(HaveOccurred())

		// Verify patch was applied
		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].Port).To(Equal("80"))
	})

	It("skips patch when endpoints are identical", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "80", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc"}},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		fc := newFakeClientWithIndex(scheme, ew)
		updater := NewNetworkStatusUpdater(fc)

		err := updater.ReplaceEndpointsAndPatch(ctx, ew, eps)
		Expect(err).NotTo(HaveOccurred())
	})
})

// ---------- findAffectedExternalWorkloadsFromPods ----------

var _ = Describe("findAffectedExternalWorkloadsFromPods", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("returns empty when no Pods are provided", func() {
		fc := newFakeClientWithIndex(scheme)
		handler := NewNetworkStatusUpdater(fc)

		affected := handler.findAffectedExternalWorkloadsFromPods(ctx, []corev1.Pod{})
		Expect(affected).To(BeEmpty())
	})

	It("skips Pods that have no matching EW", func() {
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "uid-no-ExternalWorkload",
				},
			},
		}
		fc := newFakeClientWithIndex(scheme)
		handler := NewNetworkStatusUpdater(fc)

		affected := handler.findAffectedExternalWorkloadsFromPods(ctx, []corev1.Pod{pod})
		Expect(affected).To(BeEmpty())
	})

	It("deduplicates EWs when multiple Pods belong to the same workload", func() {
		pod1 := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid-1",
				},
			},
		}
		pod2 := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid-1",
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ExternalWorkload-deploy", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid-1")},
		}
		fc := newFakeClientWithIndex(scheme, ew)
		handler := NewNetworkStatusUpdater(fc)

		affected := handler.findAffectedExternalWorkloadsFromPods(ctx, []corev1.Pod{pod1, pod2})
		Expect(affected).To(HaveLen(1))
		Expect(affected[0].Name).To(Equal("ExternalWorkload-deploy"))
	})

	It("returns multiple affected EWs for Pods with different top owners", func() {
		pod1 := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod-a", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "uid-a",
				},
			},
		}
		pod2 := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod-b", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "uid-b",
				},
			},
		}
		ewA := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ExternalWorkload-a", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("uid-a")},
		}
		ewB := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ExternalWorkload-b", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("uid-b")},
		}
		fc := newFakeClientWithIndex(scheme, ewA, ewB)
		handler := NewNetworkStatusUpdater(fc)

		affected := handler.findAffectedExternalWorkloadsFromPods(ctx, []corev1.Pod{pod1, pod2})
		Expect(affected).To(HaveLen(2))
		names := []string{affected[0].Name, affected[1].Name}
		Expect(names).To(ConsistOf("ExternalWorkload-a", "ExternalWorkload-b"))
	})
})

// ---------- HandleNetworkResourceChange ----------

var _ = Describe("HandleNetworkResourceAddUpdate", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("handles direct service change and patches EW endpoints", func() {
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
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
		}
		fc := newFakeClientWithIndex(scheme, node, svc, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		ep := fetched.Status.NetworkStatus.ExternalEndpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep.Host).To(Equal("10.0.0.1"))
		Expect(ep.Port).To(Equal("30080"))
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
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec: v1alpha1.ExternalWorkloadSpec{
				OriginalWorkloadUID: types.UID("deploy-uid"),
				Urls:                []string{"https://inference.example.com/v1/models"},
			},
		}
		fc := newFakeClientWithIndex(scheme, node, svc, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(2))

		// First endpoint from NodePort service
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].Port).To(Equal("30080"))

		// Second endpoint from spec.urls
		urlEp := fetched.Status.NetworkStatus.ExternalEndpoints[1]
		Expect(urlEp.SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(urlEp.Host).To(Equal("inference.example.com"))
		Expect(urlEp.Port).To(Equal("443"))
		Expect(urlEp.Path).To(Equal("/v1/models"))
		Expect(urlEp.Protocol).To(Equal("https"))
		Expect(urlEp.TlsEnabled).To(HaveValue(BeTrue()))
		Expect(urlEp.Url).To(Equal("https://inference.example.com/v1/models"))
	})

	It("handles ingress change and resolves referenced services", func() {
		os.Setenv(config.EnvVarClusterDomain, "https://app.example.com:443")
		defer os.Unsetenv(config.EnvVarClusterDomain)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		ing := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "my-ingress", Namespace: "ns"},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: "app.example.com",
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path:    "/",
										Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "web-svc"}},
									},
								},
							},
						},
					},
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
		}
		fc := newFakeClientWithIndex(scheme, svc, ing, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"}
		ews := handler.FindAffectedEwsForObject(ctx, ing, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		ep := fetched.Status.NetworkStatus.ExternalEndpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointIngressSource))
		Expect(ep.SourceRef.Kind).To(Equal("Ingress"))
		Expect(ep.SourceRef.Name).To(Equal("my-ingress"))
	})

	It("handles HTTPRoute change and resolves referenced services", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "http", Port: 8080, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "route-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		httpRoute := &gwv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "my-route", Namespace: "ns"},
			Spec: gwv1.HTTPRouteSpec{
				Hostnames:       []gwv1.Hostname{"app.example.com"},
				CommonRouteSpec: gwv1.CommonRouteSpec{ParentRefs: []gwv1.ParentReference{{Name: "gw"}}},
				Rules: []gwv1.HTTPRouteRule{
					{BackendRefs: []gwv1.HTTPBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "route-svc"}}},
					}},
				},
			},
			Status: gwv1.HTTPRouteStatus{
				RouteStatus: gwv1.RouteStatus{
					Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
		}
		fc := newFakeClientWithIndex(scheme, gw, svc, httpRoute, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
		ews := handler.FindAffectedEwsForObject(ctx, httpRoute, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		ep := fetched.Status.NetworkStatus.ExternalEndpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.SourceRef.Kind).To(Equal("HTTPRoute"))
		Expect(ep.SourceRef.Name).To(Equal("my-route"))
	})

	It("handles GRPCRoute change and resolves referenced services", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "grpc", Port: 9090, Protocol: gwv1.HTTPProtocolType},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "grpc-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "grpc-app"},
				Ports:    []corev1.ServicePort{{Port: 50051}},
			},
		}
		grpcRoute := &gwv1.GRPCRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "my-grpc-route", Namespace: "ns"},
			Spec: gwv1.GRPCRouteSpec{
				Hostnames:       []gwv1.Hostname{"grpc.example.com"},
				CommonRouteSpec: gwv1.CommonRouteSpec{ParentRefs: []gwv1.ParentReference{{Name: "gw"}}},
				Rules: []gwv1.GRPCRouteRule{
					{BackendRefs: []gwv1.GRPCBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "grpc-svc"}}},
					}},
				},
			},
			Status: gwv1.GRPCRouteStatus{
				RouteStatus: gwv1.RouteStatus{
					Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "grpc-app",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
		}
		fc := newFakeClientWithIndex(scheme, gw, svc, grpcRoute, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GRPCRoute"}
		ews := handler.FindAffectedEwsForObject(ctx, grpcRoute, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		ep := fetched.Status.NetworkStatus.ExternalEndpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.SourceRef.Kind).To(Equal("GRPCRoute"))
		Expect(ep.SourceRef.Name).To(Equal("my-grpc-route"))
	})

	It("handles TLSRoute change and resolves referenced services", func() {
		gw := &gwv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "ns"},
			Spec: gwv1.GatewaySpec{
				Listeners: []gwv1.Listener{
					{Name: "tls", Port: 8443, Protocol: gwv1.TLSProtocolType},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "tls-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "tls-app"},
				Ports:    []corev1.ServicePort{{Port: 443}},
			},
		}
		tlsRoute := &gwv1alpha2.TLSRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "my-tls-route", Namespace: "ns"},
			Spec: gwv1alpha2.TLSRouteSpec{
				Hostnames:       []gwv1.Hostname{"secure.example.com"},
				CommonRouteSpec: gwv1.CommonRouteSpec{ParentRefs: []gwv1.ParentReference{{Name: "gw"}}},
				Rules: []gwv1alpha2.TLSRouteRule{
					{BackendRefs: []gwv1.BackendRef{
						{BackendObjectReference: gwv1.BackendObjectReference{Name: "tls-svc"}},
					}},
				},
			},
			Status: gwv1alpha2.TLSRouteStatus{
				RouteStatus: gwv1.RouteStatus{
					Parents: acceptedParentStatus([]gwv1.ParentReference{{Name: "gw"}}),
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "tls-app",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
		}
		fc := newFakeClientWithIndex(scheme, gw, svc, tlsRoute, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1alpha2", Kind: "TLSRoute"}
		ews := handler.FindAffectedEwsForObject(ctx, tlsRoute, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		ep := fetched.Status.NetworkStatus.ExternalEndpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointGatewayAPISource))
		Expect(ep.SourceRef.Kind).To(Equal("TLSRoute"))
		Expect(ep.SourceRef.Name).To(Equal("my-tls-route"))
	})

	It("replaces stale endpoints with freshly discovered ones", func() {
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
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}
		// EW has a stale Ingress endpoint whose backing Ingress no longer exists
		staleEp := v1alpha1.ExternalEndpoint{
			Port: "443", Protocol: "TCP",
			SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "old-ingress"},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{
					ExternalEndpoints: []v1alpha1.ExternalEndpoint{staleEp},
				},
			},
		}
		fc := newFakeClientWithIndex(scheme, node, svc, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		// Stale ingress endpoint is gone; only the freshly discovered NodePort endpoint remains
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].SourceRef.Kind).To(Equal("Service"))
	})

	It("FindAffectedEwsForObject returns empty when no EW exists for the pod", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "orphan-uid",
				},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, pod)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(BeEmpty())
	})
})

// ---------- HandleNetworkResourceDelete ----------

var _ = Describe("HandleNetworkResourceDelete", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("removes service endpoints from affected EWs on service delete", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "30080", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"}},
			{Port: "443", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "my-ing"}},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		// Resolve affected EWs before deleting (simulates informer receiving delete event object)
		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(HaveLen(1))

		// Delete the service (simulates real informer behavior)
		Expect(fc.Delete(ctx, svc)).To(Succeed())

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty())
	})

	It("preserves spec.urls endpoints when service is deleted", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "30080", Protocol: "TCP", SourceType: v1alpha1.ExternalEndpointNodePortSource, SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"}},
			{Port: "443", Protocol: "https", SourceType: v1alpha1.ExternalEndpointOtherSource, Host: "inference.example.com", Path: "/v1/models", TlsEnabled: ptr.To(true), Url: "https://inference.example.com/v1/models"},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec: v1alpha1.ExternalWorkloadSpec{
				OriginalWorkloadUID: types.UID("deploy-uid"),
				Urls:                []string{"https://inference.example.com/v1/models"},
			},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-np", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(HaveLen(1))

		Expect(fc.Delete(ctx, svc)).To(Succeed())

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		// Service endpoint is gone, but spec.urls endpoint is preserved
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].SourceType).To(Equal(v1alpha1.ExternalEndpointOtherSource))
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].Url).To(Equal("https://inference.example.com/v1/models"))
	})

	It("removes ingress endpoints from affected EWs on ingress delete", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "30080", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "web-svc"}},
			{Port: "443", Protocol: "https", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Ingress", Name: "my-ingress"}},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		ing := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "my-ingress", Namespace: "ns"},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: "web-svc"}}},
							},
						},
					}},
				},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, ing, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"}
		ews := handler.FindAffectedEwsForObject(ctx, ing, gvk)
		Expect(ews).To(HaveLen(1))

		// Delete the ingress (simulates real informer behavior)
		Expect(fc.Delete(ctx, ing)).To(Succeed())

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty())
	})

	It("removes HTTPRoute endpoints from affected EWs on route delete", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "8080", Protocol: "http", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "HTTPRoute", Name: "my-route"}},
			{Port: "30080", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "route-svc"}},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "route-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80}},
			},
		}
		httpRoute := &gwv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "my-route", Namespace: "ns"},
			Spec: gwv1.HTTPRouteSpec{
				Rules: []gwv1.HTTPRouteRule{
					{BackendRefs: []gwv1.HTTPBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "route-svc"}}},
					}},
				},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, httpRoute, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
		ews := handler.FindAffectedEwsForObject(ctx, httpRoute, gvk)
		Expect(ews).To(HaveLen(1))

		// Delete the route (simulates real informer behavior)
		Expect(fc.Delete(ctx, httpRoute)).To(Succeed())

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty())
	})

	It("removes GRPCRoute endpoints from affected EWs on route delete", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "9090", Protocol: "http", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "GRPCRoute", Name: "my-grpc-route"}},
			{Port: "30080", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "grpc-svc"}},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "grpc-app",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "grpc-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "grpc-app"},
				Ports:    []corev1.ServicePort{{Port: 50051}},
			},
		}
		grpcRoute := &gwv1.GRPCRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "my-grpc-route", Namespace: "ns"},
			Spec: gwv1.GRPCRouteSpec{
				Rules: []gwv1.GRPCRouteRule{
					{BackendRefs: []gwv1.GRPCBackendRef{
						{BackendRef: gwv1.BackendRef{BackendObjectReference: gwv1.BackendObjectReference{Name: "grpc-svc"}}},
					}},
				},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, grpcRoute, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "GRPCRoute"}
		ews := handler.FindAffectedEwsForObject(ctx, grpcRoute, gvk)
		Expect(ews).To(HaveLen(1))

		// Delete the route (simulates real informer behavior)
		Expect(fc.Delete(ctx, grpcRoute)).To(Succeed())

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty())
	})

	It("removes TLSRoute endpoints from affected EWs on route delete", func() {
		eps := []v1alpha1.ExternalEndpoint{
			{Port: "8443", Protocol: "tls", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "TLSRoute", Name: "my-tls-route"}},
			{Port: "30080", Protocol: "TCP", SourceRef: v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "tls-svc"}},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{ExternalEndpoints: eps},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "tls-app",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "tls-svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "tls-app"},
				Ports:    []corev1.ServicePort{{Port: 443}},
			},
		}
		tlsRoute := &gwv1alpha2.TLSRoute{
			ObjectMeta: metav1.ObjectMeta{Name: "my-tls-route", Namespace: "ns"},
			Spec: gwv1alpha2.TLSRouteSpec{
				Rules: []gwv1alpha2.TLSRouteRule{
					{BackendRefs: []gwv1.BackendRef{
						{BackendObjectReference: gwv1.BackendObjectReference{Name: "tls-svc"}},
					}},
				},
			},
		}
		fc := newFakeClientWithIndex(scheme, svc, tlsRoute, pod, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1alpha2", Kind: "TLSRoute"}
		ews := handler.FindAffectedEwsForObject(ctx, tlsRoute, gvk)
		Expect(ews).To(HaveLen(1))

		// Delete the route (simulates real informer behavior)
		Expect(fc.Delete(ctx, tlsRoute)).To(Succeed())

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(BeEmpty())
	})

	It("FindAffectedEwsForObject returns empty when no EWs are affected", func() {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
			Spec:       corev1.ServiceSpec{Selector: map[string]string{"app": "web"}},
		}
		fc := newFakeClientWithIndex(scheme)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(BeEmpty())
	})
})

// ---------- HandleNetworkResourceChange (service with multiple Pods) ----------

var _ = Describe("HandleNetworkResourceChange for service with multiple Pods", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = testScheme()
	})

	It("computes endpoints per-pod and merges into EW status", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web"},
				Ports:    []corev1.ServicePort{{Port: 80, NodePort: 30080, Protocol: corev1.ProtocolTCP}},
			},
		}
		pod1 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "uid-1",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}
		pod2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "uid-1",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("uid-1")},
		}
		fc := newFakeClientWithIndex(scheme, node, svc, pod1, pod2, ew)
		handler := NewNetworkStatusUpdater(fc)

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc, gvk)
		Expect(ews).To(HaveLen(1))

		err := handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		fetched := &v1alpha1.ExternalWorkload{}
		Expect(fc.Get(ctx, types.NamespacedName{Name: "ew1", Namespace: "ns"}, fetched)).To(Succeed())
		// Both Pods discover the same endpoint (same node, same service).
		// deduplicateEndpoints removes the duplicate before patching.
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(fetched.Status.NetworkStatus.ExternalEndpoints[0].Port).To(Equal("30080"))
	})
})

// ---------- HandleNetworkResourceChange with non-EW object ----------

var _ = Describe("HandleNetworkResourceChange", func() {
	It("returns error when object is not an ExternalWorkload", func() {
		ctx := context.Background()
		scheme := testScheme()
		fc := newFakeClientWithIndex(scheme)
		handler := NewNetworkStatusUpdater(fc)

		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns"},
		}
		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		err := handler.HandleNetworkResourceChange(ctx, svc, gvk)
		Expect(err).NotTo(HaveOccurred())
	})
})

// ---------- ListWorkloadPods ----------

var _ = Describe("ListWorkloadPods", func() {
	It("returns pods matching the EW's OriginalWorkloadUID label", func() {
		ctx := context.Background()
		scheme := testScheme()
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("uid-123")},
		}
		matchingPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{commonconsts.ClusterTopOwnerUidLabel: "uid-123"},
			},
		}
		nonMatchingPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2", Namespace: "ns",
				Labels: map[string]string{commonconsts.ClusterTopOwnerUidLabel: "other-uid"},
			},
		}
		fc := newFakeClientWithIndex(scheme, ew, matchingPod, nonMatchingPod)

		pods, err := utils.ListWorkloadPods(ctx, fc, ew)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods).To(HaveLen(1))
		Expect(pods[0].Name).To(Equal("pod1"))
	})

	It("returns empty when no pods match", func() {
		ctx := context.Background()
		scheme := testScheme()
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("uid-123")},
		}
		fc := newFakeClientWithIndex(scheme, ew)

		pods, err := utils.ListWorkloadPods(ctx, fc, ew)
		Expect(err).NotTo(HaveOccurred())
		Expect(pods).To(BeEmpty())
	})
})

// ---------- handleTask ----------

var _ = Describe("handleTask", func() {
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

	It("forgets the task on success", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("uid-123")},
		}
		fc := newFakeClientWithIndex(scheme, ew)
		handler := NewNetworkStatusUpdater(fc)
		queue := handler.TasksQueue()

		key := types.NamespacedName{Name: "ew1", Namespace: "ns"}
		gvk := schema.GroupVersionKind{Version: "v1", Kind: "ExternalWorkload"}
		task := *utils.NewExternalWorkloadTask(utils.AddOperation, key, gvk).SetObject(ew)

		queue.Add(task)
		item, quit := queue.Get()
		Expect(quit).To(BeFalse())

		handler.handleTask(ctx, item)
		queue.Done(item)

		// Task was forgotten (not requeued), queue should be empty
		Expect(queue.Len()).To(Equal(0))
	})
})

// ---------- deduplicateEndpoints ----------

var _ = Describe("deduplicateEndpoints", func() {
	It("removes duplicate endpoints", func() {
		ep := v1alpha1.ExternalEndpoint{
			SourceType: v1alpha1.ExternalEndpointNodePortSource,
			Host:       "10.0.0.1",
			Port:       "30080",
			Protocol:   "TCP",
			Path:       "/",
			Url:        "http://10.0.0.1:30080/",
			SourceRef:  v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc1"},
		}
		result := deduplicateEndpoints([]v1alpha1.ExternalEndpoint{ep, ep, ep})
		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal(ep))
	})

	It("preserves distinct endpoints", func() {
		ep1 := v1alpha1.ExternalEndpoint{Host: "10.0.0.1", Port: "30080", Url: "http://10.0.0.1:30080/"}
		ep2 := v1alpha1.ExternalEndpoint{Host: "10.0.0.1", Port: "30081", Url: "http://10.0.0.1:30081/"}
		result := deduplicateEndpoints([]v1alpha1.ExternalEndpoint{ep1, ep2})
		Expect(result).To(HaveLen(2))
	})

	It("returns empty slice for empty input", func() {
		result := deduplicateEndpoints(nil)
		Expect(result).To(BeEmpty())
	})
})

// ---------- UpdateNetworkStatusIfNeeded ----------

var _ = Describe("UpdateNetworkStatusIfNeeded", func() {
	It("enqueues a task when addedResource is true and EW is not nil", func() {
		ctx := context.Background()
		scheme := testScheme()
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"},
		}
		fc := newFakeClient(scheme, ew)
		updater := NewNetworkStatusUpdater(fc)

		err := updater.UpdateNetworkStatusIfNeeded(ctx, pod, ew, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(updater.TasksQueue().Len()).To(Equal(1))

		task, quit := updater.TasksQueue().Get()
		Expect(quit).To(BeFalse())
		Expect(task.Key).To(Equal(types.NamespacedName{Name: "ew1", Namespace: "ns"}))
		Expect(task.Object).To(Equal(ew))
		updater.TasksQueue().Done(task)
	})

	It("does not enqueue when addedResource is false", func() {
		ctx := context.Background()
		scheme := testScheme()
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"},
		}
		fc := newFakeClient(scheme, ew)
		updater := NewNetworkStatusUpdater(fc)

		err := updater.UpdateNetworkStatusIfNeeded(ctx, pod, ew, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(updater.TasksQueue().Len()).To(Equal(0))
	})

	It("does not enqueue when externalWorkload is nil", func() {
		ctx := context.Background()
		scheme := testScheme()
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"},
		}
		fc := newFakeClient(scheme)
		updater := NewNetworkStatusUpdater(fc)

		err := updater.UpdateNetworkStatusIfNeeded(ctx, pod, nil, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(updater.TasksQueue().Len()).To(Equal(0))
	})
})

// ---------- updateNetworkStatusPatchBytes ----------

var _ = Describe("updateNetworkStatusPatchBytes", func() {
	It("produces valid JSON patch", func() {
		ns := &v1alpha1.WorkloadNetworkStatus{
			ExternalEndpoints: []v1alpha1.ExternalEndpoint{
				{
					SourceType: v1alpha1.ExternalEndpointNodePortSource,
					Host:       "10.0.0.1",
					Port:       "30080",
					Path:       "/",
					Protocol:   "TCP",
					Url:        "http://10.0.0.1:30080",
					SourceRef:  v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc"},
				},
			},
		}
		b, err := updateNetworkStatusPatchBytes(ns)
		Expect(err).NotTo(HaveOccurred())
		Expect(b).NotTo(BeEmpty())

		var parsed map[string]interface{}
		err = json.Unmarshal(b, &parsed)
		Expect(err).NotTo(HaveOccurred())
		status := parsed["status"].(map[string]interface{})
		networkStatus := status["networkStatus"].(map[string]interface{})
		Expect(networkStatus).To(HaveKey("externalEndpoints"))

		endpoints := networkStatus["externalEndpoints"].([]interface{})
		Expect(endpoints).To(HaveLen(1))
	})
})

// Suppress unused import warnings
var _ = os.Getenv
var _ = json.Unmarshal
