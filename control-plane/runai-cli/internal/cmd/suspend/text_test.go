/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package suspend_test

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/suspend"
)

var _ = Describe("Suspend Command Texts", func() {
	var (
		suspendTexts *cmd.BaseSearchableCommandTexts
		resumeTexts  *cmd.BaseSearchableCommandTexts
	)

	BeforeEach(func() {
		suspendTexts = suspend.NewFrameworkSuspendTexts(api.Training, api.Runai)
		resumeTexts = suspend.NewFrameworkResumeTexts(api.Training, api.Runai)
	})

	Describe("Suspend Texts", func() {
		Describe("GetShort", func() {
			DescribeTable("returns correct message",
				func(workloadType api.WorkloadType, framework api.Framework, expected string) {
					texts := suspend.NewFrameworkSuspendTexts(workloadType, framework)
					Expect(texts.GetShort()).To(Equal(expected))
				},
				Entry("Training Runai", api.Training, api.Runai, "suspend a standard training workload"),
				Entry("Training PyTorch", api.Training, api.PyTorch, "suspend a pytorch training workload"),
				Entry("Distributed MPI", api.Distributed, api.MPI, "suspend a mpi training workload"),
			)
		})

		Describe("GetExample", func() {
			DescribeTable("returns correct example",
				func(workloadType api.WorkloadType, framework api.Framework) {
					texts := suspend.NewFrameworkSuspendTexts(workloadType, framework)
					example := texts.GetExample()

					Expect(example).To(ContainSubstring("Suspend a workload"))
					Expect(example).To(ContainSubstring(fmt.Sprintf("runai %s %s suspend", workloadType.CLIType(), framework.CLIType())))
					Expect(example).To(ContainSubstring(fmt.Sprintf("<%s-name>", framework.CLIType())))
					Expect(example).To(ContainSubstring("--uuid"))
				},
				Entry("Training Runai", api.Training, api.Runai),
				Entry("Training PyTorch", api.Training, api.PyTorch),
				Entry("Distributed MPI", api.Distributed, api.MPI),
			)
		})

		DescribeTable("GetErrorMessages returns correct messages",
			func(workloadType api.WorkloadType, framework api.Framework) {
				texts := suspend.NewFrameworkSuspendTexts(workloadType, framework)

				// Base error
				Expect(texts.GetBaseErrorMsg()).To(Equal("failed to suspend workload"))

				// No results error
				Expect(texts.GetNoResultsErrorMsg()).To(Equal("The requested workload not found"))

				// Searchable error
				Expect(texts.GetSearchableErrorMsg()).To(Equal("failed to search for requested workload"))
			},
			Entry("Training Runai", api.Training, api.Runai),
			Entry("Training PyTorch", api.Training, api.PyTorch),
			Entry("Distributed MPI", api.Distributed, api.MPI),
		)

		DescribeTable("GetSuccessMsg returns correct message",
			func(workloadType api.WorkloadType, framework api.Framework, workloadName string) {
				texts := suspend.NewFrameworkSuspendTexts(workloadType, framework)
				msg := texts.GetSuccessMsg(workloadName)

				Expect(msg).To(ContainSubstring(fmt.Sprintf("Stopping workload '%s'", workloadName)))
				Expect(msg).To(ContainSubstring(fmt.Sprintf("runai %s %s describe %s", workloadType.CLIType(), framework.CLIType(), workloadName)))
			},
			Entry("Training Runai", api.Training, api.Runai, "test-1"),
			Entry("Training PyTorch", api.Training, api.PyTorch, "test-2"),
			Entry("Distributed MPI", api.Distributed, api.MPI, "test-3"),
		)
	})

	Describe("Resume Texts", func() {
		Describe("GetShort", func() {
			DescribeTable("returns correct message",
				func(workloadType api.WorkloadType, framework api.Framework, expected string) {
					texts := suspend.NewFrameworkResumeTexts(workloadType, framework)
					Expect(texts.GetShort()).To(Equal(expected))
				},
				Entry("Training Runai", api.Training, api.Runai, "resume standard training"),
				Entry("Training PyTorch", api.Training, api.PyTorch, "resume pytorch training"),
				Entry("Distributed MPI", api.Distributed, api.MPI, "resume mpi training"),
			)
		})

		Describe("GetExample", func() {
			DescribeTable("returns correct example",
				func(workloadType api.WorkloadType, framework api.Framework) {
					texts := suspend.NewFrameworkResumeTexts(workloadType, framework)
					example := texts.GetExample()

					Expect(example).To(ContainSubstring(fmt.Sprintf("Resume a workload")))
					Expect(example).To(ContainSubstring(fmt.Sprintf("runai %s %s resume", workloadType.CLIType(), framework.CLIType())))
					Expect(example).To(ContainSubstring(fmt.Sprintf("<%s-name>", framework.CLIType())))
					Expect(example).To(ContainSubstring("--uuid"))
				},
				Entry("Training Runai", api.Training, api.Runai),
				Entry("Training PyTorch", api.Training, api.PyTorch),
				Entry("Distributed MPI", api.Distributed, api.MPI),
			)
		})

		DescribeTable("GetErrorMessages returns correct messages",
			func(workloadType api.WorkloadType, framework api.Framework) {
				texts := suspend.NewFrameworkResumeTexts(workloadType, framework)

				// Base error
				Expect(texts.GetBaseErrorMsg()).To(Equal("failed to resume workload"))

				// No results error
				Expect(texts.GetNoResultsErrorMsg()).To(Equal("The requested workload not found"))

				// Searchable error
				Expect(texts.GetSearchableErrorMsg()).To(Equal("failed to search for requested workload"))
			},
			Entry("Training Runai", api.Training, api.Runai),
			Entry("Training PyTorch", api.Training, api.PyTorch),
			Entry("Distributed MPI", api.Distributed, api.MPI),
		)

		DescribeTable("GetSuccessMsg returns correct message",
			func(workloadType api.WorkloadType, framework api.Framework, workloadName string) {
				texts := suspend.NewFrameworkResumeTexts(workloadType, framework)
				msg := texts.GetSuccessMsg(workloadName)

				Expect(msg).To(ContainSubstring(fmt.Sprintf("Resuming workload '%s'", workloadName)))
				Expect(msg).To(ContainSubstring(fmt.Sprintf("runai %s %s describe %s", workloadType.CLIType(), framework.CLIType(), workloadName)))
			},
			Entry("Training Runai", api.Training, api.Runai, "test-1"),
			Entry("Training PyTorch", api.Training, api.PyTorch, "test-2"),
			Entry("Distributed MPI", api.Distributed, api.MPI, "test-3"),
		)
	})

	Describe("Common Methods", func() {
		It("GetLong returns empty string", func() {
			Expect(suspendTexts.GetLong()).To(BeEmpty())
			Expect(resumeTexts.GetLong()).To(BeEmpty())
		})

		It("GetAliases returns nil", func() {
			Expect(suspendTexts.GetAliases()).To(BeNil())
			Expect(resumeTexts.GetAliases()).To(BeNil())
		})
	})

})
