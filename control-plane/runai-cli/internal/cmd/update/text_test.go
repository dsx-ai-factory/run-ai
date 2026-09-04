/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/update"
)

var _ = Describe("Update Texts", func() {
	var texts *cmd.BaseSearchableCommandTexts

	Context("without Framework", func() {
		BeforeEach(func() {
			texts = update.NewWorkloadOnlyText(api.Inference)
		})

		Describe("GetShort", func() {
			DescribeTable("returns correct message",
				func(workloadType api.WorkloadType, expected string) {
					texts = update.NewWorkloadOnlyText(workloadType)
					Expect(texts.GetShort()).To(Equal(expected))
				},
				Entry("Inference", api.Inference, "update an inference workload"),
			)
		})

	})
})
