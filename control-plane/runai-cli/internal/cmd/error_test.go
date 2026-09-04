/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

var _ = Describe("Error Types", func() {
	Context("UnsupportedError", func() {
		It("should create error with correct message", func() {
			err := cmd.NewUnsupportedError(api.PyTorch, api.Workspace)
			Expect(err.Error()).To(Equal("unsupported framework pytorch and workload type workspace"))
		})

		It("should create error with correct message", func() {
			err := cmd.NewUnsupportedError(api.PyTorch, api.Inference)
			Expect(err.Error()).To(Equal("unsupported framework pytorch and workload type inference"))
		})
	})

	Context("UnsupportedAliasError", func() {
		It("should create error with correct message", func() {
			err := cmd.NewUnsupportedAliasError(api.PyTorch, api.Training)
			Expect(err.Error()).To(Equal("unsupported alias for framework pytorch and workload type training"))
		})

		It("should create error with correct message", func() {
			err := cmd.NewUnsupportedAliasError(api.Knative, api.Inference)
			Expect(err.Error()).To(Equal("unsupported alias for framework knative and workload type inference"))
		})

	})

	Context("Error Interface Compliance", func() {
		It("should implement error interface", func() {
			var _ error = &cmd.UnsupportedError{}
			var _ error = &cmd.UnsupportedAliasError{}
		})
	})
})
