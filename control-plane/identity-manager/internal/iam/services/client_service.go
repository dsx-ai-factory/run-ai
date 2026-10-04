/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/clients"
)

type ClientService interface {
	CreateClient(ctx context.Context, realm string, client *models.Client) (string, error)
	GetClient(ctx context.Context, realm string, ID string) (*models.Client, error)
	UpdateClient(ctx context.Context, realm string, client *models.Client) error
	UpdateRedirectUris(ctx context.Context, realm string, ID string, redirectUris []string) error
	DeleteClient(ctx context.Context, realm string, ID string) error
	GetClients(ctx context.Context, realm string, filter *models.ClientFilter) ([]models.Client, error)
	GetClientSecret(ctx context.Context, realm string, ID string) (string, error)
	GetClientByClientId(ctx context.Context, realm string, clientId string) (*models.Client, error)
	CreateExchangePolicy(ctx context.Context, realm, idpAlias, targetClientID string) error
	GetClientsCount(ctx context.Context, realm string, filter *models.ClientFilter) (int, error)
	RegenerateClientSecret(ctx context.Context, realm string, ID string) (string, error)
	ToggleClient(ctx context.Context, realm string, ID string, enabled bool) error
	UpdateClientStatus(ctx context.Context, realm string, id string, isEnabled bool) error
	SetEmailScopeOptional(ctx context.Context, realm string, idOfClient string) error
	SetEmailScopeDefault(ctx context.Context, realm string, idOfClient string) error
	GetAccessKeyClientById(ctx context.Context, realm string, createdBy string, id string) (*models.Client, error)
	UpdatePostLogoutRedirectUris(ctx context.Context, realm string, clientId string, customUri string) error
	DeleteAccessKeysByEmail(ctx context.Context, realm string, email string) error
}

type clientService struct {
	clientRepository clients.ClientRepository
}

func NewClientService(clientRepository clients.ClientRepository) ClientService {
	return &clientService{
		clientRepository: clientRepository,
	}
}

func (c *clientService) CreateClient(ctx context.Context, realm string, client *models.Client) (string, error) {
	return c.clientRepository.CreateClient(ctx, realm, client)
}

func (c *clientService) UpdateClient(ctx context.Context, realm string, client *models.Client) error {
	return c.clientRepository.UpdateClient(ctx, realm, client)
}

func (c *clientService) UpdateRedirectUris(ctx context.Context, realm string, ID string, redirectUris []string) error {
	return c.clientRepository.UpdateRedirectUris(ctx, realm, ID, redirectUris)
}

func (c *clientService) DeleteClient(ctx context.Context, realm string, ID string) error {
	return c.clientRepository.DeleteClient(ctx, realm, ID)
}

func (c *clientService) GetClients(ctx context.Context, realm string, filter *models.ClientFilter) ([]models.Client, error) {
	return c.clientRepository.GetClients(ctx, realm, filter)
}

func (c *clientService) GetClientSecret(ctx context.Context, realm string, ID string) (string, error) {
	return c.clientRepository.GetClientSecret(ctx, realm, ID)
}

func (c *clientService) GetClientByClientId(ctx context.Context, realm string, clientId string) (*models.Client, error) {
	return c.clientRepository.GetClientByClientId(ctx, realm, clientId)
}

func (c *clientService) CreateExchangePolicy(ctx context.Context, realm string, idpAlias, targetClientID string) error {
	return c.clientRepository.CreateExchangePolicy(ctx, realm, idpAlias, targetClientID)
}

func (c *clientService) GetClientsCount(ctx context.Context, realm string, filter *models.ClientFilter) (int, error) {
	return c.clientRepository.GetClientsCount(ctx, realm, filter)
}

func (c *clientService) RegenerateClientSecret(ctx context.Context, realm string, ID string) (string, error) {
	return c.clientRepository.RegenerateClientSecret(ctx, realm, ID)
}

func (c *clientService) ToggleClient(ctx context.Context, realm string, ID string, enabled bool) error {
	return c.clientRepository.ToggleClient(ctx, realm, ID, enabled)
}

func (c *clientService) GetClient(ctx context.Context, realm string, ID string) (*models.Client, error) {
	return c.clientRepository.GetClient(ctx, realm, ID)
}

func (c *clientService) UpdateClientStatus(ctx context.Context, realm string, id string, isEnabled bool) error {
	return c.clientRepository.UpdateClientStatus(ctx, realm, id, isEnabled)
}

func (c *clientService) SetEmailScopeOptional(ctx context.Context, realm string, idOfClient string) error {
	return c.clientRepository.SetEmailScopeOptional(ctx, realm, idOfClient)
}

func (c *clientService) SetEmailScopeDefault(ctx context.Context, realm string, idOfClient string) error {
	return c.clientRepository.SetEmailScopeDefault(ctx, realm, idOfClient)
}

func (c *clientService) GetAccessKeyClientById(ctx context.Context, realm string, createdBy string, id string) (*models.Client, error) {
	return c.clientRepository.GetAccessKeyClientById(ctx, realm, createdBy, id)
}

func (c *clientService) UpdatePostLogoutRedirectUris(ctx context.Context, realm string, clientId string, customUri string) error {
	return c.clientRepository.UpdatePostLogoutRedirectUris(ctx, realm, clientId, customUri)
}

func (c *clientService) DeleteAccessKeysByEmail(ctx context.Context, realm string, email string) error {
	return c.clientRepository.DeleteAccessKeysByEmail(ctx, realm, email)
}
