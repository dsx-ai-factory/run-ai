package attach_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/attach"
)

func TestAttachTexts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Attach Texts Suite")
}

var _ = Describe("FrameworkTexts", func() {
	DescribeTable("GetShort method",
		func(workloadType api.WorkloadType, framework api.Framework, expected string) {
			texts := attach.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetShort()).To(Equal(expected))
		},
		Entry("Training Runai", api.Training, api.Runai, "attach to a running container in a standard training workload"),
		Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "attach to a running container in a pytorch training workload"),
		Entry("Distributed JAX", api.Distributed, api.Jax, "attach to a running container in a jax training workload"),
		Entry("Distributed TF", api.Distributed, api.TF, "attach to a running container in a tf training workload"),
	)

	DescribeTable("GetExample method",
		func(workloadType api.WorkloadType, framework api.Framework, expectedSubstrings ...string) {
			texts := attach.NewFrameworkTexts(workloadType, framework)
			example := texts.GetExample()
			for _, substr := range expectedSubstrings {
				Expect(example).To(ContainSubstring(substr))
			}
			Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " " + framework.CLIType() + " attach"))
		},
		Entry("Distributed PyTorch",
			api.Distributed,
			api.PyTorch,
			"Attaching to the main worker of a workload",
			"runai training pytorch attach <workload-name> --tty --stdin",
			"Attaching to a specific pod of a workload",
			"runai training pytorch attach <workload-name> --pod <pod-name> --tty --stdin"),
		Entry("Distributed JAX",
			api.Distributed,
			api.Jax,
			"Attaching to the main worker of a workload",
			"runai training jax attach <workload-name> --tty --stdin",
			"Attaching to a specific pod of a workload",
			"runai training jax attach <workload-name> --pod <pod-name> --tty --stdin"),
	)

	Describe("GetLong method", func() {
		It("returns an empty string", func() {
			texts := attach.NewFrameworkTexts(api.Distributed, api.PyTorch)
			Expect(texts.GetLong()).To(BeEmpty())
		})
	})

	DescribeTable("GetBaseErrorMsg method",
		func(workloadType api.WorkloadType, framework api.Framework, expected string) {
			texts := attach.NewFrameworkTexts(workloadType, framework)
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		},
		Entry("Training Runai", api.Training, api.Runai, "failed to attach to a standard training workload"),
		Entry("Training PyTorch", api.Distributed, api.PyTorch, "failed to attach to a pytorch training workload"),
		Entry("Training PyTorch", api.Distributed, api.Jax, "failed to attach to a jax training workload"),
		Entry("Distributed TF", api.Distributed, api.TF, "failed to attach to a tf training workload"),
	)

	Describe("GetAliases method", func() {
		It("returns nil", func() {
			texts := attach.NewFrameworkTexts(api.Training, api.PyTorch)
			Expect(texts.GetAliases()).To(BeNil())
		})
	})
})

var _ = Describe("WorkloadOnlyTexts", func() {
	DescribeTable("GetShort method",
		func(workloadType api.WorkloadType, expected string) {
			texts := attach.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetShort()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "attach to a running container in a workspace workload"),
	)

	DescribeTable("GetExample method",
		func(workloadType api.WorkloadType, expectedSubstrings ...string) {
			texts := attach.NewWorkloadOnlyTexts(workloadType)
			example := texts.GetExample()
			for _, substr := range expectedSubstrings {
				Expect(example).To(ContainSubstring(substr))
			}
			Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " attach"))
		},
		Entry("Workspace",
			api.Workspace,
			"Attaching to the main worker of a workload",
			"runai workspace attach <workload-name> --tty --stdin",
			"Attaching to a specific pod of a workload",
			"runai workspace attach <workload-name> --pod <pod-name> --tty --stdin"),
	)

	Describe("GetLong method", func() {
		It("returns an empty string", func() {
			texts := attach.NewWorkloadOnlyTexts(api.Training)
			Expect(texts.GetLong()).To(BeEmpty())
		})
	})

	DescribeTable("GetBaseErrorMsg method",
		func(workloadType api.WorkloadType, expected string) {
			texts := attach.NewWorkloadOnlyTexts(workloadType)
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		},
		Entry("Workspace", api.Workspace, "failed to attach to a workspace workload"),
	)

	Describe("GetAliases method", func() {
		It("returns nil", func() {
			texts := attach.NewWorkloadOnlyTexts(api.Training)
			Expect(texts.GetAliases()).To(BeNil())
		})
	})
})
