/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var _ = Describe("Attach", func() {
	var (
		cmd           *cobra.Command
		attachOptions *Options
	)

	BeforeEach(func() {
		cmd = &cobra.Command{}
		attachOptions = &Options{}
	})

	Describe("FlagOutputType", func() {
		It("sets the flags on the command", func() {
			FlagOutputType(cmd, attachOptions)
			Expect(cmd.Flags().Lookup(consts.ProjectStr)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Pod)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.Container)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.AttributeNamePodRunningTimeout)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.AttributeNameStdin)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.AttributeNameTTY)).NotTo(BeNil())
			Expect(cmd.Flags().Lookup(consts.WaitTimeout)).NotTo(BeNil())
		})
	})
	Describe("validateExecOptions", func() {

		BeforeEach(func() {
			cmd = &cobra.Command{}
			attachOptions = &Options{}
		})

		It("returns an error if workload name is not set", func() {
			err := validateAttachOptions(attachOptions)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(consts.ErrorWorkloadNameIsRequired))
		})

		It("sets the workload name from args if not set in options", func() {
			attachOptions.WorkloadName = "job-name"
			err := validateAttachOptions(attachOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(attachOptions.WorkloadName).To(Equal("job-name"))
		})

		It("returns nil if all conditions are met", func() {
			attachOptions.WorkloadName = "job-name"

			err := validateAttachOptions(attachOptions)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
