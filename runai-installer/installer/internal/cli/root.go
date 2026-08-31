/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"errors"
	"os"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/spf13/cobra"
)

type globalOpts struct {
	devMode     bool
	logsDir     string
	stepTimeout time.Duration
}

var globals globalOpts

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "installer",
		Short:        "NVIDIA Run:ai Installer",
		SilenceUsage: true,
	}

	pf := root.PersistentFlags()
	pf.BoolVar(&globals.devMode, "dev", false, "development mode")
	pf.StringVar(&globals.logsDir, "logs-dir", "", "directory for installer logs")
	pf.DurationVar(&globals.stepTimeout, "step-timeout", 7*time.Minute, "timeout per step")

	root.AddCommand(newInstallCmd())
	root.AddCommand(newInstallNewCmd())
	root.AddCommand(newUpgradeCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newAirgapCmd())

	return root
}

func Execute() {
	err := NewRootCmd().Execute()
	if err == nil {
		return
	}
	if errors.Is(err, specrun.ErrPreflightFailed) {
		os.Exit(2)
	}
	os.Exit(1)
}
