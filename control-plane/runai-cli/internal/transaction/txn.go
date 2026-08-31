/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package transaction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction/info"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction/metrics"
	"golang.org/x/crypto/sha3"
)

type Reporter interface {
	Report(t *Txn) error
}

type Txn struct {
	info    *info.Info
	metrics *metrics.Metrics
	bypass  bool
}

func NewTxn() *Txn {
	return &Txn{info: info.NewInfo(), metrics: metrics.NewMetrics()}
}

func (t *Txn) GetMetrics() *metrics.Metrics {
	return t.metrics
}

func (t *Txn) GetInfo() *info.Info {
	return t.info
}

func (t *Txn) StopTxnTimer() {
	t.metrics.StopTxnTimer()
}

func (t *Txn) AddTimer(name string) int {
	id := t.metrics.CreateTimer(name)
	return id
}

func (t *Txn) StopTimer(id int) {
	t.metrics.StopTimer(id)
}

func (t *Txn) SetCounter(name string, value int) {
	t.metrics.SetCounter(name, value)
}

func (t *Txn) SetInfo(info *info.Info) {
	t.info = info
}

func (t *Txn) SetCmd(cmd string) {
	if t.info == nil {
		t.info = info.NewInfo()
	}
	t.info.Cmd = cmd
}

func (t *Txn) SetError(err error) {
	t.info.SetError(err)
}

func (t *Txn) GetError() error {
	if t.info == nil {
		return fmt.Errorf("transaction info initialize failed")
	}
	return t.info.GetError()
}

func (t *Txn) GetCtx() context.Context {
	return t.info.GetCtx()
}

func (t *Txn) Report(reporter Reporter) error {
	t.finish()
	return reporter.Report(t)
}

func (t *Txn) BypassMetricReport() {
	t.bypass = true
}

func (t *Txn) ShouldBypass() bool {
	return t.bypass
}

func (t *Txn) finish() {
	t.calculateHash()
}

func (t *Txn) calculateHash() {
	hasher := sha3.New224()
	data, _ := json.Marshal(t.info)
	hasher.Write(data)
	data, _ = json.Marshal(t.metrics)
	hasher.Write(data)
}
