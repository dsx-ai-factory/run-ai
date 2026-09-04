/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd_test

import (
	"context"
	"errors"
	"fmt"

	mocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/upgrade"
	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/upgrade/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/version"
	versionpkg "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/version"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleAutoUpdate", func() {
	var (
		checker      *mockVersionChecker
		config       *mockConfig
		printer      *mockPrinter
		logger       *zap.Logger
		testURL      string = "http://test.url"
		ctx          context.Context
		cliClient    *cli.Client
		mockCtrl     *gomock.Controller
		mockCond     *mocks.MockUpdateFlowConditioner
		outputFormat string
	)

	BeforeEach(func() {
		checker = &mockVersionChecker{}
		config = &mockConfig{autoUpdateInit: true}
		printer = &mockPrinter{messages: []string{}}
		logger, _ = zap.NewDevelopment()
		ctx = context.Background()
		var err error
		cliClient, err = cli.NewClient(testURL, "test-token")
		Expect(err).NotTo(HaveOccurred())
		mockCtrl = gomock.NewController(GinkgoT())
		mockCond = mocks.NewMockUpdateFlowConditioner(mockCtrl)
		outputFormat = "table"
	})

	JustBeforeEach(func() {
		mockCond.EXPECT().IsRunningUpdateFlowRequired(outputFormat, config.autoUpdate).Return(true).AnyTimes()
		mockCond.EXPECT().UpdateSuccessAPITimestamp().AnyTimes()
		cmd.HandleAutoUpdate(ctx, cliClient, checker, mockUpgrader(checker), config, printer, logger, testURL, mockCond, outputFormat)
	})

	Context("when version check fails", func() {
		BeforeEach(func() {
			checker.checkErr = errors.New("check failed")
		})
		Context("auto-update is enabled", func() {
			BeforeEach(func() {
				config.autoUpdate = true
			})
			It("should print error message and not attempt upgrade", func() {
				Expect(printer.messages).To(HaveLen(1))
				Expect(printer.messages[0]).To(ContainSubstring("Failed to check for CLI updates."))
				Expect(checker.upgradeErr).To(BeNil())
			})
		})
		Context("auto-update is disabled", func() {
			BeforeEach(func() {
				config.autoUpdate = false
			})
			It("should print error message and not attempt upgrade", func() {
				Expect(printer.messages).To(HaveLen(0))
				Expect(checker.upgradeErr).To(BeNil())
			})
		})

	})

	Context("when version is too high", func() {
		BeforeEach(func() {
			checker.checkErr = &version.CLIVersionTooHighError{}
		})

		It("should print version too high warning with correct format", func() {
			Expect(printer.messages).To(HaveLen(1))
			expectedMsg := fmt.Sprintf("Your CLI version (%s) is higher than the Control Plane version.\nControl Plane URL: %s\nControl Plane-compatible CLI version: %s\n\nTo download the compatible CLI version, run:\nrunai upgrade --force",
				versionpkg.GetVersion().ControlPlaneVersion,
				testURL,
				"")
			Expect(printer.messages[0]).To(Equal(expectedMsg))
		})
	})

	Context("when no update is available", func() {
		BeforeEach(func() {
			checker.hasUpdate = false
		})

		It("should not print any messages", func() {
			Expect(printer.messages).To(BeEmpty())
		})

	})

	Context("when update is available but auto-update disabled", func() {
		BeforeEach(func() {
			checker.hasUpdate = true
			checker.version = "2.0.0"
			config.autoUpdate = false
		})

		It("should print all expected messages in correct order", func() {
			Expect(printer.messages).To(HaveLen(1))
			expectedMsg := fmt.Sprintf("Warning: You are currently using version %s. A newer version %s is available.\nTo update, run:\nrunai upgrade",
				versionpkg.GetVersion().ControlPlaneVersion,
				"2.0.0")
			Expect(printer.messages[0]).To(Equal(expectedMsg))
		})

		Context("but auto-update is initialized", func() {
			BeforeEach(func() {
				config.autoUpdateInit = true
			})

			It("should only print manual update message", func() {
				Expect(printer.messages).To(HaveLen(1))
				expectedMsg := fmt.Sprintf("Warning: You are currently using version %s. A newer version %s is available.\nTo update, run:\nrunai upgrade",
					versionpkg.GetVersion().ControlPlaneVersion,
					"2.0.0")
				Expect(printer.messages[0]).To(Equal(expectedMsg))
			})
		})
	})

	Context("when update is available and auto-update enabled", func() {
		BeforeEach(func() {
			checker.hasUpdate = true
			checker.version = "2.0.0"
			config.autoUpdate = true
		})

		It("should print update messages in correct order", func() {
			Expect(printer.messages).To(HaveLen(2))
			expectedNewVersionMsg := fmt.Sprintf("A new version (%s) is available.\nUpdating now...\n",
				"2.0.0")
			expectedCompleteMsg := fmt.Sprintf("Update complete. (%s --> %s)",
				versionpkg.GetVersion().ControlPlaneVersion,
				"2.0.0")
			Expect(printer.messages[0]).To(Equal(expectedNewVersionMsg))
			Expect(printer.messages[1]).To(Equal(expectedCompleteMsg))
		})

		Context("when update fails", func() {
			BeforeEach(func() {
				checker.upgradeErr = errors.New("upgrade failed")
			})

			It("should print new version and failure messages", func() {
				Expect(printer.messages).To(HaveLen(2))
				expectedNewVersionMsg := fmt.Sprintf("A new version (%s) is available.\nUpdating now...\n",
					"2.0.0")
				Expect(printer.messages[0]).To(Equal(expectedNewVersionMsg))
				Expect(printer.messages[1]).To(Equal("Failed to update the CLI. try update manually using:\nrunai upgrade"))
			})
		})
	})

	Context("when condition is not met", func() {
		BeforeEach(func() {
			mockCond.EXPECT().IsRunningUpdateFlowRequired(outputFormat, config.autoUpdate).Return(false).AnyTimes()
			mockCond.EXPECT().UpdateSuccessAPITimestamp().AnyTimes()
		})
		It("should not run upgrade logic", func() {
			Expect(printer.messages).To(BeEmpty())
		})
	})
})
