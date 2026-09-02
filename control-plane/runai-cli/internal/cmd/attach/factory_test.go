/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
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
			Entry("MPI framework", api.MPI, api.Distributed, true),
			Entry("XGBoost framework", api.XGBoost, api.Distributed, true),
			Entry("TF framework", api.TF, api.Distributed, true),
			Entry("PyTorch framework", api.PyTorch, api.Distributed, true),
			Entry("JAX framework", api.Jax, api.Distributed, true),
			Entry("Runai framework with Training workload", api.Runai, api.Training, true),
			Entry("Runai framework with non-Training workload", api.Runai, api.Workspace, true),
			Entry("Unsupported framework", api.UndefinedFramework, api.Training, false),
		)
	})
	Describe("IsAliasSupported", func() {
		DescribeTable("should return correct support alias status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsAliasSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("MPI framework", api.MPI, api.Distributed, false),
			Entry("XGBoost framework", api.XGBoost, api.Distributed, false),
			Entry("TF framework", api.TF, api.Distributed, false),
			Entry("PyTorch framework", api.PyTorch, api.Distributed, false),
			Entry("JAX framework", api.Jax, api.Distributed, false),
			Entry("Runai framework with Training workload", api.Runai, api.Training, false),
			Entry("Runai framework with non-Training workload", api.Runai, api.Workspace, false),
			Entry("Unsupported framework", api.UndefinedFramework, api.Training, false),
		)
	})

	Describe("GetAliasCommand", func() {
		DescribeTable("should return error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.GetAliasCommand(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(framework, workloadType)))
			},
			Entry("returns error for MPI framework", api.MPI, api.Distributed),
			Entry("returns error for XGBoost framework", api.XGBoost, api.Distributed),
			Entry("returns error for TF framework", api.TF, api.Distributed),
			Entry("returns error for PyTorch framework", api.PyTorch, api.Distributed),
			Entry("returns error for JAX framework", api.Jax, api.Distributed),
			Entry("returns error for Runai framework with Training workload", api.Runai, api.Training),
			Entry("returns error for Runai framework with non-Training workload", api.Runai, api.Workspace),
		)
	})

	Describe("GetCommand", func() {
		DescribeTable("with supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				cmd, err := factory.GetCommand(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd).NotTo(BeNil())
			},
			Entry("returns a command for MPI framework", api.MPI, api.Distributed, true),
			Entry("returns a command for XGBoost framework", api.XGBoost, api.Distributed, true),
			Entry("returns a command for TF framework", api.TF, api.Distributed, true),
			Entry("returns a command for PyTorch framework", api.PyTorch, api.Distributed, true),
			Entry("returns a command for Runai framework with Training workload", api.Runai, api.Training, true),
		)

		Context("with unsupported configurations", func() {
			It("returns an error for unsupported framework", func() {
				_, err := factory.GetCommand(api.UndefinedFramework, api.Training)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
			})

			It("returns an error for Runai framework with non-Training workload", func() {
				_, err := factory.GetCommand(api.Runai, api.Distributed)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.Runai, api.Distributed)))
			})
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns texts for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(texts).To(BeAssignableToTypeOf(&cmd.BaseCommandTexts{}))
			},
			Entry("returns MPITexts for MPI framework", api.MPI, api.Distributed),
			Entry("returns TFTexts for TF framework", api.TF, api.Distributed),
			Entry("returns XGBoostTexts for XGBoost framework", api.XGBoost, api.Distributed),
			Entry("returns PyTorchTexts for PyTorch framework", api.PyTorch, api.Distributed),
			Entry("returns TrainingStandardTexts for Runai framework with Training workload", api.Runai, api.Training),
		)

		It("returns error for unsupported framework", func() {
			_, err := factory.getCommandTexts(api.UndefinedFramework, api.Training)
			Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
		})
	})
})
