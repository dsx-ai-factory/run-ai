/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/capability"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/cluster"

	argshandler "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/args"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/logs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/resolve"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/io"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/websocket"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	DefaultTailLines               = -1
	sinceFlagSet                   = "since"
	DefaultSleepTimeBetweenRetries = 5 * time.Second
)

var (
	instance  = config.GetInstance()
	cmdHelper helper.Cmd
)

type Options struct {
	Pod               string
	Container         string
	TailLines         int64
	Timestamps        bool
	Follow            bool
	Previous          bool
	LimitBytes        int64
	Since             time.Duration
	SinceTime         string
	WorkloadName      string
	WaitTimeout       time.Duration
	Project           string
	WorkloadType      api.WorkloadType
	WorkloadFramework api.Framework
}

func FlagOutputType(cmd *cobra.Command, output *Options) {
	cmd.Flags().StringVarP(&output.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	cmd.Flags().StringVar(&output.WorkloadName, consts.JobName, "", consts.JobNameAttributeExplanation)
	cmd.Flags().StringVar(&output.Pod, consts.Pod, "", consts.PodIdAttributeExplanation)
	cmd.Flags().Int64VarP(&output.TailLines, consts.Tail, "t", DefaultTailLines, consts.TailAttributeExplanation)
	cmd.Flags().StringVarP(&output.Container, consts.Container, consts.ContainerShort, "", consts.ContainerAttributeExplanation)
	cmd.Flags().BoolVar(&output.Timestamps, consts.Timestamps, false, consts.TimestampsAttributeExplanation)
	cmd.Flags().BoolVarP(&output.Follow, consts.Follow, consts.FollowShort, false, consts.FollowAttributeExplanation)
	cmd.Flags().BoolVar(&output.Previous, consts.Previous, false, consts.PreviousAttributeExplanation)
	cmd.Flags().Int64Var(&output.LimitBytes, consts.LimitBytes, 0, consts.LimitBytesAttributeExplanation)
	sinceFlags := pflag.NewFlagSet(sinceFlagSet, pflag.ExitOnError)
	sinceFlags.DurationVar(&output.Since, consts.Since, 0, consts.SinceSecondsAttributeExplanation)
	sinceFlags.StringVar(&output.SinceTime, consts.SinceTime, "", consts.SinceTimeAttributeExplanation)
	cmd.Flags().AddFlagSet(sinceFlags)
	cmd.Flags().DurationVar(&output.WaitTimeout, consts.WaitTimeout, 0, consts.WaitTimeoutAttributeExplanation)
}

func Pre(cmd *cobra.Command, args []string, logsOptions *Options) error {
	token := instance.GetToken()
	txn := instance.GetTxn()
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	controlUrl := instance.GetConfigValue(config.ControlPlaneUrl)

	if logsOptions == nil {
		// This should never happen
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsCommandError, errors.New(consts.LogsCommandError))
	}
	logsOptions.WorkloadName = argshandler.GetNameFromArgs(&args)
	if logsOptions.Project == "" {
		logsOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
	}

	capabilityChecker := capability.NewChecker(controlUrl, token, clusterId)
	isSupported, isSupportedErr := capabilityChecker.IsSupported(txn, logsOptions.WorkloadType, capability.Logs)
	if isSupportedErr != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsCommandError, isSupportedErr)
	}
	if !isSupported {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsCommandError, fmt.Errorf(capability.ClusterVersionError, capability.MinimalClusterVersionForInfwClusterApiActions))
	}

	cmdHelper.AddValidation(validators.NewProjectValidator(logsOptions.Project))
	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	err := ValidateLogsOptions(logsOptions)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsInvalidArgument, err)
	}
	return nil
}

func ValidateLogsOptions(logsOptions *Options) error {
	if logsOptions.WorkloadName == "" {
		return errors.New(consts.ErrorWorkloadNameIsRequired)
	}
	if len(logsOptions.SinceTime) > 0 && logsOptions.Since != 0 {
		return errors.New(consts.SinceTimeAndSecondsError)
	}

	if logsOptions.LimitBytes < 0 {
		return errors.New(consts.LimitBytesError)
	}

	if logsOptions.Since < 0 {
		return errors.New(consts.SinceSecondsError)
	}
	if logsOptions.TailLines < -1 {
		return errors.New(consts.TailLinesError)
	}

	// valid sinceTime is RFC3339
	if len(logsOptions.SinceTime) > 0 {
		_, err := time.Parse(time.RFC3339, logsOptions.SinceTime)
		if err != nil {
			return errors.New(consts.SinceTimeError) //nolint:staticcheck // user facing message
		}
	}
	return nil
}

func Run(cmd *cobra.Command, options *Options, txn *transaction.Txn) error {
	ctx := txn.GetCtx()
	token := instance.GetToken()

	metadata, err := cluster.GetStatus()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsCommandError, err)
	}
	clusterUrl, err := cluster.GetClusterApiUrl()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsCommandError, err)
	}
	instance.GetLogger().Debug("cluster metadata", zap.Any("metadata", metadata))

	options.WorkloadType, options.WorkloadFramework, options.WorkloadName, err = resolve.ClusterWorkload(
		instance, txn, options.WorkloadType, options.WorkloadFramework, options.Project, options.WorkloadName)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsCommandError, err)
	}

	dialer := websocket.NewTransportDialer(token)
	sleepTime := cluster.GetIterationMinimalTime(DefaultSleepTimeBetweenRetries, options.WaitTimeout)
	logSteamer := logs.NewWorkloadLogsStreamer(dialer, clusterUrl.String(), options.WaitTimeout, sleepTime, metadata)
	proxy := io.NewOneWayProxy(1 * time.Second)
	printer := logs.NewPrinter(logSteamer, proxy, os.Stdout)
	requestedWorkload := apis.NewWorkloadForClusterAPI(options.Project, options.WorkloadType, options.WorkloadName, options.WorkloadFramework)
	logsOptions := logs.NewOptions(options.Pod, options.Container,
		options.SinceTime, options.TailLines, options.Timestamps, options.Follow, options.Previous, options.LimitBytes, options.Since)
	err = printer.Print(ctx, requestedWorkload, logsOptions)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToPrintLogs, err)
	}
	return nil
}
