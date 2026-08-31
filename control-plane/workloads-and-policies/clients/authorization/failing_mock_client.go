/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization

import (
	"context"
	"net/http"

	authApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
)

// FailingMockClient is a special mock that always returns error. it is intended
// for testing error cases.
type FailingMockClient struct {
}

func (fmc *FailingMockClient) CheckPermittedScopes(_ /*ctx*/ context.Context, _ authApi.ResourceType, _ authApi.Action) (*authApi.PermittedScopes, *authApi.Error) {
	return nil, authApi.NewError(http.StatusServiceUnavailable, "mock: failed to communicate auth service")
}

// FailingMockFactory wraps FailingMockClient as a FactoryInterface for use with SetMock.
type FailingMockFactory struct{}

func (f *FailingMockFactory) Client() ClientInterface {
	return &FailingMockClient{}
}
