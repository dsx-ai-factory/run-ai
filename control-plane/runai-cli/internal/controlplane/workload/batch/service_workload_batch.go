/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package batch

import (
	"context"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type Batcher interface {
	BatchDelete(ctx context.Context, ids ...string) (*openapi.BatchResponse, error)
}

type Service struct {
	workloadClient Batcher
}

func NewService(workloadClient Batcher) *Service {
	return &Service{workloadClient: workloadClient}
}

func (s *Service) Delete(ctx context.Context, ids ...string) (*openapi.BatchResponse, error) {
	batch, err := s.workloadClient.BatchDelete(ctx, ids...)
	if err != nil {
		return nil, err
	}
	return batch, nil
}
