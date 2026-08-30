package datavolume

import (
	token "github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
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
	Use:     "datavolume",
	Aliases: []string{"dv"},
	Short:   ShortMsg,
	GroupID: consts.GroupAssetsID,
	Args:    cobra.NoArgs,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initDatavolumeInstance(cmd)
		if err != nil {
			return err
		}
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func initDatavolumeInstance(cmd *cobra.Command) error {
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
