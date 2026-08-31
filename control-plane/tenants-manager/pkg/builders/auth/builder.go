/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
	auth "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	identity_manager_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
	internalCache "github.com/run-ai/runai/control-plane/tenants-manager/internal/cache"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	tm_config "github.com/run-ai/runai/control-plane/tenants-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/postgresdao"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	builderIdentifier = "auth-realm"
)

type Builder struct {
	*tm_config.Auth
	httpClient *resty.Client
	authClient auth.IAuthorizationClient
	idmClient  identity_manager_client.IIdm
	tenantDao  dao.Tenant
}

func NewBuilder() *Builder {
	authConfig := config.Get[*tm_config.Auth]()
	client := auth.NewAuthorizationClient()
	idmClient := identity_manager_client.NewIdmClient(authConfig.IdentityServiceUrl)
	tenantDao := postgresdao.NewTenantDao(internalCache.NewTenantCacheManager())
	return &Builder{
		Auth: authConfig,
		httpClient: resty.New().
			SetHeader("Content-Type", "application/json"),
		authClient: client,
		idmClient:  idmClient,
		tenantDao:  tenantDao,
	}
}

func (b *Builder) Name() string {
	return builderIdentifier
}

func (b *Builder) Construct(ctx context.Context, tenant *builders.TenantInfo, builderResponse map[string]interface{}) (err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Constructing realm for tenant: %s ", tenant.Name)
	realm, err := b.createRunaiRealm(ctx, tenant.Name, tenant.Id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to Create realm for tenant %s: %v", tenant.Name, err.Error())
		return err
	}
	if realm.Domain != "" {
		builderResponse["tenant_domain"] = realm.Domain
	}

	err = b.CreateRealmSupplements(ctx, *tenant, realm.Clients, builderResponse)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to Create realm supplements for tenant %s: %v", tenant.Name, err.Error())
		return err
	}
	return nil
}

func (b *Builder) Destruct(ctx context.Context, tenant *builders.TenantInfo) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Destructing auth builder steps")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", tenant.Id))

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
	)
	appendErr := func(err error) {
		mu.Lock()
		errs = append(errs, err)
		mu.Unlock()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if resp, err := b.deleteRunaiRealm(ctx, tenant.Name); err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Realm %s is missing from keycloak. skipping this step... %v", tenant.Name, err.Error())
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete realm for tenant %s: %v", tenant.Name, err.Error())
				appendErr(fmt.Errorf("failed deleting realm: %w", err))
			}
		}
	}()

	if tenant.Email != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.handleAccessRulesDeletion(ctx, tenant.Email, tenant.Name); err != nil {
				appendErr(fmt.Errorf("failed deleting access rule for %s: %w", tenant.Email, err))
			}
		}()
	}

	wg.Wait()
	return errors.Join(errs...)
}

func (b *Builder) handleAccessRulesDeletion(ctx context.Context, userEmail string, tenantName string) error {
	if resp, err := b.deleteAccessRules(ctx, userEmail); err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("User %s is missing from access rules db. skipping this step... %v", userEmail, err.Error())
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete user %s for tenant %s: %v", userEmail, tenantName, err.Error())
			return err
		}
	}
	return nil
}
