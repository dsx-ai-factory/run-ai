package helper

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"
)

const (
	InteractiveMetricName = "interactive"
	InteractiveEnabled    = 1
	InteractiveDisabled   = 0
)

func NewCmd() *Cmd {
	return &Cmd{}
}

type Cmd struct {
	validationList []validators.Validation
}

func (m *Cmd) AddValidation(valid validators.Validation) {
	m.validationList = append(m.validationList, valid)
}

func (m *Cmd) Clear() {
	m.validationList = []validators.Validation{}
}

func (m *Cmd) Check() (string, error) {
	defer m.clearValidation()
	for _, valid := range m.validationList {
		if resp, err := valid.Execute(); err != nil {
			return resp, err
		}
	}
	return "", nil
}

func (m *Cmd) clearValidation() {
	m.validationList = []validators.Validation{}
}

func (m *Cmd) CheckInteractive(txn *transaction.Txn, args attributes.EnableFlag, mode config.InteractiveMode) bool {
	if args != "" {
		if args == attributes.EnableTrue {
			txn.SetCounter(InteractiveMetricName, InteractiveEnabled)
			return true
		}
		txn.SetCounter(InteractiveMetricName, InteractiveDisabled)
		return false
	}
	if mode == config.InteractiveEnabled {
		txn.SetCounter(InteractiveMetricName, InteractiveEnabled)
		return true
	}
	txn.SetCounter(InteractiveMetricName, InteractiveDisabled)
	return false
}
