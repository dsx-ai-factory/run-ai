package template_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/template"
)

var _ = Describe("Template Describe Command Validation", func() {

	Context("Project validation", func() {
		It("should return error when project is not set", func() {
			// Setup: Empty project
			options := template.DescribeOptions{
				Project: "",
			}

			// Execute validation logic directly (without cmd for testing)
			// Note: In production, PreRunE will use cmdtxn.CmdErrorPrintAndSet
			if options.Project == "" {
				// This simulates the validation behavior
				err := template.DescribeProjectValidatorPre(nil, &options)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("project is missing"))
			}
		})

		It("should succeed when project is set via options", func() {
			// Setup: Project provided via flag
			options := template.DescribeOptions{
				Project: "my-project",
			}

			// Execute validation - should not panic even without full command context
			err := template.DescribeProjectValidatorPre(nil, &options)

			// Verify: Should succeed (project is set)
			Expect(err).To(BeNil())
		})
	})

	Context("Command flags", func() {
		It("should have -p flag defined", func() {
			// Verify flag exists on the command
			describeCmd := template.Describe

			Expect(describeCmd).NotTo(BeNil())
			Expect(describeCmd.Flags().Lookup("project")).NotTo(BeNil())
		})

		It("should allow -p flag", func() {
			// Verify -p flag can be parsed
			describeCmd := template.Describe
			err := describeCmd.Flags().Parse([]string{"-p", "my-project"})
			Expect(err).To(BeNil())
		})
	})
})
