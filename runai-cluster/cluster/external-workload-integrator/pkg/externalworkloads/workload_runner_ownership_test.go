/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testWlrUID = types.UID("wlr-uid-1")

var pytorchGVK = schema.GroupVersionKind{Group: "kubeflow.org", Version: "v1", Kind: "PyTorchJob"}

func wlrOwnerRefWithUID(name string, uid types.UID) metav1.OwnerReference {
	ref := wlrOwnerRef(name)
	ref.UID = uid
	return ref
}

func newChildObject(uid types.UID, ownerRefs ...metav1.OwnerReference) *metav1.PartialObjectMetadata {
	object := newObjectWithOwnerRefs(ownerRefs...)
	object.UID = uid
	object.GetObjectKind().SetGroupVersionKind(pytorchGVK)
	return object
}

var _ = Describe("ExternalWorkloadHandler.getOwnerReferences", func() {
	It("returns the WorkloadRunner when the child is owned by one", func() {
		child := newChildObject("child-uid-1", wlrOwnerRefWithUID(testWlrName, testWlrUID))
		handler := &ExternalWorkloadHandler{}

		ownerRefs := handler.getOwnerReferences(child)

		Expect(ownerRefs).To(HaveLen(1))
		Expect(ownerRefs[0].Kind).To(Equal("WorkloadRunner"))
		Expect(ownerRefs[0].Name).To(Equal(testWlrName))
		Expect(ownerRefs[0].UID).To(Equal(testWlrUID))
	})

	It("returns the child itself when there is no WorkloadRunner owner", func() {
		child := newChildObject("child-uid-1", metav1.OwnerReference{
			APIVersion: "batch/v1", Kind: "CronJob", Name: "my-cronjob",
		})
		handler := &ExternalWorkloadHandler{}

		ownerRefs := handler.getOwnerReferences(child)

		Expect(ownerRefs).To(HaveLen(1))
		Expect(ownerRefs[0].Kind).To(Equal(pytorchGVK.Kind))
		Expect(ownerRefs[0].Name).To(Equal(testJobName))
		Expect(ownerRefs[0].UID).To(Equal(types.UID("child-uid-1")))
	})

	It("returns the child itself when there are no owner refs at all", func() {
		child := newChildObject("child-uid-1")
		handler := &ExternalWorkloadHandler{}

		ownerRefs := handler.getOwnerReferences(child)

		Expect(ownerRefs).To(HaveLen(1))
		Expect(ownerRefs[0].Kind).To(Equal(pytorchGVK.Kind))
	})
})

var _ = Describe("setWorkloadRunnerLabel UID stamping", func() {
	It("stamps the runner uid alongside the runner name", func() {
		child := newChildObject("child-uid-1", wlrOwnerRefWithUID(testWlrName, testWlrUID))
		ew := newExternalWorkload()

		setWorkloadRunnerLabel(child, ew)

		Expect(ew.Labels[commonconstants.WorkloadRunnerNameLabel]).To(Equal(testWlrName))
		Expect(ew.Labels[commonconstants.WorkloadRunnerUidLabel]).To(Equal(string(testWlrUID)))
	})

	It("stamps no runner labels when the owner is not a WorkloadRunner", func() {
		child := newChildObject("child-uid-1", metav1.OwnerReference{
			APIVersion: "batch/v1", Kind: "CronJob", Name: "my-cronjob",
		})
		ew := newExternalWorkload()

		setWorkloadRunnerLabel(child, ew)

		Expect(ew.Labels).NotTo(HaveKey(commonconstants.WorkloadRunnerUidLabel))
	})
})

// The ownerRef flip only reaches existing ExternalWorkloads if an ownerRef-only change is written.
var _ = Describe("ExternalWorkloadHandler.updateSpecIfNeeded owner references", func() {
	It("persists an ownerRef-only change", func() {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ew-pytorchjob-my-job",
				Namespace: testNamespace,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "kubeflow.org/v1", Kind: "PyTorchJob",
					Name: testJobName, UID: "child-uid-1",
				}},
			},
			Spec: v1alpha1.ExternalWorkloadSpec{OriginalWorkloadName: testJobName},
		}
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(ew).Build()

		stored := &v1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, client.ObjectKeyFromObject(ew), stored)).To(Succeed())
		updated := stored.DeepCopy()
		updated.OwnerReferences = []metav1.OwnerReference{wlrOwnerRefWithUID(testWlrName, testWlrUID)}

		handler := &ExternalWorkloadHandler{cachedClient: c}
		Expect(handler.updateSpecIfNeeded(ctx, updated, stored)).To(Succeed())

		after := &v1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, client.ObjectKeyFromObject(ew), after)).To(Succeed())
		Expect(after.OwnerReferences).To(HaveLen(1))
		Expect(after.OwnerReferences[0].Kind).To(Equal("WorkloadRunner"))
	})
})
