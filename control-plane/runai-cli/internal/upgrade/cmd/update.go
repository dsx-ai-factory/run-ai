/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	internalOs "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/os"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/upgrade"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/version"
	versionpkg "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/version"
)

const defaultNotificationTTL = 3 * time.Hour

type Checker interface {
	Check(ctx context.Context) (bool, string, error)
}

type upgraderFn func(ctx context.Context,
	cliClient *cli.Client, execShell string, quiet bool) (string, error)

type ConfigProvider interface {
	GetAutoUpdate() bool
	IsAutoUpdateInit() bool
	GetUpdateShell() string
}

type Printer interface {
	Println(a ...interface{})
	Printf(format string, a ...interface{})
}

type UpdateFlowConditioner interface {
	IsRunningUpdateFlowRequired(outputFormat string, autoUpdateEnabled bool) bool
	UpdateSuccessAPITimestamp() error
}

// HandleAutoUpdate handles the CLI update process with all dependencies injected
func HandleAutoUpdate(
	ctx context.Context,
	cliClient *cli.Client,
	checker Checker,
	upgrader upgraderFn,
	config ConfigProvider,
	printer Printer,
	logger *zap.Logger,
	cpUrl string,
	flowConditioner UpdateFlowConditioner,
	outputFormat string,
) {
	// checks if the update flow should be run based on the last successful check timestamp,
	// the output format, and whether auto-update is enabled and the control plane URL.
	if !flowConditioner.IsRunningUpdateFlowRequired(outputFormat, config.GetAutoUpdate()) {
		// If the flowConditioner condition is not met, do not check for updates
		// this prevents spamming users with checking for updates flow
		// and allows to control the frequency of update checks
		logger.Debug("Skipping update check. last successful check is still valid.")
		return
	}

	hasUpdate, latestVersion, err := checker.Check(ctx)
	if err != nil {
		var versionErr *version.CLIVersionTooHighError
		if errors.As(err, &versionErr) {
			printer.Printf(consts.CLIVersionTooHighMsg, versionpkg.GetVersion().ControlPlaneVersion, cpUrl, latestVersion)
		} else {
			logger.Debug("failed to check for CLI updates,", zap.Error(err))
			if config.GetAutoUpdate() {
				printer.Printf(consts.FailedToCheckVersion)
			}
			return
		}
	}
	// update the flowConditioner after a successful check
	// this allows to update the last successful check timestamp
	// only if the check was successful, otherwise will not update the timestamp
	err = flowConditioner.UpdateSuccessAPITimestamp()
	if err != nil {
		logger.Debug("failed to update the timestamp of the last successful check", zap.Error(err))
	}
	logger.Debug("CLI version check completed", zap.Bool("hasUpdate", hasUpdate), zap.String("latestVersion", latestVersion))
	if !hasUpdate {
		return
	}

	if !config.GetAutoUpdate() {
		printer.Printf(consts.ManualUpdateRequiredMsg, versionpkg.GetVersion().ControlPlaneVersion, latestVersion)
		return
	}

	printer.Printf(consts.NewVersionAvailableMsg, latestVersion)
	_, err = upgrader(ctx, cliClient, config.GetUpdateShell(), true)
	if err != nil {
		printer.Printf(consts.FailedToUpdateMsg)
		return
	}

	printer.Printf(consts.UpdateCompleteMsg, versionpkg.GetVersion().ControlPlaneVersion, latestVersion)
}

// PerformUpgrade executes the upgrade process
func PerformUpgrade(ctx context.Context,
	cliClient *cli.Client, execShell string, quiet bool) (string, error) {
	cliUpgradeFactory, err := upgrade.CreateFactory(cliClient, quiet)
	if err != nil {

		return "", fmt.Errorf("failed to initialize CLI update: %w", err)
	}

	upgrader, err := cliUpgradeFactory.CreateFlow(internalOs.DetectOsType())
	if err != nil {
		return "", fmt.Errorf("failed to determine upgrade requirements %w", err)
	}

	installOutput, err := upgrader.Upgrade(ctx, execShell)
	if err != nil {
		return "", err
	}

	return installOutput, nil
}

func CheckForUpdateFlow(cmd *cobra.Command, outputFormat string) {
	instance := config.GetInstance()
	logger := instance.GetLogger()
	txn := instance.GetTxn()
	tokenStr, err := auth.GetToken(txn.GetCtx())
	if err != nil {
		logger.Debug("failed to get token", zap.String("error", err.Error()))
		return
	}
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	cmdHelper := helper.NewCmd()
	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))
	if msg, err := cmdHelper.Check(); err != nil {
		logger.Debug("check failed", zap.String("msg", msg), zap.String("error", err.Error()))
		return
	}

	cliClient, err := cli.NewClient(cpUrl, tokenStr)
	if err != nil {
		logger.Debug("failed to create CLI client", zap.String("error", err.Error()))
		return
	}
	checker := version.NewCLIChecker(cliClient)
	flowCondition := upgrade.NewFlowCondition(instance.GetVersionCachePath(), defaultNotificationTTL, cpUrl, logger)
	HandleAutoUpdate(txn.GetCtx(), cliClient, checker, PerformUpgrade, instance, cmd, logger, cpUrl, flowCondition, outputFormat)
}

func ClearFlowConditionCache(instance *config.Instance, cpUrl string) error {
	logger := instance.GetLogger()
	flowCondition := upgrade.NewFlowCondition(instance.GetVersionCachePath(), defaultNotificationTTL, cpUrl, logger)
	return flowCondition.UpdateSuccessAPITimestamp()
}
