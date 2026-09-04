/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"strings"

	argsutils "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/args"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Submit Common", func() {
	Context("getExecuteString", func() {
		It("args nil", func() {
			execute, ok := argsutils.GetExecuteString(nil)
			Expect(ok).To(BeFalse())
			Expect(execute).To(Equal(""))
		})

		It("args 1 word", func() {
			args := strings.Split("one", " ")
			execute, ok := argsutils.GetExecuteString(&args)
			Expect(ok).To(BeTrue())
			Expect(execute).To(Equal("one"))
		})

		It("args more than 1", func() {
			args := strings.Split("one two three four", " ")
			execute, ok := argsutils.GetExecuteString(&args)
			Expect(ok).To(BeTrue())
			Expect(execute).To(Equal("one two three four"))
		})

		It("args with group string", func() {
			args := []string{"one", "grouped arg"}
			execute, ok := argsutils.GetExecuteString(&args)
			Expect(ok).To(BeTrue())
			Expect(execute).To(Equal("one 'grouped arg'"))
		})
	})
})
