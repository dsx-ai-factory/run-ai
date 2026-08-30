package pf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/pf"
)

var _ = Describe("Port Forward Texts", func() {
	var texts *cmd.BaseCommandTexts

	BeforeEach(func() {
		texts = pf.NewFrameworkTexts(api.Training, api.Runai)
	})

	Describe("GetShort", func() {
		DescribeTable("returns correct message for framework workloads",
			func(workloadType api.WorkloadType, framework api.Framework, expected string) {
				texts = pf.NewFrameworkTexts(workloadType, framework)
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Distributed MPI", api.Distributed, api.MPI, "forward one or more local ports to a mpi training workload"),
			Entry("Training Runai", api.Training, api.Runai, "forward one or more local ports to a standard training workload"),
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "forward one or more local ports to a pytorch training workload"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "forward one or more local ports to a jax training workload"),
			Entry("Distributed TF", api.Distributed, api.TF, "forward one or more local ports to a tf training workload"),
		)

		DescribeTable("returns correct message for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expected string) {
				texts := createTexts()
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Inference", pf.NewInferenceTexts, "forward one or more local ports to an inference workload"),
			Entry("DistributedInference", pf.NewDistributedInferenceTexts, "forward one or more local ports to a distributed inference workload"),
		)
	})

	Describe("GetExample", func() {
		DescribeTable("returns correct example for framework workloads",
			func(workloadType api.WorkloadType, framework api.Framework, expectedSubstring string) {
				texts = pf.NewFrameworkTexts(workloadType, framework)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring(expectedSubstring))
				Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " " + framework.CLIType() + " port-forward"))
				Expect(example).To(ContainSubstring("<workload-name>"))
				Expect(example).To(ContainSubstring("--port 8080:8090"))
				Expect(example).To(ContainSubstring("--address localhost"))
			},
			Entry("Distributed MPI", api.Distributed, api.MPI, "Forward connections from localhost:8080 to a workload on port 8090:"),
			Entry("Training Runai", api.Training, api.Runai, "Forward connections from localhost:8080 to a workload on port 8090:"),
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "Forward connections from localhost:8080 to a workload on port 8090:"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "Forward connections from localhost:8080 to a workload on port 8090:"),
			Entry("Distributed TF", api.Distributed, api.TF, "Forward connections from localhost:8080 to a workload on port 8090:"),
		)

		DescribeTable("returns correct example for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedCommand string) {
				texts := createTexts()
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("Forward connections from localhost:8080 to a workload on port 8090:"))
				Expect(example).To(ContainSubstring(expectedCommand))
			},
			Entry("Inference", pf.NewInferenceTexts, "runai inference port-forward"),
			Entry("DistributedInference", pf.NewDistributedInferenceTexts, "runai inference distributed port-forward"),
		)
	})

	Describe("GetBaseErrorMsg", func() {
		DescribeTable("returns correct error message",
			func(workloadType api.WorkloadType, framework api.Framework, expected string) {
				texts = pf.NewFrameworkTexts(workloadType, framework)
				Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
			},
			Entry("Distributed MPI", api.Distributed, api.MPI, "failed to set up port forwarding for a mpi training workload"),
			Entry("Training Runai", api.Training, api.Runai, "failed to set up port forwarding for a standard training workload"),
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "failed to set up port forwarding for a pytorch training workload"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "failed to set up port forwarding for a jax training workload"),
			Entry("Distributed TF", api.Distributed, api.TF, "failed to set up port forwarding for a tf training workload"),
		)
	})

	Describe("GetSuccessMsg", func() {
		DescribeTable("returns correct success message",
			func(workloadType api.WorkloadType, framework api.Framework, name string) {
				texts = pf.NewFrameworkTexts(workloadType, framework)
				Expect(texts.GetSuccessMsg(name)).To(Equal(""))
			},
			Entry("Distributed MPI", api.Distributed, api.MPI, "job-1"),
			Entry("Training Runai", api.Training, api.Runai, "job-2"),
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "job-3"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "job-4"),
			Entry("Distributed TF", api.Distributed, api.TF, "job-5"),
		)
	})

	Describe("GetLong and GetAliases", func() {
		It("GetLong returns empty string for framework workloads", func() {
			Expect(texts.GetLong()).To(BeEmpty())
		})

		DescribeTable("GetLong returns correct description for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedLong string) {
				texts := createTexts()
				Expect(texts.GetLong()).To(Equal(expectedLong))
			},
			Entry("Inference", pf.NewInferenceTexts, "If no pod is specified, a pod from the latest ready revision will be used."),
			Entry("DistributedInference", pf.NewDistributedInferenceTexts, "If no pod is specified, a leader pod will be used."),
		)

		It("GetAliases returns pf for framework workloads", func() {
			Expect(texts.GetAliases()).To(HaveLen(1))
			Expect(texts.GetAliases()).To(ContainElement("pf"))
		})

		DescribeTable("GetAliases returns pf for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts) {
				texts := createTexts()
				Expect(texts.GetAliases()).To(ContainElement("pf"))
			},
			Entry("Inference", pf.NewInferenceTexts),
			Entry("DistributedInference", pf.NewDistributedInferenceTexts),
		)
	})
	Describe("NewWorkloadOnlyTexts()", func() {
		var commandTexts *cmd.BaseCommandTexts

		BeforeEach(func() {
			commandTexts = pf.NewWorkloadOnlyTexts(api.Workspace)
		})

		DescribeTable("returns correct messages for different workload types",
			func(workloadType api.WorkloadType, expectedShort, expectedError string) {
				commandTexts = pf.NewWorkloadOnlyTexts(workloadType)

				Expect(commandTexts.GetShort()).To(Equal(expectedShort))
				Expect(commandTexts.GetBaseErrorMsg()).To(Equal(expectedError))

				// Verify example contains correct workload type
				example := commandTexts.GetExample()
				Expect(example).To(ContainSubstring("--port 8080:8090"))
				Expect(example).To(ContainSubstring("--address localhost"))

				// Verify empty long description
				Expect(commandTexts.GetLong()).To(BeEmpty())

				// Verify aliases
				Expect(commandTexts.GetAliases()).To(HaveLen(1))
				Expect(commandTexts.GetAliases()).To(ContainElement("pf"))

				// Verify empty success message
				Expect(commandTexts.GetSuccessMsg("test-job")).To(BeEmpty())
			},
			Entry("Training workload",
				api.Training,
				"forward one or more local ports to a training workload",
				"failed to set up port forwarding for a training workload",
			),
			Entry("Interactive workload",
				api.Workspace,
				"forward one or more local ports to a workspace workload",
				"failed to set up port forwarding for a workspace workload",
			),
		)

		Context("token replacement verification", func() {
			BeforeEach(func() {
				commandTexts = pf.NewWorkloadOnlyTexts(api.Training)
			})

			It("should correctly replace workload type token", func() {
				example := commandTexts.GetExample()
				Expect(example).NotTo(ContainSubstring(cmd.WorkloadTypeReplacer))
				Expect(example).NotTo(ContainSubstring(cmd.FrameworkReplacer))
			})

			It("should use frameworkless following message", func() {
				replacementTokens := []cmd.TokenPair{
					{Old: cmd.WorkloadTypeReplacer, New: api.Training.CLIType()},
					{Old: cmd.FollowingReplacer, New: cmd.FollowingFrameworkLess},
				}

				expectedTokens := cmd.ToList(replacementTokens)
				actualExample := commandTexts.GetExample()

				for i := 0; i < len(expectedTokens); i += 2 {
					Expect(actualExample).NotTo(ContainSubstring(expectedTokens[i]))
				}
			})
		})
	})

	Describe("Framework vs Frameworkless comparison", func() {
		It("should have different message formats", func() {
			frameworkText := pf.NewFrameworkTexts(api.Training, api.PyTorch)
			frameworklessText := pf.NewWorkloadOnlyTexts(api.Training)

			// Messages should be different due to framework inclusion
			Expect(frameworkText.GetShort()).NotTo(Equal(frameworklessText.GetShort()))
			Expect(frameworkText.GetExample()).NotTo(Equal(frameworklessText.GetExample()))
			Expect(frameworkText.GetBaseErrorMsg()).NotTo(Equal(frameworklessText.GetBaseErrorMsg()))

			// But should share same aliases
			Expect(frameworkText.GetAliases()).To(Equal(frameworklessText.GetAliases()))
		})
	})

})
