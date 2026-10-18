/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/smithy-go/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"golang.org/x/sync/errgroup"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	AgentClientName string = "runai_agent_"
)

func (b *Builder) CreateRealmSupplements(ctx context.Context, tenant builders.TenantInfo, clients []config.Client, builderResponse map[string]interface{}) (err error) {
	ctx = context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", tenant.Id))
	err = b.setTenantData(ctx, tenant, clients)
	if err != nil {
		return err
	}

	if tenant.Email != "" {
		_, password, err := b.createRunaiFirstUser(ctx, tenant.Email, tenant, true)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create admin user %s for tenant %s: %v", tenant.Email, tenant.Name, err.Error())
			return err
		}
		if password != "" {
			builderResponse[tenant.Email] = password
		}

		if err = b.addFirstUserAccessRules(ctx, tenant.Email, tenant); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create roles for first user %s: %s", tenant.Email, err.Error())
			return err
		}
	}

	return nil
}

func (b *Builder) setTenantData(ctx context.Context, tenant builders.TenantInfo, clients []config.Client) error {
	// Use a plain errgroup.Group (not errgroup.WithContext) so that one client's
	// failure doesn't cancel the shared ctx and abort the other clients' in-flight requests.
	var erg errgroup.Group
	for _, client := range clients {
		switch {
		case strings.Contains(client.ClientId, AgentClientName):
			client := client
			erg.Go(func() error {
				if err := b.setAgentPermissions(ctx, client, tenant.Id, tenant.Name); err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set agent permissions for tenant %s: %v", tenant.Name, err.Error())
					return err
				}
				return nil
			})
		}
	}
	return erg.Wait()
}

func (b *Builder) createRunaiRealm(ctx context.Context, realmName string, tenantId uint) (realm *config.Realm, err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Creating new keycloak realm %s", realmName)
	tenantIdStr := strconv.FormatUint(uint64(tenantId), 10)
	resp, httpResp, err := b.idmClient.GetOrCreateRealm(ctx, realmName, tenantIdStr)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to create realm %s", realmName)
		return nil, utils.CheckHttpResponse(httpResp, err, "create realm")
	}
	realm = idmRealmToConfigRealm(resp)
	return realm, nil
}

func idmRealmToConfigRealm(r *openapi.RealmCreationResponse) *config.Realm {
	if r == nil {
		return nil
	}
	clients := make([]config.Client, 0, len(r.Clients))
	for _, c := range r.Clients {
		clients = append(clients, config.Client{
			Id:       c.Id,
			ClientId: c.ClientId,
			Secret:   c.Secret,
		})
	}
	return &config.Realm{
		Name:    r.RealmName,
		Domain:  r.Domain,
		Clients: clients,
	}
}

func (b *Builder) setAgentPermissions(ctx context.Context, agent config.Client, tenantId uint, tenantName string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Set agent permissions for tenant %d", tenantId)
	reqBody := authorization_api.NewInternalCreateAccessRuleRequest(agent.ClientId, authorization_api.SUBJECTTYPE_SERVICE_ACCOUNT, authorization_api.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT)
	reqBody.SetScope(authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   fmt.Sprintf("%d", tenantId),
		Type: authorization_api.SCOPETYPE_TENANT,
		Name: ptr.String(tenantName),
	})

	request := b.authClient.CreateInternalAccessRule(ctx)
	request = request.InternalCreateAccessRuleRequest(*reqBody)
	_, resp, err := b.authClient.CreateInternalAccessRuleExecute(request)
	if err != nil && !isAlreadyExists(resp, err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set runai agent permissions for tenant id %d. Response: %+v, Error: %+v", tenantId, resp, err)
		return err
	}
	return nil
}

func isAlreadyExists(resp *http.Response, _ error) bool {
	return resp != nil && resp.StatusCode == http.StatusConflict
}

func (b *Builder) createRunaiFirstUser(ctx context.Context, tenantEmail string, tenant builders.TenantInfo, shouldNotify bool) (userId string, password string, err error) {
	tenantEmail = strings.ToLower(tenantEmail)
	resp, err := b.idmClient.DeleteInternalUserByEmail(ctx, tenant.Name, tenantEmail)
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete user %s for tenant %s: %v", tenantEmail, tenant.Name, err)
		return "", "", err
	}
	req := openapi.InternalUserCreationRequest{
		Email:         tenantEmail,
		Realm:         tenant.Name,
		ResetPassword: shouldResetUserPassword(tenant),
		Notify:        *openapi.NewNullableBool(&shouldNotify),
	}
	if tenant.Password != nil && *tenant.Password != "" {
		req.SetPassword(*tenant.Password)
	}
	userResp, httpResp, err := b.idmClient.CreateInternalUser(ctx, req)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to create first user %s for tenant id %d", tenantEmail, tenant.Id)
		return "", "", utils.CheckHttpResponse(httpResp, err, "create user")
	}
	if userResp == nil {
		return "", "", fmt.Errorf("create user returned nil response")
	}
	return userResp.Id, userResp.TempPassword, nil
}

func (b *Builder) addFirstUserAccessRules(ctx context.Context, email string, tenant builders.TenantInfo) error {
	email = strings.ToLower(email)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Creating access rules for first user %s for tenant %d", email, tenant.Id)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("time addFirstUserAccessRules: %v", time.Now())

	request, err := b.createInternalAccessRuleRequest(ctx, email, tenant)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("err time addFirstUserAccessRules: %v", time.Now())
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set access rules request for first user %s. Error: %+v", email, err)
		return err
	}
	_, resp, err := b.authClient.CreateInternalAccessRuleExecute(*request)
	if err != nil && !isAlreadyExists(resp, err) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("err time addFirstUserAccessRules: %v", time.Now())
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set access rules for first user %s. Response: %+v, Error: %+v", email, resp, err)
		return err
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("success time addFirstUserAccessRules: %v", time.Now())
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Creating access rules for first user %s: %v", email, err)
	return nil
}

func (b *Builder) createInternalAccessRuleRequest(ctx context.Context, email string, tenant builders.TenantInfo) (*authorization_api.ApiCreateInternalAccessRuleRequest, error) {
	scope := getScopeByRole(tenant.Role)
	request := b.authClient.CreateInternalAccessRule(ctx)
	reqBody := authorization_api.NewInternalCreateAccessRuleRequest(email, authorization_api.SUBJECTTYPE_USER, tenant.Role)
	reqBody.SetScope(authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   fmt.Sprintf("%d", tenant.Id),
		Type: scope,
		Name: ptr.String(tenant.Name),
	})
	request = request.InternalCreateAccessRuleRequest(*reqBody)
	return &request, nil
}

func shouldResetUserPassword(tenant builders.TenantInfo) bool {
	return tenant.Password == nil || *tenant.Password == ""
}

func getScopeByRole(roleName authorization_api.AllowedInternalCreateRoles) authorization_api.ScopeType {
	switch roleName {
	case authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR:
		return authorization_api.SCOPETYPE_TENANT
	case authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR:
		return authorization_api.SCOPETYPE_SYSTEM
	default:
		return authorization_api.SCOPETYPE_TENANT
	}
}
