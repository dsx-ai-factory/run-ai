/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/progressbar"

	"github.com/samber/lo"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"

	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/template"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/progressbar/spinner"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	commWorkload "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/workload"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/spf13/cobra"
)

type DeleteOptions struct {
	Project      string
	workloadType string
	framework    string
	isForceYes   bool
}

// DeletedWorkload used for toDelete map, map keys are the workloads names.
type DeletedWorkload struct {
	uuid  string
	found bool
}

var deleteOptions DeleteOptions

// Delete represents the delete command
var Delete = &cobra.Command{
	Use:     "delete",
	Short:   "Delete workloads",
	Aliases: []string{"del"},
	Example: BulkDeleteExamples,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if deleteOptions.Project == "" {
			deleteOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
		}
		cmdHelper.AddValidation(validators.NewProjectValidator(deleteOptions.Project))

		if msg, projectErr := cmdHelper.Check(); projectErr != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, projectErr)
		}

		if len(args) == 0 {
			cmd.HelpFunc()(cmd, args)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		if len(args) > maxDeleteThreshold {
			return cmdtxn.CmdErrorPrintAndSet(cmd,
				FailedToBulkDeletePre,
				fmt.Errorf("too many workloads to delete. limit: %d, actual: %d", maxDeleteThreshold, len(args)))
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var wlNotFoundNames []string
		var wlFoundNames []string
		var sp progressbar.ProgressBar
		confirmed, err := confirmWorkloadDeletion(args)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToBulkDeletePre, err)
		}

		if !confirmed {
			cmd.Println(BulkConfirmationAborted)
			return nil
		}

		if len(args) == 1 {
			sp = spinner.NewSpinner(fmt.Sprintf(SingleSpinnerMsg, args[0]))
		} else {
			sp = spinner.NewSpinner(BulkSpinnerMsg)
		}
		go sp.Run()
		{
			deleteList, err := getDeleteList(deleteOptions, args)
			if err != nil {
				sp.Stop()
				return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToBulkDeletePre, err)
			}
			wlNotFoundNames = lo.Keys(lo.PickBy(deleteList, func(key string, value DeletedWorkload) bool {
				return !value.found
			}))
			wlFoundNames = lo.Keys(lo.PickBy(deleteList, func(key string, value DeletedWorkload) bool {
				return value.found
			}))
			if len(wlNotFoundNames) > 0 {
				sp.Stop()
				cmd.Println(fmt.Sprintf(WorkloadsNotFound, deleteOptions.Project, wlNotFoundNames))
			}
			foundUuids := lo.FilterMap(lo.Entries(deleteList), func(entry lo.Entry[string, DeletedWorkload], _ int) (string, bool) {
				return entry.Value.uuid, entry.Value.found
			})
			res, err := bulkDelete(foundUuids)
			if err != nil {
				sp.Stop()
				return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToBulkDeletePre, err)
			}

			// in case response is missing, should not reach this, but just for the case it does.
			if res == nil {
				sp.Stop()
				//nolint:staticcheck // user facing message
				return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToBulkDeletePre, errors.New(FailedMissingResponse))
			}

			if len(res.Failed) > 0 {
				sp.Stop()
				cmd.Println(FailedToBulkDeleteSomeWorkloads)
				for _, f := range res.Failed {
					cmd.Println(fmt.Sprintf("ID: %s, Code:%d, Message:%s", f.Id, f.Code, f.Message))
				}
				return nil
			}
		}
		sp.Stop()

		if len(wlFoundNames) == 1 {
			cmd.Println(fmt.Sprintf(SingleDeleteSuccess, wlFoundNames[0]))
		} else if len(wlNotFoundNames) > 0 {
			cmd.Println(BulkDeleteSuccessNoFound)
		} else {
			cmd.Println(BulkDeleteSuccess)
		}

		return nil
	},
}

func confirmWorkloadDeletion(args []string) (bool, error) {
	if deleteOptions.isForceYes {
		return true, nil
	}
	return template.YesNoQuestion(fmt.Sprintf(BulkConfirmationMessage, len(args)))
}

func init() {
	Cmd.AddCommand(Delete)

	Delete.Flags().StringVar(&deleteOptions.workloadType, consts.WorkloadType, "", consts.AttributeTypeUsage)
	Delete.Flags().StringVar(&deleteOptions.framework, consts.WorkloadFramework, "", consts.AttributeFrameworkUsage)
	Delete.Flags().StringVarP(&deleteOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	Delete.Flags().BoolVarP(&deleteOptions.isForceYes, AttributeNameYes, "y", false, AttributeUsageYes)
	_ = Delete.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	Delete.MarkFlagsMutuallyExclusive(consts.WorkloadType, consts.WorkloadFramework)

	_ = Delete.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
	_ = Delete.RegisterFlagCompletionFunc(consts.WorkloadFramework, commWorkload.FrameworkFlagCompletion)
}

func getDeleteList(deleteOptions DeleteOptions, args []string) (map[string]DeletedWorkload, error) {
	var foundAny bool
	deleteList := make(map[string]DeletedWorkload)
	for _, arg := range args {
		deleteList[arg] = DeletedWorkload{found: false}
	}
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()
	filterBuilder, err := GetDeleteFilterBuilder(deleteOptions, cpUrl, token, clusterId)
	for chunk := range slices.Chunk(args, maxFilterSize) {
		if err != nil {
			return nil, err
		}
		for _, name := range chunk {
			filterBuilder.AppendName(name)
		}
		briefWorkloadHandler, err := workload.GetWorkloadPagination(txn, cpUrl, token, apilist.NewFilterPageOption(), filterBuilder, false)
		if err != nil {
			return nil, err
		}
		for _, resWorkload := range briefWorkloadHandler.GetRawList() {
			if entry, exists := deleteList[resWorkload.Name]; exists {
				foundAny = true
				entry.found = true
				entry.uuid = resWorkload.Id
				deleteList[resWorkload.Name] = entry
			}
		}
		filterBuilder.ClearNames()
	}
	if !foundAny {
		return nil, fmt.Errorf("no workload was found")
	}
	return deleteList, nil
}

func GetDeleteFilterBuilder(deleteOptions DeleteOptions, cpUrl string, token string, clusterId string) (*filter.Builder, error) {
	_, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, deleteOptions.Project)
	if err != nil {
		return nil, err
	}
	fb := filter.NewBuilder().WithProject(deleteOptions.Project).WithClusterID(clusterId)

	if deleteOptions.workloadType != "" {
		workloadType, err := api.ParseWorkloadType(strings.ToLower(deleteOptions.workloadType))
		if err != nil {
			return nil, err
		}
		fb.WithWorkloadType(workloadType.QueryParam())
	}

	if deleteOptions.framework != "" {
		workloadFramework, err := api.ParseWorkloadFramework(deleteOptions.framework)
		if err != nil {
			return nil, err
		}
		fb.WithWorkloadFramework(workloadFramework)
	}
	return fb, nil
}

func bulkDelete(uuidList []string) (*openapi.BatchResponse, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()

	return workload.DeleteBulkByIds(txn, cpUrl, token, uuidList)
}
