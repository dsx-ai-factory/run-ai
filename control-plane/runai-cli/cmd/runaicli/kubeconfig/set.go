package kubeconfig

import (
	kubeconfigutils "github.com/run-ai/runai/control-plane/runai-cli/internal/kubeconfig"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"

	commAttr "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attributes"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

const (
	flagNameUsername  = "username"
	flagUsageUsername = "define the user key name in the kubeconfig"

	runaiCliUserName = "runai-cli"
)

var (
	username string
)

var set = &cobra.Command{
	Use:   "set",
	Short: consts.KubeConfigSetShortMsg,
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		instance := config.GetInstance()

		txn := instance.GetTxn()
		tokenStr, err := auth.GetToken(txn.GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
		}
		instance.SetToken(tokenStr)

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := kubeconfigutils.GetConfigRealPath()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdateKubeConfigBadCFile, err)
		}
		merger, err := kubeconfigutils.GetConfigMerger(path)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdateKubeConfigBadCFile, err)
		}

		handler, err := kubeconfigutils.NewHandler(merger, path, username)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdateKubeConfigBadCFile, err)
		}
		err = updateKubeConf(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdateKubeConfigBadCFile, err)
		}

		cmd.Println(consts.KubeConfigSetSuccessfully)
		return nil
	},
}

func init() {
	Cmd.AddCommand(set)

	set.Flags().StringVar(&username, flagNameUsername, runaiCliUserName, flagUsageUsername)
}

func updateKubeConf(handler *kubeconfigutils.Handler) error {
	handler.SetTokenExecCmd()

	err := handler.SetContext()
	if err != nil {
		return err
	}
	return handler.Update()
}
