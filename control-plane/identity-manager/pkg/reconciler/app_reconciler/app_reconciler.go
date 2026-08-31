/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app_reconciler

import (
	"context"
	"fmt"
	"time"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type AppsReconcilerConfig struct {
	hostTemplate string
	appHost      string
}

type AppsReconciler struct {
	iamServices           iam.IamServiceContainer
	config                AppsReconcilerConfig
	tenantResolver        reconciler.TenantResolver
	isExternal            *bool
	version               string
	reconcilerVersionsDal business.ReconcilerVersionsDal
}

func NewAppsReconciler(ctx context.Context, iamServices *iam.IamServiceContainer, tenantRealm *config.TenantRealm,
	realmDal business.RealmDal, reconcilerVersionsDal business.ReconcilerVersionsDal, version string,
) reconciler.Reconciler {
	return newAppsReconciler(ctx, iamServices, tenantRealm, realmDal, reconcilerVersionsDal, ptr.To(true), version)
}

func NewInternalAppsReconciler(ctx context.Context, iamServices *iam.IamServiceContainer, tenantRealm *config.TenantRealm,
	realmDal business.RealmDal, reconcilerVersionsDal business.ReconcilerVersionsDal, version string,
) reconciler.Reconciler {
	return newAppsReconciler(ctx, iamServices, tenantRealm, realmDal, reconcilerVersionsDal, ptr.To(false), version)
}

func newAppsReconciler(ctx context.Context, iamServices *iam.IamServiceContainer, tenantRealm *config.TenantRealm,
	realmDal business.RealmDal, reconcilerVersionsDal business.ReconcilerVersionsDal, isExternal *bool, version string,
) reconciler.Reconciler {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Init Service Accounts Reconciler")
	tenantResolver := reconciler.NewRealmTenantResolver(ctx, realmDal)

	cr := AppsReconciler{
		iamServices:           *iamServices,
		tenantResolver:        tenantResolver,
		reconcilerVersionsDal: reconcilerVersionsDal,
		isExternal:            isExternal,
		version:               version,
		config: AppsReconcilerConfig{
			hostTemplate: tenantRealm.TenantDomainTemplate,
			appHost:      tenantRealm.AppHost,
		},
	}
	return &cr
}

func (cr *AppsReconciler) Reconcile(ctx context.Context, realm string, force bool) error {
	start := time.Now()
	isExternal := cr.getIsExternalVal()
	defer reconciler.LogReconcileDuration(ctx, start, realm, fmt.Sprintf("ServiceAccountsReconciler with external: %s. in version %s. force: %t", isExternal, cr.version, force))
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling Service Accounts for realm %s with external: %s. in version %s. force: %t", realm, isExternal, cr.version, force)

	tenantId := cr.tenantResolver.GetRealmTenantId(ctx, realm)
	if tenantId == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant id for realm %s", realm)
		return fmt.Errorf("failed to get tenant id for realm %s", realm)
	}

	if cr.shouldReconcileInternalServiceAccounts(ctx, tenantId, force) {
		err := cr.reconcileInternalServiceAccounts(ctx, realm, tenantId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile internal service accounts for realm %s: %+v", realm, err)
			return err
		}
		cr.updateInternalServiceAccountsVersionAndLogError(ctx, tenantId)
	}
	if cr.shouldReconcileServiceAccounts(ctx, tenantId, force) {
		err := cr.reconcileServiceAccounts(ctx, realm, tenantId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile service accounts for realm %s: %+v", realm, err)
			return err
		}
		cr.updateServiceAccountsVersionAndLogError(ctx, tenantId)
	}

	return nil
}

func (cr *AppsReconciler) updateInternalServiceAccountsVersionAndLogError(ctx context.Context, tenantId string) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Updating internal service accounts for tenant id %s to version %s", tenantId, cr.version)
	err := cr.reconcilerVersionsDal.UpdateInternalServiceAccountsReconcilerVersions(ctx, tenantId, cr.version)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update internal service accounts version to %s for tenant %s: %+v", cr.version, tenantId, err)
	}
}

func (cr *AppsReconciler) updateServiceAccountsVersionAndLogError(ctx context.Context, tenantId string) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Updating service accounts for tenant id %s to version %s", tenantId, cr.version)
	err := cr.reconcilerVersionsDal.UpdateServiceAccountsReconcilerVersions(ctx, tenantId, cr.version)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update service accounts version to %s for tenant %s: %+v", cr.version, tenantId, err)
	}
}

