/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"context"
)

type mockTracerProvider struct {
	shutdownCalled bool
}

func (m *mockTracerProvider) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func NewMockTracerProvider() *mockTracerProvider {
	return &mockTracerProvider{}
}

// IsShutdownCalled checks if the Shutdown method was called.
func (m *mockTracerProvider) IsShutdownCalled() bool {
	return m.shutdownCalled
}
