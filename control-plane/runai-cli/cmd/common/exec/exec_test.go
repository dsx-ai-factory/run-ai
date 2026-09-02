/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var _ = Describe("Exec", func() {
	var (
		cmd         *cobra.Command
		execOptions *Options
	)

	BeforeEach(func() {
		cmd = &cobra.Command{}
		execOptions = &Options{}
	})

	Describe("FlagOutputType", func() {
		It("sets the flags on the command", func() {
			FlagOutputType(cmd, execOptions)
			Expect(cmd.Flags().Lookup(consts.ProjectStr)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Pod)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Container)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.AttributeNamePodRunningTimeout)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.AttributeNameStdin)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.AttributeNameTTY)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.WaitTimeout)).NotTo(BeNil())
		})
	})
	Describe("ValidateExecOptions", func() {

		BeforeEach(func() {
			cmd = &cobra.Command{}
			execOptions = &Options{}
		})

		It("returns an error if workload name is not set", func() {
			err := ValidateExecOptions(execOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.ErrorWorkloadNameIsRequired))
		})

		It("sets the workload name from args if not set in options", func() {
			execOptions.WorkloadName = "job-name"
			err := ValidateExecOptions(execOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(execOptions.WorkloadName).To(Equal("job-name"))
		})

		It("returns nil if all conditions are met", func() {
			execOptions.WorkloadName = "job-name"

			err := ValidateExecOptions(execOptions)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
