/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"testing"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestModels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workload Models Suite")
}

var _ = Describe("WorkloadResourceRefArray", func() {
	It("round-trips a populated array through Value and Scan", func() {
		original := WorkloadResourceRefArray{
			{UID: "uid-1", Name: "data", Kind: string(runv1.StorageRefKindPVC)},
			{UID: "uid-2", Name: "regcred", Kind: string(runv1.StorageRefKindSecret)},
		}

		value, err := original.Value()
		Expect(err).NotTo(HaveOccurred())

		var scanned WorkloadResourceRefArray
		Expect(scanned.Scan([]byte(value.(string)))).To(Succeed())
		Expect(scanned).To(Equal(original))
	})

	It("returns a nil driver value for a nil array", func() {
		var nilArray WorkloadResourceRefArray

		value, err := nilArray.Value()
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeNil())
	})
})

var _ = Describe("WorkloadResourceConnectionsArray", func() {
	It("round-trips per-workload connections through Value and Scan", func() {
		original := WorkloadResourceConnectionsArray{
			{
				WorkloadId: "wl-1",
				ResourceRefs: WorkloadResourceRefArray{
					{UID: "uid-1", Name: "data", Kind: string(runv1.StorageRefKindPVC)},
					{UID: "uid-2", Name: "regcred", Kind: string(runv1.StorageRefKindSecret)},
				},
			},
			{
				WorkloadId: "wl-2",
				ResourceRefs: WorkloadResourceRefArray{
					{UID: "uid-1", Name: "data", Kind: string(runv1.StorageRefKindPVC)},
				},
			},
		}

		value, err := original.Value()
		Expect(err).NotTo(HaveOccurred())

		var scanned WorkloadResourceConnectionsArray
		Expect(scanned.Scan([]byte(value.(string)))).To(Succeed())
		Expect(scanned).To(Equal(original))
	})
})

var _ = Describe("Workload query fields", func() {
	It("maps preemptibility for filtering and sorting", func() {
		workloadView := WorkloadView{}

		Expect(workloadView.FilterFieldsToColumn()).To(HaveKey("preemptibility"))
		Expect(workloadView.SortFieldsToColumn()).To(HaveKey("preemptibility"))
		Expect(Workload{}.AreFilterableFromTable([]string{"preemptibility"})).To(BeTrue())
	})
})
