/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realm_reconciler

import (
	"context"
	"errors"
	"time"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamError "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	realm_updater "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type RealmReconcilerConfig struct {
	env           string
	theme         string
	smtpHost      string
	smtpPort      int
	smtpFromEmail string
}

type RealmReconciler struct {
	iamServices     iam.IamServiceContainer
	authFlowUpdater realm_updater.AuthFlowUpdaterInterface
	realmDal        business.RealmDal
	config          RealmReconcilerConfig
}

func NewRealmReconciler(ctx context.Context, tenantRealm *config.TenantRealm, iamServices iam.IamServiceContainer, realmDal business.RealmDal) reconciler.Reconciler {
	authFlowUpdater := realm_updater.NewAuthFlowUpdater(ctx, iamServices)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Init Realm Reconciler")

	return &RealmReconciler{
		iamServices:     iamServices,
		authFlowUpdater: authFlowUpdater,
		realmDal:        realmDal,
		config: RealmReconcilerConfig{
			env:           tenantRealm.Env,
			theme:         tenantRealm.LoginTheme,
			smtpHost:      tenantRealm.SmtpHost,
			smtpPort:      tenantRealm.SmtpPort,
			smtpFromEmail: tenantRealm.SmtpFromEmail,
		},
	}
}

func (rr *RealmReconciler) Reconcile(ctx context.Context, realm string, force bool) error {
	start := time.Now()
	defer reconciler.LogReconcileDuration(ctx, start, realm, "RealmReconciler")
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling realm settings for realm %s", realm)

	realmData, err := rr.iamServices.RealmService.GetRealm(ctx, realm)
	if err != nil || realmData == nil {
		if errors.Is(err, iamError.ErrNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Realm %s does not exist, creating it", realm)
			return rr.createRealm(ctx, realm)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get realm data for realm %s: %v", realm, err)
		return err
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Realm %s already exists, performing update", realm)
	defaults := &iamModels.RealmDefaults{
		Theme:         &rr.config.theme,
		SmtpHost:      rr.config.smtpHost,
		SmtpPort:      rr.config.smtpPort,
		SmtpFromEmail: rr.config.smtpFromEmail,
	}
	realmData.SetRealmDefaults(defaults)
	rr.applyTenantId(ctx, realm, realmData)

	err = rr.iamServices.RealmService.UpdateRealm(ctx, realmData)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update realm %s: %v", realm, err)
		return err
	}

	err = rr.iamServices.UserService.ToggleRequiredAction(ctx, *realmData.Name, iamModels.RequiredActionVerifyProfile, false)
	if err != nil {
		if errors.Is(err, iamError.ErrNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Skipping disable verify user action for realm %s. user action not found", realm)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to disable verify user action for realm %s. err: %v", realm, err)
			return err
		}
	}

	err = rr.iamServices.UserService.ToggleUnmanagedUsersAttributes(ctx, *realmData.Name, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to enable unmanaged user attribute for realm %s: %v", realm, err)
		return err
	}

	err = rr.authFlowUpdater.UpdateAuthFlows(ctx, realm)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile realm - update auth flow for realm %s failed: %v", realm, err)
		return err
	}

	err = rr.iamServices.RealmService.SetRealmEmailScopeOptional(ctx, realm)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set email scope optional at realm level for realm %s: %v", realm, err)
		return err
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Realm %s updated successfully", realm)
	return nil
}

func (rr *RealmReconciler) applyTenantId(ctx context.Context, realm string, realmModel *iamModels.Realm) {
	dbRealm, err := rr.realmDal.GetRealmByName(ctx, realm)
	if err != nil || dbRealm == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Could not resolve tenant id for realm %s; skipping tenant-id realm attribute. err: %v", realm, err)
		return
	}
	realmModel.TenantId = &dbRealm.TenantId
}

func (rr *RealmReconciler) createRealm(ctx context.Context, realm string) error {
	defaults := &iamModels.RealmDefaults{
		Theme:         &rr.config.theme,
		SmtpHost:      rr.config.smtpHost,
		SmtpPort:      rr.config.smtpPort,
		SmtpFromEmail: rr.config.smtpFromEmail,
	}
	realmModel := iamModels.NewRunaiRealm(realm, defaults)
	rr.applyTenantId(ctx, realm, realmModel)

	_, err := rr.iamServices.RealmService.CreateRealm(ctx, realmModel)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create realm %s: %v", realm, err)
		return err
	}
	if err := rr.iamServices.RealmService.SetRealmEmailScopeOptional(ctx, realm); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set email scope optional at realm level for realm %s: %v", realm, err)
		return err
	}
	return nil
}
