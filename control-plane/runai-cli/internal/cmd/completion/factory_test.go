/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package completion

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/list"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Factory", func() {
	var (
		listFactory *list.Factory
		instance    *config.Instance
		cmdHelper   helper.Cmd

		completionFactory Factory
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		listFactory = list.NewCommandFactory(instance, cmdHelper)

		completionFactory = NewCompletionFactory(listFactory, initInstance)
	})

	It("successfully get fn", func() {
		fn, err := completionFactory.GetCompletionFn(api.Runai, api.Training)
		Expect(err).NotTo(HaveOccurred())
		Expect(fn).NotTo(BeNil())
	})

	It("fail to get fn - not supported framework", func() {
		fn, err := completionFactory.GetCompletionFn(api.AnyDistributed, api.Training)
		Expect(err).To(HaveOccurred())
		Expect(fn).To(BeNil())
	})
})

func initInstance(_ *cobra.Command) error {
	return nil
}
