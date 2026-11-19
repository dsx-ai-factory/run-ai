/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	ctrlreconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("mapToRunner", func() {
	labelled := metav1.ObjectMeta{
		Namespace: "runai-proj",
		Labels:    map[string]string{commonconstants.WorkloadRunnerNameLabel: "llama"},
	}

	DescribeTable("maps a labelled object to its runner",
		func(object client.Object) {
			Expect(mapToRunner(context.Background(), object)).To(ConsistOf(
				ctrlreconcile.Request{NamespacedName: types.NamespacedName{Namespace: "runai-proj", Name: "llama"}}))
		},
		Entry("a replica", &corev1.Pod{ObjectMeta: labelled}),
		Entry("a SnapshotJob", &snapshotv1alpha1.SnapshotJob{ObjectMeta: labelled}),
		Entry("a PodSnapshot", &snapshotv1alpha1.PodSnapshot{ObjectMeta: labelled}),
	)

	It("maps an unlabelled object to nothing", func() {
		Expect(mapToRunner(context.Background(), &corev1.Pod{})).To(BeEmpty())
	})
})

var _ = Describe("replicaReadinessChanged", func() {
	var cold, ready, notOptedIn, notOptedInCold *corev1.Pod
	changed := replicaReadinessChanged()

	BeforeEach(func() {
		cold = replica("llama-a", "rs-1", 100, false)
		ready = replica("llama-a", "rs-1", 100, true)
		notOptedIn = ready.DeepCopy()
		delete(notOptedIn.Labels, constants.SnapshotOptInLabel)
		notOptedInCold = cold.DeepCopy()
		delete(notOptedInCold.Labels, constants.SnapshotOptInLabel)
	})

	DescribeTable("passes only a change in a replica's readiness",
		func(passes func() bool, want bool) {
			Expect(passes()).To(Equal(want))
		},
		Entry("turned Ready", func() bool {
			return changed.Update(event.UpdateEvent{ObjectOld: cold, ObjectNew: ready})
		}, true),
		Entry("stopped being Ready", func() bool {
			return changed.Update(event.UpdateEvent{ObjectOld: ready, ObjectNew: cold})
		}, true),
		Entry("stayed Ready", func() bool {
			return changed.Update(event.UpdateEvent{ObjectOld: ready, ObjectNew: ready})
		}, false),
		Entry("not opted in", func() bool {
			return changed.Update(event.UpdateEvent{ObjectOld: notOptedInCold, ObjectNew: notOptedIn})
		}, false),
		Entry("seen Ready on a restart", func() bool {
			return changed.Create(event.CreateEvent{Object: ready})
		}, true),
		Entry("created cold", func() bool {
			return changed.Create(event.CreateEvent{Object: cold})
		}, false),
		Entry("deleted", func() bool {
			return changed.Delete(event.DeleteEvent{Object: ready})
		}, false),
	)
})
