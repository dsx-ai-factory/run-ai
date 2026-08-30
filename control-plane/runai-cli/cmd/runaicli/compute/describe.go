package compute

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/describe"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type DescribeOptions struct {
	Name    string
	Project string
	Output  string
}

var (
	describeOptions DescribeOptions
)

var DescribeCmd = &cobra.Command{
	Use:     "describe [COMPUTE_NAME]",
	Short:   DescribeShortMsg,
	Aliases: []string{"desc"},
	Args:    cobra.ExactArgs(1),
	Example: describeExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		describeOptions.Name = args[0]
		if describeOptions.Project == "" {
			describeOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
		}

		cmdHelper.AddValidation(validators.NewProjectValidator(describeOptions.Project))
		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		txn := instance.GetTxn()

		projectId, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, describeOptions.Project)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeCompute, fmt.Errorf("failed to get project ID for '%s': %w", describeOptions.Project, err))
		}

		output, err := describe.Describe(txn, cpUrl, token, describeOptions.Name, projectId, clusterId)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeCompute, err)
		}

		if err := describe.RenderOutput(output, describeOptions.Output); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeCompute, err)
		}
		return nil
	},
}

func init() {
	Cmd.AddCommand(DescribeCmd)

	// Register flags
	DescribeCmd.Flags().StringVarP(&describeOptions.Output, "output", "o", "table",
		consts.DescribeOutputUsage)
	DescribeCmd.Flags().StringVarP(&describeOptions.Project, consts.ProjectStr, consts.ProjectShort, "",
		ProjectFlagUsage)

	// Register auto-completion for flags
	_ = DescribeCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = DescribeCmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
}
