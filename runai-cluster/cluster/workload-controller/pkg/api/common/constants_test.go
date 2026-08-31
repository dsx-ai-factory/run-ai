/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Phase", func() {
	ginkgo.Describe("String()", func() {
		ginkgo.DescribeTable("returns the correct string for each Phase",
			func(phase Phase, expected string) {
				gomega.Expect(phase.String()).To(gomega.Equal(expected))
			},
			ginkgo.Entry("Creating", PhaseCreating, "Creating"),
			ginkgo.Entry("Initializing", PhaseInitializing, "Initializing"),
			ginkgo.Entry("Running", PhaseRunning, "Running"),
			ginkgo.Entry("Degraded", PhaseDegraded, "Degraded"),
			ginkgo.Entry("Pending", PhasePending, "Pending"),
			ginkgo.Entry("Suspended", PhaseSuspended, "Suspended"),
			ginkgo.Entry("Completed", PhaseCompleted, "Completed"),
			ginkgo.Entry("Failed", PhaseFailed, "Failed"),
			ginkgo.Entry("Terminating", PhaseTerminating, "Terminating"),
			ginkgo.Entry("Suspending", PhaseSuspending, "Suspending"),
			ginkgo.Entry("Resuming", PhaseResuming, "Resuming"),
			ginkgo.Entry("Stopped", PhaseStopped, "Stopped"),
			ginkgo.Entry("Restarting", PhaseRestarting, "Restarting"),
			ginkgo.Entry("Unknown (empty)", Phase(""), "Unknown"),
			ginkgo.Entry("Unknown (invalid)", Phase("invalid"), "Unknown"),
		)
	})

	ginkgo.Describe("PhaseStringToEnum", func() {
		ginkgo.DescribeTable("parses string to Phase enum correctly",
			func(input string, expected Phase, shouldError bool) {
				result, err := PhaseStringToEnum(input)
				if shouldError {
					gomega.Expect(err).To(gomega.HaveOccurred())
				} else {
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
					gomega.Expect(result).To(gomega.Equal(expected))
				}
			},
			ginkgo.Entry("Creating", "Creating", PhaseCreating, false),
			ginkgo.Entry("Initializing", "Initializing", PhaseInitializing, false),
			ginkgo.Entry("Running", "Running", PhaseRunning, false),
			ginkgo.Entry("Degraded", "Degraded", PhaseDegraded, false),
			ginkgo.Entry("Pending", "Pending", PhasePending, false),
			ginkgo.Entry("Suspended", "Suspended", PhaseSuspended, false),
			ginkgo.Entry("Completed", "Completed", PhaseCompleted, false),
			ginkgo.Entry("Failed", "Failed", PhaseFailed, false),
			ginkgo.Entry("Terminating", "Terminating", PhaseTerminating, false),
			ginkgo.Entry("Suspending", "Suspending", PhaseSuspending, false),
			ginkgo.Entry("Resuming", "Resuming", PhaseResuming, false),
			ginkgo.Entry("Stopped", "Stopped", PhaseStopped, false),
			ginkgo.Entry("Restarting", "Restarting", PhaseRestarting, false),
			ginkgo.Entry("Unknown", "Unknown", Phase(""), true),
			ginkgo.Entry("empty string", "", Phase(""), true),
			ginkgo.Entry("not-a-phase", "not-a-phase", Phase(""), true),
		)
	})
})
