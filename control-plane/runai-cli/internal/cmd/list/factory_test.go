/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	var (
		factory   *Factory
		instance  *config.Instance
		cmdHelper helper.Cmd
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		factory = NewCommandFactory(instance, cmdHelper)
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
			Entry("MPI Framework with Training", api.MPI, api.Training, false),
			Entry("Runai Framework with Distributed workload", api.Runai, api.Distributed, false),
			Entry("UndefinedFramework with Training workload", api.UndefinedFramework, api.Training, false),
			Entry("UndefinedFramework with Workspace workload", api.UndefinedFramework, api.Workspace, false),
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
				Entry("returns a command for Inference workload", api.Knative, api.Inference),
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
			)
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expectWorkloadOnly bool) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				if expectWorkloadOnly {
					Expect(texts).To(BeAssignableToTypeOf(NewWorkloadOnlyListTexts(workloadType)))
				} else {
					Expect(texts).To(BeAssignableToTypeOf(NewFrameworkListTexts(workloadType, framework)))
				}
			},
			Entry("returns FrameworkListTexts for MPI Framework", api.MPI, api.Distributed, false),
			Entry("returns FrameworkListTexts for TF Framework", api.TF, api.Distributed, false),
			Entry("returns FrameworkListTexts for XGBoost Framework", api.XGBoost, api.Distributed, false),
			Entry("returns FrameworkListTexts for PyTorch Framework", api.PyTorch, api.Distributed, false),
			Entry("returns FrameworkListTexts for JAX Framework", api.Jax, api.Distributed, false),
			Entry("returns FrameworkListTexts for Runai Training", api.Runai, api.Training, false),
			Entry("returns WorkloadOnlyListTexts for Runai Workspace", api.Runai, api.Workspace, true),
			Entry("returns WorkloadOnlyListTexts for inference workloads", api.Knative, api.Inference, true),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
			Entry("Knative with Distributed", api.Knative, api.Distributed),
			Entry("Knative with Training", api.Knative, api.Training),
			Entry("UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace),
		)
	})

	Describe("IsAliasSupported", func() {
		It("always returns false", func() {
			Expect(factory.IsAliasSupported(api.MPI, api.Distributed)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.Runai, api.Training)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.UndefinedFramework, api.Workspace)).To(BeFalse())
		})
	})

	Describe("GetAliasCommand", func() {
		It("always returns unsupported error", func() {
			_, err := factory.GetAliasCommand(api.MPI, api.Distributed)

			Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(api.MPI, api.Distributed)))

			_, err = factory.GetAliasCommand(api.Runai, api.Training)
			Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(api.Runai, api.Training)))

			_, err = factory.GetAliasCommand(api.Knative, api.Inference)
			Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(api.Knative, api.Inference)))
		})
	})
})
