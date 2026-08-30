package main

import (
	"context"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	manager_reconciler "github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler/manager_reconciler"
	tenants_manager_client "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
)

func runReconcile(
	ctx context.Context,
	iamServices *iam.IamServiceContainer,
	masterReconciler reconciler.MasterReconcilerI,
	realmDal business.RealmDal,
	iamArgs *config.Iam,
) {
	tmClient := tenants_manager_client.NewTenantsManagerClient(iamArgs.TenantsManagerUrl)
	reconcileManager := manager_reconciler.NewReconcileManager(ctx, iamServices, masterReconciler, tmClient, realmDal)

	if reconcileManager.ShouldUpdateTenants(ctx) {
		reconcileManager.AddAllTenantsToLocalSchema(ctx)
	}
	err := reconcileManager.RunReconcile(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error accrued while running realm reconciler. err:  %v", err)
	}
}
