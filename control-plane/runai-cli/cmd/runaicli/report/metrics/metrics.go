package metrics

import (
	"path/filepath"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"

	"github.com/spf13/cobra"
)

const (
	Path        = "metrics"
	LogFileName = "metrics.log"
)

var (
	instance = config.GetInstance()
	txn      *transaction.Txn
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:   "metrics",
	Short: consts.MetricsShortMsg,
	Args:  cobra.NoArgs,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		txn = instance.GetTxn()
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func getMetricsLogFilePath() string {
	confPath := instance.GetOptions().ConfigPath
	configPath := filepath.Join(confPath, Path, LogFileName)
	return configPath
}

func init() {
}
