/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/realms"
)

type RealmService interface {
	CreateRealm(ctx context.Context, realm *models.Realm) (string, error)
	GetRealm(ctx context.Context, realmName string) (*models.Realm, error)
	GetRealms(ctx context.Context) ([]*models.Realm, error)
	DeleteRealm(ctx context.Context, realmName string) error
	DisableRealm(ctx context.Context, realmName string) error
	UpdateRealm(ctx context.Context, realm *models.Realm) error

	// realm settings
	UpdateAutoRedirectSSO(ctx context.Context, realm string, idpAlias *string, isEnabled bool) error
	SetRealmEmailScopeOptional(ctx context.Context, realm string) error
}

type realmService struct {
	realmRepository realms.RealmRepository
}

func NewRealmService(realmRepository realms.RealmRepository) RealmService {
	return &realmService{
		realmRepository: realmRepository,
	}
}

func (r *realmService) CreateRealm(ctx context.Context, realm *models.Realm) (string, error) {
	return r.realmRepository.CreateRealm(ctx, realm)
}

func (r *realmService) GetRealm(ctx context.Context, realmName string) (*models.Realm, error) {
	return r.realmRepository.GetRealm(ctx, realmName)
}

func (r *realmService) GetRealms(ctx context.Context) ([]*models.Realm, error) {
	return r.realmRepository.GetRealms(ctx)
}

func (r *realmService) DeleteRealm(ctx context.Context, realmName string) error {
	return r.realmRepository.DeleteRealm(ctx, realmName)
}

func (r *realmService) DisableRealm(ctx context.Context, realmName string) error {
	return r.realmRepository.DisableRealm(ctx, realmName)
}

func (r *realmService) UpdateRealm(ctx context.Context, realm *models.Realm) error {
	return r.realmRepository.UpdateRealm(ctx, *realm)
}

func (r *realmService) UpdateAutoRedirectSSO(ctx context.Context, realm string, idpAlias *string, isEnabled bool) error {
	return r.realmRepository.UpdateAutoRedirectSSO(ctx, realm, idpAlias, isEnabled)
}

func (r *realmService) SetRealmEmailScopeOptional(ctx context.Context, realm string) error {
	return r.realmRepository.SetRealmEmailScopeOptional(ctx, realm)
}
