/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"k8s.io/utils/ptr"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/networkstatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func networkTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(gwv1.Install(s))
	utilruntime.Must(gwv1alpha2.Install(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	utilruntime.Must(networkingv1.AddToScheme(s))
	return s
}

func newNetworkFakeClientWithStatus(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
		Build()
}

func newNetworkFakeClientWithIndex(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		WithStatusSubresource(&v1alpha1.ExternalWorkload{}).
		WithIndex(&v1alpha1.ExternalWorkload{},
			constants.ExternalWorkloadByOriginalUIDIndexerName,
			cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer).
		Build()
}

// ---------- updateNetworkStatusIfNeeded ----------

var _ = Describe("updateNetworkStatusIfNeeded", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = networkTestScheme()
		os.Setenv(config.EnvVarClusterDomain, "https://app.example.com:443")
	})

	AfterEach(func() {
		os.Unsetenv(config.EnvVarClusterDomain)
	})

	It("returns nil when addedResource is false", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"},
		}
		fc := newNetworkFakeClientWithStatus(scheme, ew)
		h := networkstatus.NewNetworkStatusUpdater(fc)

		err := h.UpdateNetworkStatusIfNeeded(ctx, pod, ew, false)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil when externalWorkload is nil", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"},
		}
		fc := newNetworkFakeClientWithStatus(scheme)
		h := networkstatus.NewNetworkStatusUpdater(fc)

		err := h.UpdateNetworkStatusIfNeeded(ctx, pod, nil, true)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil when both addedResource is false and externalWorkload is nil", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"},
		}
		fc := newNetworkFakeClientWithStatus(scheme)
		h := networkstatus.NewNetworkStatusUpdater(fc)

		err := h.UpdateNetworkStatusIfNeeded(ctx, pod, nil, false)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil when object is not a Pod", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "ns"},
		}
		fc := newNetworkFakeClientWithStatus(scheme, ew)
		h := networkstatus.NewNetworkStatusUpdater(fc)

		err := h.UpdateNetworkStatusIfNeeded(ctx, svc, ew, true)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil when no services match the pod", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
		}
		fc := newNetworkFakeClientWithStatus(scheme, ew)
		h := networkstatus.NewNetworkStatusUpdater(fc)

		err := h.UpdateNetworkStatusIfNeeded(ctx, pod, ew, true)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns nil when discovered endpoints already exist in status (no-op)", func() {
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
		existingEndpoint := v1alpha1.ExternalEndpoint{
			SourceType: v1alpha1.ExternalEndpointNodePortSource,
			Host:       "10.0.0.1",
			Path:       "/",
			Port:       "30080",
			Protocol:   "TCP",
			TlsEnabled: ptr.To(false),
			Url:        "http://10.0.0.1:30080/",
			SourceRef:  v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"},
		}
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew1", Namespace: "ns"},
			Status: v1alpha1.ExternalWorkloadStatus{
				NetworkStatus: v1alpha1.WorkloadNetworkStatus{
					ExternalEndpoints: []v1alpha1.ExternalEndpoint{existingEndpoint},
				},
			},
		}
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns", Labels: map[string]string{"app": "web"}},
			Spec:       corev1.PodSpec{NodeName: "node1"},
		}
		fc := newNetworkFakeClientWithStatus(scheme, ew, node, svc)
		h := networkstatus.NewNetworkStatusUpdater(fc)

		err := h.UpdateNetworkStatusIfNeeded(ctx, pod, ew, true)
		Expect(err).NotTo(HaveOccurred())
		// Verify status was not changed (endpoints remain the same)
		Expect(ew.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		Expect(ew.Status.NetworkStatus.ExternalEndpoints[0]).To(Equal(existingEndpoint))
	})

	It("enqueues task and processes to patch externalWorkload status with discovered endpoints", func() {
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
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
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
		fc := newNetworkFakeClientWithStatus(scheme, ew, node, svc, pod)
		handler := networkstatus.NewNetworkStatusUpdater(fc)

		err := handler.UpdateNetworkStatusIfNeeded(ctx, pod, ew, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(handler.TasksQueue().Len()).To(Equal(1))

		// Dequeue and process the task
		task, quit := handler.TasksQueue().Get()
		Expect(quit).To(BeFalse())
		err = handler.HandleNetworkResourceChange(ctx, task.Object, task.GVK)
		Expect(err).NotTo(HaveOccurred())
		handler.TasksQueue().Done(task)

		// Re-read the ExternalWorkload from the fake client to verify the status was patched
		freshEw := &v1alpha1.ExternalWorkload{}
		err = fc.Get(ctx, client.ObjectKeyFromObject(ew), freshEw)
		Expect(err).NotTo(HaveOccurred())
		Expect(freshEw.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))
		ep := freshEw.Status.NetworkStatus.ExternalEndpoints[0]
		Expect(ep.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep.Host).To(Equal("10.0.0.1"))
		Expect(ep.Port).To(Equal("30080"))
		Expect(ep.Protocol).To(Equal("TCP"))
		Expect(ep.Url).To(Equal("10.0.0.1:30080"))
		Expect(ep.SourceRef).To(Equal(v1alpha1.ExternalEndpointSourceRef{Kind: "Service", Name: "svc-np"}))
	})

	It("adds second service endpoints to existing ones when new service is created for deployment pods", func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
				},
			},
		}
		svc1 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-first", Namespace: "ns"},
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
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadUID: types.UID("deploy-uid")},
		}
		pod1 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod1", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}
		pod2 := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pod2", Namespace: "ns",
				Labels: map[string]string{
					"app":                          "web2",
					"run.ai/cluster-top-owner-uid": "deploy-uid",
				},
			},
			Spec: corev1.PodSpec{NodeName: "node1"},
		}

		// Start with both pods and only svc-first
		fc := newNetworkFakeClientWithIndex(scheme, ew, node, svc1, pod1, pod2)
		handler := networkstatus.NewNetworkStatusUpdater(fc)

		// Pod1 is added — enqueue and process
		err := handler.UpdateNetworkStatusIfNeeded(ctx, pod1, ew, true)
		Expect(err).NotTo(HaveOccurred())
		task, quit := handler.TasksQueue().Get()
		Expect(quit).To(BeFalse())
		err = handler.HandleNetworkResourceChange(ctx, task.Object, task.GVK)
		Expect(err).NotTo(HaveOccurred())
		handler.TasksQueue().Done(task)

		freshEw := &v1alpha1.ExternalWorkload{}
		err = fc.Get(ctx, client.ObjectKeyFromObject(ew), freshEw)
		Expect(err).NotTo(HaveOccurred())
		Expect(freshEw.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))

		// Pod2 is added — enqueue and process
		err = handler.UpdateNetworkStatusIfNeeded(ctx, pod2, freshEw, true)
		Expect(err).NotTo(HaveOccurred())
		task, quit = handler.TasksQueue().Get()
		Expect(quit).To(BeFalse())
		err = handler.HandleNetworkResourceChange(ctx, task.Object, task.GVK)
		Expect(err).NotTo(HaveOccurred())
		handler.TasksQueue().Done(task)

		freshEw = &v1alpha1.ExternalWorkload{}
		err = fc.Get(ctx, client.ObjectKeyFromObject(ew), freshEw)
		Expect(err).NotTo(HaveOccurred())
		Expect(freshEw.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(1))

		// Now a second service is created targeting the same pods
		svc2 := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc-second", Namespace: "ns"},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeNodePort,
				Selector: map[string]string{"app": "web2"},
				Ports: []corev1.ServicePort{
					{Port: 8080, NodePort: 30088, Protocol: corev1.ProtocolTCP, AppProtocol: ptr.To("http")},
				},
			},
		}
		err = fc.Create(ctx, svc2)
		Expect(err).NotTo(HaveOccurred())

		// Service add triggers FindAffectedEwsForObject + HandleNetworkResourceChange
		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Service"}
		ews := handler.FindAffectedEwsForObject(ctx, svc2, gvk)
		Expect(ews).To(HaveLen(1))
		err = handler.HandleNetworkResourceChange(ctx, ews[0], gvk)
		Expect(err).NotTo(HaveOccurred())

		// Re-read to verify both endpoints are present (old + new)
		finalEw := &v1alpha1.ExternalWorkload{}
		err = fc.Get(ctx, client.ObjectKeyFromObject(ew), finalEw)
		Expect(err).NotTo(HaveOccurred())
		Expect(finalEw.Status.NetworkStatus.ExternalEndpoints).To(HaveLen(2))

		// Build a map by source service name for easy assertion
		byService := map[string]v1alpha1.ExternalEndpoint{}
		for _, ep := range finalEw.Status.NetworkStatus.ExternalEndpoints {
			byService[ep.SourceRef.Name] = ep
		}

		ep1 := byService["svc-first"]
		Expect(ep1.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep1.Host).To(Equal("10.0.0.1"))
		Expect(ep1.Port).To(Equal("30080"))
		Expect(ep1.Protocol).To(Equal("TCP"))
		Expect(ep1.Url).To(Equal("10.0.0.1:30080"))

		ep2 := byService["svc-second"]
		Expect(ep2.SourceType).To(Equal(v1alpha1.ExternalEndpointNodePortSource))
		Expect(ep2.Host).To(Equal("10.0.0.1"))
		Expect(ep2.Port).To(Equal("30088"))
		Expect(ep2.Protocol).To(Equal("TCP"))
		Expect(ep2.Url).To(Equal("http://10.0.0.1:30088/"))
	})
})
