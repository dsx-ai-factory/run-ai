/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nvidia

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/karta/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("SnapshotJob Karta", func() {
	It("passes Karta validation", func() {
		karta := kartav1alpha1.Karta(BuildSnapshotJobV1alpha1())
		Expect(kartav1alpha1.NewKartaValidator(&karta).Validate()).To(Succeed())
	})

	// Exactly one status per state: EWI resolves overlaps by its own precedence,
	// which would hide a failure behind Running.
	DescribeTable("maps each capture state to exactly one status",
		func(conditions []any, expected kartav1alpha1.ResourceStatus) {
			Expect(snapshotJobStatus(conditions).MatchedStatuses).To(ConsistOf(expected))
		},
		Entry("created, before Snapshot's first status write",
			nil, kartav1alpha1.InitializingStatus),
		Entry("waiting for the source pod to become ready",
			snapshotJobConditions("False", "False", "False", "False"), kartav1alpha1.InitializingStatus),
		Entry("pod ready, capture in progress",
			snapshotJobConditions("True", "False", "False", "False"), kartav1alpha1.RunningStatus),
		Entry("captured, pod still ready",
			snapshotJobConditions("True", "True", "False", "False"), kartav1alpha1.RunningStatus),
		Entry("captured, source Job still finishing",
			snapshotJobConditions("False", "True", "False", "False"), kartav1alpha1.RunningStatus),
		Entry("completed",
			snapshotJobConditions("False", "True", "True", "False"), kartav1alpha1.CompletedStatus),
		Entry("failed before the pod was ready",
			snapshotJobConditions("False", "False", "False", "True"), kartav1alpha1.FailedStatus),
		Entry("failed with the pod ready",
			snapshotJobConditions("True", "False", "False", "True"), kartav1alpha1.FailedStatus),
		Entry("failed after the dump",
			snapshotJobConditions("False", "True", "False", "True"), kartav1alpha1.FailedStatus),
	)

	It("carries Snapshot's failure reason and message", func() {
		conditions := snapshotJobConditions("False", "False", "False", "True")
		// As read off a live cluster after a SnapshotJob was recreated under the same name.
		conditions[3] = map[string]any{
			"type":    "Failed",
			"status":  "True",
			"reason":  "PodSnapshotNameConflict",
			"message": `existing PodSnapshot is not owned by this SnapshotJob: PodSnapshot "runai-snap-x" belongs to a different SnapshotJob incarnation`,
		}

		Expect(snapshotJobStatus(conditions).Conditions).To(ContainElement(SatisfyAll(
			HaveField("Type", "Failed"),
			HaveField("Reason", HaveValue(Equal("PodSnapshotNameConflict"))),
			HaveField("Message", ContainSubstring("not owned by this SnapshotJob")),
		)))
	})
})

// snapshotJobConditions builds the four conditions Snapshot writes on every SnapshotJob.
func snapshotJobConditions(running, captured, completed, failed string) []any {
	return []any{
		map[string]any{"type": "Running", "status": running},
		map[string]any{"type": "Captured", "status": captured},
		map[string]any{"type": "Completed", "status": completed},
		map[string]any{"type": "Failed", "status": failed},
	}
}

func snapshotJobStatus(conditions []any) *resource.Status {
	GinkgoHelper()
	object := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "nvidia.com/v1alpha1",
		"kind":       "SnapshotJob",
		"metadata":   map[string]any{"name": "runai-snap-47e98b07e8dd", "namespace": "runai-proj"},
	}}
	if conditions != nil {
		object.Object["status"] = map[string]any{"conditions": conditions}
	}
	definition := kartav1alpha1.Karta(BuildSnapshotJobV1alpha1())
	root, err := resource.NewComponentFactoryFromObject(&definition, object).GetRootComponent()
	Expect(err).NotTo(HaveOccurred())
	status, err := root.GetStatus(context.Background())
	Expect(err).NotTo(HaveOccurred())
	Expect(status).NotTo(BeNil())
	return status
}
