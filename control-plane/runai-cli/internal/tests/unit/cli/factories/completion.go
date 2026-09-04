/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factories

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"

	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func GetCompletionFactoryMock(ctrl *gomock.Controller) *mocks.MockFactory {
	mock := mocks.NewMockFactory(ctrl)
	mock.EXPECT().GetCompletionFn(gomock.Any(), gomock.Any()).Return(GetCompletionFnMock(), nil).AnyTimes()

	return mock
}

func GetCompletionFnMock() completion.CompletionFn {
	mock := func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	return mock
}

// ConvertToCompletionTestCmd convert a cmd string to a cmd which cobra understands as needed to be completed (Tab Tab)
// provide valuePrefix if you want to test a completion to a partial value
func ConvertToCompletionTestCmd(cmdStr string, valuePrefix string) string {
	if len(cmdStr) == 0 {
		return cmdStr
	}

	// space at the end is mandatory for completion (
	completeCmd := fmt.Sprintf("__complete %s ", cmdStr)

	if len(valuePrefix) != 0 {
		completeCmd = fmt.Sprintf("%s%s", completeCmd, valuePrefix)
	}

	return completeCmd
}

func AssertCompletionResponseStatus(response string, hasError bool) {
	if hasError {
		Expect(response).To(ContainSubstring(fmt.Sprintf(":%v", cobra.ShellCompDirectiveError)))
	} else {
		Expect(response).To(ContainSubstring(fmt.Sprintf(":%v", cobra.ShellCompDirectiveNoFileComp)))
	}
}

func AssertCompletionOptions(response string, expectedOptions []string) {
	if len(expectedOptions) == 0 {
		Expect(strings.HasPrefix(response, ":")).To(BeTrue())
	}

	for _, option := range expectedOptions {
		Expect(response).To(ContainSubstring(option))
	}
}
