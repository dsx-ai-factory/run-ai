/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"fmt"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	wlpkg "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	listShortMsg   = "list inference workloads"
	listExampleMsg = `# List all inference workloads
runai inference list -A

# List workloads with default project
runai inference list

# List workloads in a specific project
runai inference list -p <project-name>

# List all workloads with a specific output format
runai inference list --yaml

# List workloads with pagination
runai inference list --limit 20 --offset 40`
	listFailedMsg = "inference list failed"
)

type ListOptions struct {
	output  listCommon.Output
	all     bool
	project string
	deleted bool
	page    apilist.PageOptions
}

var (
	listOptions ListOptions
	instance    = config.GetInstance()

	list = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   listShortMsg,
		Example: listExampleMsg,
		GroupID: consts.GroupInferenceBaseID,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if listOptions.page.Offset < 0 {
				return fmt.Errorf(apilist.ErrNegativeOffset, listOptions.page.Offset)
			}

			err := commonconfig.SetOutput(&listOptions.output)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
			}

			err = listOptions.output.RawFlagValidator()
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
			}

			err = workload.HandleColumnsFlags(listOptions.output, cmd)
			if err != nil {
				return err
			}

			if !listOptions.all && listOptions.project == "" {
				listOptions.project = instance.GetConfigValue(config.ClusterProjectName)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			txn := instance.GetTxn()
			cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
			token := instance.GetToken()
			clusterID := instance.GetConfigValue(config.ClusterUUID)

			// Validate project existence if project filter is specified (same as training list)
			if !listOptions.all && listOptions.project != "" {
				_, err := projectservice.GetProjectId(txn, cpUrl, token, clusterID, listOptions.project)
				if err != nil {
					return cmdtxn.CmdErrorPrintAndSet(cmd, listFailedMsg, err)
				}
			}

			// List standard inference, distributed inference, and NIM workload types
			fb := filter.NewBuilder().WithWorkloadType(api.Inference.QueryParam() + "||" + api.DistributedInference.QueryParam() + "||" + api.BackendNimServiceWorkloadType)
			if clusterID != "" {
				fb.WithClusterID(clusterID)
			}
			if !listOptions.all && listOptions.project != "" {
				fb.WithProject(listOptions.project)
			}

			handler, err := wlpkg.GetWorkloadPagination(txn, cpUrl, token, &listOptions.page, fb, listOptions.deleted)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, listFailedMsg, err)
			}

			var printOut string
			printOut, err = listOptions.output.Generate(handler)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, listFailedMsg, err)
			}
			cmd.Println(printOut)
			return nil
		},
	}
)

func init() {
	listCommon.FlagOutputType(list, &listOptions.output)
	listCommon.FlagOutputTypeWithColumns(list, &listOptions.output)
	list.Flags().BoolVarP(&listOptions.all, consts.ProjectAll, "A", false, consts.AttributeAllUsage)
	list.Flags().BoolVar(&listOptions.deleted, consts.DeletedWorkload, false, consts.AttributeDeletedWorkloads)
	list.Flags().StringVarP(&listOptions.project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = list.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
	apilist.SetPageFlags(list, &listOptions.page)
	list.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)
}
