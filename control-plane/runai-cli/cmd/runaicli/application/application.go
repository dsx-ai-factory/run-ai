package application

import (
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
	txn       *transaction.Txn
)

var Cmd = &cobra.Command{
	Use:     "application",
	Aliases: []string{"applications"},
	Short:   AIApplicationShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupApplicationID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := cmdfactory.InstanceInit(cmd); err != nil {
			return err
		}
		txn = instance.GetTxn()
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}
