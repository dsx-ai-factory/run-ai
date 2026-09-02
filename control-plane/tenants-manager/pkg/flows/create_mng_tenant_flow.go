/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flows

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/smithy-go/ptr"

	auth "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	identity_manager_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	tenant_provider "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/tenant"
	tm_config "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const MNG_APPLICATION_NAME = "default"

func NewCreateMngTenantFlow(tenantProvider tenant_provider.TenantProvider, idpClient identity_manager_client.IIdm, authClient auth.IAuthorizationClient) Flow {
	return &CreateMngTenantFlow{
		tenantProvider: tenantProvider,
		idpClient:      idpClient,
		authClient:     authClient,
		tenantConf:     config.Get[*tm_config.Tenant](),
	}
}

type CreateMngTenantFlow struct {
	tenantProvider tenant_provider.TenantProvider
	idpClient      identity_manager_client.IIdm
	authClient     auth.IAuthorizationClient
	tenantConf     *tm_config.Tenant
}

func (flow *CreateMngTenantFlow) Start(ctx context.Context) {
	if flow.tenantConf.MngName != "" {
		flow.createTenantFlow(ctx)
	}
}

func (flow *CreateMngTenantFlow) createTenantFlow(ctx context.Context) {
	mngName := flow.tenantConf.MngName
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("checking if mng tenant %s exists", mngName)

	action, existing := checkTenantProvisionState(ctx, flow.tenantProvider, mngName)
	switch action {
	case actionRequeueAndWait:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("mng tenant %s is Failed, requeuing provision", mngName)
		if requeueErr := flow.tenantProvider.RequeueProvision(ctx, existing, tenant_provider.ProvisionConfig{
			Email:    strings.ToLower(flow.tenantConf.MngUsername),
			Password: ptr.String(flow.tenantConf.MngPassword),
			Role:     authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR,
		}); requeueErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("failed to requeue provision for mng tenant %s: %s, restarting to retry", mngName, requeueErr.Message)
		}
		fallthrough
	case actionWait:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("mng tenant %s is provisioning, waiting for Ready", mngName)
		tenantRes := WaitForTenantReady(ctx, flow.tenantProvider, mngName)
		if tenantRes == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("mng tenant %s did not reach Ready state, restarting to retry", mngName)
		}
		flow.setupMngTenantPostReady(ctx, tenantRes)
	case actionCreate:
		flow.createMngTenant(ctx)
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("skipping mng tenant creation, tenant exists with name: %s", mngName)
	}
}

func (flow *CreateMngTenantFlow) createMngTenant(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	role := api.ROLES_CLOUD_OPERATOR
	newTenant := api.TenantCreationRequest{
		Name:         flow.tenantConf.MngName,
		Email:        strings.ToLower(flow.tenantConf.MngUsername),
		Password:     ptr.String(flow.tenantConf.MngPassword),
		ContractType: models.ContractTypeNormal,
		Role:         &role,
	}
	_, apiError := flow.tenantProvider.CreateTenant(ctx, newTenant)
	if apiError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("failed to create mng tenant: %s, err: %s. retry tenant creation..", flow.tenantConf.MngName, apiError.Message)
	}
	tenantRes := WaitForTenantReady(ctx, flow.tenantProvider, flow.tenantConf.MngName)
	if tenantRes == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Fatalf("mng tenant %s did not reach Ready state, restarting to retry", flow.tenantConf.MngName)
	}
	flow.setupMngTenantPostReady(ctx, tenantRes)
}

func (flow *CreateMngTenantFlow) setupMngTenantPostReady(ctx context.Context, tenantRes *models.Tenant) {
	tenantId := utils.UintToString(int32(tenantRes.ID))
	flow.createFirstServiceAccount(ctx, tenantId)
	if err := flow.createApplicationAccessRules(ctx, tenantId, tenantRes.Name); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create application access rules for mng tenant %s: %v", tenantId, err)
	}
	if err := flow.addConsoleAdminSystemAdminRole(ctx, tenantId, tenantRes.Name); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to add system admin role for console admin user %s: %v", flow.tenantConf.MngUsername, err)
	}
}

func (flow *CreateMngTenantFlow) createFirstServiceAccount(ctx context.Context, tenantId string) {
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, flow.tenantConf.MngUsername)
	realmName := utils.DisplayNameToRealm(flow.tenantConf.MngName)
	_, err := flow.idpClient.CreateServiceAccountInternally(ctx, MNG_APPLICATION_NAME, realmName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to create first service account for mng tenant: %s, err: %s", tenantId, err.Error())
	}
}

func (flow *CreateMngTenantFlow) createApplicationAccessRules(ctx context.Context, tenantId string, tenantName string) error {
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)

	systemScope := authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   tenantId,
		Type: authorization_api.SCOPETYPE_SYSTEM,
		Name: ptr.String(tenantName),
	}
	if err := flow.authClient.SetAppPermissions(ctx, tenantId, MNG_APPLICATION_NAME, systemScope, authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR); err != nil {
		return err
	}

	tenantScope := authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   tenantId,
		Type: authorization_api.SCOPETYPE_TENANT,
		Name: ptr.String(tenantName),
	}
	return flow.authClient.SetAppPermissions(ctx, tenantId, MNG_APPLICATION_NAME, tenantScope, authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR)
}

func (flow *CreateMngTenantFlow) addConsoleAdminSystemAdminRole(ctx context.Context, tenantId string, tenantName string) error {
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)

	reqBody := authorization_api.NewInternalCreateAccessRuleRequest(strings.ToLower(flow.tenantConf.MngUsername), authorization_api.SUBJECTTYPE_USER, authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR)
	reqBody.SetScope(authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   tenantId,
		Type: authorization_api.SCOPETYPE_TENANT,
		Name: ptr.String(tenantName),
	})

	request := flow.authClient.CreateInternalAccessRule(ctx).InternalCreateAccessRuleRequest(*reqBody)
	_, resp, err := flow.authClient.CreateInternalAccessRuleExecute(request)
	if err != nil && (resp == nil || resp.StatusCode != http.StatusConflict) {
		return err
	}
	return nil
}
