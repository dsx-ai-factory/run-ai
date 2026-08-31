/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"

	runaios "github.com/run-ai/runai/control-plane/runai-cli/internal/os"
	"github.com/spf13/viper"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	"github.com/spf13/cobra"
)

const (
	FileStr                = "file"
	configGenerateShortMsg = "generate config file"
	configGenerateExample  = `# JSON configuration file
runai config generate -f path/to/config.json
# YAML configuration file
runai config generate -f path/to/config.yaml
`
)

var (
	ofJson   bool
	ofYaml   bool
	filePath string
)

// generate represents the config command
var generate = &cobra.Command{
	Use:     "generate",
	Short:   configGenerateShortMsg,
	Example: configGenerateExample,
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" {
			return fmt.Errorf("file path is required")
		}

		filePath = filepath.Clean(filePath)
		err := generateFolder(cmd, filepath.Dir(filePath))
		if err != nil {
			return err
		}
		return generateFile(cmd, filePath)

	},
}

func generateFolder(cmd *cobra.Command, filePath string) error {
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ConfigFileGenerationFailed, err)
	}
	return nil
}

func generateFile(cmd *cobra.Command, filePath string) error {
	err := runaios.WriteViperConfigToFile(viper.GetViper(), filePath)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ConfigFileGenerationFailed, err)
	}
	cmd.Println(consts.ConfigFileGenerated)
	return nil
}

func init() {
	Config.AddCommand(generate)

	generate.Flags().StringVarP(&filePath, FileStr, "f", "", "file path for generating the configuration")
	_ = generate.MarkFlagRequired(FileStr)
}
