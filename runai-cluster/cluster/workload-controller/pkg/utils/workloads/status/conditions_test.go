/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("ExternalsUrlsBuilder", func() {

	BeforeEach(func() {

	})

	var (
		workload *jobs.JobWorkload
	)

	Context("UpsertCondition", func() {
		Context("When Conditions is empty", func() {
			BeforeEach(func() {
				workload = &jobs.JobWorkload{}
				workload.Status.Conditions = []common.WorkloadCondition{}
			})
			It("should insert condition", func() {
				UpsertCondition(&workload.Status.Conditions, common.Suspended, "True", "Reason", "Message")
				Expect(workload.Status.Conditions).To(HaveLen(1))
				Expect(string(workload.Status.Conditions[0].Type)).To(Equal(common.Suspended))
				Expect(workload.Status.Conditions[0].Status).To(Equal(v1.ConditionTrue))
			})
			It("should update condition", func() {
				UpsertCondition(&workload.Status.Conditions, common.Suspended, "True", "Reason", "Message")
				Expect(workload.Status.Conditions).To(HaveLen(1))
				UpsertCondition(&workload.Status.Conditions, common.Suspended, "False", "Reason", "Message")
				Expect(workload.Status.Conditions).To(HaveLen(1))
				Expect(string(workload.Status.Conditions[0].Type)).To(Equal(common.Suspended))
				Expect(workload.Status.Conditions[0].Status).To(Equal(v1.ConditionFalse))
			})
		})
		Context("When Conditions is not empty", func() {
			BeforeEach(func() {
				workload = &jobs.JobWorkload{}
				workload.Status.Conditions = []common.WorkloadCondition{{
					Type:   common.Suspended,
					Status: "True",
				}}
			})
			It("should insert condition", func() {
				UpsertCondition(&workload.Status.Conditions, common.Failed, "True", "Reason", "Message")
				Expect(workload.Status.Conditions).To(HaveLen(2))
			})
			It("should update condition", func() {
				UpsertCondition(&workload.Status.Conditions, common.Suspended, "False", "Reason", "Message")
				Expect(workload.Status.Conditions).To(HaveLen(1))
				Expect(string(workload.Status.Conditions[0].Type)).To(Equal(common.Suspended))
				Expect(workload.Status.Conditions[0].Status).To(Equal(v1.ConditionFalse))
			})
		})
		Context("When Conditions is nil", func() {
			BeforeEach(func() {
				workload = &jobs.JobWorkload{}
				workload.Status.Conditions = nil
			})
			It("should insert condition", func() {
				UpsertCondition(&workload.Status.Conditions, common.Failed, "True", "Reason", "Message")
				Expect(workload.Status.Conditions).To(HaveLen(1))
			})
		})
	})
	Context("UnsetCondition", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
		})
		Context("When Conditions is empty", func() {
			It("should not change conditions", func() {
				UnsetCondition(&workload.Status.Conditions, common.Suspended)
				Expect(workload.Status.Conditions).To(HaveLen(0))
			})
		})
		Context("When Conditions is not empty", func() {
			BeforeEach(func() {
				workload.Status.Conditions = []common.WorkloadCondition{{
					Type:   common.Suspended,
					Status: "True",
				}, {
					Type:   common.Failed,
					Status: "True",
				}, {
					Type:   common.Ready,
					Status: "True",
				}}

			})
			It("should remove condition", func() {
				lenConditions := len(workload.Status.Conditions)
				UnsetCondition(&workload.Status.Conditions, common.Suspended)
				Expect(workload.Status.Conditions).To(HaveLen(lenConditions - 1))
			})
			It("should not remove other conditions", func() {
				lenConditions := len(workload.Status.Conditions)
				UnsetCondition(&workload.Status.Conditions, common.Initialized)
				Expect(workload.Status.Conditions).To(HaveLen(lenConditions))
			})
		})
	})
})

func BenchmarkUnsetCondition(b *testing.B) {
	conditions := []common.WorkloadCondition{{
		Type:   common.Suspended,
		Status: "True",
	}, {
		Type:   common.Failed,
		Status: "True",
	}, {
		Type:   common.Ready,
		Status: "True",
	}}
	for i := 0; i < b.N; i++ {
		UnsetCondition(&conditions, common.Pending)
		//UpsertCondition(&conditions, common.Ready, "True", "Reason", "Message")
	}
}
