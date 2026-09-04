/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submitdistbc

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command", func() {
	var (
		instance     *config.Instance
		cmdHelper    helper.Cmd
		commandTexts *cmd.BaseCommandTexts
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		commandTexts = NewFrameworkText(api.Distributed, api.MPI)
	})

	Describe("NewCommand", func() {
		It("should create a new Command", func() {
			command := NewCommand(api.MPI, spec.NewDistributedBCSpec(api.MPI), nil, commandTexts, instance, cmdHelper)
			Expect(command).NotTo(BeNil())
			Expect(command.Framework).To(Equal(api.MPI))
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			distributedCmd := NewCommand(api.MPI, spec.NewDistributedBCSpec(api.MPI), nil, commandTexts, instance, cmdHelper)
			err := distributedCmd.RegisterFlags()
			Expect(err).To(BeNil())
		})
	})
})
