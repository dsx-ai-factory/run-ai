/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"errors"

	"go.uber.org/zap"

	versionpkg "github.com/run-ai/runai/control-plane/runai-cli/pkg/version"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	token "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	internalOs "github.com/run-ai/runai/control-plane/runai-cli/internal/os"
	upgradecmd "github.com/run-ai/runai/control-plane/runai-cli/internal/upgrade/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/version"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
	txn       *transaction.Txn
	force     bool
	execShell string
)

const (
	execShellStr = "exec"
	aliasCommand = "update"
)

// Cmd represents the base command when called without any subcommands
var Cmd = &cobra.Command{
	Use:     consts.UpgradeCommand,
	Short:   consts.UpgradeShortMsg,
	Aliases: []string{aliasCommand},
	Args:    cobra.NoArgs,
	GroupID: consts.GroupConfigID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)

		cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		tokenStr, err := token.GetToken(txn.GetCtx())
		if err != nil {
			return err
		}
		instance.SetToken(tokenStr)
		if !instance.GetAutoUpdate() {
			cmd.Println(consts.EnableAutoUpdate)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		instance := config.GetInstance()

		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()
		ctx := txn.GetCtx()

		cliClient, err := cli.NewClient(cpUrl, token)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToCheckVersion, err)
		}

		if !force {
			cliVersion := version.NewCLIChecker(cliClient)
			isNewVersion, versionName, err := cliVersion.Check(ctx)
			if err != nil {

				var cliVersionTooHighError *version.CLIVersionTooHighError
				if errors.As(err, &cliVersionTooHighError) {
					cmd.Printf(consts.CLIVersionTooHighMsg,
						versionpkg.GetVersion().ControlPlaneVersion,
						cpUrl, cliVersionTooHighError.ControlPlaneVersion)
					return err
				}
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToCheckVersion, err)
			}
			if !isNewVersion {
				cmd.PrintErrln(consts.NoNewVersion, versionName)
				return nil
			}
			cmd.Println(consts.UpgradingCLIVersion, versionName)
		} else {
			cmd.Println(consts.ForceUpgrade)
		}

		installOutput, err := upgradecmd.PerformUpgrade(ctx, cliClient, execShell, false)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdateCLI, err)
		}
		cmd.Println(installOutput)

		err = upgradecmd.ClearFlowConditionCache(instance, cpUrl)
		if err != nil {
			instance.GetLogger().Debug(consts.FailedToClearUpdateCache, zap.String("error", err.Error()))
		}
		return nil
	},
}

func init() {
	Cmd.Flags().BoolVar(&force, "force", false, "upgrade CLI without checking for new version")

	if internalOs.DetectOsType() != internalOs.Windows {
		internalInstance := config.GetInstance()
		Cmd.Flags().StringVar(&execShell, execShellStr, internalInstance.GetUpdateShell(), "the shell that will be used to run the upgrade")
		_ = Cmd.RegisterFlagCompletionFunc(execShellStr, execFlagCompletion)
	}
}

// execFlagCompletion auto completion support for exec command
func execFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"bash\trun the upgrade with bash command",
		"dash\trun the upgrade with dash command",
		"sh\trun the upgrade with sh command",
		"fish\trun the upgrade with fish command",
		"zsh\trun the upgrade with zsh command",
	}, cobra.ShellCompDirectiveNoFileComp
}
