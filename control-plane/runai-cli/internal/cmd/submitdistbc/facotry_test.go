/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submitdistbc

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Submit Factory", func() {
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
			Entry("MPI Framework with Distributed", api.MPI, api.Distributed, true),
			Entry("XGBoost Framework with Distributed", api.XGBoost, api.Distributed, true),
			Entry("TF Framework with Distributed", api.TF, api.Distributed, true),
			Entry("PyTorch Framework with Distributed", api.PyTorch, api.Distributed, true),
			Entry("MPI Framework with Training", api.MPI, api.Training, false),
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
				Entry("returns a command for MPI Framework with Distributed", api.MPI, api.Distributed),
				Entry("returns a command for TF Framework with Distributed", api.TF, api.Distributed),
				Entry("returns a command for XGBoost Framework with Distributed", api.XGBoost, api.Distributed),
				Entry("returns a command for PyTorch Framework with Distributed", api.PyTorch, api.Distributed),
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
				Expect(texts).To(BeAssignableToTypeOf(NewFrameworkText(workloadType, framework)))

				submitTexts, _ := texts.(*cmd.BaseCommandTexts)
				Expect(submitTexts.GetExample()).To(ContainSubstring(framework.CLIType()))
			},
			Entry("returns FrameworkSubmitTexts for MPI Framework with Distributed", api.MPI, api.Distributed),
			Entry("returns FrameworkSubmitTexts for TF Framework with Distributed", api.TF, api.Distributed),
			Entry("returns FrameworkSubmitTexts for XGBoost Framework with Distributed", api.XGBoost, api.Distributed),
			Entry("returns FrameworkSubmitTexts for PyTorch Framework with Distributed", api.PyTorch, api.Distributed),
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
			Expect(factory.IsAliasSupported(api.MPI, api.Distributed)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.Runai, api.Training)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.UndefinedFramework, api.Workspace)).To(BeFalse())
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
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
		)
	})
})
