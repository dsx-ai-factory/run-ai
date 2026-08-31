/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
)

type Ingestor struct {
	handlerChain Handler
}

func NewIngestor(batcher *Batcher[models.AuditLog], consumer MessageConsumer) (*Ingestor, error) {
	handler, err := NewMessageConsumerHandler(NewDbSinkHandler(batcher), consumer)
	if err != nil {
		return nil, err
	}
	return &Ingestor{
		handlerChain: handler,
	}, nil
}

func (i *Ingestor) Stop() error {
	return i.handlerChain.Stop()
}
