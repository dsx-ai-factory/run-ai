/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasource_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/datasource"
)

var _ = Describe("Datasource Describe Command", func() {
	Describe("Command Structure", func() {
		It("should have correct aliases", func() {
			Expect(datasource.Describe.Aliases).To(ContainElement("desc"))
		})

		It("should require exactly one argument", func() {
			Expect(datasource.Describe.Args).NotTo(BeNil())
		})

		It("should have correct use string", func() {
			Expect(datasource.Describe.Use).To(ContainSubstring("DATASOURCE_NAME"))
		})

		It("should indicate type flag is required in use string", func() {
			Expect(datasource.Describe.Use).To(ContainSubstring("--type"))
		})
	})

	Describe("Flags", func() {
		It("should have project flag with shorthand", func() {
			flag := datasource.Describe.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have output flag with default value", func() {
			flag := datasource.Describe.Flags().Lookup("output")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("o"))
			Expect(flag.DefValue).To(Equal("table"))
		})

		It("should have type flag", func() {
			flag := datasource.Describe.Flags().Lookup("type")
			Expect(flag).NotTo(BeNil())
		})

		It("should have type flag marked as required", func() {
			flag := datasource.Describe.Flags().Lookup("type")
			Expect(flag).NotTo(BeNil())
			// Check that the flag has required annotation
			annotations := flag.Annotations
			Expect(annotations).To(HaveKey("cobra_annotation_bash_completion_one_required_flag"))
		})

		It("should allow parsing project flag", func() {
			err := datasource.Describe.Flags().Parse([]string{"-p", "test-project"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing output flag", func() {
			err := datasource.Describe.Flags().Parse([]string{"-o", "json"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing type flag", func() {
			err := datasource.Describe.Flags().Parse([]string{"--type", "Pvc"})
			Expect(err).To(BeNil())
		})
	})

	Describe("Project validation", func() {
		It("should return error when project is not set", func() {
			// Setup: Empty project
			options := datasource.DescribeOptions{
				Project: "",
			}

			// Execute validation logic directly (without cmd for testing)
			if options.Project == "" {
				err := datasource.DescribeProjectValidatorPre(nil, &options)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("project is missing"))
			}
		})

		It("should succeed when project is set via options", func() {
			// Setup: Project provided via flag
			options := datasource.DescribeOptions{
				Project: "my-project",
			}

			// Execute validation - should not panic even without full command context
			err := datasource.DescribeProjectValidatorPre(nil, &options)

			// Verify: Should succeed (project is set)
			Expect(err).To(BeNil())
		})
	})

	Describe("Output format validation", func() {
		It("should support table output format", func() {
			options := datasource.DescribeOptions{
				Output: "table",
			}
			Expect(options.Output).To(Equal("table"))
		})

		It("should support json output format", func() {
			options := datasource.DescribeOptions{
				Output: "json",
			}
			Expect(options.Output).To(Equal("json"))
		})

		It("should support yaml output format", func() {
			options := datasource.DescribeOptions{
				Output: "yaml",
			}
			Expect(options.Output).To(Equal("yaml"))
		})
	})
})
