/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("cleanup", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	// cleanUp reconciles testRunner's name.
	cleanUp := func(c client.Client) {
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "runai-proj", Name: "llama"}}
		_, err := (&cleaner{client: c, apiReader: c}).Reconcile(ctx, request)
		Expect(err).NotTo(HaveOccurred())
	}
	named := func(key string) gomegatypes.GomegaMatcher {
		return HaveField("ObjectMeta.Name", snapshotkey.SnapshotName(key, "llama"))
	}

	It("deletes a deleted runner's SnapshotJobs and PodSnapshots", func() {
		c := captureClient(capturedJob("k1", true), capturedPodSnapshot("k1", "runner-uid"))

		cleanUp(c)

		Expect(snapshotJobs(ctx, c)).To(BeEmpty())
		Expect(podSnapshots(ctx, c)).To(BeEmpty())
	})

	It("keeps the current runner's", func() {
		c := captureClient(testRunner(), capturedJob("k1", true), capturedPodSnapshot("k1", "runner-uid"))

		cleanUp(c)

		Expect(snapshotJobs(ctx, c)).To(ConsistOf(named("k1")))
		Expect(podSnapshots(ctx, c)).To(ConsistOf(named("k1")))
	})

	// A workload deleted and resubmitted under the same name gets a new runner.
	It("deletes an earlier runner's of the same name", func() {
		earlier := capturedJob("old-key", true)
		earlier.Labels[commonconstants.WorkloadRunnerUidLabel] = "earlier-runner-uid"
		c := captureClient(testRunner(),
			earlier, capturedPodSnapshot("old-key", "earlier-runner-uid"),
			capturedJob("k1", true), capturedPodSnapshot("k1", "runner-uid"))

		cleanUp(c)

		Expect(snapshotJobs(ctx, c)).To(ConsistOf(named("k1")))
		Expect(podSnapshots(ctx, c)).To(ConsistOf(named("k1")))
	})

	// The workload is resubmitted, and its new runner captures, while the old
	// runner's captures are being cleaned up.
	DescribeTable("leaves a capture made by a new runner during the cleanup",
		func(beforeLookup bool) {
			earlier := capturedJob("old-key", true)
			earlier.Labels[commonconstants.WorkloadRunnerUidLabel] = "earlier-runner-uid"
			c := captureClient(earlier)
			resubmit := func() {
				Expect(c.Create(ctx, testRunner())).To(Succeed())
				Expect(c.Create(ctx, capturedJob("k1", true))).To(Succeed())
			}
			lookup := lookupHook{Reader: c, before: func() {}, after: func() {}}
			if beforeLookup {
				lookup.before = resubmit
			} else {
				lookup.after = resubmit
			}

			request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "runai-proj", Name: "llama"}}
			_, err := (&cleaner{client: c, apiReader: lookup}).Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Expect(snapshotJobs(ctx, c)).To(ConsistOf(named("k1")))
		},
		Entry("between the listing and the runner lookup", true),
		Entry("right after the runner lookup", false),
	)

	It("leaves what the plugin didn't make, and other runners'", func() {
		unlabelled := capturedJob("k1", true)
		delete(unlabelled.Labels, commonconstants.WorkloadRunnerUidLabel)
		otherName := capturedJob("k2", true)
		otherName.Labels[commonconstants.WorkloadRunnerNameLabel] = "mistral"
		otherName.Labels[commonconstants.WorkloadRunnerUidLabel] = "mistral-uid"
		otherNamespace := capturedJob("k3", true)
		otherNamespace.Namespace = "runai-other"
		otherNamespace.Labels[commonconstants.WorkloadRunnerUidLabel] = "other-llama-uid"
		c := captureClient(testRunner(), unlabelled, otherName, otherNamespace)

		cleanUp(c)

		Expect(snapshotJobs(ctx, c)).To(ConsistOf(named("k1"), named("k2"), named("k3")))
	})

	It("checks a runner when it is deleted", func() {
		runner := testRunner()
		deletes := onlyDeletes()

		Expect(deletes.Delete(event.DeleteEvent{Object: runner})).To(BeTrue())
		Expect(deletes.Create(event.CreateEvent{Object: runner})).To(BeFalse())
		Expect(deletes.Update(event.UpdateEvent{ObjectOld: runner, ObjectNew: runner})).To(BeFalse())
	})

	// The startup list arrives as creates, so a runner deleted while the
	// controller was down is still cleaned up.
	It("checks the runner of a SnapshotJob or PodSnapshot when it is created", func() {
		job := capturedJob("k1", true)
		creates := onlyCreates()

		Expect(creates.Create(event.CreateEvent{Object: job})).To(BeTrue())
		Expect(creates.Update(event.UpdateEvent{ObjectOld: job, ObjectNew: job})).To(BeFalse())
		Expect(creates.Delete(event.DeleteEvent{Object: job})).To(BeFalse())
	})
})

// lookupHook runs before and after each read it passes on.
type lookupHook struct {
	client.Reader
	before, after func()
}

func (h lookupHook) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	h.before()
	defer h.after()
	return h.Reader.Get(ctx, key, obj, opts...)
}

// capturedPodSnapshot is the PodSnapshot Snapshot creates for a capture of key,
// with the labels the SnapshotJob passes on.
func capturedPodSnapshot(key, runnerUID string) *snapshotv1alpha1.PodSnapshot {
	return &snapshotv1alpha1.PodSnapshot{ObjectMeta: metav1.ObjectMeta{
		Name:      snapshotkey.SnapshotName(key, "llama"),
		Namespace: "runai-proj",
		Labels: map[string]string{
			constants.SnapshotKeyLabel:              key,
			commonconstants.WorkloadRunnerNameLabel: "llama",
			commonconstants.WorkloadRunnerUidLabel:  runnerUID,
		},
	}}
}

func podSnapshots(ctx context.Context, c client.Client) []snapshotv1alpha1.PodSnapshot {
	list := &snapshotv1alpha1.PodSnapshotList{}
	Expect(c.List(ctx, list)).To(Succeed())
	return list.Items
}
