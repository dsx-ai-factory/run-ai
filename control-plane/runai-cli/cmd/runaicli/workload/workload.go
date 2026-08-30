package workload

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"

	token "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cmdHelper    helper.Cmd
	instance     = config.GetInstance()
	txn          *transaction.Txn
	WorkloadType string
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "workload",
	Short:   consts.WorkloadShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupWorkloadID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initWorkloadInstance(cmd)
		if err != nil {
			// err is already formatted
			return err
		}
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func initWorkloadInstance(cmd *cobra.Command) error {
	instance = config.GetInstance()

	txn = instance.GetTxn()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterUrl := instance.GetConfigValue(config.ClusterDomain)
	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))
	cmdHelper.AddValidation(validators.NewClusterValidator(clusterUrl))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}

	tokenStr, err := token.GetToken(txn.GetCtx())
	if err != nil {
		return err
	}
	instance.SetToken(tokenStr)

	return nil
}
