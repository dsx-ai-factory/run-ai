/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory_test

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Commands", func() {
	var (
		commands  factory.Commands
		instance  *config.Instance
		cmdHelper helper.Cmd
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		commands = factory.NewCommands(instance, cmdHelper)
	})

	Describe("NewCommands", func() {
		It("should create a new Commands struct with the provided instance and cmdHelper", func() {
			Expect(commands.Instance).To(Equal(instance))
			Expect(commands.CmdHelper).To(Equal(cmdHelper))
		})
	})

	Describe("GetAllSupportedCommands", func() {
		// for distributed we have +1 we have submit and submit-dist
		Context("when all commands are supported for the framework", func() {
			It("should return all supported commands for mpi distributed", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.MPI, api.Distributed)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(12))
			})
			It("should return all supported commands for training runai", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.Runai, api.Training)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(11))
			})
			It("should return all supported commands for workspace runai", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.Runai, api.Workspace)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(11))
			})
			It("should return all supported commands for pytorch distributed", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.PyTorch, api.Distributed)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(12))
			})
			It("should return all supported commands for tf distributed", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.TF, api.Distributed)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(12))
			})
			It("should return all supported commands for xgboost distributed", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.XGBoost, api.Distributed)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(12))
			})

		})

		Context("when some commands are not supported for the framework", func() {
			It("should return only supported commands without error", func() {
				supportedCommands, err := commands.GetAllSupportedCommands(api.AnyDistributed, api.Distributed)
				Expect(err).ToNot(HaveOccurred())
				Expect(supportedCommands).To(HaveLen(0))
			})
		})

	})
})
