/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit_test

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/submit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Submit Texts", func() {
	var texts *cmd.BaseCommandTexts
	Context("with Framework", func() {
		BeforeEach(func() {
			texts = submit.NewFrameworkText(api.Training, api.Runai)
		})

		Describe("GetShort", func() {
			DescribeTable("returns correct message",
				func(workloadType api.WorkloadType, framework api.Framework, expected string) {
					texts = submit.NewFrameworkText(workloadType, framework)
					Expect(texts.GetShort()).To(Equal(expected))
				},
				Entry("Training Runai", api.Training, api.Runai, "submit a standard training workload"),
				Entry("Training PyTorch", api.Distributed, api.PyTorch, "submit a pytorch training workload"),
				Entry("Training JAX", api.Distributed, api.Jax, "submit a jax training workload"),
				Entry("Distributed MPI", api.Distributed, api.MPI, "submit a mpi training workload"),
				Entry("Distributed TF", api.Distributed, api.TF, "submit a tf training workload"),
			)
		})

		Describe("GetExample", func() {
			Context("with non-Distributed workload types", func() {
				It("includes Jupyter notebook example", func() {
					texts = submit.NewFrameworkText(api.Training, api.Runai)
					example := texts.GetExample()

					// Base examples
					Expect(example).To(ContainSubstring("Submit a workload"))
					Expect(example).To(ContainSubstring("runai training standard submit"))
					Expect(example).To(ContainSubstring("-- ls -la"))
					Expect(example).To(ContainSubstring("--command -- echo"))

					// Jupyter example
					Expect(example).To(ContainSubstring(" Submit a workload"))
					Expect(example).To(ContainSubstring("jupyter/scipy-notebook"))
					Expect(example).To(ContainSubstring("--gpu-devices-request 1"))
					Expect(example).To(ContainSubstring("--NotebookApp.base_url"))
				})
			})

			Context("with Distributed workload type", func() {
				It("excludes Jupyter notebook example", func() {
					texts = submit.NewFrameworkText(api.Distributed, api.MPI)
					example := texts.GetExample()

					// Base examples should be present
					Expect(example).To(ContainSubstring("Submit a workload"))
					Expect(example).To(ContainSubstring("runai training mpi submit"))
					Expect(example).To(ContainSubstring("-- ls -la"))
					Expect(example).To(ContainSubstring("--command -- echo"))

					// Jupyter example should not be present
					Expect(example).NotTo(ContainSubstring("jupyter notebook"))
					Expect(example).NotTo(ContainSubstring("jupyter/scipy-notebook"))
					Expect(example).NotTo(ContainSubstring("--gpu-devices-request"))
				})
			})
		})

		Describe("GetBaseErrorMsg", func() {
			DescribeTable("returns correct error message",
				func(workloadType api.WorkloadType, framework api.Framework, expected string) {
					texts = submit.NewFrameworkText(workloadType, framework)
					Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
				},
				Entry("Training Runai", api.Training, api.Runai, "a standard training submission failed"),
				Entry("Distributed PyTorch", api.Distributed, api.PyTorch, "a pytorch training submission failed"),
				Entry("Distributed JAX", api.Distributed, api.Jax, "a jax training submission failed"),
				Entry("Distributed MPI", api.Distributed, api.MPI, "a mpi training submission failed"),
				Entry("Distributed TF", api.Distributed, api.TF, "a tf training submission failed"),
			)
		})

		Describe("GetLong and GetAliases", func() {
			It("GetLong returns empty string", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("GetAliases returns nil", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})

		Describe("Framework and Workload Type combinations", func() {
			DescribeTable("handles all valid combinations correctly",
				func(workloadType api.WorkloadType, framework api.Framework) {
					texts = submit.NewFrameworkText(workloadType, framework)

					// Check short message format
					shortMsg := texts.GetShort()
					Expect(shortMsg).To(ContainSubstring(framework.CLIType()))
					Expect(shortMsg).To(ContainSubstring(workloadType.CLIType()))

					// Check example format
					example := texts.GetExample()
					Expect(example).To(ContainSubstring(fmt.Sprintf("runai %s %s", workloadType.CLIType(), framework.CLIType())))
					Expect(example).To(ContainSubstring("Submit a workload"))

					// Check error message format
					errorMsg := texts.GetBaseErrorMsg()
					Expect(errorMsg).To(ContainSubstring(framework.CLIType()))
					Expect(errorMsg).To(ContainSubstring(workloadType.CLIType()))
				},
				Entry("Training Runai", api.Training, api.Runai),
				Entry("Training PyTorch", api.Training, api.PyTorch),
				Entry("Training MPI", api.Training, api.MPI),
				Entry("Training TF", api.Training, api.TF),
				Entry("Training XGBoost", api.Training, api.XGBoost),
				Entry("Distributed Runai", api.Distributed, api.Runai),
				Entry("Distributed PyTorch", api.Distributed, api.PyTorch),
				Entry("Distributed MPI", api.Distributed, api.MPI),
				Entry("Distributed TF", api.Distributed, api.TF),
				Entry("Distributed XGBoost", api.Distributed, api.XGBoost),
			)
		})
	})
	Context("without Framework", func() {
		BeforeEach(func() {
			texts = submit.NewWorkloadOnlyText(api.Workspace)
		})

		Describe("GetShort", func() {
			DescribeTable("returns correct message",
				func(workloadType api.WorkloadType, expected string) {
					texts = submit.NewWorkloadOnlyText(workloadType)
					Expect(texts.GetShort()).To(Equal(expected))
				},
				Entry("Workspace Runai", api.Workspace, "submit a workspace workload"),
				Entry("Inference Knative", api.Inference, "submit an inference workload"),
			)
		})

	})
})
