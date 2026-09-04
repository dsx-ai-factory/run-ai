/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reporter

import (
	"encoding/json"
	"log"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/natefinch/lumberjack"
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
