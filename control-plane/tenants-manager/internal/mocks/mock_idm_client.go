/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mocks

import (
	"context"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
)

type MockIdmClient struct{}

func NewMockIdmClient() *MockIdmClient { return &MockIdmClient{} }

func (n *MockIdmClient) MintInstallerBootstrapToken(_ context.Context, _ string, _ string) (string, error) {
	return "", nil
}

func (n *MockIdmClient) GetInstallerServiceAccount(_ context.Context) (*openapi.ServiceAccountPostResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) GetInstallerServiceAccountForTenantId(_ context.Context, _ string) (*openapi.ServiceAccountPostResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) GetInternalClientByFilter(_ context.Context, _ string) ([]openapi.Client, error) {
	return nil, nil
}

func (n *MockIdmClient) AppendRedirectUriToInternalClient(_ context.Context, _ openapi.Client, _ string) error {
	return nil
}

func (n *MockIdmClient) DeleteRealm(_ context.Context, _ string, _ bool) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func (n *MockIdmClient) CreateRealm(_ context.Context, _ string, _ string) (*openapi.RealmCreationResponse, *http.Response, error) {
	return nil, &http.Response{StatusCode: http.StatusOK}, nil
}

func (n *MockIdmClient) CreateInternalUser(_ context.Context, _ openapi.InternalUserCreationRequest) (*openapi.UserPostResponse, *http.Response, error) {
	return nil, &http.Response{StatusCode: http.StatusCreated}, nil
}

func (n *MockIdmClient) CreateInternalClient(_ context.Context, _ openapi.ClientCreationRequest) (*openapi.CreateInternalClient201Response, *http.Response, error) {
	return nil, nil, nil
}

func (n *MockIdmClient) DeleteClient(_ context.Context, _ string) (*http.Response, error) {
	return nil, nil
}

func (n *MockIdmClient) CreateClient(_ context.Context, _ string, _ string, _ *string) (*string, error) {
	return nil, nil
}

func (n *MockIdmClient) CreateServiceAccountInternally(_ context.Context, _ string, _ string) (*openapi.ServiceAccountPostResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) ClientCredentialsLogin(_ context.Context, _ string, _ string) (*openapi.TokenResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) ClientCredentialsLoginV2(_ context.Context, _ string, _ string) (*openapi.OAuth2TokenResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) TokenExchange(_ context.Context, _ string, _ string, _ string, _ string) (*openapi.TokenResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) RefreshToken(_ context.Context, _ string, _ string) (*openapi.TokenResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) UserLogin(_ context.Context, _ string, _ string, _ string) (*openapi.TokenResponse, error) {
	return nil, nil
}

func (n *MockIdmClient) DeleteInternalUserByEmail(_ context.Context, _ string, _ string) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func (n *MockIdmClient) GetOrCreateRealm(_ context.Context, _ string, _ string) (*openapi.RealmCreationResponse, *http.Response, error) {
	return nil, &http.Response{StatusCode: http.StatusOK}, nil
}
