/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"

	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"

	commWorkload "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"

	internalApi "github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type DescribeCmdOptions struct {
	Output            string
	Project           string
	ShowGeneral       bool
	ShowRuntime       bool
	ShowCompute       bool
	ShowStorage       bool
	ShowPods          bool
	ShowEvents        bool
	ShowNetworks      bool
	IncludeContainers bool
	PodLimit          int32
	EventLimit        int32
	Framework         internalApi.Framework
	WorkloadType      internalApi.WorkloadType
	SortOrder         string
}

var (
	DescribeFlagsOptions DescribeCmdOptions
	frameworkType        string
)

var Describe = &cobra.Command{
	Use:   "describe WORKLOAD_NAME",
	Short: "Describe a workload",
	Args:  cobra.ExactArgs(1),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if WorkloadType != "" {

			workloadTypeEnum, InnerError := internalApi.ParseWorkloadType(WorkloadType)
			if InnerError != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToGetWorkloadType, errors.New(consts.FailedToGetWorkloadType))
			}
			DescribeFlagsOptions.WorkloadType = workloadTypeEnum
		}

		if frameworkType != "" {
			framework, err := internalApi.ParseWorkloadFramework(frameworkType)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, fmt.Sprintf(consts.InvalidFrameworkType, internalApi.GetWorkloadFramework()), err)
			}
			DescribeFlagsOptions.Framework = framework
		}

		return DescribeProjectValidatorPre(cmd, &DescribeFlagsOptions)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return RunWorkloadDescribeCMD(cmd, args, DescribeFlagsOptions, txn)
	},

	ValidArgsFunction: GetWorkloadNamesForArg,
}

func DescribeProjectValidatorPre(cmd *cobra.Command, options *DescribeCmdOptions) error {
	if options.Project == "" {
		DescribeFlagsOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
		options.Project = DescribeFlagsOptions.Project
	}
	cmdHelper.AddValidation(validators.NewProjectValidator(options.Project))

	if msg, projectErr := cmdHelper.Check(); projectErr != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, projectErr)
	}
	return nil
}

func DescribeInitializerCMD(args []string, options DescribeCmdOptions, txn *transaction.Txn) (
	*api.Workload, *describe.Service, *displayer.DisplayOptions, error) {
	id := txn.AddTimer(consts.TimerDescribeWorkload)
	defer txn.StopTimer(id)

	workloadName := args[0]
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)

	token := instance.GetToken()

	_, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, options.Project)
	if err != nil {
		return nil, nil, nil, err
	}

	requestedWorkload := api.Workload{
		Name:      workloadName,
		ClusterID: clusterId,
		Project:   options.Project,
		Type:      options.WorkloadType,
		Framework: options.Framework,
	}
	workloadClient, err := client.NewWorkloadServiceClient(cpUrl, token)
	if err != nil {
		return nil, nil, nil, err
	}
	workloadV2Client, err := client.NewWorkloadV2(cpUrl, token)
	if err != nil {
		return nil, nil, nil, err
	}
	podsClient, err := client.NewBackendWorkloadPodsClient(cpUrl, token)
	if err != nil {
		return nil, nil, nil, err
	}
	eventsClient, err := client.NewBackendWorkloadEventsClient(cpUrl, token, options.SortOrder)
	if err != nil {
		return nil, nil, nil, err
	}
	describeClients := describe.NewService(workloadClient, workloadV2Client, podsClient, eventsClient)

	workloadDescribeOptions := displayer.DisplayOptions{
		ShowGeneral:       options.ShowGeneral,
		ShowCompute:       options.ShowCompute,
		ShowStorage:       options.ShowStorage,
		ShowPods:          options.ShowPods,
		ShowEvents:        options.ShowEvents,
		ShowNetworks:      options.ShowNetworks,
		IncludeContainers: options.IncludeContainers,
	}

	return &requestedWorkload, describeClients, &workloadDescribeOptions, nil
}

func RunWorkloadDescribeCMD(cmd *cobra.Command, args []string, options DescribeCmdOptions, txn *transaction.Txn) error {
	requestedWorkload, describeClients, workloadDescribeOptions, err := DescribeInitializerCMD(args, options, txn)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDescribeWorkload, err)
	}

	workloadDetails, err := describeClients.GetWorkload(txn.GetCtx(), requestedWorkload, options.PodLimit, options.EventLimit)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDescribeWorkload, err)
	}

	displayer := describe.NewWorkloadDisplayer(workloadDetails, *workloadDescribeOptions)
	err = RenderDescription(options, displayer)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDescribeWorkload, err)
	}

	return nil
}

func init() {
	Cmd.AddCommand(Describe)
	Describe.Flags().StringVar(&WorkloadType, consts.WorkloadType, "", consts.WorkloadTypeUsage)
	Describe.Flags().StringVar(&frameworkType, consts.WorkloadFramework, "", consts.AttributeFrameworkUsage)
	BindDescribeFlags(Describe, &DescribeFlagsOptions)
	_ = Describe.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
	_ = Describe.RegisterFlagCompletionFunc(consts.WorkloadFramework, commWorkload.FrameworkFlagCompletion)

	Describe.MarkFlagsMutuallyExclusive(consts.WorkloadType, consts.WorkloadFramework)
}

func BindDescribeFlags(describe *cobra.Command, options *DescribeCmdOptions) {
	describe.Flags().StringVarP(&options.Output, "output", "o", "table", consts.DescribeOutputUsage)
	describe.Flags().StringVarP(&options.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = describe.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	describe.Flags().BoolVar(&options.ShowGeneral, "general", true, consts.DescribeGeneralUsage)
	describe.Flags().BoolVar(&options.ShowCompute, "compute", true, consts.DescribeComputeUsage)
	describe.Flags().BoolVar(&options.ShowStorage, "storage", true, consts.DescribeStorageUsage)
	describe.Flags().BoolVar(&options.ShowPods, "pods", true, consts.DescribePodsUsage)
	describe.Flags().BoolVar(&options.ShowEvents, "events", true, consts.DescribeEventsUsage)
	describe.Flags().BoolVar(&options.ShowNetworks, "networks", true, consts.DescribeNetworksUsage)
	describe.Flags().BoolVar(&options.IncludeContainers, "containers", false, consts.DescribeContainersUsage)
	describe.Flags().Int32Var(&options.PodLimit, "pod-limit", 10, consts.DescribePodLimitUsage)
	describe.Flags().Int32Var(&options.EventLimit, "event-limit", 50, consts.DescribeEventLimitUsage)
	describe.Flags().StringVar(&options.SortOrder, consts.DescribeEventsSortOrderFlag, client.EventsDefaultSortOrder.String(), consts.DescribeEventsSortOrderUsage)
	_ = describe.RegisterFlagCompletionFunc(consts.DescribeEventsSortOrderFlag, EventsSortOrderFlagCompletion)
}

func EventsSortOrderFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return client.GetEventSortOrderOptions(), cobra.ShellCompDirectiveNoFileComp
}

func RenderDescription(options DescribeCmdOptions, displayer describe.Displayer) error {
	var err error
	switch options.Output {
	case "table":
		err = displayer.AsTable()
	case "json":
		err = displayer.AsJSON()
	case "yaml":
		err = displayer.AsYAML()
	default:
		err = fmt.Errorf("unsupported output format: %s", options.Output)
	}
	return err
}
