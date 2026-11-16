/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"errors"

	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("HandleRunnerDeleted", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	name := types.NamespacedName{Namespace: "runai-proj", Name: "llama"}
	named := func(key string) gomegatypes.GomegaMatcher {
		return HaveField("ObjectMeta.Name", snapshotkey.SnapshotName(key, "llama"))
	}

	It("deletes the runner's SnapshotJobs and PodSnapshots, and an earlier runner's of the same name", func() {
		earlier := capturedJob("old-key", true)
		earlier.Labels[commonconstants.WorkloadRunnerUidLabel] = "earlier-runner-uid"
		c := captureClient(
			earlier, capturedPodSnapshot("old-key", "earlier-runner-uid"),
			capturedJob("k1", true), capturedPodSnapshot("k1", "runner-uid"))

		Expect(capturePlugin().HandleRunnerDeleted(ctx, name, c)).To(Succeed())

		Expect(snapshotJobs(ctx, c)).To(BeEmpty())
		Expect(podSnapshots(ctx, c)).To(BeEmpty())
	})

	It("leaves what the plugin didn't make, and other runners'", func() {
		unlabelled := capturedJob("k1", true)
		delete(unlabelled.Labels, commonconstants.WorkloadRunnerUidLabel)
		otherName := capturedJob("k2", true)
		otherName.Labels[commonconstants.WorkloadRunnerNameLabel] = "mistral"
		otherName.Labels[commonconstants.WorkloadRunnerUidLabel] = "mistral-uid"
		otherNamespace := capturedJob("k3", true)
		otherNamespace.Namespace = "runai-other"
		otherNamespace.Labels[commonconstants.WorkloadRunnerUidLabel] = "other-llama-uid"
		c := captureClient(unlabelled, otherName, otherNamespace)

		Expect(capturePlugin().HandleRunnerDeleted(ctx, name, c)).To(Succeed())

		Expect(snapshotJobs(ctx, c)).To(ConsistOf(named("k1"), named("k2"), named("k3")))
	})

	// dromos retries a reconcile that fails.
	It("fails when a delete fails, to be tried again", func() {
		c := failingDeletes(captureClient(capturedJob("k1", true)), errors.New("the API server is unreachable"))

		Expect(capturePlugin().HandleRunnerDeleted(ctx, name, c)).To(MatchError(ContainSubstring("unreachable")))
	})

	It("takes one already deleted for done", func() {
		gone := apierrors.NewNotFound(schema.GroupResource{Group: "nvidia.com", Resource: "snapshotjobs"}, "k1")
		c := failingDeletes(captureClient(capturedJob("k1", true)), gone)

		Expect(capturePlugin().HandleRunnerDeleted(ctx, name, c)).To(Succeed())
	})

	// Listing a kind whose CRD is missing fails, so it would fail for every runner.
	It("does nothing without Snapshot", func() {
		c := captureClient(capturedJob("k1", true))
		plugin := capturePlugin()
		plugin.snapshotInstalled = false

		Expect(plugin.HandleRunnerDeleted(ctx, name, c)).To(Succeed())

		Expect(snapshotJobs(ctx, c)).To(HaveLen(1))
	})
})

// failingDeletes is c with every delete failing with err.
func failingDeletes(c client.Client, err error) client.Client {
	return interceptor.NewClient(c.(client.WithWatch), interceptor.Funcs{
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return err
		},
	})
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
