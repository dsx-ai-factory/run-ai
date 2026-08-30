package workload

import (
	"fmt"

	workloadcommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	workloadSubmit "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/submit"
	"github.com/spf13/cobra"
)

var (
	submitOptions workloadSubmit.SubmitOptions
)

var Submit = &cobra.Command{
	Use:     submitUse,
	Short:   submitShortMsg,
	Long:    submitLongMsg,
	Example: submitExample,
	Args:    cobra.MaximumNArgs(1),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			err := argshandler.BothAreSet(true, submitOptions.NamePrefix != "")
			if err != nil {
				return err
			}
			submitOptions.Name = args[0]
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {

		workloadName, workloadKind, err := workloadSubmit.Submit(txn, instance, &submitOptions)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSubmit, err)
		}

		cmd.Println(fmt.Sprintf(successWorkloadSubmitted, workloadKind, workloadName))
		return nil
	},
}

func init() {
	Cmd.AddCommand(Submit)

	Submit.Flags().StringVarP(&submitOptions.File, flagFile, consts.FileShort, "", flagFileUsage)
	Submit.Flags().StringVarP(&submitOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	Submit.Flags().StringVar(&submitOptions.Priority, flagPriority, "", consts.AttributeUsagePriorityClass)
	Submit.Flags().StringVar(&submitOptions.Category, flagCategory, "", flagCategoryUsage)
	Submit.Flags().StringVar(&submitOptions.NamePrefix, flagNamePrefix, "", flagNamePrefixUsage)
	Submit.Flags().Var(cli.NewNullablePreemptibilityFlagWorkloadV2(&submitOptions.Preemptibility),
		consts.AttributeNamePreemptibility,
		consts.AttributeUsagePreemptibility)
	Submit.Flags().Var(cli.NewNullableMnnvlModeFlagWorkloadV2(&submitOptions.Mnnvl),
		flagMnnvl,
		flagMnnvlUsage)

	_ = Submit.MarkFlagRequired(flagFile)
	_ = Submit.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
	_ = Submit.RegisterFlagCompletionFunc(consts.AttributeNamePreemptibility, workloadcommon.PreemptibilityFlagCompletion)
	_ = Submit.RegisterFlagCompletionFunc(flagMnnvl, workloadcommon.MnnvlFlagCompletion)
}
