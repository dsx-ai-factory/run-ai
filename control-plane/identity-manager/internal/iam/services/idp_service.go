/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"io"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/idps"
)

type IdpService interface {
	CreateIdp(ctx context.Context, realm string, idp models.Idp) (*string, error)
	GetIdps(ctx context.Context, realm string) ([]*models.Idp, error)
	GetIdp(ctx context.Context, realm string, idpAlias string) (*models.Idp, error)
	GetIdpAliasForExchangeToken(ctx context.Context, realm string) (*string, error)
	UpdateIdp(ctx context.Context, realm string, idp models.Idp) error
	DeleteIdp(ctx context.Context, realm string, idpAlias string) error
	GetIdpConfig(ctx context.Context, realm string, url string, providerId string) (map[string]string, error)
	GetIdpConfigFromFile(ctx context.Context, realm string, providerId string, fileName string, fileBody io.Reader) (map[string]string, error)
	SetBrowserRedirector(ctx context.Context, realm string, alias string) error
	GetSPSamlMetadataUrl(ctx context.Context, realm string, idpAlias string) string

	// idp mappers
	GetIdpMappers(ctx context.Context, realm, alias string) ([]*models.IdpMapper, error)
	UpdateIdpMappers(ctx context.Context, realm, alias string, mappers []*models.IdpMapper) error

	// idp settings
	SetMandatotyClaimMapper(ctx context.Context, realm, alias string, mapper *models.IdpMapper) error
	DeleteMandatoryClaimMapper(ctx context.Context, realm, alias, name string) error
}

type idpService struct {
	idpRepository idps.IdpRepository
}

func NewIdpService(idpRepository idps.IdpRepository) IdpService {
	return &idpService{
		idpRepository: idpRepository,
	}
}

func (i *idpService) CreateIdp(ctx context.Context, realm string, idp models.Idp) (*string, error) {
	return i.idpRepository.CreateIdp(ctx, realm, idp)
}

func (i *idpService) GetIdps(ctx context.Context, realm string) ([]*models.Idp, error) {
	return i.idpRepository.GetIdps(ctx, realm)
}

func (i *idpService) GetIdp(ctx context.Context, realm string, idpAlias string) (*models.Idp, error) {
	return i.idpRepository.GetIdp(ctx, realm, idpAlias)
}

func (i *idpService) UpdateIdp(ctx context.Context, realm string, idp models.Idp) error {
	return i.idpRepository.UpdateIdp(ctx, realm, idp)
}

func (i *idpService) DeleteIdp(ctx context.Context, realm string, idpAlias string) error {
	return i.idpRepository.DeleteIdp(ctx, realm, idpAlias)
}

func (i *idpService) GetIdpMappers(ctx context.Context, realm, alias string) ([]*models.IdpMapper, error) {
	return i.idpRepository.GetIdpMappers(ctx, realm, alias)
}

func (i *idpService) UpdateIdpMappers(ctx context.Context, realm, alias string, mappers []*models.IdpMapper) error {
	return i.idpRepository.UpdateIdpMappers(ctx, realm, alias, mappers)
}

func (i *idpService) GetIdpConfig(ctx context.Context, realm string, url string, providerId string) (map[string]string, error) {
	return i.idpRepository.GetIdpConfig(ctx, realm, url, providerId)
}

func (i *idpService) GetIdpConfigFromFile(ctx context.Context, realm string, providerId string, fileName string, fileBody io.Reader) (map[string]string, error) {
	return i.idpRepository.GetIdpConfigFromFile(ctx, realm, providerId, fileName, fileBody)
}

func (i *idpService) SetBrowserRedirector(ctx context.Context, realm string, alias string) error {
	return i.idpRepository.SetBrowserRedirector(ctx, realm, alias)
}

func (i *idpService) GetSPSamlMetadataUrl(ctx context.Context, realm string, idpAlias string) string {
	return i.idpRepository.GetSPSamlMetadataUrl(ctx, realm, idpAlias)
}

func (i *idpService) SetMandatotyClaimMapper(ctx context.Context, realm, alias string, mapper *models.IdpMapper) error {
	return i.idpRepository.SetMandatotyClaimMapper(ctx, realm, alias, mapper)
}

func (i *idpService) DeleteMandatoryClaimMapper(ctx context.Context, realm, alias, name string) error {
	return i.idpRepository.DeleteMandatoryClaimMapper(ctx, realm, alias, name)
}

func (i *idpService) GetIdpAliasForExchangeToken(ctx context.Context, realm string) (*string, error) {
	return i.idpRepository.GetIdpAliasForExchangeToken(ctx, realm)
}
