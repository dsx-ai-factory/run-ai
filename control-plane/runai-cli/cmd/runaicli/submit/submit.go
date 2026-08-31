/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"os"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/submit"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/spf13/cobra"
)

const (
	AttributeNameFlag = "name"
	BCPvcFlag         = "pvc"
	BCCustomUrlFlag   = "custom-url"
	BCServiceTypeFlag = "service-type"
)

var (
	SubmitSpec    = spec.NewSubmitSpec()
	BCStylePvcUse []string
	BCServiceType []string
	BCCustomUrl   string
	workloadName  string
)

// Cmd represents the submit command
// Deprecated: Use the specific submit commands instead
var Cmd = &cobra.Command{
	Use:        "submit",
	Short:      consts.SubmitShortMsg,
	GroupID:    consts.GroupDeprecatedID,
	Deprecated: consts.DepSubmitMsg,

	PersistentPreRunE: workload.Cmd.PersistentPreRunE,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// backward compatibility
		// this for backward compatibility command (`runai submit --interactive ....`)
		if SubmitSpec.Interactive {
			SubmitSpec.SetType(spectype.Workspace)
		} else {
			SubmitSpec.SetType(spectype.Training)
		}

		if workloadName != "" {
			// if name is not empty, add it to the args list as the first argument
			args = append([]string{workloadName}, args...)
			os.Args = append([]string{os.Args[0], os.Args[1], workloadName}, os.Args[2:]...)
		}

		// using temporary attribute, the function changes the list, and we don't want this at this point.
		tempArgs := args
		// yes this is duplicate inside submit.Pre, but this code one day will be removed.
		name, err := argshandler.ValidateArgsAndExtractName(cmd, &tempArgs)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}
		err = SubmitSpec.BackwardCompatabilityMutate(name, BCStylePvcUse, BCServiceType, BCCustomUrl)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}
		submit.Spec = &SubmitSpec.GeneralSpec
		return submit.Pre(cmd, args)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return submit.Run(cmd)
	},
}

func init() {
	// workaround to support backward-compatible `name` flag (--name)
	Cmd.Flags().StringVar(&workloadName, AttributeNameFlag, "", "Job name")

	// Deprecated flags
	Cmd.Flags().StringArrayVar(&BCStylePvcUse, BCPvcFlag, nil, "[Deprecated] To mount a persistent volume")
	_ = Cmd.Flags().MarkDeprecated(BCPvcFlag, "To mount a persistent volume,"+
		" use --existing-pvc or --new-pvc to create a new persistent volume")
	Cmd.Flags().StringVar(&BCCustomUrl, BCCustomUrlFlag, "",
		"[Deprecated] Custom URL to use as the external URL for the job."+
			" To be used with --service-type (if not provided there)."+
			" If not provided, the system will generate a URL automatically")
	_ = Cmd.Flags().MarkDeprecated(BCCustomUrlFlag, "To set a custom URL, use external-url flag")
	Cmd.Flags().StringArrayVar(&BCServiceType, BCServiceTypeFlag, nil,
		"[Deprecated] External access type to jobs. One of: nodeport, external-url."+
			" Supports more than one service,"+
			" e.g: --service-type nodeport,port=31000:8080 --service-type external-url,port=8888")
	_ = Cmd.Flags().MarkDeprecated(BCServiceTypeFlag, "To set external access,"+
		" use port or external-url.")

	// Add flags (all flags including the backward compatibility flags)
	SubmitSpec.SetFlagExtended(Cmd, &submit.Project)

	// Mark flags as required if needed
	Cmd.MarkFlagsMutuallyExclusive(consts.AttributeNameRequiredPodAffinityKey, consts.AttributeNamePreferredPodAffinityKey)

	_ = Cmd.MarkFlagRequired(consts.AttributeNameImage)
}
