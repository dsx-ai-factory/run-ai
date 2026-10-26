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

		DescribeTable("preserves literal single quotes",
			func(input []string, expected string) {
				execute, ok := argsutils.GetExecuteString(&input)
				Expect(ok).To(BeTrue())
				Expect(execute).To(Equal(expected))
			},
			Entry("shell literal", []string{"sh", "-c", "echo 'value is $HOME'"}, `sh -c "echo 'value is $HOME'"`),
			Entry("MATLAB string", []string{"matlab", "-batch", "cd('/some/long/path'); exit"}, `matlab -batch "cd('/some/long/path'); exit"`),
			Entry("mixed quotes", []string{"sh", "-c", `printf "%s" 'value'`}, `sh -c "printf \"%s\" 'value'"`),
			Entry("single word", []string{"can't"}, `"can't"`),
		)
	})
})
