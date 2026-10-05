/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	inferencev1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/internal/render"
)

const (
	testWeightKey     = "weight"
	profiledRouteName = "profiled-route"
	testAppLabel      = "app"
	profiledEppName   = "profiled-epp"
	testTypeKey       = "type"
	testNameKey       = "name"
	testQueueType     = "queue"
	runtimeProfile    = "llm-d-v0.10"
	balancedProfile   = "balanced"
	queueProfile      = "queue-focused"
)

func runtimeEpp() map[string]any {
	return map[string]any{
		"profile": runtimeProfile,
		"scheduling": map[string]any{
			"defaultProfile": balancedProfile, "activeProfile": balancedProfile,
			"profiles": []any{
				map[string]any{testNameKey: balancedProfile, "scorers": []any{
					map[string]any{testTypeKey: testQueueType, testWeightKey: int64(10)},
					map[string]any{testTypeKey: "kvCache", testWeightKey: int64(30)},
				}},
				map[string]any{testNameKey: queueProfile, "scorers": []any{
					map[string]any{testTypeKey: testQueueType, testWeightKey: int64(60)},
					map[string]any{testTypeKey: "kvCache", testWeightKey: int64(30)},
				}},
			},
		},
	}
}

var _ = Describe("Runtime scheduling profiles", func() {
	var ctx context.Context
	var candidate *unstructured.Unstructured
	var reconciler *InferenceEndpointReconciler
	var request reconcile.Request
	BeforeEach(func() {
		ctx = context.Background()
		ns := fmt.Sprintf("profiles-%d", time.Now().UnixNano())
		Expect(k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})).To(Succeed())
		base := &inferencev1alpha1.InferenceEndpoint{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns},
			Spec: inferencev1alpha1.InferenceEndpointSpec{
				Workload: inferencev1alpha1.WorkloadSpec{Selector: inferencev1alpha1.WorkloadSelector{MatchLabels: map[string]string{testAppLabel: "backend"}}, Port: 8000},
				Routing:  inferencev1alpha1.RoutingSpec{GatewayRef: inferencev1alpha1.GatewayReference{Name: gatewayName}, Hostnames: []string{"profiles.test"}},
			},
		}
		candidate = endpointWithEpp(base, runtimeEpp())
		reconciler = &InferenceEndpointReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		request = reconcile.Request{NamespacedName: client.ObjectKeyFromObject(candidate)}
	})

	It("admits and retains a catalog and its selected profile", func() {
		Expect(k8sClient.Create(ctx, candidate)).To(Succeed())
		got := &unstructured.Unstructured{}
		got.SetGroupVersionKind(candidate.GroupVersionKind())
		Expect(k8sClient.Get(ctx, request.NamespacedName, got)).To(Succeed())
		epp, _, err := unstructured.NestedMap(got.Object, "spec", "epp")
		Expect(err).NotTo(HaveOccurred())
		Expect(epp).To(Equal(runtimeEpp()))
	})

	It("switches the route without rolling EPP and waits for current Gateway acknowledgement", func() {
		Expect(k8sClient.Create(ctx, candidate)).To(Succeed())
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		key := client.ObjectKey{Namespace: candidate.GetNamespace(), Name: profiledEppName}
		beforeCM, beforeDep := &corev1.ConfigMap{}, &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, key, beforeCM)).To(Succeed())
		Expect(k8sClient.Get(ctx, key, beforeDep)).To(Succeed())
		acceptRoute(ctx, key.Namespace, profiledRouteName)
		markDeploymentAvailable(ctx, key.Namespace, key.Name)
		By("keeping the cached route stale while the API server receives the active switch")
		cachedRoute := &unstructured.Unstructured{}
		cachedRoute.SetGroupVersionKind(routeGVK)
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: key.Namespace, Name: profiledRouteName}, cachedRoute)).To(Succeed())
		reconciler.Client = &staleRouteClient{Client: k8sClient, route: cachedRoute}
		reconciler.APIReader = k8sClient
		Expect(k8sClient.Get(ctx, request.NamespacedName, candidate)).To(Succeed())
		Expect(unstructured.SetNestedField(candidate.Object, queueProfile, "spec", "epp", "scheduling", "activeProfile")).To(Succeed())
		Expect(k8sClient.Update(ctx, candidate)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		afterCM, afterDep := &corev1.ConfigMap{}, &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, key, afterCM)).To(Succeed())
		Expect(k8sClient.Get(ctx, key, afterDep)).To(Succeed())
		Expect(afterCM.Data).To(Equal(beforeCM.Data))
		Expect(afterDep.Spec).To(Equal(beforeDep.Spec))
		Expect(afterDep.Generation).To(Equal(beforeDep.Generation))
		Expect(afterDep.Spec.Template.Annotations[render.ConfigHashAnnotation]).NotTo(BeEmpty())
		route := &unstructured.Unstructured{}
		route.SetGroupVersionKind(routeGVK)
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: key.Namespace, Name: profiledRouteName}, route)).To(Succeed())
		rules, _, err := unstructured.NestedSlice(route.Object, "spec", "rules")
		Expect(err).NotTo(HaveOccurred())
		filters := rules[0].(map[string]any)["filters"].([]any)
		Expect(filters).To(Equal([]any{map[string]any{
			testTypeKey: "RequestHeaderModifier", "requestHeaderModifier": map[string]any{
				"set": []any{map[string]any{testNameKey: "EPP-Profile", "value": queueProfile}},
			},
		}}))
		endpoint := &inferencev1alpha1.InferenceEndpoint{}
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionFalse(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())
		acceptRoute(ctx, key.Namespace, profiledRouteName)
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionTrue(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())
	})

	It("rolls EPP for a catalog edit and removes the profile filter when returning to legacy", func() {
		Expect(k8sClient.Create(ctx, candidate)).To(Succeed())
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		key := client.ObjectKey{Namespace: candidate.GetNamespace(), Name: profiledEppName}
		before := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, key, before)).To(Succeed())
		acceptRoute(ctx, key.Namespace, profiledRouteName)
		markDeploymentAvailable(ctx, key.Namespace, key.Name)
		Expect(k8sClient.Get(ctx, request.NamespacedName, candidate)).To(Succeed())
		epp := runtimeEpp()
		profiles := epp["scheduling"].(map[string]any)["profiles"].([]any)
		profiles[0].(map[string]any)["scorers"].([]any)[0].(map[string]any)[testWeightKey] = int64(20)
		Expect(unstructured.SetNestedMap(candidate.Object, epp, "spec", "epp")).To(Succeed())
		Expect(k8sClient.Update(ctx, candidate)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		after := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, key, after)).To(Succeed())
		Expect(after.Generation).To(BeNumerically(">", before.Generation))
		Expect(after.Spec.Template.Annotations[render.ConfigHashAnnotation]).NotTo(Equal(before.Spec.Template.Annotations[render.ConfigHashAnnotation]))
		endpoint := &inferencev1alpha1.InferenceEndpoint{}
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.FindStatusCondition(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady).Reason).To(Equal(inferencev1alpha1.ReasonEppNotAvailable))

		By("returning to the legacy implementation and removing its obsolete route filter")
		Expect(k8sClient.Get(ctx, request.NamespacedName, candidate)).To(Succeed())
		Expect(unstructured.SetNestedMap(candidate.Object, map[string]any{"profile": profileName}, "spec", "epp")).To(Succeed())
		Expect(k8sClient.Update(ctx, candidate)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, key, after)).To(Succeed())
		Expect(after.Spec.Template.Spec.Containers[0].Image).To(Equal(inferencev1alpha1.EppImageGIEV15))
		route := &unstructured.Unstructured{}
		route.SetGroupVersionKind(routeGVK)
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: key.Namespace, Name: profiledRouteName}, route)).To(Succeed())
		rules, _, err := unstructured.NestedSlice(route.Object, "spec", "rules")
		Expect(err).NotTo(HaveOccurred())
		Expect(rules[0].(map[string]any)).NotTo(HaveKey("filters"))
	})

	It("rejects an invalid active-profile update without changing the stored endpoint", func() {
		Expect(k8sClient.Create(ctx, candidate)).To(Succeed())
		generation := candidate.GetGeneration()
		Expect(unstructured.SetNestedField(candidate.Object, "missing", "spec", "epp", "scheduling", "activeProfile")).To(Succeed())
		Expect(apierrors.IsInvalid(k8sClient.Update(ctx, candidate))).To(BeTrue())
		Expect(k8sClient.Get(ctx, request.NamespacedName, candidate)).To(Succeed())
		active, _, err := unstructured.NestedString(candidate.Object, "spec", "epp", "scheduling", "activeProfile")
		Expect(err).NotTo(HaveOccurred())
		Expect(active).To(Equal(balancedProfile))
		Expect(candidate.GetGeneration()).To(Equal(generation))
	})

	DescribeTable("rejects invalid catalogs at admission", func(change func(map[string]any)) {
		epp := runtimeEpp()
		change(epp)
		Expect(unstructured.SetNestedMap(candidate.Object, epp, "spec", "epp")).To(Succeed())
		Expect(apierrors.IsInvalid(k8sClient.Create(ctx, candidate))).To(BeTrue())
	},
		Entry("unknown active", func(e map[string]any) { e["scheduling"].(map[string]any)["activeProfile"] = "missing" }),
		Entry("missing default", func(e map[string]any) { delete(e["scheduling"].(map[string]any), "defaultProfile") }),
		Entry("missing active", func(e map[string]any) { delete(e["scheduling"].(map[string]any), "activeProfile") }),
		Entry("unknown default", func(e map[string]any) { e["scheduling"].(map[string]any)["defaultProfile"] = "missing" }),
		Entry("empty catalog", func(e map[string]any) { e["scheduling"].(map[string]any)["profiles"] = []any{} }),
		Entry("missing catalog", func(e map[string]any) { delete(e, "scheduling") }),
		Entry("legacy implementation", func(e map[string]any) { e["profile"] = profileName }),
		Entry("image-only implementation", func(e map[string]any) { delete(e, "profile"); e["image"] = customImage }),
		Entry("incompatible image", func(e map[string]any) { e["image"] = customImage }),
		Entry("ambiguous scorers", func(e map[string]any) {
			e["scheduling"].(map[string]any)["scorers"] = []any{map[string]any{testTypeKey: testQueueType, testWeightKey: int64(10)}}
		}),
		Entry("duplicate profile names", func(e map[string]any) {
			profiles := e["scheduling"].(map[string]any)["profiles"].([]any)
			profiles[1].(map[string]any)[testNameKey] = balancedProfile
		}),
		Entry("empty scorers", func(e map[string]any) {
			e["scheduling"].(map[string]any)["profiles"].([]any)[0].(map[string]any)["scorers"] = []any{}
		}),
		Entry("zero-only scorers", func(e map[string]any) {
			e["scheduling"].(map[string]any)["profiles"].([]any)[0].(map[string]any)["scorers"] = []any{map[string]any{testTypeKey: testQueueType, testWeightKey: int64(0)}}
		}),
		Entry("invalid profile name", func(e map[string]any) {
			e["scheduling"].(map[string]any)["profiles"].([]any)[0].(map[string]any)[testNameKey] = "invalid name"
		}),
	)
})

// staleRouteClient reproduces informer lag after SSA. Writes and other reads
// use the real envtest API server; only the route read remains at its old state.
type staleRouteClient struct {
	client.Client
	route *unstructured.Unstructured
}

func (c *staleRouteClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if route, ok := obj.(*unstructured.Unstructured); ok && route.GroupVersionKind() == routeGVK {
		route.Object = c.route.DeepCopy().Object
		return nil
	}
	return c.Client.Get(ctx, key, obj, opts...)
}
