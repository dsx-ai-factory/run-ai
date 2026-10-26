/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package write

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

type writer interface {
	Write(log openapi.AuditLog) error
}

type FallbackWriter struct {
	primary  writer
	fallback writer
}

func NewFallbackWriter(primary, fallback writer) *FallbackWriter {
	return &FallbackWriter{primary: primary, fallback: fallback}
}

func (fw *FallbackWriter) Write(log openapi.AuditLog) error {
	err := fw.primary.Write(log)
	if err != nil {
		logging.Errorf("Primary writer failed, falling back to secondary writer: %v", err)
		return fw.fallback.Write(log)
	}
	return nil
}
