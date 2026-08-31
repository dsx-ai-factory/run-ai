/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	errWorkloadNotFound        = errors.New("workload not found")
	errScalingCancelled        = errors.New("scaling cancelled by user")
	scalingMsg                 = "%s %s scaled successfully\n"
	dryRunMsg                  = "%s"
	EnvFormat                  = "%s=%s"
	autoscalingWarningTemplate = `Warning: Autoscaling is currently enabled for this workload (min=%d, max=%d).
Scaling to %d replicas will disable autoscaling and switch to fixed replicas.
Continue? [y/N] `
)

// AutoscalingInfo holds information about workload autoscaling configuration
type AutoscalingInfo struct {
	Enabled     bool
	MinReplicas int32
	MaxReplicas int32
}

// GetAutoscalingInfoFn is a function type for fetching autoscaling info
type GetAutoscalingInfoFn func(txn *transaction.Txn, domain, token, uuid string) (*AutoscalingInfo, error)

type Command struct {
	cmd.BaseCommand
	generalSpec          *spec.GeneralSpec
	commandTexts         cmd.SearchableCommandTexts
	searchWorkloadFn     cmd.SearchWorkloadFn
	uuid                 string
	updateWorkloadFn     cmd.UpdateWorkloadFn
	project              string
	workloadType         api.WorkloadType
	completionFn         func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
	force                bool
	getAutoscalingInfoFn GetAutoscalingInfoFn // Optional: for NIM autoscaling warning
}

func NewCommand(framework api.Framework, workloadType api.WorkloadType, workloadSpec *spec.GeneralSpec, initBaseCommandFn func() *cobra.Command,
	commandTexts cmd.SearchableCommandTexts, instance *config.Instance, cmdHelper helper.Cmd,
	searchWorkloadFn cmd.SearchWorkloadFn, updateWorkloadFn cmd.UpdateWorkloadFn, completionFn func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)) *Command {
	return NewCommandWithAutoscalingCheck(framework, workloadType, workloadSpec, initBaseCommandFn, commandTexts, instance, cmdHelper, searchWorkloadFn, updateWorkloadFn, completionFn, nil)
}

