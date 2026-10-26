/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	authflow "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/auth_flow"
)

type AuthFlowService interface {
	CreateAuthFlow(ctx context.Context, realm string, flow models.AuthenticationFlowRepresentation) error
	GetAuthFlow(ctx context.Context, realm string, flowAlias string) (*models.AuthenticationFlowRepresentation, error)
	GetAuthExecutions(ctx context.Context, realm string, flowAlias string) ([]*models.ModifyAuthenticationExecutionRepresentation, error)
	CreateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionRepresentation) error
	UpdateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.ModifyAuthenticationExecutionRepresentation) error
	CreateAuthExecutionFlow(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionFlowRepresentation) error
}

type authFlowService struct {
	authFlowRepository authflow.AuthFlowRepository
}

func NewAuthFlowService(authFlowRepository authflow.AuthFlowRepository) AuthFlowService {
	return &authFlowService{
		authFlowRepository: authFlowRepository,
	}
}

// CreateAuthExecution implements AuthFlowService.
func (a *authFlowService) CreateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionRepresentation) error {
	return a.authFlowRepository.CreateAuthExecution(ctx, realm, flowAlias, execution)
}

// CreateAuthExecutionFlow implements AuthFlowService.
func (a *authFlowService) CreateAuthExecutionFlow(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionFlowRepresentation) error {
	return a.authFlowRepository.CreateAuthExecutionFlow(ctx, realm, flowAlias, execution)
}

// CreateAuthFlow implements AuthFlowService.
func (a *authFlowService) CreateAuthFlow(ctx context.Context, realm string, flow models.AuthenticationFlowRepresentation) error {
	return a.authFlowRepository.CreateAuthFlow(ctx, realm, flow)
}

// GetAuthExecutions implements AuthFlowService.
func (a *authFlowService) GetAuthExecutions(ctx context.Context, realm string, flowAlias string) ([]*models.ModifyAuthenticationExecutionRepresentation, error) {
	return a.authFlowRepository.GetAuthExecutions(ctx, realm, flowAlias)
}

// GetAuthFlow implements AuthFlowService.
func (a *authFlowService) GetAuthFlow(ctx context.Context, realm string, flowAlias string) (*models.AuthenticationFlowRepresentation, error) {
	return a.authFlowRepository.GetAuthFlow(ctx, realm, flowAlias)
}

// UpdateAuthExecution implements AuthFlowService.
func (a *authFlowService) UpdateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.ModifyAuthenticationExecutionRepresentation) error {
	return a.authFlowRepository.UpdateAuthExecution(ctx, realm, flowAlias, execution)
}
