/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"
	"errors"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type AuthManagerInterface interface {
	ValidatePrivilegeEscalation(ctx context.Context, authzClient auth_client.IAuthorizationClient, subjectId string) error
	GetRealmAndTenantID(ctx context.Context, tenantId string) (*string, *string, error)
}

type AuthManager struct {
	realmDal business.RealmDal
}

func NewAuthManager(realmDal business.RealmDal) AuthManagerInterface {
	return &AuthManager{
		realmDal: realmDal,
	}
}

func (a *AuthManager) GetRealmAndTenantID(ctx context.Context, tenantId string) (*string, *string, error) {
	if tenantId == "" {
		realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
		tenant := runaictx.GetString(ctx, runaictx.CfTenantId)
		return &realm, tenant, nil
	}
	if !authorization_utils.HasSystemPermissions(ctx) {
		return nil, nil, errors.New("unauthorized to manage security settings")
	}
	realm, err := a.realmDal.GetRealmByTenantId(ctx, tenantId)
	if err == nil && realm != nil {
		return &realm.Realm, &tenantId, nil
	}
	return nil, nil, err
}

func (authManager *AuthManager) ValidatePrivilegeEscalation(ctx context.Context, authzClient auth_client.IAuthorizationClient, subjectId string) error {
	_, err := authzClient.PrivilegeEscalationEvaluate(ctx).SubjectId(subjectId).GroupIds([]string{}).Execute()
	return err
}
