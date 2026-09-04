/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Update Factory", func() {
	var (
		factory   *Factory
		instance  *config.Instance
		cmdHelper helper.Cmd
	)

	BeforeEach(func() {
		instance = config.GetInstance()
		cmdHelper = helper.Cmd{}
		factory = NewCommandFactory(instance, cmdHelper)
	})

	Describe("IsSupported", func() {
		DescribeTable("returns correct support status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("Knative Framework with Inference", api.Knative, api.Inference, true),
			Entry("MPI Framework with Distributed", api.MPI, api.Distributed, false),
			Entry("XGBoost Framework with Distributed", api.XGBoost, api.Distributed, false),
			Entry("TF Framework with Distributed", api.TF, api.Distributed, false),
			Entry("PyTorch Framework with Distributed", api.PyTorch, api.Distributed, false),
			Entry("Runai Framework with Training workload", api.Runai, api.Training, false),
			Entry("MPI Framework with Training", api.MPI, api.Training, false),
			Entry("Runai Framework with Workspace", api.Runai, api.Workspace, false),
			Entry("Runai Framework with Distributed", api.Runai, api.Distributed, false),
			Entry("Unsupported distributed framework", api.AnyDistributed, api.Distributed, false),
			Entry("UndefinedFramework", api.UndefinedFramework, api.Training, false),
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
				Entry("returns a command for Knative Framework with Inference workload", api.Knative, api.Inference),
			)
		})

		Context("with unsupported configurations", func() {
			DescribeTable("returns appropriate error",
				func(framework api.Framework, workloadType api.WorkloadType) {
					_, err := factory.GetCommand(framework, workloadType)
					Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
				},
				Entry("MPI with Training", api.MPI, api.Training),
				Entry("Runai with Distributed", api.Runai, api.Distributed),
				Entry("UndefinedFramework", api.UndefinedFramework, api.Training),
			)
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(texts).To(BeAssignableToTypeOf(NewWorkloadOnlyText(workloadType)))

				updateTexts, _ := texts.(*cmd.BaseSearchableCommandTexts)
				Expect(updateTexts.GetExample()).To(ContainSubstring(workloadType.CLIType()))
			},
			Entry("returns UpdateTexts for Knative Framework with Inference", api.Knative, api.Inference),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
			Entry("UndefinedFramework", api.UndefinedFramework, api.Training),
			Entry("Unsupported distributed framework", api.AnyDistributed, api.Distributed),
		)
	})

	Describe("IsAliasSupported", func() {
		It("always returns false", func() {
			Expect(factory.IsAliasSupported(api.Knative, api.Inference)).To(BeFalse())
		})
	})

	Describe("GetAliasCommand", func() {
		DescribeTable("always returns unsupported error",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.GetAliasCommand(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(framework, workloadType)))
			},
			Entry("Knative Framework with Inference", api.Knative, api.Inference),
		)
	})
})
