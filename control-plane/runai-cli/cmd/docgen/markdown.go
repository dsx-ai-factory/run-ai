/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	runaiCli "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewMarkDownCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "markdown",
		Aliases: []string{"md"},
		Short:   markdownShortMsg,
		Run: func(cmd *cobra.Command, args []string) {
			GenerateMarkdown(cmd)
		},
	}
	return cmd
}

func GenerateMarkdown(cmd *cobra.Command) {
	cmd.Printf(generateMDStartMsg)
	dirPath, _ := cmd.Flags().GetString(outputDirFlag)
	runaiCli.RootCmd.DisableAutoGenTag = true
	err := doc.GenMarkdownTree(runaiCli.RootCmd, dirPath)
	if err != nil {
		cmd.PrintErrln(createMDError, err)
	}
	cmd.Printf(generateMDCompleteMsg)
}
