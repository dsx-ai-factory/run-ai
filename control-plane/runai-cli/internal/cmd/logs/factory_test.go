/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
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
			Entry("MPI Framework with Training", api.MPI, api.Training, false),
			Entry("Runai Framework with Distributed workload", api.Runai, api.Distributed, false),
			Entry("Runai Framework with Workspace workload", api.Runai, api.Workspace, true),
			Entry("Knative Framework with Inference workload", api.Knative, api.Inference, true),
			Entry("UndefinedFramework with Training workload", api.UndefinedFramework, api.Training, false),
			Entry("UndefinedFramework with Workspace workload", api.UndefinedFramework, api.Workspace, false),
			Entry("UndefinedFramework with Inference workload", api.UndefinedFramework, api.Workspace, false),
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
				Entry("UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace),
				Entry("UndefinedFramework with Inference", api.UndefinedFramework, api.Inference),
			)
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(texts).To(BeAssignableToTypeOf(NewFrameworkLogsTexts(workloadType, framework)))
			},
			Entry("returns FrameworkLogsTexts for MPI Framework", api.MPI, api.Distributed),
			Entry("returns FrameworkLogsTexts for TF Framework", api.TF, api.Distributed),
			Entry("returns FrameworkLogsTexts for XGBoost Framework", api.XGBoost, api.Distributed),
			Entry("returns FrameworkLogsTexts for PyTorch Framework", api.PyTorch, api.Distributed),
			Entry("returns FrameworkLogsTexts for JAX Framework", api.Jax, api.Distributed),
			Entry("returns FrameworkLogsTexts for Runai Training", api.Runai, api.Training),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
			Entry("UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace),
			Entry("UndefinedFramework with Inference", api.UndefinedFramework, api.Inference),
		)
	})

	Describe("IsAliasSupported", func() {
		It("always returns false", func() {
			Expect(factory.IsAliasSupported(api.MPI, api.Distributed)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.Runai, api.Training)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.UndefinedFramework, api.Workspace)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.UndefinedFramework, api.Inference)).To(BeFalse())
		})
	})

	Describe("GetAliasCommand", func() {
		DescribeTable("always returns unsupported error",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.GetAliasCommand(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(framework, workloadType)))
			},
			Entry("MPI Framework with Distributed", api.MPI, api.Distributed),
			Entry("Runai Framework with Training", api.Runai, api.Training),
			Entry("UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace),
			Entry("UndefinedFramework with Inference", api.UndefinedFramework, api.Inference),
		)
	})
})
