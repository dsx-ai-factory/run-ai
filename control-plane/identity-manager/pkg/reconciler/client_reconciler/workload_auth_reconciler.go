/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_reconciler

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
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

type WorkloadAuthReconcilerConfig struct {
	hostTemplate string
	appHost      string
}

type WorkloadAuthReconciler struct {
	iamServices            iam.IamServiceContainer
	config                 WorkloadAuthReconcilerConfig
	tenantResolver         reconciler.TenantResolver
	authClient             authorization_client.IAuthorizationClient
	clientReconcilerHelper ClientReconcilerHelperInterface
}

func NewWorkloadAuthReconciler(ctx context.Context, iamServices *iam.IamServiceContainer, tenantRealm *config.TenantRealm,
	realmDal business.RealmDal,
) reconciler.Reconciler {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Init workload auth client reconciler")
	tenantResolver := reconciler.NewRealmTenantResolver(ctx, realmDal)
	clientReconcilerHelper := NewClientReconcilerHelper(ctx, *iamServices, tenantResolver)

	wr := WorkloadAuthReconciler{
		iamServices:            *iamServices,
		tenantResolver:         tenantResolver,
		clientReconcilerHelper: clientReconcilerHelper,
		authClient:             authorization_client.NewAuthorizationClient(),
		config: WorkloadAuthReconcilerConfig{
			hostTemplate: tenantRealm.TenantDomainTemplate,
			appHost:      tenantRealm.AppHost,
		},
	}
	return &wr
}

func (wr *WorkloadAuthReconciler) Reconcile(ctx context.Context, realm string, force bool) error {
	start := time.Now()
	defer reconciler.LogReconcileDuration(ctx, start, realm, "WorkloadAuthReconciler")
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling workload auth clients for realm %s", realm)

	tenantId := wr.tenantResolver.GetRealmTenantId(ctx, realm)
	if tenantId == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant id for realm %s", realm)
		return fmt.Errorf("failed to get tenant id for realm %s", realm)
	}

	err := wr.reconcileWorkloadAuth(ctx, realm, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile workload auth clients for realm %s: %v", realm, err)
		return err
	}

	return nil
}

func (wr *WorkloadAuthReconciler) reconcileWorkloadAuth(ctx context.Context, realm string, tenantId string) error {
	clientType := utils.ClientTypeWorkloadsAuth
	filters := &iamModels.ClientFilter{
		ClientType: &clientType,
	}
	clients, err := wr.iamServices.ClientService.GetClients(ctx, realm, filters)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get workload auth clients for realm %s: %+v", realm, err)
		return err
	}
	for _, client := range clients {
		var clusterId string
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Checking cluster id for workload auth client: %s", *client.ClientID)
		if client.Attributes != nil && client.Attributes.ClusterId != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Found cluster id %s for client %s", *client.Attributes.ClusterId, *client.ClientID)
			clusterId = *client.Attributes.ClusterId
		}
		runaiClient := realm_client.NewWorkloadsAuthClient(*client.ClientID, tenantId, clusterId, wr.config.appHost, wr.config.hostTemplate)
		err := wr.clientReconcilerHelper.ReconcileClient(ctx, realm, runaiClient)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile worklaod auth client %s in realm %s: %+v", *client.ClientID, realm, err)
			return err
		}
	}
	return nil
}
