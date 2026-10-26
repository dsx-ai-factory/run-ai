/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type ScopeQueryParams struct {
	// The scope of the query
	ScopeType string `json:"scopeType,omitempty"`
	// The scope id of the query
	ScopeId string `json:"scopeId,omitempty"`

	AncestorsResolver *orgUnitClient.AncestorsResolver `gorm:"-"`
}

type ListOptions struct {
	IncludeAncestors   bool
	IncludeDescendants bool
	IncludePrivate     bool
}

const lettersAndNumbersPattern = `^[a-zA-Z0-9_\-]+$`

var LettersAndNumbersRegexp = regexp.MustCompile(lettersAndNumbersPattern)

func NewScopeQueryParams(ancestorsResolver *orgUnitClient.AncestorsResolver) *ScopeQueryParams {
	return &ScopeQueryParams{
		ScopeType: string(ancestorsResolver.ScopeType()),
		ScopeId:   ancestorsResolver.ScopeId(),

		AncestorsResolver: ancestorsResolver,
	}
}

// BuildScopeQuery - build the query to get the resources from the scope and its ancestors/descendants
// For example if the scope is a project and project id is 1, the query will be:
// (scopeType = ? AND scopeId = ?) OR (scopeType = ? AND scopeId = ?) ... with bound args
func (sq *ScopeQueryParams) BuildScopeQuery(ctx context.Context, includeAncestors bool, includeDescendants bool) (string, []interface{}, *api.Error) {
	if sq.ScopeType == "" || sq.ScopeId == "" {
		return "", nil, api.NewError(http.StatusInternalServerError, "Failed to build scope query, scopeId and scopeType must be provided", nil)
	}

	if err := common.ValidateScopeType(sq.ScopeType); err != nil {
		return "", nil, err
	}

	tenantId, ctxError := common.TenantIdFromContext(ctx)
	if ctxError != nil {
		return "", nil, api.NewError(int(ctxError.Code), ctxError.Message, nil)
	}

	if !LettersAndNumbersRegexp.MatchString(sq.ScopeId) {
		return "", nil, api.NewError(http.StatusBadRequest, "Invalid scopeId", nil)
	}

	query := fmt.Sprintf("(%s = ?)", TenantIdColumn)
	args := []interface{}{tenantId}

	// TODO: this implementation of includesDescendants supports only cluster scope
	if includeDescendants && sq.ScopeType == string(authorizationClient.SCOPETYPE_CLUSTER) {
		// returns all the objects that are in the cluster regardless of the scope type
		query += fmt.Sprintf(" AND (%s = ?)", ClusterIdColumn)
		args = append(args, sq.ScopeId)
		return query, args, nil
	}

	query += fmt.Sprintf(" AND ((%s = ? AND %s = ?)", ScopeTypeColumn, ScopeIdColumn)
	args = append(args, sq.ScopeType, sq.ScopeId)

	if sq.ScopeType == string(authorizationClient.SCOPETYPE_TENANT) || !includeAncestors {
		query += ")"
		return query, args, nil
	}

	if sq.AncestorsResolver == nil {
		sq.AncestorsResolver = orgUnitClient.NewAncestorsResolver(orgUnitApi.ScopeType(sq.ScopeType), sq.ScopeId)
	}

	ancestors, err := sq.AncestorsResolver.Get(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to obtain ancestors of %s/%s", sq.ScopeType, sq.ScopeId)
		return "", nil, api.NewError(int(err.Code), err.Message, nil)
	}

	ancestorsByScope := map[authorizationClient.ScopeType][]struct {
		scopeType authorizationClient.ScopeType
		id        string
	}{
		authorizationClient.SCOPETYPE_CLUSTER: {
			{authorizationClient.SCOPETYPE_TENANT, ancestors.GetTenantId()},
		},
		authorizationClient.SCOPETYPE_DEPARTMENT: {
			{authorizationClient.SCOPETYPE_CLUSTER, ancestors.GetClusterUuid()},
			{authorizationClient.SCOPETYPE_TENANT, ancestors.GetTenantId()},
		},
		authorizationClient.SCOPETYPE_PROJECT: {
			{authorizationClient.SCOPETYPE_DEPARTMENT, ancestors.GetDepartmentId()},
			{authorizationClient.SCOPETYPE_CLUSTER, ancestors.GetClusterUuid()},
			{authorizationClient.SCOPETYPE_TENANT, ancestors.GetTenantId()},
		},
	}

	for _, s := range ancestorsByScope[authorizationClient.ScopeType(sq.ScopeType)] {
		if !LettersAndNumbersRegexp.MatchString(s.id) {
			return "", nil, api.NewError(http.StatusBadRequest,
				fmt.Sprintf("Invalid %s id", s.scopeType), nil)
		}
		query += fmt.Sprintf(" OR (%s = ? AND %s = ?)", ScopeTypeColumn, ScopeIdColumn)
		args = append(args, s.scopeType, s.id)
	}

	query += ")"
	return query, args, nil
}
