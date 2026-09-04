/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"

	argshandler "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/args"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/logs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// WorkloadInfo holds the detected workload type, framework, Uuid and external name
type WorkloadInfo struct {
	Type                 string
	Framework            string
	ExternalWorkloadName string
	Uuid                 string
}

var (
	LogsFlagsOptions logs.Options
)

var Logs = &cobra.Command{
	Use:   "logs",
	Long:  logsExampleMsg,
	Short: logsShortMsg,
	Args:  cobra.ExactArgs(1),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		LogsFlagsOptions.WorkloadName = argshandler.GetNameFromArgs(&args)
		return LogsValidatorPre(cmd, &LogsFlagsOptions)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return RunWorkloadLogsCMD(cmd, args, LogsFlagsOptions, txn)
	},

	ValidArgsFunction: GetWorkloadNamesForArg,
}

func LogsValidatorPre(cmd *cobra.Command, options *logs.Options) error {
	if options.Project == "" {
		LogsFlagsOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
		options.Project = LogsFlagsOptions.Project
	}
	cmdHelper.AddValidation(validators.NewProjectValidator(options.Project))

	if msg, projectErr := cmdHelper.Check(); projectErr != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, projectErr)
	}

	// Validate logs options
	if err := logs.ValidateLogsOptions(options); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LogsInvalidArgument, err)
	}

	return nil
}

func LogsInitializerCMD(args *[]string, options *logs.Options, txn *transaction.Txn) (*WorkloadInfo, error) {
	id := txn.AddTimer(consts.TimerDescribeWorkload)
	defer txn.StopTimer(id)

	workloadName := (*args)[0]
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

	// Get the workload to detect its type and framework
	fb := filter.NewBuilder().
		WithClusterID(clusterId).
		WithProject(options.Project).
		AppendName(workloadName)

	workloadList, err := workloadClient.List(txn.GetCtx(), 10, 0, fb.StringList(), false)
	if err != nil {
		return nil, err
	}

	if len(workloadList.Workloads) == 0 {
		return nil, fmt.Errorf("workload %s not found in project %s", workloadName, options.Project)
	}

	workloadFromList := workloadList.Workloads[0]
	workloadId := workloadFromList.Id
	workload, err := workloadClient.GetWorkload(txn.GetCtx(), workloadId)
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

	// For external workloads, and any workloads, we need to use a special approach
	// These workloads use "external" for both type and framework in the cluster API URL
	// URL format: /api/v1/{project}/workloads/external/external/{workloadName}/logs
	if workload.ExternalWorkloadName != nil {
		workloadInfo.ExternalWorkloadName = *workload.ExternalWorkloadName
		options.WorkloadFramework, options.WorkloadType, options.WorkloadName =
			api.UndefinedFramework, api.Training, workloadInfo.ExternalWorkloadName
	}

	return workloadInfo, nil
}

func RunWorkloadLogsCMD(cmd *cobra.Command, args []string, options logs.Options, txn *transaction.Txn) error {
	workloadInfo, err := LogsInitializerCMD(&args, &options, txn)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToGetWorkloadType, err)
	}

	if workloadInfo.ExternalWorkloadName == "" {
		options.WorkloadType, options.WorkloadFramework, err = DetermineTypeAndFramework(workloadInfo)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToGetWorkloadType, err)
		}
	}
	return logs.Run(cmd, &options, txn)
}

// DetermineTypeAndFramework maps workload info to internal API types and frameworks
// Exported for testing
func DetermineTypeAndFramework(workloadInfo *WorkloadInfo) (api.WorkloadType, api.Framework, error) {

	switch workloadInfo.Type {
	case api.Workspace.StringCapitalize():
		return api.Workspace, api.Runai, nil

	case api.Inference.StringCapitalize():
		return api.Inference, api.Knative, nil

	case api.DistributedInference.StringCapitalize():
		return api.DistributedInference, api.LWS, nil

	case api.Training.StringCapitalize():
		return api.Training, api.Runai, nil

	case api.Distributed.StringCapitalize():
		switch workloadInfo.Framework {
		case clusterconstants.KindPyTorchJob:
			return api.Distributed, api.PyTorch, nil
		case clusterconstants.KindMpiJob:
			return api.Distributed, api.MPI, nil
		case clusterconstants.KindTFJob:
			return api.Distributed, api.TF, nil
		case clusterconstants.KindXGBoostJob:
			return api.Distributed, api.XGBoost, nil
		case clusterconstants.KindJAXJob:
			return api.Distributed, api.Jax, nil
		case clusterconstants.KindRunaiJob, "":
			// Standard training workload (RunaiJob) but marked as Distributed
			return api.Distributed, api.Runai, nil
		default:
			// Fallback to distributed with AnyDistributed framework
			return api.Distributed, api.AnyDistributed, nil
		}

	default:
		return -1, -1, fmt.Errorf("unknown workload type: %s", workloadInfo.Type)
	}
}

func init() {
	Cmd.AddCommand(Logs)
	Logs.Flags().StringVarP(&LogsFlagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = Logs.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	Logs.Flags().StringVar(&LogsFlagsOptions.Pod, consts.Pod, "", consts.PodIdAttributeExplanation)
	Logs.Flags().Int64VarP(&LogsFlagsOptions.TailLines, consts.Tail, "t", -1, consts.TailAttributeExplanation)
	Logs.Flags().StringVarP(&LogsFlagsOptions.Container, consts.Container, consts.ContainerShort, "", consts.ContainerAttributeExplanation)
	Logs.Flags().BoolVar(&LogsFlagsOptions.Timestamps, consts.Timestamps, false, consts.TimestampsAttributeExplanation)
	Logs.Flags().BoolVarP(&LogsFlagsOptions.Follow, consts.Follow, consts.FollowShort, false, consts.FollowAttributeExplanation)
	Logs.Flags().BoolVar(&LogsFlagsOptions.Previous, consts.Previous, false, consts.PreviousAttributeExplanation)
	Logs.Flags().Int64Var(&LogsFlagsOptions.LimitBytes, consts.LimitBytes, 0, consts.LimitBytesAttributeExplanation)

	sinceFlags := pflag.NewFlagSet("since", pflag.ExitOnError)
	sinceFlags.DurationVar(&LogsFlagsOptions.Since, consts.Since, 0, consts.SinceSecondsAttributeExplanation)
	sinceFlags.StringVar(&LogsFlagsOptions.SinceTime, consts.SinceTime, "", consts.SinceTimeAttributeExplanation)
	Logs.Flags().AddFlagSet(sinceFlags)

	Logs.Flags().DurationVar(&LogsFlagsOptions.WaitTimeout, consts.WaitTimeout, 0, consts.WaitTimeoutAttributeExplanation)
}
