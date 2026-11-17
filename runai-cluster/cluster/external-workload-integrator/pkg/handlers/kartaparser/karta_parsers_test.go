/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kartaparser

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

var _ = Describe("TestRidHandler", Ordered, func() {
	Describe("Other funcs tests", func() {
		It("sumReplicas doesn't fail on nil", func() {
			sumReplicas(nil, nil)
			// if we got here, it means it didn't panic
			sumReplicas(nil, ptr.To(int32(5)))

			source := ptr.To(int32(5))
			sumReplicas(&source, nil)
			Expect(*source).To(BeEquivalentTo(int32(5)), "Expected source to remain unchanged when adding nil")
			sumReplicas(&source, ptr.To(int32(2)))
			Expect(*source).To(BeEquivalentTo(int32(7)), "Expected source to be 5 + 2 = 7")
		})
	})

	Describe("mapResourceStatusToEwPhase", func() {
		DescribeTable("maps matched Karta statuses to an ExternalWorkload phase",
			func(statuses []kartav1alpha1.ResourceStatus, expected v1alpha1.ExternalWorkloadPhase) {
				Expect((&EWConverter{}).mapResourceStatusToEwPhase(statuses)).To(Equal(expected))
			},
			Entry("no matched status", nil, v1alpha1.ExWUnknownPhase),
			Entry("undefined", []kartav1alpha1.ResourceStatus{kartav1alpha1.UndefinedStatus}, v1alpha1.ExWUnknownPhase),
			Entry("initializing", []kartav1alpha1.ResourceStatus{kartav1alpha1.InitializingStatus}, v1alpha1.ExWInitializingPhase),
			Entry("running", []kartav1alpha1.ResourceStatus{kartav1alpha1.RunningStatus}, v1alpha1.ExWRunningPhase),
			Entry("completed", []kartav1alpha1.ResourceStatus{kartav1alpha1.CompletedStatus}, v1alpha1.ExWCompletedPhase),
			Entry("failed", []kartav1alpha1.ResourceStatus{kartav1alpha1.FailedStatus}, v1alpha1.ExWFailedPhase),
			Entry("degraded", []kartav1alpha1.ResourceStatus{kartav1alpha1.DegradedStatus}, v1alpha1.ExWDegradedPhase),
			Entry("suspending", []kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendingStatus}, v1alpha1.ExWStoppingPhase),
			Entry("suspended", []kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendedStatus}, v1alpha1.ExWStoppedPhase),
			Entry("resuming", []kartav1alpha1.ResourceStatus{kartav1alpha1.ResumingStatus}, v1alpha1.ExWInitializingPhase),

			Entry("running keeps precedence over suspending",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendingStatus, kartav1alpha1.RunningStatus}, v1alpha1.ExWRunningPhase),
			Entry("failed keeps precedence over suspended",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendedStatus, kartav1alpha1.FailedStatus}, v1alpha1.ExWFailedPhase),
			Entry("completed keeps precedence over suspended",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendedStatus, kartav1alpha1.CompletedStatus}, v1alpha1.ExWCompletedPhase),
			Entry("running keeps precedence over degraded",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.DegradedStatus, kartav1alpha1.RunningStatus}, v1alpha1.ExWRunningPhase),
			Entry("running keeps precedence over failed",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.FailedStatus, kartav1alpha1.RunningStatus}, v1alpha1.ExWRunningPhase),
			Entry("resuming wins over suspended",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendedStatus, kartav1alpha1.ResumingStatus}, v1alpha1.ExWInitializingPhase),
			Entry("suspending wins over suspended",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.SuspendedStatus, kartav1alpha1.SuspendingStatus}, v1alpha1.ExWStoppingPhase),
			Entry("suspended wins over degraded",
				[]kartav1alpha1.ResourceStatus{kartav1alpha1.DegradedStatus, kartav1alpha1.SuspendedStatus}, v1alpha1.ExWStoppedPhase),
		)
	})
})
