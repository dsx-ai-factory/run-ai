/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package inferenceendpoint

import (
	"context"
	"encoding/json"
	"fmt"

	ep "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const spec = `{"workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},"routing":{"gatewayRef":{"name":"gateway"},"hostnames":["model.example.test"]},"epp":{"profile":"gie-v1.5"}}`

var _ = Describe("Inference endpoint lifecycle", func() {
	var (
		ctx    context.Context
		runner *dromos.WorkloadRunner
		c      client.WithWatch
		p      *Plugin
		state  *reconcile.State
		scheme *runtime.Scheme
	)
	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(dromos.AddToScheme(scheme)).To(Succeed())
		scheme.AddKnownTypeWithName(ep.InferenceEndpointGVK, &unstructured.Unstructured{})
		scheme.AddKnownTypeWithName(ep.InferenceEndpointGVK.GroupVersion().WithKind("InferenceEndpointList"), &unstructured.UnstructuredList{})
		runner = &dromos.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "model", Namespace: "project", UID: "runner-uid", Generation: 2}}
		Expect(ep.SetInferenceEndpoint(runner, &ep.InferenceEndpointConfig{Spec: json.RawMessage(spec)})).To(Succeed())
		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{ep.InferenceEndpointGVK.GroupVersion()})
		mapper.Add(ep.InferenceEndpointGVK, meta.RESTScopeNamespace)
		c = fake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(mapper).Build()
		p = NewPlugin(c, c.RESTMapper())
		state = reconcile.New()
	})
	reconcileEndpoint := func() error { return p.HandleAdditionalResources(ctx, runner, nil, nil, c, state) }
	getEndpoint := func() *unstructured.Unstructured {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(ep.InferenceEndpointGVK)
		name, err := ep.InferenceEndpointName(runner)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Get(ctx, client.ObjectKey{Namespace: runner.Namespace, Name: name}, obj)).To(Succeed())
		return obj
	}
	It("creates an endpoint owned by the runner and does not create it twice", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		Expect(metav1.IsControlledBy(obj, runner)).To(BeTrue())
		rv := obj.GetResourceVersion()
		Expect(reconcileEndpoint()).To(Succeed())
		Expect(getEndpoint().GetResourceVersion()).To(Equal(rv))
	})
	It("updates the spec while preserving unrelated metadata and status", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		obj.SetAnnotations(map[string]string{"keep": "value"})
		obj.Object["status"] = map[string]any{"observedGeneration": int64(1)}
		Expect(c.Update(ctx, obj)).To(Succeed())
		Expect(ep.SetInferenceEndpoint(runner, &ep.InferenceEndpointConfig{Spec: json.RawMessage(replacePort(spec, 9000))})).To(Succeed())
		Expect(reconcileEndpoint()).To(Succeed())
		obj = getEndpoint()
		Expect(obj.GetAnnotations()).To(HaveKeyWithValue("keep", "value"))
		Expect(obj.Object).To(HaveKey("status"))
		port, _, err := unstructured.NestedInt64(obj.Object, "spec", "workload", "port")
		Expect(err).NotTo(HaveOccurred())
		Expect(port).To(Equal(int64(9000)))
	})
	It("refuses to adopt an endpoint owned by another runner", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		owners := obj.GetOwnerReferences()
		owners[0].UID = "other"
		obj.SetOwnerReferences(owners)
		Expect(c.Update(ctx, obj)).To(Succeed())
		Expect(reconcileEndpoint()).To(MatchError(ContainSubstring("not controlled")))
		Expect(getEndpoint().GetOwnerReferences()[0].UID).To(Equal(owners[0].UID))
	})
	It("cleans only its own endpoint when configuration is removed", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		delete(runner.Annotations, ep.InferenceEndpointAnnotation)
		Expect(reconcileEndpoint()).To(Succeed())
		Expect(apierrors.IsNotFound(c.Get(ctx, client.ObjectKeyFromObject(obj), obj))).To(BeTrue())
		Expect(reconcileEndpoint()).To(Succeed())
	})
	It("does not delete a foreign endpoint when configuration is absent", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		obj.SetOwnerReferences(nil)
		Expect(c.Update(ctx, obj)).To(Succeed())
		delete(runner.Annotations, ep.InferenceEndpointAnnotation)
		Expect(reconcileEndpoint()).To(Succeed())
		Expect(getEndpoint().GetOwnerReferences()).To(BeEmpty())
	})
	It("does not recreate endpoints for a deleting runner", func() {
		runner.DeletionTimestamp = &metav1.Time{Time: metav1.Now().Time}
		Expect(reconcileEndpoint()).To(Succeed())
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(ep.InferenceEndpointGVK.GroupVersion().WithKind("InferenceEndpointList"))
		Expect(c.List(ctx, list)).To(Succeed())
		Expect(list.Items).To(BeEmpty())
	})
	It("records API failures so readiness cannot hide a failed apply", func() {
		c = interceptor.NewClient(c, interceptor.Funcs{Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
			return fmt.Errorf("write denied")
		}})
		Expect(reconcileEndpoint()).To(MatchError(ContainSubstring("write denied")))
		s := &dromos.PluginStatus{}
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhaseFailed))
	})
	It("is disabled without intent and pending until an endpoint exists", func() {
		s := &dromos.PluginStatus{}
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhasePending))
		delete(runner.Annotations, ep.InferenceEndpointAnnotation)
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhaseDisabled))
		Expect(s.AdditionalInfo).To(BeNil())
	})
	It("reports readiness only for the desired spec and current generation", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		obj.SetGeneration(3)
		obj.Object["status"] = map[string]any{"observedGeneration": int64(3), "conditions": []any{map[string]any{"type": "Ready", "status": "True", "observedGeneration": int64(2), "reason": "Ready", "message": "ready"}}}
		Expect(c.Update(ctx, obj)).To(Succeed())
		s := &dromos.PluginStatus{}
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhasePending))
		obj.Object["status"].(map[string]any)["conditions"].([]any)[0].(map[string]any)["observedGeneration"] = int64(3)
		Expect(c.Update(ctx, obj)).To(Succeed())
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhaseReady))
		Expect(s.AdditionalInfo).NotTo(BeNil())
		Expect(string(s.AdditionalInfo.Raw)).To(ContainSubstring("model.example.test"))
		Expect(ep.SetInferenceEndpoint(runner, &ep.InferenceEndpointConfig{Spec: json.RawMessage(replacePort(spec, 9000))})).To(Succeed())
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhasePending))
	})
	It("does not trust Ready on a foreign endpoint", func() {
		Expect(reconcileEndpoint()).To(Succeed())
		obj := getEndpoint()
		obj.SetOwnerReferences(nil)
		Expect(c.Update(ctx, obj)).To(Succeed())
		s := &dromos.PluginStatus{}
		Expect(p.Status(ctx, runner, c, s, state)).To(Succeed())
		Expect(s.Phase).To(Equal(dromos.PluginPhaseFailed))
	})
	It("rejects malformed intent during admission", func() {
		runner.Annotations[ep.InferenceEndpointAnnotation] = "{}"
		_, errs := p.WebhookValidate(ctx, nil, runner)
		Expect(errs).NotTo(BeEmpty())
	})
	It("allows unrelated workloads when the endpoint CRD is absent", func() {
		mapper := meta.NewDefaultRESTMapper(nil)
		p = NewPlugin(c, mapper)
		b := &builder.Builder{}
		got, err := p.Watchers(b)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeIdenticalTo(b))
		delete(runner.Annotations, ep.InferenceEndpointAnnotation)
		_, errs := p.WebhookValidate(ctx, nil, runner)
		Expect(errs).To(BeEmpty())
	})
	It("reports a missing CRD for configured workloads", func() {
		p = NewPlugin(c, meta.NewDefaultRESTMapper(nil))
		_, errs := p.WebhookValidate(ctx, nil, runner)
		Expect(errs).NotTo(BeEmpty())
		Expect(errs.ToAggregate().Error()).To(ContainSubstring("InferenceEndpoint"))
	})
	It("rejects a builder without a usable client", func() {
		_, err := New(ctx, framework.PluginDeps{})
		Expect(err).To(HaveOccurred())
	})
})

func replacePort(raw string, port int) string {
	var v map[string]any
	Expect(json.Unmarshal([]byte(raw), &v)).To(Succeed())
	v["workload"].(map[string]any)["port"] = port
	b, err := json.Marshal(v)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}
