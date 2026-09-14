/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

const (
	oldChildUID  = types.UID("child-uid-old")
	newChildUID  = types.UID("child-uid-new")
	testEwUID    = types.UID("ew-uid-1")
	testWorkload = "cp-workload-id-1"
)

var originalCreation = metav1.NewTime(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))

// survivingExternalWorkload is the ExternalWorkload left behind after the runner deleted its child.
func survivingExternalWorkload(runnerUID types.UID, workloadID string) *v1alpha1.ExternalWorkload {
	ew := &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GenerateExternalWorkloadName(pytorchGVK, testJobName),
			Namespace: testNamespace,
			UID:       testEwUID,
			Labels: map[string]string{
				commonconstants.WorkloadRunnerNameLabel: testWlrName,
			},
			Annotations:     map[string]string{},
			OwnerReferences: []metav1.OwnerReference{wlrOwnerRefWithUID(testWlrName, testWlrUID)},
		},
		Spec: v1alpha1.ExternalWorkloadSpec{
			OriginalWorkloadUID:  oldChildUID,
			OriginalWorkloadName: testJobName,
			OriginalWorkloadKind: pytorchGVK.Kind,
			Group:                pytorchGVK.Group,
			Version:              pytorchGVK.Version,
			CreationTimestamp:    originalCreation,
		},
	}
	if runnerUID != "" {
		ew.Labels[commonconstants.WorkloadRunnerUidLabel] = string(runnerUID)
	}
	if workloadID != "" {
		ew.Annotations[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload] = workloadID
	}
	return ew
}

func rebindHandler(objects ...client.Object) (*ExternalWorkloadHandler, client.Client) {
	scheme := runtime.NewScheme()
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	c := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		WithIndex(&v1alpha1.ExternalWorkload{}, constants.ExternalWorkloadByOriginalUIDIndexerName,
			cachedclient.ExternalWorkloadByOriginalWorkloadUIDIndexer).
		Build()
	return &ExternalWorkloadHandler{cachedClient: c}, c
}

func storedExternalWorkload(ctx context.Context, c client.Client) *v1alpha1.ExternalWorkload {
	stored := &v1alpha1.ExternalWorkload{}
	Expect(c.Get(ctx, types.NamespacedName{
		Name:      utils.GenerateExternalWorkloadName(pytorchGVK, testJobName),
		Namespace: testNamespace,
	}, stored)).To(Succeed())
	return stored
}

// A restart recreates the child with a new uid, so the uid index no longer finds the surviving EW.
var _ = Describe("ExternalWorkloadHandler.getOrRebindExternalWorkloadOfObject", func() {
	var ctx context.Context
	var recreatedChild *metav1.PartialObjectMetadata

	BeforeEach(func() {
		ctx = context.Background()
		recreatedChild = newChildObject(newChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))
	})

	It("re-points the surviving external workload at the recreated child", func() {
		handler, c := rebindHandler(survivingExternalWorkload(testWlrUID, testWorkload))

		rebound, err := handler.getOrRebindExternalWorkloadOfObject(ctx, recreatedChild, newChildUID)

		Expect(err).NotTo(HaveOccurred())
		Expect(rebound).NotTo(BeNil())
		Expect(rebound.Spec.OriginalWorkloadUID).To(Equal(newChildUID))

		all := &v1alpha1.ExternalWorkloadList{}
		Expect(c.List(ctx, all)).To(Succeed())
		Expect(all.Items).To(HaveLen(1), "the surviving external workload must be reused, not duplicated")

		stored := storedExternalWorkload(ctx, c)
		Expect(stored.Spec.OriginalWorkloadUID).To(Equal(newChildUID))
		Expect(stored.UID).To(Equal(testEwUID), "external workload identity must survive the restart")
		Expect(stored.Spec.CreationTimestamp.UTC()).To(Equal(originalCreation.UTC()),
			"the original submission time must be preserved across restarts")
		Expect(stored.Annotations).To(
			HaveKeyWithValue(commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload, testWorkload))
	})

	It("refuses to re-bind an external workload belonging to a different runner", func() {
		handler, c := rebindHandler(survivingExternalWorkload("some-other-runner-uid", testWorkload))

		_, err := handler.getOrRebindExternalWorkloadOfObject(ctx, recreatedChild, newChildUID)

		Expect(apierrors.IsNotFound(err)).To(BeTrue(),
			"a resubmitted workload must not adopt a dead workload's history")
		Expect(storedExternalWorkload(ctx, c).Spec.OriginalWorkloadUID).To(Equal(oldChildUID))
	})

	It("refuses to re-bind when the external workload has no control-plane workload id", func() {
		handler, c := rebindHandler(survivingExternalWorkload(testWlrUID, ""))

		_, err := handler.getOrRebindExternalWorkloadOfObject(ctx, recreatedChild, newChildUID)

		Expect(apierrors.IsNotFound(err)).To(BeTrue(),
			"without the annotation the workload id would follow the child uid")
		Expect(storedExternalWorkload(ctx, c).Spec.OriginalWorkloadUID).To(Equal(oldChildUID))
	})

	It("refuses to re-bind when the external workload carries no runner uid", func() {
		handler, _ := rebindHandler(survivingExternalWorkload("", testWorkload))

		_, err := handler.getOrRebindExternalWorkloadOfObject(ctx, recreatedChild, newChildUID)

		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("does not re-bind a child that has no runner owner", func() {
		handler, _ := rebindHandler(survivingExternalWorkload(testWlrUID, testWorkload))
		orphanChild := newChildObject(newChildUID)

		_, err := handler.getOrRebindExternalWorkloadOfObject(ctx, orphanChild, newChildUID)

		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("re-binds an external workload whose name carries a random suffix", func() {
		suffixed := survivingExternalWorkload(testWlrUID, testWorkload)
		suffixed.Name = "ew-pytorchjob-a-very-long-workload-name-that-got-truncat-9f3k"
		handler, c := rebindHandler(suffixed)

		rebound, err := handler.getOrRebindExternalWorkloadOfObject(ctx, recreatedChild, newChildUID)

		Expect(err).NotTo(HaveOccurred())
		Expect(rebound).NotTo(BeNil())
		Expect(rebound.Spec.OriginalWorkloadUID).To(Equal(newChildUID))

		all := &v1alpha1.ExternalWorkloadList{}
		Expect(c.List(ctx, all)).To(Succeed())
		Expect(all.Items).To(HaveLen(1), "a truncated name must not cause a duplicate external workload")
	})

	It("still finds the external workload by the child uid index", func() {
		handler, _ := rebindHandler(survivingExternalWorkload(testWlrUID, testWorkload))
		child := newChildObject(oldChildUID, wlrOwnerRefWithUID(testWlrName, testWlrUID))

		found, err := handler.getOrRebindExternalWorkloadOfObject(ctx, child, oldChildUID)

		Expect(err).NotTo(HaveOccurred())
		Expect(found.UID).To(Equal(testEwUID))
	})
})
