/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pf

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var _ = Describe("PortForwardCommand", func() {
	var (
		portForwardCmd *PortForwardCommand
		mockInstance   *config.Instance
		mockCmdHelper  helper.Cmd
		texts          *cmd.BaseCommandTexts
	)

	BeforeEach(func() {
		mockInstance = config.GetInstance()
		mockCmdHelper = helper.Cmd{}
		texts = NewFrameworkTexts(api.Distributed, api.MPI)
		portForwardCmd = NewPortForwardCommand(api.MPI, cmd.NewEmptyCobraCommand, api.Distributed, texts, mockInstance, mockCmdHelper, factories.GetCompletionFnMock())
	})

	Describe("NewPortForwardCommand", func() {
		It("should create a new PortForwardCommand", func() {
			Expect(portForwardCmd).ToNot(BeNil())
			Expect(portForwardCmd.Framework).To(Equal(api.MPI))
			Expect(portForwardCmd.workloadType).To(Equal(api.Distributed))
		})
	})

	Describe("GetCommand", func() {
		It("should return a cobra.Command with correct properties", func() {
			command := portForwardCmd.GetCommand()
			Expect(command).To(BeAssignableToTypeOf(&cobra.Command{}))
			Expect(command.Use).To(Equal("port-forward [WORKLOAD_NAME]"))
			Expect(command.Short).To(Equal(texts.GetShort()))
			Expect(command.Example).To(Equal(texts.GetExample()))
			Expect(command.Args).To(BeAssignableToTypeOf(cobra.ExactArgs(1)))
			Expect(command.ValidArgsFunction).ToNot(BeNil())
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			err := portForwardCmd.RegisterFlags()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
