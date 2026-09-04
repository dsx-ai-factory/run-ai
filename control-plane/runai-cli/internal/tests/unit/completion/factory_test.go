/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package completion

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

var (
	instance        = config.GetInstance()
	errExpected     = errors.New("prepare to die")
	expectedOptions = []string{"a", "b", "c"}
)

var _ = Describe("Completion Factory Tests", func() {
	var (
		logs         *observer.ObservedLogs
		resourceType = completion.Project

		args       = []string{}
		toComplete = ""
		cmd        = &cobra.Command{}
	)

	BeforeEach(func() {
		logger, logsCapture := setupLogsCapture()
		instance.SetLogger(logger)
		logs = logsCapture
	})

	DescribeTableSubtree("test factory methods",
		func(factoryFn func(completion.ResourceType, func(*cobra.Command) error, func(*cobra.Command) ([]string, error)) completion.CompletionFn) {
			It("initInstance function is used - with log err", func() {
				fn := factoryFn(resourceType, initInstanceErr, optionsLister)
				completions, directive := fn(cmd, args, toComplete)
				Expect(completions).To(BeNil())
				Expect(directive).To(Equal(cobra.ShellCompDirectiveError))
				validateLogs(resourceType, errExpected, logs)
			})

			It("optionLister function is used - with log err", func() {
				fn := factoryFn(resourceType, initInstance, optionsListerErr)
				completions, directive := fn(cmd, args, toComplete)
				Expect(completions).To(BeNil())
				Expect(directive).To(Equal(cobra.ShellCompDirectiveError))
				validateLogs(resourceType, errExpected, logs)
			})

			It("no err", func() {
				fn := factoryFn(resourceType, initInstance, optionsLister)
				completions, directive := fn(cmd, args, toComplete)
				Expect(completions).To(Equal(expectedOptions))
				Expect(directive).To(Equal(cobra.ShellCompDirectiveNoFileComp))
				validateNoLogs(logs)
			})
		},
		Entry("arg", completion.GetArgCompletionFn),
		Entry("flag", completion.GetFlagCompletionFn),
	)
})

func initInstanceErr(_ *cobra.Command) error {
	return errExpected
}

func initInstance(_ *cobra.Command) error {
	return nil
}

func optionsListerErr(_ *cobra.Command) ([]string, error) {
	return nil, errExpected
}

func optionsLister(_ *cobra.Command) ([]string, error) {
	return expectedOptions, nil
}

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.DebugLevel)
	return zap.New(core), logs
}

func validateLogs(resourceType completion.ResourceType, err error, observed *observer.ObservedLogs) {
	Expect(observed.Len()).To(Equal(1))
	logEntry := observed.All()[0]
	Expect(logEntry.Message).To(Equal(consts.FailedToAutoComplete))
	logContext := logEntry.ContextMap()
	Expect(logContext["resource"]).To(Equal(resourceType.String()))
	Expect(logContext["error"]).To(Equal(err.Error()))
}

func validateNoLogs(observed *observer.ObservedLogs) {
	Expect(observed.Len()).To(Equal(0))
}
