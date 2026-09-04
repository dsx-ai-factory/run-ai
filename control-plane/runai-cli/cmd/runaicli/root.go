/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaicli

/*
Copyright © 2024 Fox Team fox-team@run.ai
*/

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/compute"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/credential"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/datasource"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/datavolume"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/diagnostics"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/environment"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/kubeconfig"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/pvc"
	token "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	tokenManager "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/token"
	updateCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/upgrade/cmd"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/submitdist"

	runaios "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/shortcuts"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/describe"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/logout"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/mycredential"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/report"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/upgrade"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/whoami"

	commAttr "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attributes"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/submit"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/application"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/list"
	templateCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/template"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/training"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
	workloadtypeCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workloadtype"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workspace"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction/reporter"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/nodepool"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/node"

	departmentCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/department"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	attachlegacy "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/attach"
	execlegacy "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/exec"
	logslegacy "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/logs"
	pflegacy "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/pf"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/cluster"
	configCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/login"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/version"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	UseConfig     = "config"
	UseVersion    = "version"
	UseCompletion = "completion"
)

var (
	options = config.Options{
		ConfigPath:  runaios.GetDefaultConfigPath(),
		ConfigFile:  consts.DefaultConfigFile,
		Interactive: config.InteractiveUnset,
		Debug:       false,
		Verbose:     false,
		Quiet:       false,
	}
	errInit  error
	instance = config.GetInstance()
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:          consts.AppName,
	Short:        consts.RootShortMsg,
	Long:         consts.RootLongMsg,
	SilenceUsage: true,

	// Would be run before any child command, ***unless*** it has a PersistentPreRun of its own (found out the hard way)
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if options.Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}

		}

		if instance.GetLogger() == nil {
			if errInit != nil {
				zap.L().Error("failed to initialize logger",
					zap.String("error", errInit.Error()),
				)
			} else {
				zap.L().Error("failed to initialize logger")
			}
			return fmt.Errorf("failed to initialize logger")
		}

		if noCommand(cmd) && errInit != nil {
			return nil
		}

		// config command (we can ignore configuration issue)
		if cmdConfig(cmd) {
			return nil
		}
		if cmdVersion(cmd) {
			return nil
		}

		return errInit
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},

	// there could be only one persistent post run, otherwise the metrics will not be generated, so in case of need, add this code to the function call
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if cmdCompletion(cmd) {
			return nil
		}
		outputFormat := cli.GetOutputTypeFromFlags(cmd)
		updateCmd.CheckForUpdateFlow(cmd, outputFormat)
		return reporter.Report(cmd, args)
	},
}

func noCommand(cmd *cobra.Command) bool {
	return cmd.Parent() == nil
}

func cmdConfig(cmd *cobra.Command) bool {
	if cmd.Use == UseConfig {
		return true
	}

	if cmd.Parent() != nil {
		return cmd.Parent().Use == UseConfig
	}
	return false
}

func cmdVersion(cmd *cobra.Command) bool {
	if cmd.Use == UseVersion {
		return true
	}

	if cmd.Parent() != nil {
		return cmd.Parent().Use == UseVersion
	}
	return false
}

// cmdCompletion returns true if the executed command is the shell completion script generator
// (e.g. "runai completion zsh").
func cmdCompletion(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Use == UseCompletion {
			return true
		}
	}
	return false
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&options.ConfigPath, consts.ConfigPath, "", "config path; can be set by environment variable RUNAI_CLI_CONFIG_PATH")
	RootCmd.PersistentFlags().StringVar(&options.ConfigFile, consts.ConfigFile, options.ConfigFile, "config file name; can be set by environment variable RUNAI_CLI_CONFIG_FILE")
	RootCmd.PersistentFlags().BoolVarP(&options.Debug, consts.Debug, "d", false, "enable debug mode")
	RootCmd.PersistentFlags().BoolVar(&options.Verbose, consts.Verbose, false, "enable verbose mode")
	RootCmd.PersistentFlags().BoolVarP(&options.Quiet, consts.Quiet, "q", false, "enable quiet mode, suppress all output except error messages")

	// init on each command
	cobra.OnInitialize(initRoutine)

	// command section
	addCommands()
	RootCmd.SetOut(os.Stdout)

}

