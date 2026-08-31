/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/logs"
)

var _ = Describe("Logs Texts", func() {
	var (
		texts        *cmd.BaseCommandTexts
		workloadType api.WorkloadType
		framework    api.Framework
	)

	BeforeEach(func() {
		workloadType = api.Training
		framework = api.Runai
		texts = logs.NewFrameworkLogsTexts(workloadType, framework)
	})

	Describe("GetShort", func() {
		It("should return correct short message for Runai framework", func() {
			expected := "view logs of a standard training workload"
			Expect(texts.GetShort()).To(Equal(expected))
		})

		It("should return correct short message for other frameworks", func() {
			texts = logs.NewFrameworkLogsTexts(workloadType, api.PyTorch)
			expected := "view logs of a pytorch training workload"
			Expect(texts.GetShort()).To(Equal(expected))
			texts = logs.NewFrameworkLogsTexts(workloadType, api.Jax)
			expected = "view logs of a jax training workload"
			Expect(texts.GetShort()).To(Equal(expected))
		})

		It("should handle Distributed workload type correctly", func() {
			texts = logs.NewFrameworkLogsTexts(api.Distributed, framework)
			expected := "view logs of a standard training workload"
			Expect(texts.GetShort()).To(Equal(expected))
		})

		DescribeTable("returns correct short message for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedShort string) {
				texts := createTexts()
				Expect(texts.GetShort()).To(Equal(expectedShort))
			},
			Entry("Inference", logs.NewInferenceLogsTexts, "view logs of an inference workload"),
			Entry("DistributedInference", logs.NewDistributedInferenceLogsTexts, "view logs of a distributed inference workload"),
		)
	})

	Describe("GetExample", func() {
		It("should return correct example message for Runai framework", func() {
			example := texts.GetExample()
			Expect(example).To(ContainSubstring("Get logs for a workload"))
			Expect(example).To(ContainSubstring("runai training standard logs <workload-name>"))
			Expect(example).To(ContainSubstring("--pod=<pod-name>"))
			Expect(example).To(ContainSubstring("--container=<container-name>"))
		})

		It("should return correct example message for other frameworks", func() {
			texts = logs.NewFrameworkLogsTexts(workloadType, api.MPI)
			example := texts.GetExample()
			Expect(example).To(ContainSubstring("Get logs for a workload"))
			Expect(example).To(ContainSubstring("runai training mpi logs <workload-name>"))
			Expect(example).To(ContainSubstring("--pod=<pod-name>"))
			Expect(example).To(ContainSubstring("--container=<container-name>"))
		})

		It("should handle Distributed workload type correctly", func() {
			texts = logs.NewFrameworkLogsTexts(api.Distributed, framework)
			example := texts.GetExample()
			Expect(example).To(ContainSubstring("Get logs for a workload"))
			Expect(example).To(ContainSubstring("--pod=<pod-name>"))
		})

		DescribeTable("returns correct example for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedCommand string) {
				texts := createTexts()
				example := texts.GetExample()
				Expect(example).To(ContainSubstring("Get logs for a workload"))
				Expect(example).To(ContainSubstring(expectedCommand))
			},
			Entry("Inference", logs.NewInferenceLogsTexts, "runai inference logs <workload-name>"),
			Entry("DistributedInference", logs.NewDistributedInferenceLogsTexts, "runai inference distributed logs <workload-name>"),
		)
	})

	Describe("GetBaseErrorMsg", func() {
		It("should return correct base error message for Runai framework", func() {
			expected := "failed to get logs for a standard training workload"
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		})

		It("should return correct base error message for other frameworks", func() {
			texts = logs.NewFrameworkLogsTexts(workloadType, api.TF)
			expected := "failed to get logs for a tf training workload"
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		})

		It("should handle Distributed workload type correctly", func() {
			texts = logs.NewFrameworkLogsTexts(api.Distributed, framework)
			expected := "failed to get logs for a standard training workload"
			Expect(texts.GetBaseErrorMsg()).To(Equal(expected))
		})
	})

	Describe("GetLong and GetAliases", func() {
		It("should return empty string for GetLong for framework workloads", func() {
			Expect(texts.GetLong()).To(BeEmpty())
		})

		DescribeTable("GetLong returns correct description for inference workloads",
			func(createTexts func() *cmd.BaseCommandTexts, expectedLong string) {
				texts := createTexts()
				Expect(texts.GetLong()).To(Equal(expectedLong))
			},
			Entry("Inference", logs.NewInferenceLogsTexts, "If no pod is specified, a pod from the latest ready revision will be used."),
			Entry("DistributedInference", logs.NewDistributedInferenceLogsTexts, "If no pod is specified, a leader pod will be used."),
		)

		It("should return nil for GetAliases", func() {
			Expect(texts.GetAliases()).To(BeNil())
		})
	})

	Describe("Edge cases", func() {
		It("should handle all valid workload types with frameworks", func() {
			workloadTypes := []api.WorkloadType{api.Training, api.Distributed}
			frameworks := []api.Framework{api.Runai, api.MPI, api.PyTorch, api.TF, api.XGBoost}

			for _, wt := range workloadTypes {
				for _, fw := range frameworks {
					texts = logs.NewFrameworkLogsTexts(wt, fw)

					shortMsg := texts.GetShort()
					Expect(shortMsg).To(ContainSubstring(fw.CLIType()))
					Expect(shortMsg).To(ContainSubstring(wt.CLIType()))

					example := texts.GetExample()
					Expect(example).To(ContainSubstring(fmt.Sprintf("runai %s %s", wt.CLIType(), fw.CLIType())))
					Expect(example).To(ContainSubstring("<workload-name>"))

					errorMsg := texts.GetBaseErrorMsg()
					Expect(errorMsg).To(ContainSubstring(fw.CLIType()))
					Expect(errorMsg).To(ContainSubstring(wt.CLIType()))
				}
			}
		})
	})

	Describe("WorkloadOnly Logs", func() {
		var workloadOnlyTexts *cmd.BaseCommandTexts

		BeforeEach(func() {
			workloadOnlyTexts = logs.NewWorkloadOnlyLogsTexts(api.Training)
		})

		It("should return correct short message", func() {
			expected := "view logs of a training workload"
			Expect(workloadOnlyTexts.GetShort()).To(Equal(expected))
		})

		It("should return correct example message", func() {
			example := workloadOnlyTexts.GetExample()
			Expect(example).To(ContainSubstring("Get logs for a workload"))
			Expect(example).To(ContainSubstring("runai training logs <workload-name>"))
			Expect(example).To(ContainSubstring("--pod=<pod-name>"))
			Expect(example).To(ContainSubstring("--container=<container-name>"))
		})

		It("should return correct base error message", func() {
			expected := "failed to get logs for a training workload"
			Expect(workloadOnlyTexts.GetBaseErrorMsg()).To(Equal(expected))
		})

		It("should handle workload types correctly", func() {
			workloadTypes := []api.WorkloadType{api.Training, api.Workspace, api.Distributed}

			for _, wt := range workloadTypes {
				workloadOnlyTexts = logs.NewWorkloadOnlyLogsTexts(wt)

				shortMsg := workloadOnlyTexts.GetShort()
				Expect(shortMsg).To(ContainSubstring(wt.CLIType()))
				Expect(shortMsg).NotTo(ContainSubstring("framework"))

				example := workloadOnlyTexts.GetExample()
				Expect(example).To(ContainSubstring(fmt.Sprintf("runai %s logs", wt.CLIType())))
				Expect(example).To(ContainSubstring("<workload-name>"))

				errorMsg := workloadOnlyTexts.GetBaseErrorMsg()
				Expect(errorMsg).To(ContainSubstring(wt.CLIType()))
				Expect(errorMsg).NotTo(ContainSubstring("framework"))
			}
		})

		It("should return empty long description", func() {
			Expect(workloadOnlyTexts.GetLong()).To(BeEmpty())
		})

		It("should return nil aliases", func() {
			Expect(workloadOnlyTexts.GetAliases()).To(BeNil())
		})

		It("should verify example contains all common options", func() {
			example := workloadOnlyTexts.GetExample()
			commonOptions := []string{
				"--tail=100",
				"--timestamps",
				"--follow",
				"--previous",
				"--limit-bytes=1024",
				"--since=300s",
				"--since-time=2023-05-30T10:00:00Z",
				"--wait-timeout=30s",
			}

			for _, option := range commonOptions {
				Expect(example).To(ContainSubstring(option))
			}
		})
	})
})
