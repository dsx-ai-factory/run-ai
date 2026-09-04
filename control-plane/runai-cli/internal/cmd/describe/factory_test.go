/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

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
			Entry("MPI framework with Distributed", api.MPI, api.Distributed, true),
			Entry("XGBoost framework with Distributed", api.XGBoost, api.Distributed, true),
			Entry("TF framework with Distributed", api.TF, api.Distributed, true),
			Entry("PyTorch framework with Distributed", api.PyTorch, api.Distributed, true),
			Entry("Runai framework with Training workload", api.Runai, api.Training, true),
			Entry("Runai framework with Workspace workload", api.Runai, api.Workspace, true),
			Entry("UndefinedFramework with Workspace workload", api.UndefinedFramework, api.Workspace, false),
			Entry("MPI framework with Training", api.MPI, api.Training, false),
			Entry("Runai framework with Distributed workload", api.Runai, api.Distributed, false),
			Entry("UndefinedFramework with Training workload", api.UndefinedFramework, api.Training, false),
			Entry("Knative framework with inference workload", api.Knative, api.Inference, true),
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
				Entry("returns a command for MPI framework", api.MPI, api.Distributed),
				Entry("returns a command for TF framework", api.TF, api.Distributed),
				Entry("returns a command for XGBoost framework", api.XGBoost, api.Distributed),
				Entry("returns a command for PyTorch framework", api.PyTorch, api.Distributed),
				Entry("returns a command for Runai framework with Training workload", api.Runai, api.Training),
				Entry("returns a command for Runai framework with Workspace workload", api.Runai, api.Workspace),
				Entry("returns a command for Knative framework with Inference workload", api.Knative, api.Inference),
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
				Entry("Runai with Inference", api.Runai, api.Inference),
			)
		})
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
			Entry("Runai framework with Training workload", api.Runai, api.Training, false),
			Entry("Runai framework with Workspace workload", api.Runai, api.Workspace, false),
			Entry("UndefinedFramework with Workspace", api.UndefinedFramework, api.Workspace, false),
			Entry("Knative with Inference", api.Knative, api.Inference, false),
		)
	})

	Describe("GetAliasCommand", func() {
		DescribeTable("should return error for all configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.GetAliasCommand(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(framework, workloadType)))
			},
			Entry("MPI framework", api.MPI, api.Distributed),
			Entry("XGBoost framework", api.XGBoost, api.Distributed),
			Entry("TF framework", api.TF, api.Distributed),
			Entry("PyTorch framework", api.PyTorch, api.Distributed),
			Entry("Runai framework with Training workload", api.Runai, api.Training),
			Entry("Runai framework with Workspace workload", api.Runai, api.Workspace),
			Entry("Knative framework with Inference workload", api.Knative, api.Inference),
		)
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expectWorkloadOnly bool) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				if expectWorkloadOnly {
					Expect(texts).To(BeAssignableToTypeOf(NewWorkloadOnlyTexts(workloadType)))
				} else {
					Expect(texts).To(BeAssignableToTypeOf(NewFrameworkTexts(workloadType, framework)))
				}
			},
			Entry("returns FrameworkTexts for MPI framework", api.MPI, api.Distributed, false),
			Entry("returns FrameworkTexts for TF framework", api.TF, api.Distributed, false),
			Entry("returns FrameworkTexts for XGBoost framework", api.XGBoost, api.Distributed, false),
			Entry("returns FrameworkTexts for PyTorch framework", api.PyTorch, api.Distributed, false),
			Entry("returns FrameworkTexts for Runai framework with Training", api.Runai, api.Training, false),
			Entry("returns WorkloadOnlyTexts for Runai with Workspace", api.Runai, api.Workspace, true),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
			Entry("Runai with Inference", api.Runai, api.Inference),
		)
	})
})
