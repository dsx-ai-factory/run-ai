/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var _ = Describe("Logs", func() {
	var (
		cmd         *cobra.Command
		logsOptions *Options
	)

	BeforeEach(func() {
		cmd = &cobra.Command{}
		logsOptions = &Options{}
	})

	Describe("FlagOutputType", func() {
		It("sets the flags on the command", func() {
			FlagOutputType(cmd, logsOptions)
			Expect(cmd.Flags().Lookup(consts.ProjectStr)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.JobName)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Pod)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Tail)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Container)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Timestamps)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Follow)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Previous)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.LimitBytes)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Since)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.SinceTime)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.WaitTimeout)).NotTo(BeNil())
		})
	})
	Describe("ValidateLogsOptions", func() {

		BeforeEach(func() {
			cmd = &cobra.Command{}
			logsOptions = &Options{}
		})

		It("returns an error if workload name is not set", func() {
			err := ValidateLogsOptions(logsOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.ErrorWorkloadNameIsRequired))
		})

		It("sets the workload name from args if not set in options", func() {
			logsOptions.WorkloadName = "job-name"
			err := ValidateLogsOptions(logsOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(logsOptions.WorkloadName).To(Equal("job-name"))
		})

		It("returns an error if both since time and since seconds are set", func() {
			logsOptions.SinceTime = "2023-05-29T12:00:00Z"
			logsOptions.Since = 300
			logsOptions.WorkloadName = "job-name"

			err := ValidateLogsOptions(logsOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.SinceTimeAndSecondsError))
		})

		It("returns an error if limit bytes is negative", func() {
			logsOptions.LimitBytes = -100
			logsOptions.WorkloadName = "job-name"

			err := ValidateLogsOptions(logsOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.LimitBytesError))
		})

		It("returns an error if since seconds is negative", func() {
			logsOptions.Since = -300
			logsOptions.WorkloadName = "job-name"

			err := ValidateLogsOptions(logsOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.SinceSecondsError))
		})

		It("returns an error if tail lines is less than -1", func() {
			logsOptions.TailLines = -2
			logsOptions.WorkloadName = "job-name"

			err := ValidateLogsOptions(logsOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.TailLinesError))
		})

		It("returns nil if all conditions are met", func() {
			logsOptions.WorkloadName = "job-name"

			err := ValidateLogsOptions(logsOptions)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
