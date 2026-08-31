/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var _ = Describe("Factory", func() {
	var (
		factory   *Factory
		instance  *config.Instance
		cmdHelper helper.Cmd
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		factory = NewCommandFactory(instance, cmdHelper, factories.GetCompletionFactoryMock(ctrl))
	})

	Describe("IsSupported", func() {
		DescribeTable("returns correct support status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("MPI Framework with Distributed", api.MPI, api.Distributed, true),
			Entry("XGBoost Framework with Distributed", api.XGBoost, api.Distributed, true),
			Entry("TF Framework with Distributed", api.TF, api.Distributed, true),
			Entry("PyTorch Framework with Distributed", api.PyTorch, api.Distributed, true),
			Entry("JAX Framework with Distributed", api.Jax, api.Distributed, true),
			Entry("Runai Framework with Training workload", api.Runai, api.Training, true),
			Entry("Runai Framework with Workspace workload", api.Runai, api.Workspace, true),
			Entry("UndefinedFramework with Workspace workload", api.UndefinedFramework, api.Workspace, false),
			Entry("MPI Framework with Training", api.MPI, api.Training, false),
			Entry("Runai Framework with Distributed workload", api.Runai, api.Distributed, false),
			Entry("UndefinedFramework with Training workload", api.UndefinedFramework, api.Training, false),
			Entry("Knative Framework with Training workload", api.Knative, api.Training, false),
			Entry("Knative Framework with Workspace workload", api.Knative, api.Workspace, false),
			Entry("Knative Framework with Inference workload", api.Knative, api.Inference, true),
		)
	})

	Describe("GetCommand", func() {
		Context("with supported configurations", func() {
			DescribeTable("returns command successfully",
				func(framework api.Framework, workloadType api.WorkloadType) {
					cmd, err := factory.GetCommand(framework, workloadType)
					Expect(err).NotTo(HaveOccurred())
					Expect(cmd).NotTo(BeNil())
				},
				Entry("returns a command for MPI Framework", api.MPI, api.Distributed),
				Entry("returns a command for TF Framework", api.TF, api.Distributed),
				Entry("returns a command for XGBoost Framework", api.XGBoost, api.Distributed),
				Entry("returns a command for PyTorch Framework", api.PyTorch, api.Distributed),
				Entry("returns a command for JAX Framework", api.Jax, api.Distributed),
				Entry("returns a command for Runai Framework with Training workload", api.Runai, api.Training),
				Entry("returns a command for Runai Framework with Workspace workload", api.Runai, api.Workspace),
				Entry("returns a command for Knative Framework with Inference workload", api.Knative, api.Inference),
			)
		})

		Context("with unsupported configurations", func() {
			DescribeTable("returns appropriate error",
				func(framework api.Framework, workloadType api.WorkloadType) {
					_, err := factory.GetCommand(framework, workloadType)
					Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
				},
				Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
				Entry("MPI with Training", api.MPI, api.Training),
				Entry("Runai with Distributed", api.Runai, api.Distributed),
				Entry("returns a command for UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace),
				Entry("Knative with Distributed", api.Knative, api.Distributed),
				Entry("Knative with Training", api.Knative, api.Training),
			)
		})
	})

	Describe("GetAliasCommand", func() {
		Context("with supported configurations", func() {
			DescribeTable("returns correct command",
				func(framework api.Framework, workloadType api.WorkloadType) {
					cmd, err := factory.GetAliasCommand(framework, workloadType)
					Expect(err).NotTo(HaveOccurred())
					Expect(cmd).NotTo(BeNil())
					Expect(len(cmd)).To(Equal(1))

					command, ok := cmd[0].(*Command)
					Expect(ok).To(BeTrue(), "Expected command to be a BashCommand")
					Expect(command).NotTo(BeNil())
					Expect(command.preFunc).NotTo(BeNil())
					Expect(command.ArgsFunc).To(BeAssignableToTypeOf(cobra.ExactArgs(1)))
					Expect(command.useName).To(Equal(BashUseAlias))
				},
				Entry("returns a command for MPI Framework", api.MPI, api.Distributed),
				Entry("returns a command for TF Framework", api.TF, api.Distributed),
				Entry("returns a command for XGBoost Framework", api.XGBoost, api.Distributed),
				Entry("returns a command for PyTorch Framework", api.PyTorch, api.Distributed),
				Entry("returns a command for JAX Framework", api.Jax, api.Distributed),
				Entry("returns a command for Runai Framework with Training workload", api.Runai, api.Training),
				Entry("returns a command for Runai Framework with Workspace workload", api.Runai, api.Workspace),
				Entry("returns a command for Knative Framework with Inference workload", api.Knative, api.Inference),
			)
		})

		Context("with unsupported configurations", func() {
			DescribeTable("returns appropriate error",
				func(framework api.Framework, workloadType api.WorkloadType) {
					_, err := factory.GetAliasCommand(framework, workloadType)
					Expect(err).To(BeAssignableToTypeOf(cmd.NewUnsupportedAliasError(framework, workloadType)))
				},
				Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
				Entry("MPI with Training", api.MPI, api.Training),
				Entry("Runai with Distributed", api.Runai, api.Distributed),
				Entry("returns a command for UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace),
			)
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expectWorkloadOnly bool) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				if expectWorkloadOnly {
					Expect(texts).To(BeAssignableToTypeOf(NewWorkloadOnlyExecTexts(workloadType)))
				} else {
					Expect(texts).To(BeAssignableToTypeOf(NewFrameworkExecTexts(workloadType, framework)))
				}
			},
			Entry("returns FrameworkExecTexts for MPI Framework", api.MPI, api.Distributed, false),
			Entry("returns FrameworkExecTexts for TF Framework", api.TF, api.Distributed, false),
			Entry("returns FrameworkExecTexts for XGBoost Framework", api.XGBoost, api.Distributed, false),
			Entry("returns FrameworkExecTexts for PyTorch Framework", api.PyTorch, api.Distributed, false),
			Entry("returns FrameworkExecTexts for JAX Framework", api.Jax, api.Distributed, false),
			Entry("returns FrameworkExecTexts for Runai Training", api.Runai, api.Training, false),
			Entry("returns WorkloadOnlyExecTexts for Runai Workspace", api.Runai, api.Workspace, true),
			Entry("returns WorkloadOnlyExecTexts for Knative Inference", api.Knative, api.Inference, true),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
		)
	})

	Describe("getCommandAliasTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expectWorkloadOnly bool) {
				texts, err := factory.getCommandAliasTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				if expectWorkloadOnly {
					Expect(texts).To(BeAssignableToTypeOf(NewWorkloadOnlyBashTexts(workloadType)))
				} else {
					Expect(texts).To(BeAssignableToTypeOf(NewFrameworkBashTexts(workloadType, framework)))
				}
			},
			Entry("returns FrameworkBashTexts for MPI Framework", api.MPI, api.Distributed, false),
			Entry("returns FrameworkBashTexts for TF Framework", api.TF, api.Distributed, false),
			Entry("returns FrameworkBashTexts for XGBoost Framework", api.XGBoost, api.Distributed, false),
			Entry("returns FrameworkBashTexts for PyTorch Framework", api.PyTorch, api.Distributed, false),
			Entry("returns FrameworkBashTexts for JAX Framework", api.Jax, api.Distributed, false),
			Entry("returns FrameworkBashTexts for Runai Training", api.Runai, api.Training, false),
			Entry("returns WorkloadOnlyBashTexts for Runai Workspace", api.Runai, api.Workspace, true),
			Entry("returns WorkloadOnlyBashTexts for Knative Inference", api.Knative, api.Inference, true),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandAliasTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
		)
	})
})
