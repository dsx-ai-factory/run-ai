package list

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	token "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	List = "list"
)

var (
	instance  = config.GetInstance()
	cmdHelper helper.Cmd
)

// Cmd represents the list command
// Deprecated: Use the specific list commands instead
var Cmd = &cobra.Command{
	Use:        List,
	Short:      consts.ListShortMsg,
	Args:       cobra.NoArgs,
	GroupID:    consts.GroupDeprecatedID,
	Deprecated: consts.ListDepMsg,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		tokenStr, err := token.GetToken(instance.GetTxn().GetCtx())
		if err != nil {
			return err
		}
		instance.SetToken(tokenStr)
		if cmd.CalledAs() == List {
			return jobs.PersistentPreRunE(cmd, args)
		}
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == List {
			return jobs.PreRunE(cmd, args)
		}
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == List {
			return jobs.RunE(cmd, args)
		}
		return nil
	},
}

func init() {
	Cmd.Flags().BoolVarP(&workload.FlagsOptions.All, "all-projects", "A", false, consts.AttributeAllUsage)
	Cmd.Flags().StringVarP(&workload.FlagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
}
