/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_flow

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	gocloakInterface "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	helpers "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/auth_flow/helpers"
)

type KeycloakAuthFlowRepository struct {
	auth     auth.Auth
	kcClient gocloakInterface.GoCloak
}

func NewKeycloakAuthRepository(client *auth.KeycloakAuthenticator) *KeycloakAuthFlowRepository {
	return &KeycloakAuthFlowRepository{
		auth:     client,
		kcClient: client.KcClient,
	}
}

func (k *KeycloakAuthFlowRepository) CreateAuthFlow(ctx context.Context, realm string, flow models.AuthenticationFlowRepresentation) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}
	kcFlow := helpers.FromAuthFlowRepModel(&flow)
	err = k.kcClient.CreateAuthenticationFlow(ctx, *token, realm, *kcFlow)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakAuthFlowRepository) GetAuthFlow(ctx context.Context, realm string, flowAlias string) (*models.AuthenticationFlowRepresentation, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, err
	}
	authFlows, err := k.kcClient.GetAuthenticationFlows(ctx, *token, realm)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	if len(authFlows) == 0 {
		return nil, nil
	}

	for _, flow := range authFlows {
		if *flow.Alias == flowAlias {
			return helpers.ToAuthFlowRepModel(flow), nil
		}
	}
	return nil, nil
}

func (k *KeycloakAuthFlowRepository) GetAuthExecutions(ctx context.Context, realm string, flowAlias string) ([]*models.ModifyAuthenticationExecutionRepresentation, error) {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, err
	}
	executions, err := k.kcClient.GetAuthenticationExecutions(ctx, *token, realm, flowAlias)
	if err != nil {
		return nil, errors.HandleKeyCloakError(err)
	}
	var modifyAuthRep []*models.ModifyAuthenticationExecutionRepresentation
	for _, exec := range executions {
		modifyAuthRep = append(modifyAuthRep, helpers.ToModifyAutheExecutionRepModel(exec))
	}
	return modifyAuthRep, nil
}

func (k *KeycloakAuthFlowRepository) CreateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionRepresentation) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}
	err = k.kcClient.CreateAuthenticationExecution(ctx, *token, realm, flowAlias, gocloak.CreateAuthenticationExecutionRepresentation{
		Provider: execution.Provider,
	})
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakAuthFlowRepository) UpdateAuthExecution(ctx context.Context, realm string, flowAlias string, execution models.ModifyAuthenticationExecutionRepresentation) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}
	kcExecution := helpers.FromModifyAutheExecutionRepModel(&execution)
	err = k.kcClient.UpdateAuthenticationExecution(ctx, *token, realm, flowAlias, *kcExecution)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}

func (k *KeycloakAuthFlowRepository) CreateAuthExecutionFlow(ctx context.Context, realm string, flowAlias string, execution models.CreateAuthenticationExecutionFlowRepresentation) error {
	token, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return err
	}
	kcExecution := helpers.FromCreateAutheExecutionRepModel(&execution)
	err = k.kcClient.CreateAuthenticationExecutionFlow(ctx, *token, realm, flowAlias, *kcExecution)
	if err != nil {
		return errors.HandleKeyCloakError(err)
	}
	return nil
}
