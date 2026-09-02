/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package delete

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var _ = Describe("Delete", func() {
	var (
		deleteCommand *Command
		texts         *cmd.BaseSearchableCommandTexts
		instance      *config.Instance
		cmdHelper     helper.Cmd
	)

	BeforeEach(func() {
		texts = NewFrameworkTexts(api.Distributed, api.MPI)
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		deleteCommand = NewDeleteCommand(api.MPI, cmd.NewEmptyCobraCommand, texts, instance, cmdHelper, distributed.Delete, distributed.GetByNameAndProj, api.Distributed, factories.GetCompletionFnMock())
	})

	Describe("NewDeleteCommand", func() {
		It("should create a new Command with correct properties", func() {
			Expect(deleteCommand.Framework).To(Equal(api.MPI))
			Expect(deleteCommand.GetTxnFn).ToNot(BeNil())
			Expect(deleteCommand.Instance).To(Equal(instance))
			Expect(deleteCommand.CmdHelper).To(Equal(cmdHelper))
			Expect(deleteCommand.OperationType).To(Equal("delete"))
			Expect(deleteCommand.OperationFunc).To(BeAssignableToTypeOf(distributed.Delete))
			Expect(deleteCommand.GetWorkloadFn).To(BeAssignableToTypeOf(distributed.GetByNameAndProj))
			Expect(deleteCommand.Texts).To(Equal(texts))
		})
	})

	Describe("Inherited SearchBaseCommand properties", func() {
		It("should have the correct inherited properties", func() {

			Expect(deleteCommand.WorkloadType).To(Equal(api.Distributed))
			Expect(deleteCommand.Options).ToNot(BeNil())
		})
	})

	Describe("Command properties", func() {
		It("should have the correct Cobra command properties", func() {
			cobraCmd := deleteCommand.GetCommand()
			Expect(cobraCmd).To(BeAssignableToTypeOf(&cobra.Command{}))
			Expect(cobraCmd.Use).To(Equal("delete [WORKLOAD_NAME]"))
			Expect(cobraCmd.Short).To(Equal(texts.GetShort()))
			Expect(cobraCmd.Args).To(BeAssignableToTypeOf(cobra.MaximumNArgs(1)))
			Expect(cobraCmd.ValidArgsFunction).ToNot(BeNil())
		})
	})

})
