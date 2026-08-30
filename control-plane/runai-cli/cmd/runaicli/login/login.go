package login

import (
	"fmt"

	commAttr "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attributes"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/network"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	upgradecmd "github.com/run-ai/runai/control-plane/runai-cli/internal/upgrade/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	charLimitation    = 64
	browserTimerName  = "browser_login"
	listenHost        = "listen-host"
	listenPort        = "listen-port"
	defaultListenPort = 43121
	defaultListenHost = "localhost"
)

type browserSettings struct {
	host      string
	port      int
	address   string
	printOnly bool
}

var (
	cmdHelper    helper.Cmd
	instance     = config.GetInstance()
	txn          *transaction.Txn
	authFilePath string
	browser      browserSettings
)

// Cmd represents the login command
var Cmd = &cobra.Command{
	Use:     "login",
	Short:   consts.LoginShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupUserManagementID,
	Example: consts.ExampleLogin,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		authFilePath = instance.GetAuthenticationPath()

		cmdHelper.AddValidation(validators.NewAuthValidator(instance.GetConfigValue(config.AuthenticationUrl)))
		cmdHelper.AddValidation(validators.NewCpValidator(instance.GetConfigValue(config.ControlPlaneUrl)))
		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}
		return nil
	},

	// PreRunE will only be executed the interactive mode if the user has not set the login method
	PreRunE: func(cmd *cobra.Command, args []string) error {
		browser.address = fmt.Sprintf("%s:%d", browser.host, browser.port)
		cmdHelper.AddValidation(network.NewAddressValidator(browser.address))
		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		authenticator, err := auth.InitBrowserLogin(instance.GetConfigValue(config.AuthenticationUrl), browser.address, browser.printOnly)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LoginFailed, err)
		}

		err = auth.Login(txn, browserTimerName, authenticator, authFilePath)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, err)
		}
		cmd.Println(consts.SuccessfullyLoggedIn)
		return nil
	},

	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		tokenStr, err := auth.GetToken(txn.GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
		}

		setClusterIfNeeded(cmd, txn, cpUrl, tokenStr)
		outputFormat := cli.GetOutputTypeFromFlags(cmd)
		upgradecmd.CheckForUpdateFlow(cmd, outputFormat)
		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&browser.host, listenHost, defaultListenHost, consts.ListenHostUsage)
	Cmd.Flags().IntVar(&browser.port, listenPort, defaultListenPort, consts.ListenPortUsage)
	Cmd.Flags().BoolVar(&browser.printOnly, consts.AttributeNamePrintOnlyLogin, false, consts.AttributePrintOnlyLoginUsage)
	_ = Cmd.Flags().MarkHidden(consts.AttributeNamePrintOnlyLogin)
}

func setClusterIfNeeded(cmd *cobra.Command, txn *transaction.Txn, cpUrl, tokenStr string) {
	logger := config.GetInstance().GetLogger()
	handler, err := cluster.GetClusterList(txn, cpUrl, tokenStr)
	if err != nil {
		logger.Debug(consts.FailedToListClusters, zap.String("error", err.Error()))
		return
	}

	clusterList := handler.GetList()
	if len(clusterList) != 1 {
		logger.Debug(consts.MoreOrLessThanOneCluster)
		return
	}

	currentCluster := instance.GetConfigValue(config.ClusterUUID)
	candidateCluster := clusterList[0]
	if currentCluster == candidateCluster.Uuid {
		logger.Debug(consts.ClusterAlreadySet)
		return
	}

	err = instance.SetCluster(candidateCluster.Uuid, candidateCluster.Name, candidateCluster.Domain)
	if err != nil {
		logger.Debug(consts.FailedToSetCluster, zap.String("error", err.Error()))
		return
	}
	cmd.Println(fmt.Sprintf(consts.SingleClusterDefined, candidateCluster.Name))
}
