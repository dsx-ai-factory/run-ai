/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Command", func() {
	var (
		instance         *config.Instance
		cmdHelper        helper.Cmd
		commandTexts     *cmd.BaseSearchableCommandTexts
		searchWorkloadFn cmd.SearchWorkloadFn
		updateWorkloadFn cmd.UpdateWorkloadFn
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		commandTexts = NewWorkloadOnlyText(api.Inference)
	})

	Describe("NewCommand", func() {
		It("should create a new Command", func() {
			command := NewCommand(api.Knative, spec.NewInferenceSpec(), nil, commandTexts, instance, cmdHelper, searchWorkloadFn, updateWorkloadFn)
			Expect(command).NotTo(BeNil())
			Expect(command.Framework).To(Equal(api.Knative))
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			command := NewCommand(api.Knative, spec.NewInferenceSpec(), nil, commandTexts, instance, cmdHelper, searchWorkloadFn, updateWorkloadFn)
			err := command.RegisterFlags()
			Expect(err).To(BeNil())
		})
	})
})