// NewCommandWithAutoscalingCheck creates a scale command with optional autoscaling check (for NIM workloads)
func NewCommandWithAutoscalingCheck(framework api.Framework, workloadType api.WorkloadType, workloadSpec *spec.GeneralSpec, initBaseCommandFn func() *cobra.Command,
	commandTexts cmd.SearchableCommandTexts, instance *config.Instance, cmdHelper helper.Cmd,
	searchWorkloadFn cmd.SearchWorkloadFn, updateWorkloadFn cmd.UpdateWorkloadFn, completionFn func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective),
	getAutoscalingInfoFn GetAutoscalingInfoFn) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &Command{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: framework,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		generalSpec:          workloadSpec,
		commandTexts:         commandTexts,
		searchWorkloadFn:     searchWorkloadFn,
		updateWorkloadFn:     updateWorkloadFn,
		workloadType:         workloadType,
		completionFn:         completionFn,
		getAutoscalingInfoFn: getAutoscalingInfoFn,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = "scale [WORKLOAD_NAME]"
	c.Cmd.Short = c.commandTexts.GetShort()
	c.Cmd.Example = c.commandTexts.GetExample()
	c.Cmd.ValidArgsFunction = c.completionFn
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		token := c.Instance.GetToken()
		cpUrl := c.Instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := c.Instance.GetConfigValue(config.ClusterUUID)
		instance := config.GetInstance()

		project := c.project
		if project == "" {
			project = instance.GetConfigValue(config.ClusterProjectName)
		}

		var cmdHelper helper.Cmd
		cmdHelper.AddValidation(validators.NewProjectValidator(project))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		Name, err := argshandler.ValidateArgsAndExtractName(cmd, &args)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdate, err)
		}

		if Name == "" {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), errors.New(consts.ErrorWorkloadNameIsRequired))
		}

		fb := filter.NewBuilder().WithWorkloadFramework(c.Framework).WithWorkloadType(c.workloadType.QueryParam())
		handler, err := c.searchWorkloadFn(c.GetTxnFn(), cpUrl, token, Name, project, clusterId, fb)

		if err != nil {
			_ = cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetSearchableErrorMsg(), err)

			if handler != nil && len(handler.GetList()) > 1 {
				cmd.Println(consts.PreFoundMultipleWorkloadMessage)
				cmd.Println(handler.GenerateTable())
			}
			return err
		}
		if handler != nil && len(handler.GetList()) != 1 {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetNoResultsErrorMsg(), errWorkloadNotFound)
		}
		c.uuid = handler.GetList()[0].Id

		return nil
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Validate that --replicas flag is set (required for distributed inference scaling)
		if !cmd.Flags().Changed(consts.AttributeNameReplicas) {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(),
				errors.New("--replicas flag is required for scaling"))
		}

		instance := config.GetInstance()
		txn := instance.GetTxn()
		token := instance.GetToken()
		cpUrl := c.Instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := c.Instance.GetConfigValue(config.ClusterUUID)

		// Check for autoscaling and prompt for confirmation if not forced (NIM only).
		shouldRemoveAutoscaling := false
		if c.getAutoscalingInfoFn != nil && (!c.generalSpec.IsDryRun() || c.force) {
			autoscalingInfo, err := c.getAutoscalingInfoFn(txn, cpUrl, token, c.uuid)
			if err != nil {
				// Log warning but continue - don't fail scaling if we can't get autoscaling info
				cmd.PrintErrf("Warning: Could not check autoscaling status: %v\n", err)
			} else if autoscalingInfo != nil && autoscalingInfo.Enabled {
				// Get requested replicas from spec
				requestedReplicas := c.generalSpec.GetFlags().Nim.Replicas

				if !c.generalSpec.IsDryRun() && !c.force {
					cmd.Printf(autoscalingWarningTemplate,
						autoscalingInfo.MinReplicas,
						autoscalingInfo.MaxReplicas,
						requestedReplicas)

					// Read user input
					reader := bufio.NewReader(os.Stdin)
					response, err := reader.ReadString('\n')
					if err != nil {
						return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
					}

					response = strings.TrimSpace(strings.ToLower(response))
					if response != "y" && response != "yes" {
						return cmdtxn.CmdErrorPrintAndSet(cmd, "", errScalingCancelled)
					}
				}
				shouldRemoveAutoscaling = true
			}
		}

		// If scaling from autoscaling to fixed replicas, add autoscaling to remove list
		if shouldRemoveAutoscaling {
			c.generalSpec.GetFlags().Nim.Remove = append(c.generalSpec.GetFlags().Nim.Remove, "autoscaling")
		}

		c.setCliCommandAsAnnotation()

		workloadName, err := c.updateWorkloadFn(txn, cpUrl, token, c.uuid, clusterId, c.generalSpec)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
		}

		if c.generalSpec.IsDryRun() {
			cmd.Printf(dryRunMsg, workloadName)
		} else {
			cmd.Printf(scalingMsg, c.workloadType, workloadName)
		}
		return nil
	}
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	c.generalSpec.SetScaleFlags(c.Cmd)
	c.Cmd.Flags().StringVarP(&c.project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	// Add --force flag for NIM workloads only (for skipping autoscaling confirmation prompt)
	if c.getAutoscalingInfoFn != nil {
		c.Cmd.Flags().BoolVar(&c.force, consts.AttributeNameForce, false, consts.AttributeUsageForceScale)
	}

	return nil
}

func (c *Command) setCliCommandAsAnnotation() {
	osArgs := os.Args
	arguments, _ := argshandler.GetExecuteString(&osArgs)
	// for now we will use the same annotation key as the update command
	c.generalSpec.SetCliCommandAsAnnotation(fmt.Sprintf(EnvFormat, spec.UpdateCommandAnnotationKey, arguments))
}
