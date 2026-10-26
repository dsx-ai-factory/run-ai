/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authz

import (
	"context"
	"fmt"

	authorizationclienttypes "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	idmclienttypes "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"k8s.io/utils/ptr"
)

var accessRuleRoleByClientType = map[idmclienttypes.ClientType]authorizationclienttypes.AllowedInternalCreateRoles{
	idmclienttypes.CLIENTTYPE_AGENT:          authorizationclienttypes.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
	idmclienttypes.CLIENTTYPE_WORKLOADS_AUTH: authorizationclienttypes.ALLOWEDINTERNALCREATEROLES_CLUSTER_SERVICE_ACCOUNT,
	idmclienttypes.CLIENTTYPE_INSTALLER:      authorizationclienttypes.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT,
}

func (a *AuthzManager) CreateAccessRule(ctx context.Context, clientID, clusterUuid, clusterName string, clientType idmclienttypes.ClientType) error {
	roleName, ok := accessRuleRoleByClientType[clientType]
	if !ok {
		return fmt.Errorf("no access rule role configured for client type %s", clientType)
	}

	req := a.authorizationClient.CreateInternalAccessRule(ctx).InternalCreateAccessRuleRequest(authorizationclienttypes.InternalCreateAccessRuleRequest{
		SubjectId:   clientID,
		SubjectType: authorizationclienttypes.SUBJECTTYPE_SERVICE_ACCOUNT,
		Scope: *authorizationclienttypes.NewNullableInternalCreateAccessRuleRequestScope(&authorizationclienttypes.InternalCreateAccessRuleRequestScope{
			Id:   clusterUuid,
			Type: authorizationclienttypes.SCOPETYPE_CLUSTER,
			Name: ptr.To(clusterName),
		}),
		RoleName: roleName,
	})

	_, resp, err := a.authorizationClient.CreateInternalAccessRuleExecute(req)
	if err != nil {

		if resp != nil && resp.StatusCode == 409 {
			return nil
		}
		return fmt.Errorf("failed to create access rule for client %s and cluster %s: %w", clientID, clusterUuid, err)
	}

	return nil
}

func (a *AuthzManager) DeleteAccessRules(ctx context.Context, clusterUuid string) error {
	req := a.authorizationClient.DeleteScopeAccessRules(ctx)
	req.ScopeType("cluster")
	req.ScopeId(clusterUuid)

	_, err := a.authorizationClient.DeleteScopeAccessRulesExecute(req)
	if err != nil {
		return fmt.Errorf("failed to delete access rules for cluster %s: %w", clusterUuid, err)
	}

	return nil
}
