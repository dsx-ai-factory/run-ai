/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/access_key_reconciler"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/client_reconciler"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/realm_reconciler"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/service_account_reconciler"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

func BuildMasterReconciler(
	ctx context.Context,
	iamServices iam.IamServiceContainer,
	realmDal business.RealmDal,
	reconcilerVersionsDal business.ReconcilerVersionsDal,
	iamArgs *config.Iam,
) reconciler.MasterReconcilerI {
	reconcilerConfig := reconciler.NewReconcilerConfig(
		iamArgs.ReconcilerInternalServiceAccountsVersion,
		iamArgs.ReconcilerServiceAccountsVersion,
		iamArgs.ReconcilerAccessKeysVersion,
	)

	tenantRealm := runai_config.Get[*config.TenantRealm]()
	reconcilers := reconciler.NewMasterReconciler(ctx)
	reconcilers.AddReconciler(api.RECONCILERTYPE_REALMS, realm_reconciler.NewRealmReconciler(ctx, tenantRealm, iamServices, realmDal))
	reconcilers.AddReconciler(api.RECONCILERTYPE_CLIENTS, client_reconciler.NewClientsReconciler(ctx, &iamServices, tenantRealm, realmDal))
	reconcilers.AddReconciler(api.RECONCILERTYPE_SERVICE_ACCOUNTS, service_account_reconciler.NewServiceAccountsReconciler(ctx, &iamServices, tenantRealm, realmDal, reconcilerVersionsDal, reconcilerConfig.ServiceAccountsVersion))
	reconcilers.AddReconciler(api.RECONCILERTYPE_INTERNAL_SERVICE_ACCOUNTS, service_account_reconciler.NewInternalServiceAccountsReconciler(ctx, &iamServices, tenantRealm, realmDal, reconcilerVersionsDal, reconcilerConfig.InternalServiceAccountsVersion))
	reconcilers.AddReconciler(api.RECONCILERTYPE_ACCESS_KEYS, access_key_reconciler.NewAccessKeysReconciler(ctx, &iamServices, tenantRealm, realmDal, reconcilerVersionsDal, reconcilerConfig.AccessKeysVersion))
	reconcilers.AddReconciler(api.RECONCILERTYPE_WORKLOAD_AUTH, client_reconciler.NewWorkloadAuthReconciler(ctx, &iamServices, tenantRealm, realmDal))

	return reconcilers
}
