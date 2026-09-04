/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app_reconciler_by_type

import (
	"context"
	"fmt"
	"time"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type AppsReconcilerConfig struct {
	hostTemplate string
	appHost      string
}

type AppsReconciler struct {
	iamServices           iam.IamServiceContainer
	config                AppsReconcilerConfig
	tenantResolver        reconciler.TenantResolver
	authClient            authorization_client.IAuthorizationClient
	version               string
	reconcilerVersionsDal business.ReconcilerVersionsDal
}

func NewAppReconcilerByType(ctx context.Context, iamServices *iam.IamServiceContainer, tenantRealm *config.TenantRealm,
	realmDal business.RealmDal, reconcilerVersionsDal business.ReconcilerVersionsDal, version string,
) reconciler.Reconciler {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Init Access Keys Reconciler")
	tenantResolver := reconciler.NewRealmTenantResolver(ctx, realmDal)

	ar := AppsReconciler{
		iamServices:           *iamServices,
		tenantResolver:        tenantResolver,
		reconcilerVersionsDal: reconcilerVersionsDal,
		authClient:            authorization_client.NewAuthorizationClient(),
		version:               version,
		config: AppsReconcilerConfig{
			hostTemplate: tenantRealm.TenantDomainTemplate,
			appHost:      tenantRealm.AppHost,
		},
	}
	return &ar
}

func (a *AppsReconciler) Reconcile(ctx context.Context, realm string, force bool) error {
	start := time.Now()
	defer reconciler.LogReconcileDuration(ctx, start, realm, fmt.Sprintf("Access keys reconciler. in version %s. force: %t", a.version, force))
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling Access Keys for realm %s. in version %s. force: %t", realm, a.version, force)

	tenantId := a.tenantResolver.GetRealmTenantId(ctx, realm)
	if tenantId == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant id for realm %s", realm)
		return fmt.Errorf("failed to get tenant id for realm %s", realm)
	}

	if a.shouldReconcileAccessKeys(ctx, tenantId, force) {
		err := a.reconcileAccessKeys(ctx, realm, tenantId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile external apps for realm %s: %+v", realm, err)
			return err
		}
		a.updateAccessKeysVersionAndLogError(ctx, tenantId)
	}

	return nil
}

func (r *AppsReconciler) shouldReconcileAccessKeys(ctx context.Context, tenantId string, force bool) bool {
	if force {
		return true
	}
	versions, err := r.reconcilerVersionsDal.GetReconcilerVersionsByTenantId(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("failed to read reconciler versions for tenant %s, reconciling access keys", tenantId)
		return true
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("reconciler access keys version set to %s for tenant %s", versions.AccessKeys, tenantId)
	return versions.AccessKeys != r.version
}

func (cr *AppsReconciler) updateAccessKeysVersionAndLogError(ctx context.Context, tenantId string) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Updating access keys for tenant id %s to version %s", tenantId, cr.version)
	err := cr.reconcilerVersionsDal.UpdateAccessKeysReconcilerVersions(ctx, tenantId, cr.version)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update access keys version to %s for tenant %s: %+v", cr.version, tenantId, err)
	}
}

func (cr *AppsReconciler) reconcileAccessKeys(ctx context.Context, realm string, tenantId string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling access keys for realm %s", realm)

	clientType := realm_client.ClientTypeUserApp
	filters := &iamModels.ClientFilter{
		ClientType: &clientType,
	}
	apps, err := cr.iamServices.ClientService.GetClients(ctx, realm, filters)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get apps for realm %s: %+v", realm, err)
		return err
	}
	for _, app := range apps {
		if app.Attributes == nil || app.Attributes.CreatedBy == nil {
			return fmt.Errorf("failed to reconcile access key %s in realm %s. missing 'createdBy' attribute", *app.ClientID, realm)
		}
		runaiClient := realm_client.NewRunaiServiceAccountClient(*app.ClientID, tenantId, nil, realm_client.ClientTypeUserApp)
		err := cr.reconcileApplication(ctx, realm, runaiClient)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile access key %s in realm %s: %+v", *app.ClientID, realm, err)
			return err
		}
	}
	return nil
}

func (cr *AppsReconciler) reconcileApplication(ctx context.Context, realm string, client realm_client.Client) error {
	clientID := client.GetClientID()
	existsClient, err := cr.iamServices.ClientService.GetClientByClientId(ctx, realm, clientID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get App %s in realm %s: %+v", clientID, realm, err)
		return err
	}

	newClient := client.GetClientData(ctx, existsClient)
	err = cr.iamServices.ClientService.UpdateClient(ctx, realm, newClient)
	if err != nil {
		return err
	}
	return nil
}
