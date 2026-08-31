/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"

	commonexec "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/exec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/capability"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	ExecFlagsOptions commonexec.Options
)

var Exec = &cobra.Command{
	Use:   "exec",
	Long:  execExampleMsg,
	Short: execShortMsg,
	Args:  cobra.MinimumNArgs(1),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		ExecFlagsOptions.WorkloadName, err = argshandler.ValidateArgsAndExtractName(cmd, &args)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecInvalidArgument, err)
		}
		ExecFlagsOptions.Command = args
		return ExecValidatorPre(cmd, &ExecFlagsOptions)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return RunWorkloadExecCMD(cmd, ExecFlagsOptions, txn)
	},

	ValidArgsFunction: GetWorkloadNamesForArg,
}

func ExecValidatorPre(cmd *cobra.Command, options *commonexec.Options) error {
	if options.Project == "" {
		ExecFlagsOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
		options.Project = ExecFlagsOptions.Project
	}
	cmdHelper.AddValidation(validators.NewProjectValidator(options.Project))

	if msg, projectErr := cmdHelper.Check(); projectErr != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, projectErr)
	}

	if err := commonexec.ValidateExecOptions(options); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecInvalidArgument, err)
	}

	return nil
}

func ExecInitializerCMD(options *commonexec.Options, txn *transaction.Txn) (*WorkloadInfo, error) {
	id := txn.AddTimer(consts.TimerDescribeWorkload)
	defer txn.StopTimer(id)

	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	_, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, options.Project)
	if err != nil {
		return nil, err
	}

	workloadClient, err := client.NewWorkloadServiceClient(cpUrl, token)
	if err != nil {
		return nil, err
	}

	fb := filter.NewBuilder().
		WithClusterID(clusterId).
		WithProject(options.Project).
		AppendName(options.WorkloadName)

	workloadList, err := workloadClient.List(txn.GetCtx(), 10, 0, fb.StringList(), false)
	if err != nil {
		return nil, err
	}

	if len(workloadList.Workloads) == 0 {
		return nil, fmt.Errorf("workload %s not found in project %s", options.WorkloadName, options.Project)
	}

	workloadFromList := workloadList.Workloads[0]
	workload, err := workloadClient.GetWorkload(txn.GetCtx(), workloadFromList.Id)
	if err != nil {
		return nil, err
	}

	workloadInfo := &WorkloadInfo{
		Type: workload.Type,
		Uuid: workload.Id,
	}

	if workload.DistributedFramework != nil {
		workloadInfo.Framework = *workload.DistributedFramework
	}

	if workload.ExternalWorkloadName != nil {
		workloadInfo.ExternalWorkloadName = *workload.ExternalWorkloadName
		options.WorkloadFramework, options.WorkloadType, options.WorkloadName =
			api.UndefinedFramework, api.Training, workloadInfo.ExternalWorkloadName
	}

	return workloadInfo, nil
}

func RunWorkloadExecCMD(cmd *cobra.Command, options commonexec.Options, txn *transaction.Txn) error {
	workloadInfo, err := ExecInitializerCMD(&options, txn)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToGetWorkloadType, err)
	}

	if workloadInfo.ExternalWorkloadName == "" {
		options.WorkloadType, options.WorkloadFramework, err = DetermineTypeAndFramework(workloadInfo)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToGetWorkloadType, err)
		}
	}

	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	capabilityChecker := capability.NewChecker(cpUrl, token, clusterId)
	isSupported, capErr := capabilityChecker.IsSupported(txn, options.WorkloadType, capability.Exec)
	if capErr != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError, capErr)
	}
	if !isSupported {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ExecCommandError,
			fmt.Errorf(capability.ClusterVersionError, capability.MinimalClusterVersionForInfwClusterApiActions))
	}

	return commonexec.Run(cmd, &options, txn)
}

func init() {
	Cmd.AddCommand(Exec)
	commonexec.FlagOutputType(Exec, &ExecFlagsOptions)
}
