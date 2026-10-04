/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	auditconfig "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/write"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

type contextField string

const (
	cfAuditLog contextField = "audit-log"
)

var (
	once                         sync.Once
	singleInstance               Writer = &write.EmptyWriter{}
	isInitialized                bool
	enablePrintAuditLogsToStdout bool
)

type Writer interface {
	Write(log openapi.AuditLog) error
}

func CtxWithAudit(ctx context.Context, al *AuditLog) context.Context {
	return context.WithValue(ctx, cfAuditLog, al)
}

func AuditFromCtx(ctx context.Context) *AuditLog {
	if ctx.Value(cfAuditLog) == nil {
		return &AuditLog{}
	}
	return ctx.Value(cfAuditLog).(*AuditLog)
}

func AuditFromAudit(al *AuditLog) *AuditLog {
	var copied openapi.AuditLog
	b, _ := json.Marshal(al.Log)
	_ = json.Unmarshal(b, &copied)
	return &AuditLog{Log: copied}
}

func Init(conf *auditconfig.Config) error {
	var err error
	once.Do(func() {
		if err = auditconfig.InitConfig(conf); err != nil {
			return
		}

		c := config.Get[*auditconfig.Config]()
		if !c.Enabled {
			return
		}

		enablePrintAuditLogsToStdout = c.EnablePrintAuditLogsToStdout

		singleInstance, err = createWriter()
		if err == nil {
			isInitialized = true
		}
	})
	if err != nil {
		logging.Errorf("failed initializing audit writer. err: %+v", err)
	}
	return err
}

func IsInitialized() bool {
	return isInitialized
}

func SetMock(writer Writer) {
	singleInstance = writer
}

func createWriter() (Writer, error) {
	apiWriter := write.NewAPIWriter()

	pubsubWriter, err := write.NewPubSubWriter()
	if err != nil {
		logging.Warnf("Failed initializing pubsub writer, using only API writer. error: %v", err)
		return apiWriter, fmt.Errorf("failed initializing pubsub writer, using only API writer. error: %w", err)
	}

	return write.NewFallbackWriter(pubsubWriter, apiWriter), nil
}
