/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"fmt"
	"strconv"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"

	commAttr "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attributes"

	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	token "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
)

var (
	cmdHelper   helper.Cmd
	interactive attributes.EnableFlag
	instance    = config.GetInstance()
	txn         *transaction.Txn
)

// Cmd represents the project command
var Cmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"projects"},
	Short:   consts.ProjectShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupOrganizationID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initProjectInstance(cmd)
		if err != nil {
			// err is already formatted
			return err
		}

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}

		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		options := instance.GetOptions()
		if !cmdHelper.CheckInteractive(txn, interactive, options.Interactive) {
			cmd.HelpFunc()(cmd, args)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		tokenStr, err := token.GetToken(txn.GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
		}

		fb := filter.NewBuilder().WithClusterID(clusterId)
		handler, err := projectservice.GetProjectList(txn, cpUrl, tokenStr, &PageOptions, fb)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, err)
		}

		// Sanity check: this code path should never be reached.
		if handler == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, ErrParseList)
		}

		selectedProject, err := table.Selection(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, err)
		}
		if selectedProject == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, fmt.Errorf(consts.ErrorSelectionEmpty))
		}

		selectedProjectName, exist := (*selectedProject)["Name"]
		if !exist {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, fmt.Errorf("project name not found"))
		}

		foundProject := false
		for _, project := range handler.GetList() {
			if project.Name == selectedProjectName {
				projectId, err := strconv.Atoi(project.Id)
				if err != nil {
					return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, err)
				}
				foundProject = true
				err = instance.SetProject(project.Name, int32(projectId))
				if err != nil {
					return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, err)
				}
				break
			}
		}

		if !foundProject {
			err = fmt.Errorf(consts.ErrorNoProjectFound, selectedProjectName)
		}

		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListProject, err)
		}
		cmd.Printf(consts.SetProjectProperties, selectedProjectName)
		return nil
	},
}

func init() {
	Cmd.Flags().Var(&interactive, consts.Interactive, consts.InteractiveArgumentUsage)
	_ = Cmd.RegisterFlagCompletionFunc(consts.Interactive, attributes.EnableFlagCompletion)
	list.SetPageFlags(Cmd, &PageOptions)
}

func initProjectInstance(cmd *cobra.Command) error {
	instance = config.GetInstance()

	txn = instance.GetTxn()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterUrl := instance.GetConfigValue(config.ClusterDomain)
	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))
	cmdHelper.AddValidation(validators.NewClusterValidator(clusterUrl))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}

	tokenStr, err := token.GetToken(txn.GetCtx())
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
	}
	instance.SetToken(tokenStr)

	return nil
}
