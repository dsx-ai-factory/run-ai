/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization_client

import (
	"context"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
)

type ScopeValidator interface {
	GetPermittedScopesFromCtx(ctx context.Context) (*authorization_api.PermittedScopes, bool)
	HasPermittedScopes(ctx context.Context) error
	IsPermittedScopesEmpty(permittedScopes *authorization_api.PermittedScopes) bool
	GetPermittedScopesDbFilter(permittedScopes authorization_api.PermittedScopes) string
	GetPermittedScopesDbFilterWithColumnNames(
		permittedScopes authorization_api.PermittedScopes,
		tenantIdColumnName string,
		clusterIdColumnName string,
		departmentIdColumnName string,
		projectIdColumnName string,
	) string
}

type ScopeValidatorImp struct {
}

func NewScopeValidator() ScopeValidator {
	return &ScopeValidatorImp{}
}

func (s *ScopeValidatorImp) GetPermittedScopesFromCtx(ctx context.Context) (*authorization_api.PermittedScopes, bool) {
	return helpers.GetPermittedScopesFromCtx(ctx)
}

func (s *ScopeValidatorImp) HasPermittedScopes(ctx context.Context) error {
	return helpers.HasPermittedScopes(ctx)
}

func (s *ScopeValidatorImp) IsPermittedScopesEmpty(permittedScopes *authorization_api.PermittedScopes) bool {
	return helpers.IsPermittedScopesEmpty(permittedScopes)
}

func (s *ScopeValidatorImp) GetPermittedScopesDbFilter(permittedScopes authorization_api.PermittedScopes) string {
	return helpers.GetPermittedScopesDbFilter(permittedScopes)
}

func (s *ScopeValidatorImp) GetPermittedScopesDbFilterWithColumnNames(
	permittedScopes authorization_api.PermittedScopes,
	tenantIdColumnName string,
	clusterIdColumnName string,
	departmentIdColumnName string,
	projectIdColumnName string,
) string {
	return helpers.GetPermittedScopesDbFilterWithColumnNames(permittedScopes, tenantIdColumnName, clusterIdColumnName, departmentIdColumnName, projectIdColumnName)
}
