/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec_test

import (
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

func TestExecTexts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Exec Texts Suite")
}

var _ = Describe("Command Texts", func() {
	Describe("Framework Exec Texts", func() {
		var texts cmd.CommandTexts

		DescribeTable("GetShort method for framework workloads",
			func(workloadType api.WorkloadType, framework api.Framework, expected string) {
				texts = exec.NewFrameworkExecTexts(workloadType, framework)
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "execute a command in a pytorch training workload"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "execute a command in a jax training workload"),
			Entry("Distributed TF", api.Distributed, api.TF, "execute a command in a tf training workload"),
		)

		DescribeTable("GetShort method for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expected string) {
				texts := createTexts()
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Inference", exec.NewInferenceExecTexts, "execute a command in an inference workload"),
			Entry("DistributedInference", exec.NewDistributedInferenceExecTexts, "execute a command in a distributed inference workload"),
		)

		DescribeTable("GetExample method for framework workloads",
			func(workloadType api.WorkloadType, framework api.Framework, expectedSubstring string) {
				texts = exec.NewFrameworkExecTexts(workloadType, framework)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring(expectedSubstring))
				Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " " + framework.CLIType() + " exec"))
				Expect(example).To(ContainSubstring("Execute bash in the"))
				Expect(example).To(ContainSubstring("Execute ls command in the"))
				Expect(example).To(ContainSubstring("Execute a command in a specific"))
			},
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "training pytorch"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "training jax"),
			Entry("Distributed TF", api.Distributed, api.TF, "training tf"),
		)

		DescribeTable("GetExample method for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedCommand string) {
				texts := createTexts()
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("Execute bash in the workload"))
				Expect(example).To(ContainSubstring(expectedCommand))
			},
			Entry("Inference", exec.NewInferenceExecTexts, "runai inference exec"),
			Entry("DistributedInference", exec.NewDistributedInferenceExecTexts, "runai inference distributed exec"),
		)

		DescribeTable("GetBaseErrorMsg method",
			func(workloadType api.WorkloadType, framework api.Framework, expected string) {
				texts = exec.NewFrameworkExecTexts(workloadType, framework)
				Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
			},
			Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "failed to execute command in a pytorch training workload"),
			Entry("Distributed JAX", api.Distributed, api.Jax, "failed to execute command in a jax training workload"),
			Entry("Distributed TF", api.Distributed, api.TF, "failed to execute command in a tf training workload"),
		)

		Context("with any workload type and framework", func() {
			BeforeEach(func() {
				texts = exec.NewFrameworkExecTexts(api.Training, api.PyTorch)
			})

			It("returns empty string for GetLong", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("returns nil for GetAliases", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})

		DescribeTable("GetLong method for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedLong string) {
				texts := createTexts()
				Expect(texts.GetLong()).To(Equal(expectedLong))
			},
			Entry("Inference", exec.NewInferenceExecTexts, "If no pod is specified, a pod from the latest ready revision will be used."),
			Entry("DistributedInference", exec.NewDistributedInferenceExecTexts, "If no pod is specified, a leader pod will be used."),
		)
	})

	Describe("Workload-Only Exec Texts", func() {
		var texts cmd.CommandTexts

		DescribeTable("GetShort method",
			func(workloadType api.WorkloadType, expected string) {
				texts = exec.NewWorkloadOnlyExecTexts(workloadType)
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "execute a command in a workspace workload"),
		)

		DescribeTable("GetExample method",
			func(workloadType api.WorkloadType, expectedSubstring string) {
				texts = exec.NewWorkloadOnlyExecTexts(workloadType)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring(expectedSubstring))
				Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " exec"))
				Expect(example).To(ContainSubstring("Execute bash in the"))
				Expect(example).To(ContainSubstring("Execute ls command in the"))
				Expect(example).To(ContainSubstring("Execute a command in a specific"))
			},
			Entry("Workspace", api.Workspace, "workspace"),
		)

		DescribeTable("GetBaseErrorMsg method",
			func(workloadType api.WorkloadType, expected string) {
				texts = exec.NewWorkloadOnlyExecTexts(workloadType)
				Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "failed to execute command in a workspace workload"),
		)

		Context("with any workload type", func() {
			BeforeEach(func() {
				texts = exec.NewWorkloadOnlyExecTexts(api.Workspace)
			})

			It("returns empty string for GetLong", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("returns nil for GetAliases", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})
	})

	Describe("Framework Bash Texts", func() {
		var texts cmd.CommandTexts

		DescribeTable("GetShort method for framework workloads",
			func(workloadType api.WorkloadType, framework api.Framework, expected string) {
				texts = exec.NewFrameworkBashTexts(workloadType, framework)
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Training PyTorch", api.Training, api.PyTorch, "open a bash shell in a pytorch training workload"),
			Entry("Distributed TF", api.Distributed, api.TF, "open a bash shell in a tf training workload"),
		)

		DescribeTable("GetShort method for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expected string) {
				texts := createTexts()
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Inference", exec.NewInferenceBashTexts, "open a bash shell in an inference workload"),
			Entry("DistributedInference", exec.NewDistributedInferenceBashTexts, "open a bash shell in a distributed inference workload"),
		)

		DescribeTable("GetExample method for framework workloads",
			func(workloadType api.WorkloadType, framework api.Framework, expectedSubstring string) {
				texts = exec.NewFrameworkBashTexts(workloadType, framework)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring(expectedSubstring))
				Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " " + framework.CLIType() + " bash"))
				Expect(example).To(ContainSubstring("Open a bash shell in the"))
				Expect(example).To(ContainSubstring("Open a bash shell in a specific"))
			},
			Entry("Training PyTorch", api.Training, api.PyTorch, "training pytorch"),
			Entry("Distributed TF", api.Distributed, api.TF, "training tf"),
		)

		DescribeTable("GetExample method for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedCommand string) {
				texts := createTexts()
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("Open a bash shell in the workload"))
				Expect(example).To(ContainSubstring(expectedCommand))
			},
			Entry("Inference", exec.NewInferenceBashTexts, "runai inference bash"),
			Entry("DistributedInference", exec.NewDistributedInferenceBashTexts, "runai inference distributed bash"),
		)

		DescribeTable("GetBaseErrorMsg method",
			func(workloadType api.WorkloadType, framework api.Framework, expected string) {
				texts = exec.NewFrameworkBashTexts(workloadType, framework)
				Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
			},
			Entry("Training PyTorch", api.Training, api.PyTorch, "failed to open bash shell in a pytorch training workload"),
			Entry("Distributed TF", api.Distributed, api.TF, "failed to open bash shell in a tf training workload"),
		)

		Context("with any workload type and framework", func() {
			BeforeEach(func() {
				texts = exec.NewFrameworkBashTexts(api.Training, api.PyTorch)
			})

			It("returns empty string for GetLong", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("returns nil for GetAliases", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})

		DescribeTable("GetLong method for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedLong string) {
				texts := createTexts()
				Expect(texts.GetLong()).To(Equal(expectedLong))
			},
			Entry("Inference", exec.NewInferenceBashTexts, "If no pod is specified, a pod from the latest ready revision will be used."),
			Entry("DistributedInference", exec.NewDistributedInferenceBashTexts, "If no pod is specified, a leader pod will be used."),
		)
	})

	Describe("Workload-Only Bash Texts", func() {
		var texts cmd.CommandTexts

		DescribeTable("GetShort method",
			func(workloadType api.WorkloadType, expected string) {
				texts = exec.NewWorkloadOnlyBashTexts(workloadType)
				Expect(texts.GetShort()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "open a bash shell in a workspace workload"),
		)

		DescribeTable("GetExample method",
			func(workloadType api.WorkloadType, expectedSubstring string) {
				texts = exec.NewWorkloadOnlyBashTexts(workloadType)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring(expectedSubstring))
				Expect(example).To(ContainSubstring("runai " + workloadType.CLIType() + " bash"))
				Expect(example).To(ContainSubstring("Open a bash shell in the"))
				Expect(example).To(ContainSubstring("Open a bash shell in a specific"))
			},
			Entry("Workspace", api.Workspace, "workspace"),
		)

		DescribeTable("GetBaseErrorMsg method",
			func(workloadType api.WorkloadType, expected string) {
				texts = exec.NewWorkloadOnlyBashTexts(workloadType)
				Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "failed to open bash shell in a workspace workload"),
		)

		Context("with any workload type", func() {
			BeforeEach(func() {
				texts = exec.NewWorkloadOnlyBashTexts(api.Workspace)
			})

			It("returns empty string for GetLong", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("returns nil for GetAliases", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})
	})
})
