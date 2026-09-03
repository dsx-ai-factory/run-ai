/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale_test

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/scale"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Scale Command", func() {
	var (
		scaleCommand *scale.Command
		texts        cmd.SearchableCommandTexts
		instance     *config.Instance
		cmdHelper    helper.Cmd
		workloadSpec *spec.GeneralSpec
	)

	BeforeEach(func() {
		texts = scale.NewScaleText(api.DistributedInference, api.LWS)
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		workloadSpec = spec.NewDistributedSpec(api.LWS)
		scaleCommand = scale.NewCommand(api.LWS, api.DistributedInference, workloadSpec, cmd.NewEmptyCobraCommand, texts, instance, cmdHelper, inference.GetByNameAndProj, inference.Update, factories.GetCompletionFnMock())
	})

	Describe("NewCommand", func() {
		It("should create a new Command with correct properties", func() {
			Expect(scaleCommand.Framework).To(Equal(api.LWS))
			Expect(scaleCommand.GetTxnFn).ToNot(BeNil())
			Expect(scaleCommand.Instance).To(Equal(instance))
			Expect(scaleCommand.CmdHelper).To(Equal(cmdHelper))
		})
	})

	Describe("GetCommand", func() {
		It("should return a cobra.Command with correct properties", func() {
			command := scaleCommand.GetCommand()
			Expect(command).To(BeAssignableToTypeOf(&cobra.Command{}))
			Expect(command.Use).To(Equal("scale [WORKLOAD_NAME]"))
			Expect(command.Short).To(Equal(texts.GetShort()))
			Expect(command.Example).To(Equal(texts.GetExample()))
			Expect(command.ValidArgsFunction).ToNot(BeNil())
		})

		It("should have auto completion function", func() {
			command := scaleCommand.GetCommand()
			Expect(command.ValidArgsFunction).ToNot(BeNil())
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			err := scaleCommand.RegisterFlags()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
