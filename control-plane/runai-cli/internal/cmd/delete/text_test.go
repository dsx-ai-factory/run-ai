package delete_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/delete"
)

var _ = Describe("FrameworkTexts", func() {
	DescribeTable("GetShort method",
		func(workloadType api.WorkloadType, framework api.Framework, expected string) {
			texts := delete.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetShort()).To(Equal(expected))
		},
		Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "delete a pytorch training workload"),
		Entry("Distributed JAX", api.Distributed, api.Jax, "delete a jax training workload"),
		Entry("Distributed TF", api.Distributed, api.TF, "delete a tf training workload"),
	)

	DescribeTable("GetExample method",
		func(workloadType api.WorkloadType, framework api.Framework, expectedSubstrings ...string) {
			texts := delete.NewFrameworkTexts(workloadType, framework)
			example := texts.GetExample()
			for _, substr := range expectedSubstrings {
				Expect(example).To(ContainSubstring(substr))
			}
			Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " " + framework.CLIType() + " delete"))
		},
		Entry("Distributed PyTorch",
			api.Distributed,
			api.PyTorch,
			"Delete a workload with a default project",
			"runai training pytorch delete <workload-name>",
			"Delete a workload with a specific project",
			"Delete a workload by UUID",
			"runai training pytorch delete --uuid=<workload-id> -p <project-name>"),
		Entry("Distributed JAX",
			api.Distributed,
			api.Jax,
			"Delete a workload with a default project",
			"runai training jax delete <workload-name>",
			"Delete a workload with a specific project",
			"Delete a workload by UUID",
			"runai training jax delete --uuid=<workload-id> -p <project-name>"),
	)

	Describe("GetLong method", func() {
		It("returns an empty string", func() {
			texts := delete.NewFrameworkTexts(api.Training, api.PyTorch)
			Expect(texts.GetLong()).To(BeEmpty())
		})
	})

	DescribeTable("GetBaseErrorMsg method",
		func(workloadType api.WorkloadType, framework api.Framework) {
			texts := delete.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetBaseErrorMsg()).To(Equal("failed to delete workload"))
		},
		Entry("Distributed PyTorch", api.Distributed, api.PyTorch),
		Entry("Distributed JAX", api.Distributed, api.Jax),
		Entry("Distributed TF", api.Distributed, api.TF),
	)

	Describe("GetAliases method", func() {
		It("returns [del]", func() {
			texts := delete.NewFrameworkTexts(api.Training, api.PyTorch)
			Expect(texts.GetAliases()).To(Equal([]string{"del"}))
		})
	})

	DescribeTable("GetNoResultsErrorMsg method",
		func(workloadType api.WorkloadType, framework api.Framework) {
			texts := delete.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetNoResultsErrorMsg()).To(Equal("The requested workload not found"))
		},
		Entry("Training PyTorch", api.Training, api.PyTorch),
		Entry("Distributed TF", api.Distributed, api.TF),
	)

	DescribeTable("GetSearchableErrorMsg method",
		func(workloadType api.WorkloadType, framework api.Framework) {
			texts := delete.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetSearchableErrorMsg()).To(Equal("failed to search for requested workload"))
		},
		Entry("Training PyTorch", api.Training, api.PyTorch),
		Entry("Distributed TF", api.Distributed, api.TF),
	)

	DescribeTable("GetSuccessMsg method",
		func(workloadType api.WorkloadType, framework api.Framework, workloadName, expectedSubstring string) {
			texts := delete.NewFrameworkTexts(workloadType, framework)
			successMsg := texts.GetSuccessMsg(workloadName)
			Expect(successMsg).To(ContainSubstring(expectedSubstring))
			Expect(successMsg).To(ContainSubstring(workloadName))
		},
		Entry("Training PyTorch", api.Training, api.PyTorch, "pytorch-job", "Deleting workload 'pytorch-job'"),
		Entry("Distributed TF", api.Distributed, api.TF, "tf-job", "Deleting workload 'tf-job'"),
	)
})

var _ = Describe("WorkloadOnlyTexts", func() {
	DescribeTable("GetShort method",
		func(workloadType api.WorkloadType, expected string) {
			texts := delete.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetShort()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "delete a workspace workload"),
		Entry("Inference", api.Inference, "delete an inference workload"),
	)

	DescribeTable("GetExample method",
		func(workloadType api.WorkloadType, expectedSubstrings ...string) {
			texts := delete.NewWorkloadOnlyTexts(workloadType)
			example := texts.GetExample()
			for _, substr := range expectedSubstrings {
				Expect(example).To(ContainSubstring(substr))
			}
			Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " delete"))
		},
		Entry("Workspace",
			api.Workspace,
			"Delete a workload with a default project",
			"runai workspace delete <workload-name>",
			"Delete a workload with a specific project",
			"Delete a workload by UUID",
			"runai workspace delete --uuid=<workload-id> -p <project-name>"),
	)

	Describe("GetLong method", func() {
		It("returns an empty string", func() {
			texts := delete.NewWorkloadOnlyTexts(api.Workspace)
			Expect(texts.GetLong()).To(BeEmpty())
		})
	})

	DescribeTable("GetBaseErrorMsg method",
		func(workloadType api.WorkloadType, expected string) {
			texts := delete.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "failed to delete workload"),
	)

	Describe("GetAliases method", func() {
		It("returns [del]", func() {
			texts := delete.NewWorkloadOnlyTexts(api.Workspace)
			Expect(texts.GetAliases()).To(Equal([]string{"del"}))
		})
	})

	DescribeTable("GetNoResultsErrorMsg method",
		func(workloadType api.WorkloadType, expected string) {
			texts := delete.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetNoResultsErrorMsg()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "The requested workload not found"),
	)

	DescribeTable("GetSearchableErrorMsg method",
		func(workloadType api.WorkloadType, expected string) {
			texts := delete.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetSearchableErrorMsg()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "failed to search for requested workload"),
	)

	DescribeTable("GetSuccessMsg method",
		func(workloadType api.WorkloadType, workloadName, expectedSubstring string) {
			texts := delete.NewWorkloadOnlyTexts(workloadType)
			successMsg := texts.GetSuccessMsg(workloadName)
			Expect(successMsg).To(ContainSubstring(expectedSubstring))
			Expect(successMsg).To(ContainSubstring(workloadName))
		},
		Entry("Workspace", api.Workspace, "test-workload", "Deleting workload 'test-workload'"),
	)
})
