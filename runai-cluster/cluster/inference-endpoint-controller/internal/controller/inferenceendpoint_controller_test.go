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
	profileName  = string(inferencev1alpha1.EppProfileGIEV15)
	profileField = "profile"
	imageField   = "image"
	customImage  = "example.com/custom-epp:v1"
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
				Epp: inferencev1alpha1.EppSpec{Image: "registry.k8s.io/gateway-api-inference-extension/epp:v1.5.0"},
			},
		}
		Expect(k8sClient.Create(ctx, endpoint)).To(Succeed())

		reconciler = &InferenceEndpointReconciler{Client: k8sClient, Scheme: k8sClient.Scheme()}
		request = reconcile.Request{NamespacedName: types.NamespacedName{Name: endpointName, Namespace: ns}}
	})

	It("creates the ten child objects, owned and labelled by the endpoint", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())

		for _, c := range expectedChildren() {
			key := types.NamespacedName{Name: c.name, Namespace: ns}
			Expect(k8sClient.Get(ctx, key, c.obj)).To(Succeed(), "%T %s should exist", c.obj, c.name)
			Expect(metav1.IsControlledBy(c.obj, endpoint)).To(BeTrue(), "%T %s should be controlled by the endpoint", c.obj, c.name)
			Expect(c.obj.GetLabels()).To(HaveKeyWithValue(render.EndpointLabel, endpointName))
		}
	})

	It("creates the existing EPP stack from a profile without an image", func() {
		candidate := endpointWithEpp(endpoint, map[string]any{profileField: profileName})
		Expect(k8sClient.Create(ctx, candidate)).To(Succeed())

		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(candidate)})
		Expect(err).NotTo(HaveOccurred())
		dep := &appsv1.Deployment{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: profiledEppName, Namespace: ns}, dep)).To(Succeed())
		Expect(dep.Spec.Template.Spec.Containers[0].Image).To(Equal(inferencev1alpha1.EppImageGIEV15))
		cm := &corev1.ConfigMap{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: profiledEppName, Namespace: ns}, cm)).To(Succeed())
		Expect(cm.Data["default-plugins.yaml"]).To(ContainSubstring("kind: EndpointPickerConfig"))
	})

	DescribeTable("resolves the EPP profile to its pinned image",
		func(epp inferencev1alpha1.EppSpec, image string, wantErr bool) {
			endpoint.Spec.Epp = epp
			in, err := inputFor(endpoint)
			if wantErr {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(in.EppImage).To(Equal(image))
		},
		Entry("profile alone", inferencev1alpha1.EppSpec{Profile: inferencev1alpha1.EppProfileGIEV15}, inferencev1alpha1.EppImageGIEV15, false),
		Entry("profile with its pinned image", inferencev1alpha1.EppSpec{Profile: inferencev1alpha1.EppProfileGIEV15, Image: inferencev1alpha1.EppImageGIEV15}, inferencev1alpha1.EppImageGIEV15, false),
		Entry("approved image only", inferencev1alpha1.EppSpec{Image: inferencev1alpha1.EppImageGIEV15}, inferencev1alpha1.EppImageGIEV15, false),
		Entry("unapproved image only", inferencev1alpha1.EppSpec{Image: customImage}, "", true),
		Entry("missing profile and image", inferencev1alpha1.EppSpec{}, "", true),
		Entry("unknown profile", inferencev1alpha1.EppSpec{Profile: "unsupported", Image: customImage}, "", true),
		Entry("incompatible image override", inferencev1alpha1.EppSpec{Profile: inferencev1alpha1.EppProfileGIEV15, Image: customImage}, "", true),
	)

	DescribeTable("validates EPP profile and image combinations at admission",
		func(epp map[string]any, valid bool) {
			err := k8sClient.Create(ctx, endpointWithEpp(endpoint, epp))
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(errors.IsInvalid(err)).To(BeTrue(), "expected admission rejection, got %v", err)
			}
		},
		Entry("profile alone", map[string]any{profileField: profileName}, true),
		Entry("profile with its pinned image", map[string]any{profileField: profileName, imageField: inferencev1alpha1.EppImageGIEV15}, true),
		Entry("approved image only", map[string]any{imageField: inferencev1alpha1.EppImageGIEV15}, true),
		Entry("unapproved image only", map[string]any{imageField: customImage}, false),
		Entry("missing profile and image", map[string]any{}, false),
		Entry("unknown profile", map[string]any{profileField: "unsupported", imageField: customImage}, false),
		Entry("empty profile", map[string]any{profileField: "", imageField: customImage}, false),
		Entry("incompatible image override", map[string]any{profileField: profileName, imageField: customImage}, false),
	)

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
		Expect(dep.Spec.Template.Spec.Containers[0].Image).To(Equal(endpoint.Spec.Epp.Image))
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

	It("creates, updates and removes the InferenceModelRewrite with the rewrites field", func() {
		_, err := reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		rewrite := &unstructured.Unstructured{}
		rewrite.SetGroupVersionKind(rewriteGVK)
		key := types.NamespacedName{Name: endpointName + "-rewrite", Namespace: ns}
		Expect(errors.IsNotFound(k8sClient.Get(ctx, key, rewrite))).To(BeTrue(), "no rewrite object without rewrites")

		By("adding a canary rewrite")
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		endpoint.Spec.Rewrites = []inferencev1alpha1.RewriteSpec{{
			Model:   "llama3",
			Targets: []inferencev1alpha1.RewriteTarget{{Model: "llama3", Weight: 90}, {Model: "llama3-v2", Weight: 10}},
		}}
		Expect(k8sClient.Update(ctx, endpoint)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient.Get(ctx, key, rewrite)).To(Succeed())
		Expect(metav1.IsControlledBy(rewrite, endpoint)).To(BeTrue())
		rules, _, _ := unstructured.NestedSlice(rewrite.Object, "spec", "rules")
		Expect(rules).To(HaveLen(1))
		targets := rules[0].(map[string]any)["targets"].([]any)
		Expect(targets).To(HaveLen(2))
		Expect(targets[1].(map[string]any)["modelRewrite"]).To(Equal("llama3-v2"))
		Expect(targets[1].(map[string]any)["weight"]).To(Equal(int64(10)))

		By("removing the rewrites again")
		Expect(k8sClient.Get(ctx, request.NamespacedName, endpoint)).To(Succeed())
		endpoint.Spec.Rewrites = nil
		Expect(k8sClient.Update(ctx, endpoint)).To(Succeed())
		_, err = reconciler.Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(errors.IsNotFound(k8sClient.Get(ctx, key, &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": rewriteGVK.GroupVersion().String(), "kind": rewriteGVK.Kind}}))).To(BeTrue(), "rewrite object removed with the field")
	})

	It("does nothing when the endpoint no longer exists", func() {
		_, err := reconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: "never-created", Namespace: ns},
		})
		Expect(err).NotTo(HaveOccurred())
	})
})

// endpointWithEpp creates an unstructured resource to exercise CRD admission,
// including the distinction between absent fields and explicitly empty strings.
func endpointWithEpp(endpoint *inferencev1alpha1.InferenceEndpoint, epp map[string]any) *unstructured.Unstructured {
	spec, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&endpoint.Spec)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	spec["epp"] = epp
	candidate := &unstructured.Unstructured{Object: map[string]any{"spec": spec}}
	candidate.SetAPIVersion(inferencev1alpha1.GroupVersion.String())
	candidate.SetKind("InferenceEndpoint")
	candidate.SetName("profiled")
	candidate.SetNamespace(endpoint.Namespace)
	return candidate
}

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
