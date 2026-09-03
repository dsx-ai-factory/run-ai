/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flows

import (
	"context"
	"strings"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	tenant_provider "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/tenant"
	tm_config "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

func NewCreateTenantFlow(tenantProvider tenant_provider.TenantProvider) Flow {
	return &CreateTenantFlow{
		tenantProvider: tenantProvider,
		tenantConf:     config.Get[*tm_config.Tenant](),
	}
}

type CreateTenantFlow struct {
	tenantProvider tenant_provider.TenantProvider
	tenantConf     *tm_config.Tenant
}

func (flow *CreateTenantFlow) Start(ctx context.Context) {
	if flow.tenantConf.Name != "" {
		flow.createTenantFlow(ctx)
	}
}

func (flow *CreateTenantFlow) createTenantFlow(ctx context.Context) {
	tenantName := flow.tenantConf.Name
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("checking if a tenant with the name %s exists", tenantName)

	action, existing := checkTenantProvisionState(ctx, flow.tenantProvider, tenantName)
	switch action {
	case actionRequeueAndWait:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("tenant %s is Failed, requeuing provision", tenantName)
		if apiErr := flow.tenantProvider.RequeueProvision(ctx, existing, tenant_provider.ProvisionConfig{
			Email:    strings.ToLower(flow.tenantConf.Username),
			Password: ptr.To(flow.tenantConf.Password),
			Role:     authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR,
		}); apiErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("failed to requeue provision for tenant %s: %s, restarting to retry", tenantName, apiErr.Message)
		}
		fallthrough
	case actionWait:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("tenant %s is provisioning, waiting for Ready", tenantName)
		if WaitForTenantReady(ctx, flow.tenantProvider, tenantName) == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("tenant %s did not reach Ready state, restarting to retry", tenantName)
		}
	case actionCreate:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("no local tenant %s exists, creating one", tenantName)
		flow.createTenant(ctx)
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("skipping tenant creation, tenant: %s", tenantName)
	}
}

func (flow *CreateTenantFlow) createTenant(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	newTenant := api.TenantCreationRequest{Name: flow.tenantConf.Name, Email: strings.ToLower(flow.tenantConf.Username), Password: ptr.To(flow.tenantConf.Password), ContractType: models.ContractTypeNormal}
	_, apiError := flow.tenantProvider.CreateTenant(ctx, newTenant)
	if apiError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("failed to create tenant: %s, err: %s. retry tenant creation..", flow.tenantConf.Name, apiError.Message)
	}
	if WaitForTenantReady(ctx, flow.tenantProvider, flow.tenantConf.Name) == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("tenant %s did not reach Ready state, restarting to retry", flow.tenantConf.Name)
	}
}
