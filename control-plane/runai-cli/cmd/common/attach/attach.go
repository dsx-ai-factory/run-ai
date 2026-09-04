/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"context"
	"errors"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/status"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/io"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/pseudo"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/attach"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/cluster"

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

func Pre(cmd *cobra.Command, args []string, attachOptions *Options) error {
	if attachOptions == nil {
		// This should never happen
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.AttachCommandError, errors.New(consts.AttachCommandError))
	}
	attachOptions.WorkloadName = args[0]
	if attachOptions.Project == "" {
		attachOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
	}
	cmdHelper.AddValidation(validators.NewProjectValidator(attachOptions.Project))
	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	err := validateAttachOptions(attachOptions)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.AttachInvalidArgument, err)
	}
	return nil
}

func validateAttachOptions(attachOptions *Options) error {
	if attachOptions.WorkloadName == "" {
		return errors.New(consts.ErrorWorkloadNameIsRequired)
	}
	return nil
}

func Run(cmd *cobra.Command, options *Options, txn *transaction.Txn) error {
	ctx := txn.GetCtx()
	token := instance.GetToken()

	metadata, err := cluster.GetStatus()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.AttachCommandError, err)
	}
	clusterUrl, err := cluster.GetClusterApiUrl()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.AttachCommandError, err)
	}
	instance.GetLogger().Debug("cluster metadata", zap.Any("metadata", metadata))

	options.WorkloadType, options.WorkloadFramework, options.WorkloadName, err = resolve.ClusterWorkload(
		instance, txn, options.WorkloadType, options.WorkloadFramework, options.Project, options.WorkloadName)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.AttachCommandError, err)
	}

	err = Attach(ctx, cmd, options, token, clusterUrl.String(), metadata)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToPrintAttachOutput, err)
	}
	return nil
}

func Attach(ctx context.Context, cmd *cobra.Command, options *Options, token, clusterAPIUrl string, metadata *status.Metadata) error {
	dialer := websocket.NewTransportDialer(token)
	sleepTime := cluster.GetIterationMinimalTime(DefaultSleepTimeBetweenRetries, options.WaitTimeout)
	attachReq := attach.NewWorkloadAttach(dialer, clusterAPIUrl, options.WaitTimeout, sleepTime, metadata)
	term := pseudo.NewTerminal(attachReq, cmd.OutOrStdout())
	workload := apis.NewWorkload(options.Project, options.WorkloadType, options.WorkloadName, options.WorkloadFramework)
	attachOptions, err := pseudo.NewOptions(options.Pod, options.Container, options.Stdin, options.TTY)
	if err != nil {
		return err
	}
	cmd.Println("If you don't see a command prompt, try pressing enter.")
	return term.Exec(ctx, workload, attachOptions, io.NewStd(cmd))
}
