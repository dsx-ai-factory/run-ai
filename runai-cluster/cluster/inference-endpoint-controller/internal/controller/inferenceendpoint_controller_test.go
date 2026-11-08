/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	inferencev1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/internal/render"
)

// Names of the endpoint under test and of the children it produces.
const (
	endpointName = "llama3"
	poolName     = endpointName + "-pool"
	eppName      = endpointName + "-epp"
	routeName    = endpointName + "-route"
	gatewayName  = "inference-agw"
	// testAppLabel is the selector label of the workload pods.
	testAppLabel = "app"
	testNameKey  = "name"
	testTypeKey  = "type"
	// metricsNamespace stands in for the --metrics-scrape-namespace flag.
	metricsNamespace = "runai"
)

// child pairs a typed (or unstructured) object with the name it should have.
type child struct {
	obj  client.Object
	name string
}

func expectedChildren() []child {
	pool := &unstructured.Unstructured{}
	pool.SetGroupVersionKind(poolGVK)
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(routeGVK)
	return []child{
		{pool, poolName},
		{&corev1.ServiceAccount{}, eppName},
		{&corev1.ConfigMap{}, eppName},
		{&rbacv1.Role{}, eppName + "-non-sa"},
		{&rbacv1.Role{}, eppName + "-sa"},
		{&rbacv1.RoleBinding{}, eppName + "-non-sa"},
		{&rbacv1.RoleBinding{}, eppName + "-sa"},
		{&corev1.Service{}, eppName},
		{&appsv1.Deployment{}, eppName},
		{&networkingv1.NetworkPolicy{}, eppName},
		{route, routeName},
	}
}

