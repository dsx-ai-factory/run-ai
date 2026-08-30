package reporter

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	ReportTimer = "report_timer"
)

var (
	instance = config.GetInstance()
	txn      *transaction.Txn
)

func ReportCmd(cmd *cobra.Command, err error) error {
	cmd.Root().SilenceErrors = true
	txn = instance.GetTxn()
	txn.SetError(err)
	return Report(cmd, []string{})
}

func Report(cmd *cobra.Command, _ []string) error {
	logger := instance.GetLogger()
	txn = instance.GetTxn()

	// we want to stop the timer as soon as the transaction finished the command, what happens from here is CLI time, not transaction
	txn.StopTxnTimer()
	rtId := txn.AddTimer(ReportTimer)
	txn.SetCmd(cmd.CommandPath())

	if txn == nil {
		logger.Debug("no transaction found, bypassing metrics reporting")
		return fmt.Errorf("no transaction found, bypassing metrics reporting")
	}

	txnConfig, err := instance.GetTxnConfig()
	if err != nil {
		logger.Debug("no metrics config found, bypassing metrics reporting")
		return fmt.Errorf("no metrics config found, bypassing metrics reporting")
	}

	if !txnConfig.ReportEnabled() || txn.ShouldBypass() {
		return txn.GetError()
	}

	reporter, err := Factory(txnConfig)
	if err != nil {
		logger.Debug("failed to report metrics", zap.Error(err))
		return nil
	}

	txn.StopTimer(rtId)
	err = txn.Report(reporter)
	if err != nil {
		logger.Debug("failed to report metrics", zap.Error(err))
	}
	return txn.GetError()
}
