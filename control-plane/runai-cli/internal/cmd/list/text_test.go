/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/list"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("List Texts", func() {
	Describe("FrameworkListTexts", func() {
		var texts *cmd.BaseCommandTexts

		BeforeEach(func() {
			texts = list.NewFrameworkListTexts(api.Training, api.Runai)
		})

		Describe("GetShort", func() {
			It("should return correct short message for Runai framework", func() {
				Expect(texts.GetShort()).To(Equal("list standard training workloads"))
			})

			It("should return correct short message for other frameworks", func() {
				texts = list.NewFrameworkListTexts(api.Distributed, api.PyTorch)
				Expect(texts.GetShort()).To(Equal("list pytorch training workloads"))
			})

			It("should return correct short message for other frameworks", func() {
				texts = list.NewWorkloadOnlyListTexts(api.Inference)
				Expect(texts.GetShort()).To(Equal("list inference workloads"))
			})
		})

		Describe("GetExample", func() {
			It("should return correct example message for Runai framework", func() {
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("List all standard training workloads"))
				Expect(example).To(ContainSubstring("runai training standard list -A"))
				Expect(example).To(ContainSubstring("List workloads with default project"))
				Expect(example).To(ContainSubstring("List workloads in a specific project"))
				Expect(example).To(ContainSubstring("List workloads with pagination"))
			})

			It("should return correct example message for other frameworks", func() {
				texts = list.NewFrameworkListTexts(api.Training, api.MPI)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("List all mpi training workloads"))
				Expect(example).To(ContainSubstring("runai training mpi list -A"))
			})
		})

		Describe("GetBaseErrorMsg", func() {
			It("should return correct base error message for Runai framework", func() {
				Expect(texts.GetBaseErrorMsg()).To(Equal("failed to list a standard training"))
			})

			It("should return correct base error message for other frameworks", func() {
				texts = list.NewFrameworkListTexts(api.Training, api.TF)
				Expect(texts.GetBaseErrorMsg()).To(Equal("failed to list a tf training"))
			})

			It("should return correct base error message for Inference (knative framework)", func() {
				texts = list.NewWorkloadOnlyListTexts(api.Inference)
				Expect(texts.GetBaseErrorMsg()).To(Equal("failed to list an inference"))
			})
		})

		Describe("GetLong and GetAliases", func() {
			It("should return empty string for GetLong", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("should return nil for GetAliases", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})
	})

	Describe("WorkloadOnlyListTexts", func() {
		var texts *cmd.BaseCommandTexts

		BeforeEach(func() {
			texts = list.NewWorkloadOnlyListTexts(api.Workspace)
		})

		Describe("GetShort", func() {
			It("should return correct short message", func() {
				Expect(texts.GetShort()).To(Equal("list workspace workloads"))
			})
		})

		Describe("GetExample", func() {
			It("should return correct example message", func() {
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("List all workspace workloads"))
				Expect(example).To(ContainSubstring("runai workspace list -A"))
				Expect(example).To(ContainSubstring("List workloads with default project"))
				Expect(example).To(ContainSubstring("List workloads in a specific project"))
				Expect(example).To(ContainSubstring("List all workloads with a specific output format"))
				Expect(example).To(ContainSubstring("List workloads with pagination"))
			})

			It("should return correct example message for Inference (knative framework)", func() {
				texts := list.NewWorkloadOnlyListTexts(api.Inference)
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("List all inference workloads"))
				Expect(example).To(ContainSubstring("runai inference list -A"))
				Expect(example).To(ContainSubstring("List workloads with default project"))
				Expect(example).To(ContainSubstring("List workloads in a specific project"))
				Expect(example).To(ContainSubstring("List workloads with pagination"))
			})
		})

		Describe("GetBaseErrorMsg", func() {
			It("should return correct base error message", func() {
				Expect(texts.GetBaseErrorMsg()).To(Equal("failed to list a workspace"))
			})
		})

		Describe("GetLong and GetAliases", func() {
			It("should return empty string for GetLong", func() {
				Expect(texts.GetLong()).To(BeEmpty())
			})

			It("should return nil for GetAliases", func() {
				Expect(texts.GetAliases()).To(BeNil())
			})
		})
	})

	Describe("Edge cases", func() {
		Context("FrameworkListTexts", func() {
			It("should handle all framework combinations", func() {
				workloadTypes := []api.WorkloadType{api.Workspace, api.Training, api.Distributed}
				frameworks := []api.Framework{api.Runai, api.MPI, api.PyTorch, api.TF, api.XGBoost}

				for _, wt := range workloadTypes {
					for _, fw := range frameworks {
						texts := list.NewFrameworkListTexts(wt, fw)
						Expect(texts.GetShort()).To(ContainSubstring(fw.CLIType()))
						Expect(texts.GetShort()).To(ContainSubstring(wt.CLIType()))
						Expect(texts.GetExample()).To(ContainSubstring(fw.CLIType()))
						Expect(texts.GetExample()).To(ContainSubstring(wt.CLIType()))
						Expect(texts.GetBaseErrorMsg()).To(ContainSubstring(fw.CLIType()))
						Expect(texts.GetBaseErrorMsg()).To(ContainSubstring(wt.CLIType()))
					}
				}
			})
		})

		Context("WorkloadOnlyListTexts", func() {
			It("should handle all workload types", func() {
				workloadTypes := []api.WorkloadType{api.Workspace, api.Training, api.Distributed, api.Inference}

				for _, wt := range workloadTypes {
					texts := list.NewWorkloadOnlyListTexts(wt)
					Expect(texts.GetShort()).To(ContainSubstring(wt.CLIType()))
					Expect(texts.GetExample()).To(ContainSubstring(wt.CLIType()))
					Expect(texts.GetBaseErrorMsg()).To(ContainSubstring(wt.CLIType()))
				}
			})
		})
	})
})
