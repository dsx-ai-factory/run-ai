/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package knativescaling

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMinimum(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Knative minimum replica updates")
}

var _ = Describe("updating an existing inference Service", func() {
	var service, config, revision, crd, desired *unstructured.Unstructured
	var c client.Client
	var ctx context.Context

	child := func(parent *unstructured.Unstructured, kind, name string) *unstructured.Unstructured {
		u := object(servingAPI, kind)
		u.SetNamespace(parent.GetNamespace())
		u.SetName(name)
		u.SetUID(types.UID("uid-" + kind))
		u.SetOwnerReferences([]metav1.OwnerReference{*metav1.NewControllerRef(parent, parent.GroupVersionKind())})
		return u
	}
	minimum := func(u *unstructured.Unstructured) string {
		GinkgoHelper()
		value, ok, err := unstructured.NestedString(u.Object, minimumPath...)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
		return value
	}
	preview := func(_ context.Context, input *unstructured.Unstructured) (*unstructured.Unstructured, error) {
		input.SetUID(service.GetUID())
		return input, nil
	}
	buildClient := func() client.Client {
		return fake.NewClientBuilder().WithObjects(service, config, revision, crd).Build()
	}
	currentMin := func() string {
		GinkgoHelper()
		stored := object(servingAPI, "Revision")
		Expect(c.Get(ctx, client.ObjectKeyFromObject(revision), stored)).To(Succeed())
		return stored.GetAnnotations()[minScale]
	}
	changeMin := func(value string) *unstructured.Unstructured {
		GinkgoHelper()
		u := service.DeepCopy()
		u.SetUID("")
		u.SetResourceVersion("")
		delete(u.Object, "status")
		Expect(unstructured.SetNestedField(u.Object, value, minimumPath...)).To(Succeed())
		return u
	}

	BeforeEach(func() {
		ctx = context.Background()
		runner := object("run.ai/v1alpha1", "WorkloadRunner")
		runner.SetName("inference")
		runner.SetNamespace("project")
		runner.SetUID("runner")
		service = child(runner, "Service", "inference")
		service.SetGeneration(4)
		service.Object["status"] = map[string]any{
			"observedGeneration": int64(4), "latestReadyRevisionName": "rev", "latestCreatedRevisionName": "rev",
			"traffic":    []any{map[string]any{"revisionName": "rev", "percent": int64(100)}},
			"conditions": []any{map[string]any{"type": "Ready", "status": "True"}},
		}
		Expect(unstructured.SetNestedField(service.Object, "1", minimumPath...)).To(Succeed())
		config = child(service, "Configuration", "inference")
		revision = child(config, "Revision", "rev")
		revision.SetAnnotations(map[string]string{minScale: "1"})
		crd = object("apiextensions.k8s.io/v1", "CustomResourceDefinition")
		crd.SetName("revisions.serving.knative.dev")
		crd.SetLabels(map[string]string{"app.kubernetes.io/version": "1.21.0"})
		desired = changeMin("0")
		c = buildClient()
	})

	It("lowers the serving floor, preserves it on retry, restores it on reversal, and carries it into an image update", func() {
		for range 2 {
			desired = changeMin("0")
			Expect(Prepare(ctx, c, desired, preview)).To(Succeed())
			Expect(currentMin()).To(Equal("0"))
			Expect(minimum(desired)).To(Equal("1"))
			Expect(desired.GetResourceVersion()).NotTo(BeEmpty())
		}
		desired = changeMin("1")
		Expect(Prepare(ctx, c, desired, preview)).To(Succeed())
		Expect(currentMin()).To(Equal("1"))

		desired = changeMin("0")
		Expect(unstructured.SetNestedField(desired.Object, []any{map[string]any{"image": "model:v2"}}, "spec", "template", "spec", "containers")).To(Succeed())
		Expect(Prepare(ctx, c, desired, preview)).To(Succeed())
		Expect(minimum(desired)).To(Equal("0"))
		Expect(currentMin()).To(Equal("1"))
	})

	It("retries if another writer changes the Revision during service admission", func() {
		err := Prepare(ctx, c, desired, func(ctx context.Context, input *unstructured.Unstructured) (*unstructured.Unstructured, error) {
			stored := object(servingAPI, "Revision")
			Expect(c.Get(ctx, client.ObjectKeyFromObject(revision), stored)).To(Succeed())
			stored.SetLabels(map[string]string{"serving.knative.dev/routingState": "active"})
			Expect(c.Update(ctx, stored)).To(Succeed())
			return preview(ctx, input)
		})
		Expect(apierrors.IsConflict(err)).To(BeTrue(), "error: %v", err)
		Expect(currentMin()).To(Equal("1"))
		Expect(minimum(desired)).To(Equal("0"))
	})

	It("runs admission on the requested minimum before changing the Revision", func() {
		rejected := errors.New("policy forbids zero")
		err := Prepare(ctx, c, desired, func(_ context.Context, input *unstructured.Unstructured) (*unstructured.Unstructured, error) {
			Expect(minimum(input)).To(Equal("0"))
			return nil, rejected
		})
		Expect(err).To(MatchError(ContainSubstring(rejected.Error())))
		Expect(currentMin()).To(Equal("1"))
	})

	It("does not roll a new Revision when a previously lowered service temporarily loses Ready", func() {
		Expect(unstructured.SetNestedSlice(service.Object, []any{map[string]any{"type": "Ready", "status": "Unknown"}}, "status", "conditions")).To(Succeed())
		revision.SetAnnotations(map[string]string{minScale: "0"})
		c = buildClient()
		Expect(Prepare(ctx, c, desired, preview)).To(Succeed())
		Expect(minimum(desired)).To(Equal("1"))
		Expect(currentMin()).To(Equal("0"))
	})

	DescribeTable("retains normal Knative rollout behavior",
		func(modify func()) {
			modify()
			c = buildClient()
			Expect(Prepare(ctx, c, desired, preview)).To(Succeed())
			Expect(minimum(desired)).To(Equal("0"))
			Expect(currentMin()).To(Equal("1"))
		},
		Entry("when the installed Knative cannot update autoscaling in place", func() { crd.SetLabels(map[string]string{"app.kubernetes.io/version": "1.20.4"}) }),
		Entry("when using a different autoscaler", func() {
			revision.SetAnnotations(map[string]string{minScale: "1", "autoscaling.knative.dev/class": "hpa.autoscaling.knative.dev"})
		}),
		Entry("when the Knative version is unknown", func() { crd.SetLabels(nil) }),
		Entry("when stale status names a deleted Revision", func() {
			revision.SetName("some-other-revision")
		}),
		Entry("when the original Service has never become Ready", func() { delete(service.Object, "status") }),
		Entry("when a different Revision is already rolling out", func() {
			Expect(unstructured.SetNestedField(service.Object, "rev-next", "status", "latestCreatedRevisionName")).To(Succeed())
		}),
		Entry("when traffic is split", func() {
			Expect(unstructured.SetNestedSlice(service.Object, []any{map[string]any{"revisionName": "rev", "percent": int64(70)}, map[string]any{"revisionName": "older", "percent": int64(30)}}, "status", "traffic")).To(Succeed())
		}),
	)

	It("does not patch a Revision belonging to a different Configuration", func() {
		refs := revision.GetOwnerReferences()
		refs[0].UID = "different-config"
		revision.SetOwnerReferences(refs)
		c = buildClient()
		Expect(Prepare(ctx, c, desired, preview)).To(MatchError(ContainSubstring("no longer belongs")))
		Expect(currentMin()).To(Equal("1"))
	})
})
