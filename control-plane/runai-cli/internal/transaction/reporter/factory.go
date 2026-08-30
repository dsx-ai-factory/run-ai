package reporter

import (
	"fmt"

	config2 "github.com/run-ai/runai/control-plane/runai-cli/internal/transaction/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

func Factory(txnConf *config2.Txn) (transaction.Reporter, error) {
	switch txnConf.ReportType() {
	case attributes.Logger:
		return NewLoggerReporter(config.GetInstance().GetLogger()), nil
	case attributes.Local:
		return NewLocalReporter(txnConf.GetLocation(), txnConf.GetSize(), txnConf.GetFiles(), txnConf.GetMaxAge()), nil
	default:
		return nil, fmt.Errorf("invalid reporter type")
	}
}
