/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import "context"

type DummyObservabilityHandler struct {
}

func NewDummyObservabilityHandler() *DummyObservabilityHandler {
	return &DummyObservabilityHandler{}
}

func (h *DummyObservabilityHandler) WithObservability(ctx context.Context, storeName string, key string, operationType OperationType, operation func(ctx context.Context) (bool, error)) error {
	_, err := operation(ctx)
	return err
}
