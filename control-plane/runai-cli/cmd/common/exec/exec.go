/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/capability"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/io"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/pseudo"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/exec"

	argshandler "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/args"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/resolve"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/websocket"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	DefaultSleepTimeBetweenRetries = 5 * time.Second
	bashCmd                        = "/bin/bash"
)

var (
	instance  = config.GetInstance()
	cmdHelper helper.Cmd
)

type Options struct {
	WorkloadName      string
	Pod               string
	Container         string
	Project           string
	WorkloadType      api.WorkloadType
	WorkloadFramework api.Framework

	PodRunning time.Duration
	Stdin      bool
	TTY        bool
	Command    []string

	WaitTimeout time.Duration
}

func FlagOutputType(cmd *cobra.Command, output *Options) {
	cmd.Flags().StringVarP(&output.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	cmd.Flags().StringVar(&output.Pod, consts.Pod, "", consts.PodIdAttributeExplanation)
	cmd.Flags().StringVarP(&output.Container, consts.Container, consts.ContainerShort, "", consts.ContainerAttributeExplanation)

	cmd.Flags().DurationVar(&output.PodRunning, consts.AttributeNamePodRunningTimeout, 0, consts.AttributeUsagePodRunningTimeout)
	cmd.Flags().BoolVarP(&output.Stdin, consts.AttributeNameStdin, "i", false, consts.AttributeUsageStdin)
	cmd.Flags().BoolVarP(&output.TTY, consts.AttributeNameTTY, "t", false, consts.AttributeUsageTTY)

	cmd.Flags().DurationVar(&output.WaitTimeout, consts.WaitTimeout, 0, consts.WaitTimeoutAttributeExplanation)
}

func Pre(cmd *cobra.Command, args []string, execOptions *Options) error {
	var err error
	token := instance.GetToken()
	txn := instance.GetTxn()
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	controlUrl := instance.GetConfigValue(config.ControlPlaneUrl)

	if execOptions == nil {
		// This should never happen
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, errors.New(consts.ExecCommandError))
	}
	if execOptions.Project == "" {
		execOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
	}

	capabilityChecker := capability.NewChecker(controlUrl, token, clusterId)
	isSupported, isSupportedErr := capabilityChecker.IsSupported(txn, execOptions.WorkloadType, capability.Exec)
	if isSupportedErr != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, isSupportedErr)
	}
	if !isSupported {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, fmt.Errorf(capability.ClusterVersionError, capability.MinimalClusterVersionForInfwClusterApiActions))
	}

	cmdHelper.AddValidation(validators.NewProjectValidator(execOptions.Project))
	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	execOptions.WorkloadName, err = argshandler.ValidateArgsAndExtractName(cmd, &args)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecInvalidArgument, err)
	}
	execOptions.Command = args
	err = ValidateExecOptions(execOptions)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecInvalidArgument, err)
	}
	return nil
}

func ValidateExecOptions(execOptions *Options) error {
	if execOptions.WorkloadName == "" {
		return errors.New(consts.ErrorWorkloadNameIsRequired)
	}
	return nil
}

func Run(cmd *cobra.Command, options *Options, txn *transaction.Txn) error {
	ctx := txn.GetCtx()
	token := instance.GetToken()

	metadata, err := cluster.GetStatus()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, err)
	}
	clusterUrl, err := cluster.GetClusterApiUrl()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, err)
	}
	instance.GetLogger().Debug("cluster metadata", zap.Any("metadata", metadata))

	options.WorkloadType, options.WorkloadFramework, options.WorkloadName, err = resolve.ClusterWorkload(
		instance, txn, options.WorkloadType, options.WorkloadFramework, options.Project, options.WorkloadName)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, err)
	}

	dialer := websocket.NewTransportDialer(token)
	sleepTime := cluster.GetIterationMinimalTime(DefaultSleepTimeBetweenRetries, options.WaitTimeout)
	execReq := exec.NewWorkloadExec(dialer, clusterUrl.String(), options.WaitTimeout, sleepTime, metadata)
	term := pseudo.NewTerminal(execReq, cmd.OutOrStdout())
	requestedWorkload := apis.NewWorkloadForClusterAPI(options.Project, options.WorkloadType, options.WorkloadName, options.WorkloadFramework)
	execOptions, err := pseudo.NewOptions(options.Pod, options.Container, options.Stdin, options.TTY, options.Command...)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToPrintExecOutput, err)
	}
	err = term.Exec(ctx, requestedWorkload, execOptions, io.NewStd(cmd))
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToPrintExecOutput, err)
	}
	return nil
}

func BashPreRun(cmd *cobra.Command, args []string, execOptions *Options) error {
	if len(args) < 1 {
		// This should never happen
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecInvalidArgument, errors.New(consts.ErrorWorkloadNameIsRequired))
	}
	execOptions.WorkloadName = args[0]
	execOptions.Command = []string{bashCmd}
	execOptions.TTY = true
	execOptions.Stdin = true
	err := ProjectHandler(cmd, execOptions)
	if err != nil {
		return err
	}
	return nil

}

func ProjectHandler(cmd *cobra.Command, execOptions *Options) error {
	if execOptions == nil {
		// This should never happen
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, errors.New(consts.ExecCommandError))
	}
	if execOptions.Project == "" {
		execOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
	}
	cmdHelper.AddValidation(validators.NewProjectValidator(execOptions.Project))
	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	return nil
}
