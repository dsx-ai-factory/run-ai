package reporter

import (
	"encoding/json"
	"log"

	"github.com/natefinch/lumberjack"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

type LocalReporter struct {
}

func NewLocalReporter(location string, size, backups, age int) *LocalReporter {
	log.SetOutput(&lumberjack.Logger{
		Filename:   location,
		MaxSize:    size, // megabytes
		MaxBackups: backups,
		MaxAge:     age, // days
	})

	return &LocalReporter{}
}

func (lr *LocalReporter) Report(t *transaction.Txn) error {
	infoMarshal, err := json.Marshal(t.GetInfo())
	if err != nil {
		return err
	}
	metricsMarshal, err := json.Marshal(t.GetMetrics())
	if err != nil {
		return err
	}
	log.Printf("info=%s, metrics=%s", string(infoMarshal), string(metricsMarshal))
	return nil
}
