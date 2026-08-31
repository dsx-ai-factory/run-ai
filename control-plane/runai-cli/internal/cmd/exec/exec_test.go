/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/exec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Exec Command", func() {
	var (
		execCommand *exec.Command
		texts       *cmd.BaseCommandTexts
		instance    *config.Instance
		cmdHelper   helper.Cmd
	)

	BeforeEach(func() {
		texts = exec.NewFrameworkExecTexts(api.Distributed, api.MPI)
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		execCommand = exec.NewExecCommand(api.MPI, cmd.NewEmptyCobraCommand, api.Distributed, texts, instance, cmdHelper, factories.GetCompletionFnMock(), nil, nil)
	})

	Describe("NewExecCommand", func() {
		It("should create a new Command with correct properties", func() {
			Expect(execCommand.Framework).To(Equal(api.MPI))
			Expect(execCommand.GetTxnFn).ToNot(BeNil())
			Expect(execCommand.Instance).To(Equal(instance))
			Expect(execCommand.CmdHelper).To(Equal(cmdHelper))
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			err := execCommand.RegisterFlags()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	It("should have auto completion function", func() {
		command := execCommand.GetCommand()
		Expect(command.ValidArgsFunction).ToNot(BeNil())
	})
})
