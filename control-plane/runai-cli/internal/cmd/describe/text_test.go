package describe_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/describe"
)

var _ = Describe("FrameworkTexts", func() {
	DescribeTable("GetShort method",
		func(workloadType api.WorkloadType, framework api.Framework, expected string) {
			texts := describe.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetShort()).To(Equal(expected))
		},
		Entry("Training PyTorch", api.Training, api.PyTorch, "describe a pytorch training workload"),
		Entry("Distributed TF", api.Distributed, api.TF, "describe a tf training workload"),
	)

	DescribeTable("GetExample method",
		func(workloadType api.WorkloadType, framework api.Framework, expectedSubstrings ...string) {
			texts := describe.NewFrameworkTexts(workloadType, framework)
			example := texts.GetExample()
			for _, substr := range expectedSubstrings {
				Expect(example).To(ContainSubstring(substr))
			}
			Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " " + framework.CLIType() + " describe"))
		},
		Entry("Training PyTorch",
			api.Training,
			api.PyTorch,
			"Describe a workload with a default project",
			"runai training pytorch describe <workload-name>",
			"Describe a workload by UUID",
			"runai training pytorch describe --uuid=<workload-id>"),
		Entry("Distributed TF",
			api.Distributed,
			api.TF,
			"Describe a workload with a default project",
			"runai training tf describe <workload-name>",
			"Describe a workload by UUID",
			"runai training tf describe --uuid=<workload-id>"),
	)

	Describe("GetLong method", func() {
		It("returns an empty string", func() {
			texts := describe.NewFrameworkTexts(api.Training, api.PyTorch)
			Expect(texts.GetLong()).To(BeEmpty())
		})
	})

	DescribeTable("GetBaseErrorMsg method",
		func(workloadType api.WorkloadType, framework api.Framework, expected string) {
			texts := describe.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		},
		Entry("Training PyTorch", api.Training, api.PyTorch, "failed to describe a pytorch training"),
		Entry("Distributed TF", api.Distributed, api.TF, "failed to describe a tf training"),
	)

	Describe("GetAliases method", func() {
		It("returns nil", func() {
			texts := describe.NewFrameworkTexts(api.Training, api.PyTorch)
			Expect(texts.GetAliases()).To(BeNil())
		})
	})
})

var _ = Describe("WorkloadOnlyTexts", func() {
	DescribeTable("GetShort method",
		func(workloadType api.WorkloadType, expected string) {
			texts := describe.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetShort()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "describe a workspace workload"),
		Entry("Inference", api.Inference, "describe an inference workload"),
	)

	DescribeTable("GetExample method",
		func(workloadType api.WorkloadType, expectedSubstrings ...string) {
			texts := describe.NewWorkloadOnlyTexts(workloadType)
			example := texts.GetExample()
			for _, substr := range expectedSubstrings {
				Expect(example).To(ContainSubstring(substr))
			}
			Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " describe"))
		},
		Entry("Workspace",
			api.Workspace,
			"Describe a workload with a default project",
			"runai workspace describe <workload-name>",
			"Describe a workload by UUID",
			"runai workspace describe --uuid=<workload-id>"),
	)

	Describe("GetLong method", func() {
		It("returns an empty string", func() {
			texts := describe.NewWorkloadOnlyTexts(api.Workspace)
			Expect(texts.GetLong()).To(BeEmpty())
		})
	})

	DescribeTable("GetBaseErrorMsg method",
		func(workloadType api.WorkloadType, expected string) {
			texts := describe.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "failed to describe a workspace"),
	)

	Describe("GetAliases method", func() {
		It("returns nil", func() {
			texts := describe.NewWorkloadOnlyTexts(api.Workspace)
			Expect(texts.GetAliases()).To(BeNil())
		})
	})
})
