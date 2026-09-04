/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/io"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	oscli "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/utility"

	"github.com/spf13/cobra"
)

const (
	cpUrlName           = "cp-url"
	authUrlName         = "auth-url"
	autoUpdateFlag      = "auto-update"
	updateShellFlag     = "update-shell"
	outputFlag          = "output"
	cacheDirectory      = "cache-directory"
	listFormatAlignFlag = "list-format-align"

	// Flag descriptions
	cpUrlDesc             = "set the control plane URL"
	authUrlDesc           = "set the authorization URL; most likely the same as the control plane URL"
	outputDesc            = "set the default output type"
	autoUpdateDesc        = "enable/disable automatic CLI updates"
	updateShellDesc       = "set the shell to use for updates (e.g., bash, zsh)"
	cacheDirDeprecatedMsg = "modifying the cache directory is deprecated and will be removed in a future release"
	listFormatAlignDesc   = "enable/disable aligned output formatting for list commands to improve readability (disabled by default)."
)

type Options struct {
	interactive     attributes.EnableFlag
	cpUrl           string
	authUrl         string
	output          string
	statusDuration  string
	cacheDir        string
	autoUpdate      attributes.EnableFlag
	updateShell     string
	listFormatAlign attributes.EnableFlag
}

var (
	options Options
)

// set represents the config command
var set = &cobra.Command{
	Use:     "set",
	Short:   "Set configuration values",
	Args:    cobra.NoArgs,
	Example: consts.ConfigSetExample,

	RunE: func(cmd *cobra.Command, args []string) error {
		// check sanity to prevent set of some of the attributes
		logger := instance.GetLogger()
		attribute, err := utility.SanityCheck(options.cpUrl, options.authUrl)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSetSuffix(cmd, fmt.Sprintf(consts.ConfigSetFailed, attribute), err, consts.NL)
		}

		// clear credentials and delete authentication file in case we set them
		if options.cpUrl != "" || options.authUrl != "" {
			err = instance.ClearCredentials()
			if err != nil {
				logger.Warn("failed to clear credentials from config file", zap.Error(err))
			}
			err = io.CheckExistAndDeleteFile(instance.GetAuthenticationPath())
			if err != nil {
				logger.Warn("failed to remove authentication file", zap.Error(err))
			}
		}

		err = checkOption(options, cmd)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ConfigFileGenerationFailed, err)
		}
		cmd.Println(consts.ConfigSetSuccess)
		return nil
	},
}

func checkOption(options Options, cmd *cobra.Command) error {
	var err error
	logger := instance.GetLogger()
	if options.interactive != "" {
		logger.Debug("setting interactive mode", zap.String(consts.Interactive, string(options.interactive)))
		if options.interactive.IsEnabled() {
			err = instance.SetInteractive(true)
		} else {
			err = instance.SetInteractive(false)
		}
		if err != nil {
			return err
		}
	}

	if options.statusDuration != "" {
		_, err = time.ParseDuration(options.statusDuration)
		if err != nil {
			return err
		}
		err = instance.SetStatusDuration(options.statusDuration)
		if err != nil {
			return err
		}
	}

	if options.cpUrl != "" {
		logger.Debug("setting control plane URL", zap.String("cp_url", options.cpUrl))
		err = instance.SetCPURL(options.cpUrl)
		if err != nil {
			return err
		}
		err = instance.ClearCluster()
		if err != nil {
			return err
		}

		// remove cache file if exists
		err = oscli.RemoveFile(instance.GetVersionCachePath())
		if err != nil {
			return err
		}
	}
	if options.cacheDir != "" {
		cmd.Println(cacheDirDeprecatedMsg)
		cmd.Println("cache directory is aligned with config directory and cannot be set separately")
	}

	if options.authUrl != "" {
		logger.Debug("setting authentication URL", zap.String("authentication", options.authUrl))
		return instance.SetAuthURL(options.authUrl)
	} else if options.cpUrl != "" {
		logger.Debug("setting authentication URL as the control plane URL",
			zap.String("authentication", options.cpUrl))
		return instance.SetAuthURL(options.cpUrl)
	}

	if options.autoUpdate != "" {
		if options.autoUpdate.IsEnabled() {
			logger.Debug("enabling auto-update for CLI")
			err = instance.SetAutoUpdate(true)
		} else {
			logger.Debug("disabling auto-update for CLI")
			err = instance.SetAutoUpdate(false)
		}
		if err != nil {
			return err
		}
	}

	if options.updateShell != "" {
		logger.Debug("setting update shell", zap.String("shell", options.updateShell))
		err = instance.SetUpdateShell(options.updateShell)
		if err != nil {
			return err
		}
	}
	if options.listFormatAlign != "" {
		if options.listFormatAlign.IsEnabled() {
			logger.Debug("enabling list format alignment")
			err = settingsInstance.SetListFormatAlign(true)
		} else {
			logger.Debug("disabling list format alignment")
			err = settingsInstance.SetListFormatAlign(false)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	Config.AddCommand(set)

	set.Flags().StringVar(&options.cpUrl, cpUrlName, "", cpUrlDesc)
	set.Flags().StringVar(&options.authUrl, authUrlName, "", authUrlDesc)
	set.Flags().StringVar(&options.output, outputFlag, "", outputDesc)
	set.Flags().StringVar(&options.cacheDir, cacheDirectory, "", "set the cache directory")

	_ = set.Flags().MarkDeprecated(cacheDirectory, cacheDirDeprecatedMsg)
	set.Flags().Var(&options.interactive, consts.Interactive, consts.InteractiveArgumentUsage)
	set.Flags().StringVar(&options.statusDuration, consts.StatusTimeoutDuration, "", consts.StatusTimeoutDurationArgumentUsage)
	set.Flags().Var(&options.autoUpdate, autoUpdateFlag, autoUpdateDesc)
	set.Flags().StringVar(&options.updateShell, updateShellFlag, "", updateShellDesc)
	set.Flags().Var(&options.listFormatAlign, listFormatAlignFlag, listFormatAlignDesc)
	set.MarkFlagsOneRequired(cpUrlName, authUrlName, cacheDirectory, consts.Interactive, consts.StatusTimeoutDuration, autoUpdateFlag, updateShellFlag, listFormatAlignFlag)

	_ = set.RegisterFlagCompletionFunc(consts.Interactive, attributes.EnableFlagCompletion)
	_ = set.RegisterFlagCompletionFunc(autoUpdateFlag, autoUpdateFlagCompletion)
	_ = set.RegisterFlagCompletionFunc(updateShellFlag, updateShellFlagCompletion)
	_ = set.RegisterFlagCompletionFunc(listFormatAlignFlag, listAlignFlagCompletion)
}
