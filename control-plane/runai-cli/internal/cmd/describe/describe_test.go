/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var _ = Describe("Command", func() {
	var (
		distributedCommand *describe.Command
		texts              *cmd.BaseCommandTexts
		instance           *config.Instance
		cmdHelper          helper.Cmd
	)

	BeforeEach(func() {
		texts = describe.NewFrameworkTexts(api.Distributed, api.MPI)
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		distributedCommand = describe.NewCommand(api.MPI, api.Distributed, cmd.NewEmptyCobraCommand, texts, instance, cmdHelper, factories.GetCompletionFnMock(), DescribeFnMock)
	})

	Describe("NewDistributedCommand", func() {
		It("should create a new Command with correct properties", func() {
			Expect(distributedCommand.Framework).To(Equal(api.MPI))
			Expect(distributedCommand.GetTxnFn).ToNot(BeNil())
			Expect(distributedCommand.Instance).To(Equal(instance))
			Expect(distributedCommand.CmdHelper).To(Equal(cmdHelper))
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			err := distributedCommand.RegisterFlags()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("Command execution", func() {
		It("should have PreRunE function", func() {
			command := distributedCommand.GetCommand()
			Expect(command.PreRunE).ToNot(BeNil())
		})

		It("should have RunE function", func() {
			command := distributedCommand.GetCommand()
			Expect(command.RunE).ToNot(BeNil())
		})

		It("should have auto completion function", func() {
			command := distributedCommand.GetCommand()
			Expect(command.ValidArgsFunction).ToNot(BeNil())
		})
	})
})

func DescribeFnMock(cmd *cobra.Command, args []string, options workload.DescribeCmdOptions, txn *transaction.Txn) error {
	return nil
}