func (cr *AppsReconciler) reconcileInternalServiceAccounts(ctx context.Context, realm string, tenantId string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling internal service accounts for realm %s", realm)
	return cr.reconcileServiceAccountsByType(ctx, realm, tenantId, false)
}

func (cr *AppsReconciler) reconcileServiceAccounts(ctx context.Context, realm string, tenantId string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling service accounts for realm %s", realm)
	return cr.reconcileServiceAccountsByType(ctx, realm, tenantId, true)
}

func (cr *AppsReconciler) reconcileServiceAccountsByType(ctx context.Context, realm string, tenantId string, isExternal bool) error {
	filters := &iamModels.ClientFilter{
		IsExternal: &isExternal,
	}
	serviceAccounts, err := cr.iamServices.ClientService.GetClients(ctx, realm, filters)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get service accounts for realm %s: %+v", realm, err)
		return err
	}
	for _, serviceAccount := range serviceAccounts {
		var clusterId *string
		clientType := realm_client.ClientTypeTenantApp
		if !isExternal {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Checking cluster id for internal service account %s", *serviceAccount.ClientID)
			if serviceAccount.Attributes != nil && serviceAccount.Attributes.ClusterId != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Found cluster id %s for service account %s", *serviceAccount.Attributes.ClusterId, *serviceAccount.ClientID)
				clusterId = serviceAccount.Attributes.ClusterId
				clientType = realm_client.ClientTypeAgent
			} else {
				clientType = realm_client.ClientTypeInstaller
			}
		}
		runaiClient := realm_client.NewRunaiServiceAccountClient(*serviceAccount.ClientID, tenantId, clusterId, clientType)
		err := cr.reconcileServiceAccount(ctx, realm, runaiClient)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile service account %s in realm %s: %+v", *serviceAccount.ClientID, realm, err)
			return err
		}
	}
	return nil
}

func (cr *AppsReconciler) reconcileServiceAccount(ctx context.Context, realm string, client realm_client.Client) error {
	clientID := client.GetClientID()
	existsClient, err := cr.iamServices.ClientService.GetClientByClientId(ctx, realm, clientID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get service account %s in realm %s: %+v", clientID, realm, err)
		return err
	}

	newClient := client.GetClientData(ctx, existsClient)
	if newClient.ID != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Service account %s already exists in realm %s, performing update", *newClient.ClientID, realm)
		err = cr.iamServices.ClientService.UpdateClient(ctx, realm, newClient)
		return err
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Service account %s does not exist in realm %s, performing create", *newClient.ClientID, realm)
	_, err = cr.iamServices.ClientService.CreateClient(ctx, realm, newClient)
	return err
}

func (r *AppsReconciler) shouldReconcileInternalServiceAccounts(ctx context.Context, tenantId string, force bool) bool {
	if r.isExternal != nil && *r.isExternal {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("reconciler set to run on service accounts skipping internal service accounts for tenant %s", tenantId)
		return false
	}
	if force {
		return true
	}
	versions, err := r.reconcilerVersionsDal.GetReconcilerVersionsByTenantId(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("failed to read reconciler versions for tenant %s, reconciling internal service accounts", tenantId)
		return true
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("reconciler internal service accounts version set to %s for tenant %s", versions.InternalServiceAccounts, tenantId)
	return versions.InternalServiceAccounts != r.version
}

func (r *AppsReconciler) shouldReconcileServiceAccounts(ctx context.Context, tenantId string, force bool) bool {
	if r.isExternal != nil && !*r.isExternal {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("reconciler set to run on internal service accounts skipping service accounts for tenant %s", tenantId)
		return false
	}
	if force {
		return true
	}
	versions, err := r.reconcilerVersionsDal.GetReconcilerVersionsByTenantId(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("failed to read reconciler versions for tenant %s, reconciling service accounts", tenantId)
		return true
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("reconciler service accounts version set to %s for tenant %s", versions.ServiceAccounts, tenantId)
	return versions.ServiceAccounts != r.version
}

func (cr *AppsReconciler) getIsExternalVal() interface{} {
	if cr.isExternal == nil {
		return "nil"
	}
	if *cr.isExternal {
		return "true"
	}
	return "false"
}
