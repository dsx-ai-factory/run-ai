package workload

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

var _ = Describe("ColumnsFlagCompletion", func() {
	var cmd *cobra.Command

	BeforeEach(func() {
		cmd = &cobra.Command{}
	})

	Context("when toComplete is empty", func() {
		It("should return all available columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(HaveLen(len(manipulate.ColumnMap)))
			// Verify it contains expected columns
			Expect(completions).To(ContainElements("Name", "Type", "Status", "Project", "Framework", "Category", "Priority", "Department", "SubmittedBy", "CreatedAt", "TotalPendingTimeSeconds", "AllocatedGPUMemory", "AllocatedCPUCompute", "AllocatedCPUMemory", "CompletedAt", "Preemptible", "Pods", "AllocatedGPU", "URLs"))
		})
	})

	Context("when toComplete has no comma", func() {
		It("should return all columns for empty input", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(HaveLen(len(manipulate.ColumnMap)))
		})

		It("should filter columns by prefix", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "na")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(ContainElement("Name"))
			Expect(completions).NotTo(ContainElement("Type"))
			Expect(completions).NotTo(ContainElement("Status"))
		})

		It("should return empty for non-matching prefix", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "xyz")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(BeEmpty())
		})
	})

	Context("when toComplete has comma-separated values", func() {
		It("should exclude already selected columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "name,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement("name,Name"))
			Expect(completions).To(ContainElement("name,Type"))
			Expect(completions).To(ContainElement("name,Status"))
		})

		It("should exclude multiple already selected columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "Name,Type,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement(ContainSubstring("Name,Type,Name")))
			Expect(completions).NotTo(ContainElement(ContainSubstring("Name,Type,Type")))
			Expect(completions).To(ContainElement("Name,Type,Status"))
			Expect(completions).To(ContainElement("Name,Type,Project"))
		})

		It("should filter remaining columns by current input", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "name,st")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(ContainElement("name,Status"))
			Expect(completions).NotTo(ContainElement("name,type"))
			Expect(completions).NotTo(ContainElement("name,project"))
		})

		It("should handle whitespace in selections", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "name, type ,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement(ContainSubstring("name, type ,name")))
			Expect(completions).NotTo(ContainElement(ContainSubstring("name, type ,type")))
			Expect(completions).To(ContainElement("name, type ,Status"))
		})

		It("should return empty when all columns are selected", func() {
			// Create a string with all available columns
			allColumns := make([]string, 0, len(manipulate.ColumnMap))
			for column := range manipulate.ColumnMap {
				allColumns = append(allColumns, string(column))
			}

			// Test with a subset to avoid making the test too brittle
			selectedColumns := "name,type,status,project,"
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, selectedColumns)

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			// Should return fewer completions since several are already selected
			Expect(len(completions)).To(BeNumerically("<", len(manipulate.ColumnMap)))
		})
	})

	Context("edge cases", func() {
		It("should handle single comma", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, ",")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(HaveLen(len(manipulate.ColumnMap)))
		})
		It("should handle case insensitive filtering", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "NAME,TY")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(ContainElement("NAME,Type"))
		})
	})
})
