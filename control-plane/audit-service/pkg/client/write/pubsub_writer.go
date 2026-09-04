/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package write

import (
	"context"
	"errors"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

type PubsubWriter struct{}

func NewPubSubWriter() (*PubsubWriter, error) {
	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		return nil, err
	}

	return &PubsubWriter{}, nil
}

func (psw *PubsubWriter) Write(log openapi.AuditLog) error {
	if !runaipubsub.IsInitialized() {
		return errors.New("pubsub is not initialized")
	}
	return runaipubsub.Publish(context.Background(), &log)
}