func addCommands() {
	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupConfigID,
		Title: consts.GroupConfigTitle,
	})
	RootCmd.AddCommand(configCmd.Config)
	RootCmd.AddCommand(kubeconfig.Cmd)
	RootCmd.AddCommand(version.Cmd)
	RootCmd.AddCommand(report.Cmd)
	RootCmd.AddCommand(upgrade.Cmd)
	RootCmd.AddCommand(auth.Cmd)
	RootCmd.AddCommand(diagnostics.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupUserManagementID,
		Title: consts.GroupUserManagementTitle,
	})
	RootCmd.AddCommand(login.Cmd)
	RootCmd.AddCommand(whoami.Cmd)
	RootCmd.AddCommand(logout.Cmd)
	RootCmd.AddCommand(mycredential.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupResourcesID,
		Title: consts.GroupResourcesTitle,
	})
	RootCmd.AddCommand(cluster.Cmd)
	RootCmd.AddCommand(node.Cmd)
	RootCmd.AddCommand(nodepool.Cmd)
	RootCmd.AddCommand(pvc.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupAssetsID,
		Title: consts.GroupAssetsTitle,
	})
	RootCmd.AddCommand(environment.Cmd)
	RootCmd.AddCommand(compute.Cmd)
	RootCmd.AddCommand(datasource.Cmd)
	RootCmd.AddCommand(datavolume.Cmd)
	RootCmd.AddCommand(credential.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupOrganizationID,
		Title: consts.GroupOrganizationTitle,
	})
	RootCmd.AddCommand(project.Cmd)
	RootCmd.AddCommand(departmentCmd.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupApplicationID,
		Title: consts.GroupApplicationTitle,
	})
	RootCmd.AddCommand(application.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupWorkloadID,
		Title: consts.GroupWorkloadTitle,
	})
	RootCmd.AddCommand(workload.Cmd)
	RootCmd.AddCommand(training.Cmd)
	RootCmd.AddCommand(workspace.Cmd)
	RootCmd.AddCommand(inference.Cmd)
	RootCmd.AddCommand(shortcuts.Tensorflow)
	RootCmd.AddCommand(shortcuts.Pytorch)
	RootCmd.AddCommand(shortcuts.Xgboost)
	RootCmd.AddCommand(shortcuts.Mpi)
	RootCmd.AddCommand(shortcuts.Jax)
	RootCmd.AddCommand(workloadtypeCmd.Cmd)

	RootCmd.AddCommand(templateCmd.Cmd)

	RootCmd.AddGroup(&cobra.Group{
		ID:    consts.GroupDeprecatedID,
		Title: consts.GroupDeprecatedTitle,
	})

	// legacy commands
	RootCmd.AddCommand(list.Cmd)
	RootCmd.AddCommand(submit.Cmd)
	RootCmd.AddCommand(submitdist.Cmd)
	RootCmd.AddCommand(pflegacy.Cmd)
	RootCmd.AddCommand(logslegacy.Cmd)
	RootCmd.AddCommand(execlegacy.Cmd)   //nolint:staticcheck // our deprecate
	RootCmd.AddCommand(attachlegacy.Cmd) //nolint:staticcheck // our deprecate
	RootCmd.AddCommand(describe.Cmd)     //nolint:staticcheck // our deprecate
}

// initConfig reads in config file and ENV variables if set.
func initRoutine() {

	// init instance
	errInit = config.GetInstance().Initializing(options)
	if errInit != nil {
		fmt.Printf("Error: %v\n", errInit)
		os.Exit(1)
	}
	// init settings instance (automatically inherits config path, and settings file name is constant)
	errInit = settings.GetInstance().Initializing(config.GetInstance().GetOptions().ConfigPath)
	if errInit != nil {
		fmt.Printf("Error: %v\n", errInit)
		os.Exit(1)
	}

	// init cache dir
	initCacheDir()

	// init token manager
	manager := token.GetManager()
	if !manager.IsInitialized() {
		manager.Init(tokenManager.NewSimpleTokenServiceFromConfig, instance)
	}
}

func initCacheDir() {
	logger := instance.GetLogger()
	cacheDir := config.GetInstance().GetCacheDirPath()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		errMkdir := os.MkdirAll(cacheDir, os.ModePerm)
		if errMkdir != nil {
			logger.Debug("failed to create cache dir", zap.String("cacheDir", cacheDir), zap.Error(err))
		}
	}
}
