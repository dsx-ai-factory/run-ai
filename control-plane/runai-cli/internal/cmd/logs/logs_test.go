/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs_test

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/logs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logs Command", func() {
	var (
		logsCommand *logs.Command
		mpiLogsText *cmd.BaseCommandTexts
		instance    *config.Instance
		cmdHelper   helper.Cmd
	)

	BeforeEach(func() {
		mpiLogsText = logs.NewFrameworkLogsTexts(api.Distributed, api.MPI)
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		logsCommand = logs.NewLogsCommand(api.MPI, cmd.NewEmptyCobraCommand, api.Distributed, mpiLogsText, instance, cmdHelper, factories.GetCompletionFnMock())
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			err := logsCommand.RegisterFlags()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	It("should have auto completion function", func() {
		command := logsCommand.GetCommand()
		Expect(command.ValidArgsFunction).ToNot(BeNil())
	})
})
