/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package suspend

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	var (
		instance              *config.Instance
		cmdHelper             helper.Cmd
		completionFactoryMock *mocks.MockFactory
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		instance = config.GetInstance()
		cmdHelper = helper.Cmd{}
		completionFactoryMock = factories.GetCompletionFactoryMock(ctrl)
	})

	Describe("IsSupported", func() {
		var factory *Factory

		BeforeEach(func() {
			factory = NewSuspendFactory(instance, cmdHelper, completionFactoryMock)
		})

		DescribeTable("returns correct support status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("MPI framework", api.MPI, api.Distributed, true),
			Entry("XGBoost framework", api.XGBoost, api.Distributed, true),
			Entry("TF framework", api.TF, api.Distributed, true),
			Entry("PyTorch framework", api.PyTorch, api.Distributed, true),
			Entry("Runai framework with Training workload", api.Runai, api.Training, true),
			Entry("Runai framework with non-Training workload", api.Runai, api.Workspace, true),
			Entry("Unsupported framework", api.UndefinedFramework, api.Distributed, false),
		)
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
				Entry("returns error for Runai framework with Training workload", api.Runai, api.Training),
				Entry("returns error for Runai framework with non-Training workload", api.Runai, api.Workspace),
			)
		})
	})

	Describe("GetCommand", func() {
		Context("SuspendFactory", func() {
			var factory *Factory

			BeforeEach(func() {
				factory = NewSuspendFactory(instance, cmdHelper, completionFactoryMock)
			})

			DescribeTable("should return a command",
				func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
					cmd, err := factory.GetCommand(api.Runai, api.Training)
					Expect(err).NotTo(HaveOccurred())
					Expect(cmd).NotTo(BeNil())
				},
				Entry("MPI framework", api.MPI, api.Distributed, false),
				Entry("XGBoost framework", api.XGBoost, api.Distributed, false),
				Entry("TF framework", api.TF, api.Distributed, false),
				Entry("PyTorch framework", api.PyTorch, api.Distributed, false),
				Entry("Runai framework with Training workload", api.Runai, api.Training, false),
			)

			It("returns an error for unsupported framework", func() {
				_, err := factory.GetCommand(api.UndefinedFramework, api.Training)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
			})

		})

		Context("ResumeFactory", func() {
			var factory *Factory

			BeforeEach(func() {
				factory = NewResumeFactory(instance, cmdHelper, completionFactoryMock)
			})

			DescribeTable("should return a command",
				func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
					cmd, err := factory.GetCommand(api.Runai, api.Training)
					Expect(err).NotTo(HaveOccurred())
					Expect(cmd).NotTo(BeNil())
				},
				Entry("MPI framework", api.MPI, api.Distributed, false),
				Entry("XGBoost framework", api.XGBoost, api.Distributed, false),
				Entry("TF framework", api.TF, api.Distributed, false),
				Entry("PyTorch framework", api.PyTorch, api.Distributed, false),
				Entry("Runai framework with Training workload", api.Runai, api.Training, false),
			)

			It("returns an error for unsupported framework", func() {
				_, err := factory.GetCommand(api.UndefinedFramework, api.Training)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
			})

		})
	})
})

var _ = Describe("Helper Functions", func() {
	Describe("getSuspendCommandTexts", func() {
		DescribeTable("returns correct SuspendText",
			func(framework api.Framework, workloadType api.WorkloadType) {
				texts, err := getSuspendCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(texts).To(BeAssignableToTypeOf(&cmd.BaseSearchableCommandTexts{}))
				text, _ := texts.(*cmd.BaseSearchableCommandTexts)
				Expect(text.GetExample()).To(ContainSubstring(framework.CLIType()))
				Expect(text.GetExample()).To(ContainSubstring(workloadType.CLIType()))
			},
			Entry("MPI framework", api.MPI, api.Distributed),
			Entry("XGBoost framework", api.XGBoost, api.Distributed),
			Entry("TF framework", api.TF, api.Distributed),
			Entry("PyTorch framework", api.PyTorch, api.Distributed),
			Entry("Runai framework with Training workload", api.Runai, api.Training),
		)

		It("returns error for unsupported framework", func() {
			_, err := getSuspendCommandTexts(api.UndefinedFramework, api.Distributed)
			Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Distributed)))
		})

	})

	Describe("getResumeCommandTexts", func() {
		DescribeTable("returns correct SuspendText",
			func(framework api.Framework, workloadType api.WorkloadType) {
				texts, err := getResumeCommandText(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(texts).To(BeAssignableToTypeOf(&cmd.BaseSearchableCommandTexts{}))
				text, _ := texts.(*cmd.BaseSearchableCommandTexts)
				Expect(text.GetExample()).To(ContainSubstring(framework.CLIType()))
				Expect(text.GetExample()).To(ContainSubstring(workloadType.CLIType()))
			},
			Entry("MPI framework", api.MPI, api.Distributed),
			Entry("XGBoost framework", api.XGBoost, api.Distributed),
			Entry("TF framework", api.TF, api.Distributed),
			Entry("PyTorch framework", api.PyTorch, api.Distributed),
			Entry("Runai framework with Training workload", api.Runai, api.Training),
		)

		It("returns error for unsupported framework", func() {
			_, err := getResumeCommandText(api.UndefinedFramework, api.Distributed)
			Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Distributed)))
		})
	})

	Describe("getSuspendFn", func() {
		DescribeTable("returns error for unsupported framework",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := getSuspendFn(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("returns error distributed.Suspend for UndefinedFramework framework", api.UndefinedFramework, api.Distributed),
		)

		DescribeTable("returns suspend fn",
			func(framework api.Framework, workloadType api.WorkloadType) {
				fn, err := getSuspendFn(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(fn).NotTo(BeNil())
			},
			Entry("MPI framework", api.MPI, api.Distributed),
			Entry("XGBoost framework", api.XGBoost, api.Distributed),
			Entry("TF framework", api.TF, api.Distributed),
			Entry("PyTorch framework", api.PyTorch, api.Distributed),
			Entry("Runai framework with Training workload", api.Runai, api.Training),
		)
	})

	Describe("getResumeFn", func() {
		DescribeTable("returns error for unsupported framework",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := getResumeFn(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("returns error distributed.Resume for UndefinedFramework framework", api.UndefinedFramework, api.Distributed),
		)

		DescribeTable("returns resume fn",
			func(framework api.Framework, workloadType api.WorkloadType) {
				fn, err := getResumeFn(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(fn).NotTo(BeNil())
			},
			Entry("MPI framework", api.MPI, api.Distributed),
			Entry("XGBoost framework", api.XGBoost, api.Distributed),
			Entry("TF framework", api.TF, api.Distributed),
			Entry("PyTorch framework", api.PyTorch, api.Distributed),
			Entry("Runai framework with Training workload", api.Runai, api.Training),
		)

		It("returns error for unsupported framework", func() {
			_, err := getResumeFn(api.UndefinedFramework, api.Training)
			Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
		})
	})
})