var _ = Describe("InferenceEndpoint Controller", func() {
	var (
		ctx        context.Context
		ns         string
		endpoint   *inferencev1alpha1.InferenceEndpoint
		reconciler *InferenceEndpointReconciler
		request    reconcile.Request
	)

	// envtest has no garbage collector, so every test gets its own namespace and
	// nothing is cleaned up between them.
	BeforeEach(func() {
		ctx = context.Background()
		ns = fmt.Sprintf("test-%d", time.Now().UnixNano())
		Expect(k8sClient.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})).To(Succeed())

		endpoint = &inferencev1alpha1.InferenceEndpoint{
			ObjectMeta: metav1.ObjectMeta{Name: endpointName, Namespace: ns},
			Spec: inferencev1alpha1.InferenceEndpointSpec{
				Workload: inferencev1alpha1.WorkloadSpec{
					Selector: inferencev1alpha1.WorkloadSelector{MatchLabels: map[string]string{testAppLabel: "vllm-llama3"}},
					Port:     8000,
				},
				Routing: inferencev1alpha1.RoutingSpec{
					GatewayRef: inferencev1alpha1.GatewayReference{Name: gatewayName},
					Hostnames:  []string{"llama3.test"},
				},
				Epp: inferencev1alpha1.EppSpec{Profile: inferencev1alpha1.EppProfileGIEV15},
			},
		}
		Expect(k8sClient.Create(ctx, endpoint)).To(Succeed())

		reconciler = &InferenceEndpointReconciler{Client: k8sClient, Scheme: k8sClient.Scheme(), MetricsNamespace: metricsNamespace}
		request = reconcile.Request{NamespacedName: types.NamespacedName{Name: endpointName, Namespace: ns}}
	})

	It("creates the eleven child objects, owned and labelled by the endpoint", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		for _, c := range expectedChildren() {
			key := types.NamespacedName{Name: c.name, Namespace: ns}
			Expect(k8sClient.Get(ctx, key, c.obj)).To(Succeed(), "%T %s should exist", c.obj, c.name)
			Expect(metav1.IsControlledBy(c.obj, endpoint)).To(BeTrue(), "%T %s should be controlled by the endpoint", c.obj, c.name)
			Expect(c.obj.GetLabels()).To(HaveKeyWithValue(render.EndpointLabel, endpointName))
		}
	})

	It("restricts EPP ingress to the gateway and metrics namespaces", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		policy := &networkingv1.NetworkPolicy{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: eppName, Namespace: ns}, policy)).To(Succeed())
		Expect(metav1.IsControlledBy(policy, endpoint)).To(BeTrue())
		dep := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: eppName, Namespace: ns}, dep)).To(Succeed())
		Expect(policy.Spec.PodSelector.MatchLabels).To(Equal(dep.Spec.Template.Labels))
		Expect(policy.Spec.PolicyTypes).To(Equal([]networkingv1.PolicyType{networkingv1.PolicyTypeIngress}))
		Expect(policy.Spec.Ingress).To(HaveLen(2))
		Expect(policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels).To(HaveKeyWithValue(corev1.LabelMetadataName, ns), "gateway namespace defaults to the endpoint namespace")
		Expect(policy.Spec.Ingress[0].Ports).To(HaveLen(2))
		Expect(policy.Spec.Ingress[1].From[0].NamespaceSelector.MatchLabels).To(HaveKeyWithValue(corev1.LabelMetadataName, metricsNamespace))
		Expect(policy.Spec.Ingress[1].Ports[0].Port.IntValue()).To(Equal(9090))
	})

	It("points the route at the pool and the pool at the EPP", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		pool := &unstructured.Unstructured{}
		pool.SetGroupVersionKind(poolGVK)
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: poolName, Namespace: ns}, pool)).To(Succeed())
		selector, _, _ := unstructured.NestedStringMap(pool.Object, "spec", "selector", "matchLabels")
		Expect(selector).To(Equal(map[string]string{testAppLabel: "vllm-llama3"}))
		eppRef, _, _ := unstructured.NestedString(pool.Object, "spec", "endpointPickerRef", testNameKey)
		Expect(eppRef).To(Equal(eppName))

		route := &unstructured.Unstructured{}
		route.SetGroupVersionKind(routeGVK)
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: routeName, Namespace: ns}, route)).To(Succeed())
		hostnames, _, _ := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
		Expect(hostnames).To(Equal([]string{"llama3.test"}))
		parents, _, _ := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
		Expect(parents).To(HaveLen(1))
		Expect(parents[0].(map[string]any)[testNameKey]).To(Equal(gatewayName))
		Expect(parents[0].(map[string]any)["namespace"]).To(Equal(ns), "gateway namespace defaults to the endpoint namespace")
	})

	It("routes the path prefix, /v1 by default, to the pool and follows a change", func() {
		routePrefix := func() string {
			_, err := reconciler.Reconcile(ctx, request)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			route := &unstructured.Unstructured{}
			route.SetGroupVersionKind(routeGVK)
			ExpectWithOffset(1, k8sClient.Get(ctx, types.NamespacedName{Name: routeName, Namespace: ns}, route)).To(Succeed())
			rules, _, err := unstructured.NestedSlice(route.Object, "spec", "rules")
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(1, rules).To(HaveLen(1))
			rule := rules[0].(map[string]any)
			ExpectWithOffset(1, rule["backendRefs"].([]any)[0]).To(HaveKeyWithValue("kind", "InferencePool"))
			path := rule["matches"].([]any)[0].(map[string]any)["path"].(map[string]any)
			ExpectWithOffset(1, path).To(HaveKeyWithValue("type", "PathPrefix"))
			return path["value"].(string)
		}
		Expect(routePrefix()).To(Equal("/v1"))

		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		endpoint.Spec.Routing.PathPrefix = "/openai/v1"
		Expect(k8sClient.Update(ctx, endpoint)).To(Succeed())
		Expect(routePrefix()).To(Equal("/openai/v1"))
	})

	It("stays not Ready until the Gateway accepts the route and the EPP is available", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		By("reporting RouteNotAccepted while the route has no parent status")
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		ready := meta.FindStatusCondition(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)
		Expect(ready).NotTo(BeNil())
		Expect(ready.Status).To(Equal(metav1.ConditionFalse))
		Expect(ready.Reason).To(Equal(inferencev1alpha1.ReasonRouteNotAccepted))
		Expect(endpoint.Status.ObservedGeneration).To(Equal(endpoint.Generation))

		By("simulating the Gateway accepting the route")
		acceptRoute(ctx, ns, routeName)
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		By("reporting EppNotAvailable while the Deployment has no available replica")
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		ready = meta.FindStatusCondition(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)
		Expect(ready.Status).To(Equal(metav1.ConditionFalse))
		Expect(ready.Reason).To(Equal(inferencev1alpha1.ReasonEppNotAvailable))
		routeAccepted := meta.FindStatusCondition(endpoint.Status.Conditions, inferencev1alpha1.ConditionRouteAccepted)
		Expect(routeAccepted).NotTo(BeNil())
		Expect(routeAccepted.Status).To(Equal(metav1.ConditionTrue))

		By("simulating the EPP Deployment becoming available")
		markDeploymentAvailable(ctx, ns, eppName)
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		ready = meta.FindStatusCondition(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)
		Expect(ready.Status).To(Equal(metav1.ConditionTrue))
		Expect(ready.Reason).To(Equal(inferencev1alpha1.ReasonRouteAccepted))
	})

	It("reports Ready=False ApplyFailed when the children cannot be applied", func() {
		// An empty scheme cannot resolve the owner kind, so setting the owner
		// reference on the first child fails.
		broken := &InferenceEndpointReconciler{Client: k8sClient, Scheme: runtime.NewScheme(), MetricsNamespace: metricsNamespace}
		_, err := broken.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		ready := meta.FindStatusCondition(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)
		Expect(ready).NotTo(BeNil())
		Expect(ready.Status).To(Equal(metav1.ConditionFalse))
		Expect(ready.Reason).To(Equal(inferencev1alpha1.ReasonApplyFailed))
		Expect(endpoint.Status.ObservedGeneration).To(Equal(endpoint.Generation))
	})

	It("recreates a child that was deleted", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		dep := &appsv1.Deployment{}
		key := types.NamespacedName{Name: eppName, Namespace: ns}
		Expect(k8sClient.Get(ctx, key, dep)).To(Succeed())
		Expect(k8sClient.Delete(ctx, dep)).To(Succeed())
		Eventually(func() bool {
			return errors.IsNotFound(k8sClient.Get(ctx, key, &appsv1.Deployment{}))
		}, 10*time.Second, 100*time.Millisecond).Should(BeTrue())

		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, key, &appsv1.Deployment{})).To(Succeed())
	})

	It("reverts a drifted child field on the next reconcile", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		dep := &appsv1.Deployment{}
		key := types.NamespacedName{Name: eppName, Namespace: ns}
		Expect(k8sClient.Get(ctx, key, dep)).To(Succeed())
		dep.Spec.Template.Spec.Containers[0].Image = "someone-else/epp:hacked"
		Expect(k8sClient.Update(ctx, dep)).To(Succeed())

		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, key, dep)).To(Succeed())
		Expect(dep.Spec.Template.Spec.Containers[0].Image).To(Equal(inferencev1alpha1.EppImageGIEV15))
	})

	It("writes scorer weights into the EPP ConfigMap and rolls the Deployment", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		dep := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: eppName, Namespace: ns}, dep)).To(Succeed())
		hashBefore := dep.Spec.Template.Annotations[render.ConfigHashAnnotation]
		Expect(hashBefore).NotTo(BeEmpty())

		By("changing the weights on the endpoint")
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		endpoint.Spec.Epp.Scheduling = &inferencev1alpha1.SchedulingSpec{Scorers: []inferencev1alpha1.ScorerSpec{
			{Type: inferencev1alpha1.ScorerPrefixCache, Weight: 100},
		}}
		Expect(k8sClient.Update(ctx, endpoint)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		cm := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: eppName, Namespace: ns}, cm)).To(Succeed())
		Expect(cm.Data["default-plugins.yaml"]).To(MatchRegexp(`pluginRef: "?prefix-cache-scorer"?\n\s+weight: 100\n`))
		Expect(cm.Data["default-plugins.yaml"]).NotTo(ContainSubstring("queue-scorer"))

		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: eppName, Namespace: ns}, dep)).To(Succeed())
		Expect(dep.Spec.Template.Annotations[render.ConfigHashAnnotation]).NotTo(Equal(hashBefore), "pod template must change so the EPP restarts")
	})

	It("waits for the updated EPP Deployment before reporting new scorer settings Ready", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		acceptRoute(ctx, ns, routeName)
		markDeploymentAvailable(ctx, ns, eppName)
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionTrue(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())

		By("changing scorers while the previous EPP replica is still available")
		endpoint.Spec.Epp.Scheduling = &inferencev1alpha1.SchedulingSpec{Scorers: []inferencev1alpha1.ScorerSpec{
			{Type: inferencev1alpha1.ScorerPrefixCache, Weight: 100},
		}}
		Expect(k8sClient.Update(ctx, endpoint)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionFalse(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())

		By("observing the new generation before its replica is updated")
		dep := &appsv1.Deployment{}
		key := types.NamespacedName{Name: eppName, Namespace: ns}
		Expect(k8sClient.Get(ctx, key, dep)).To(Succeed())
		dep.Status.ObservedGeneration = dep.Generation
		dep.Status.UpdatedReplicas = 0
		Expect(k8sClient.Status().Update(ctx, dep)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionFalse(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())

		By("waiting until the old replica is gone")
		Expect(k8sClient.Get(ctx, key, dep)).To(Succeed())
		dep.Status.UpdatedReplicas = 1
		dep.Status.Replicas = 2
		Expect(k8sClient.Status().Update(ctx, dep)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionFalse(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())

		By("finishing the EPP rollout")
		markDeploymentAvailable(ctx, ns, eppName)
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		Expect(meta.IsStatusConditionTrue(endpoint.Status.Conditions, inferencev1alpha1.ConditionReady)).To(BeTrue())
		Expect(endpoint.Status.ObservedGeneration).To(Equal(endpoint.Generation))
	})

	It("does nothing when the endpoint no longer exists", func() {
		_, err := reconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: "never-created", Namespace: ns},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})

// acceptRoute writes the parent status a Gateway controller would write:
// Accepted=True and ResolvedRefs=True for our parentRef.
func acceptRoute(ctx context.Context, ns, name string) {
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(routeGVK)
	ExpectWithOffset(1, k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, route)).To(Succeed())
	now := metav1.Now().UTC().Format(time.RFC3339)
	condition := func(t string) map[string]any {
		return map[string]any{
			testTypeKey: t, "status": "True", "reason": t, "message": "test",
			"lastTransitionTime": now, "observedGeneration": route.GetGeneration(),
		}
	}
	parents := []any{map[string]any{
		"controllerName": "agentgateway.dev/agentgateway",
		"parentRef":      map[string]any{testNameKey: gatewayName, "namespace": ns},
		"conditions":     []any{condition("Accepted"), condition("ResolvedRefs")},
	}}
	ExpectWithOffset(1, unstructured.SetNestedSlice(route.Object, parents, "status", "parents")).To(Succeed())
	ExpectWithOffset(1, k8sClient.Status().Update(ctx, route)).To(Succeed())
}

// markDeploymentAvailable writes the status the deployment controller would
// write once the EPP pod passes its readiness probe.
func markDeploymentAvailable(ctx context.Context, ns, name string) {
	dep := &appsv1.Deployment{}
	ExpectWithOffset(1, k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, dep)).To(Succeed())
	dep.Status.Replicas = 1
	dep.Status.UpdatedReplicas = 1
	dep.Status.ObservedGeneration = dep.Generation
	dep.Status.ReadyReplicas = 1
	dep.Status.AvailableReplicas = 1
	ExpectWithOffset(1, k8sClient.Status().Update(ctx, dep)).To(Succeed())
}
