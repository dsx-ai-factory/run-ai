/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package dao (file scope.go): shared scope-hierarchy filtering for asset list
// queries. Any asset kind whose list API exposes the clusterId/departmentId/
// projectId anchor plus includeAncestors/includeDescendants can reuse
// extractScopeAnchor + buildScopeQuery; nothing here is specific to a single
// asset kind.

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
)

// scopeColumn* are the table-qualified assets_metadata columns used to anchor
// scope-hierarchy filtering. They are qualified because the asset list queries
// join other tables, so a bare "cluster_id" could be ambiguous.
var (
	scopeColumnCluster    = fmt.Sprintf("%s.cluster_id", dbtables.TableNameAssetsMeta)
	scopeColumnDepartment = fmt.Sprintf("%s.department_id", dbtables.TableNameAssetsMeta)
	scopeColumnProject    = fmt.Sprintf("%s.project_id", dbtables.TableNameAssetsMeta)
)

// scopeAnchorEqualityOperator is the SQL operator that queryapi produces for the
// API "==" operator (see queryapi's operator table). Scope anchors must use it.
const scopeAnchorEqualityOperator = "="

// scopeAnchor is the single scope filter requested by the caller, together with
// the ancestor ids resolved from the org-unit tree (when includeAncestors is
// set). It is the input to the scope-hierarchy SQL builder.
type scopeAnchor struct {
	scope                api.Scope
	id                   string
	ancestorDepartmentId *string
	ancestorClusterId    *string
}

// extractScopeAnchor pulls the (at most one) scope anchor out of the parsed
// filterBy predicates and returns the remaining generic predicates alongside
// it. When includeAncestors is set, the anchor's ancestor ids are resolved from
// the org-unit tree so the scope-hierarchy SQL can include parent scopes. A
// query without a scope anchor returns a nil anchor (list everything the RBAC
// block permits).
func extractScopeAnchor(ctx context.Context, filters []queryapi.Filter, includeAncestors bool) ([]queryapi.Filter, *scopeAnchor, *api.Error) {
	remaining := make([]queryapi.Filter, 0, len(filters))
	anchors := make([]scopeAnchor, 0, 1)

	for _, filter := range filters {
		var anchorScope api.Scope
		var anchorField string
		switch filter.Field {
		case scopeColumnCluster:
			anchorScope, anchorField = api.SCOPE_CLUSTER, "clusterId"
		case scopeColumnDepartment:
			anchorScope, anchorField = api.SCOPE_DEPARTMENT, "departmentId"
		case scopeColumnProject:
			anchorScope, anchorField = api.SCOPE_PROJECT, "projectId"
		default:
			remaining = append(remaining, filter)
			continue
		}

		// A scope anchor identifies a single scope node, so only equality makes
		// sense; range/substring operators (!=, <=, >=, =@, ...) are rejected
		// rather than silently treated as equality.
		if filter.Operator != scopeAnchorEqualityOperator {
			return nil, nil, api.NewValidationError(fmt.Sprintf("the %s scope filter must use the '==' operator", anchorField))
		}

		anchors = append(anchors, scopeAnchor{scope: anchorScope, id: firstFilterValue(filter)})
	}

	if len(anchors) == 0 {
		return remaining, nil, nil
	}
	if len(anchors) > 1 {
		return nil, nil, api.NewValidationError("only one of clusterId, departmentId or projectId may be used to filter by scope")
	}

	anchor := anchors[0]

	if anchor.scope == api.SCOPE_PROJECT {
		if _, convErr := strconv.Atoi(anchor.id); convErr != nil {
			return nil, nil, api.NewValidationError("projectId filter value must be a numeric project id")
		}
	}

	if includeAncestors {
		switch anchor.scope {
		case api.SCOPE_PROJECT:
			resp, ancestorsErr := org_unit.Get().GetProjectAncestors(ctx, anchor.id)
			if ancestorsErr != nil {
				return nil, nil, api.NewError(http.StatusServiceUnavailable, "failed to interrogate the ancestors of the project", errors.New(ancestorsErr.Message))
			}
			anchor.ancestorDepartmentId = resp.DepartmentId.Get()
			anchor.ancestorClusterId = resp.ClusterUuid.Get()
		case api.SCOPE_DEPARTMENT:
			resp, ancestorsErr := org_unit.Get().GetDepartmentAncestors(ctx, anchor.id)
			if ancestorsErr != nil {
				return nil, nil, api.NewError(int(ancestorsErr.GetCode()), "failed to interrogate the ancestors of the department", errors.New(ancestorsErr.Message))
			}
			anchor.ancestorClusterId = resp.ClusterUuid.Get()
		}
	}

	return remaining, &anchor, nil
}

// buildScopeQuery builds the parenthesized scope-hierarchy OR-group for the
// requested anchor. The base term is the anchor scope itself; includeAncestors
// adds the parent scopes (resolved ids plus tenant/system), and
// includeDescendants adds the child scopes (keyed off the denormalized
// department_id/cluster_id columns carried by every descendant row).
func buildScopeQuery(anchor *scopeAnchor, includeAncestors, includeDescendants bool) (*database.ParametrizedQuery, *api.Error) {
	meta := dbtables.TableNameAssetsMeta
	scopeQuery := database.NewParametrizedQuery()

	switch anchor.scope {
	case api.SCOPE_PROJECT:
		projectId, convErr := strconv.Atoi(anchor.id)
		if convErr != nil {
			return nil, api.NewValidationError("projectId filter value must be a numeric project id")
		}
		scopeQuery.Add(fmt.Sprintf("(%s.scope = ? and %s.project_id = ?)", meta, meta), api.SCOPE_PROJECT, projectId)
		if includeAncestors {
			if anchor.ancestorDepartmentId != nil {
				scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ? and %s.department_id = ?)", meta, meta), api.SCOPE_DEPARTMENT, *anchor.ancestorDepartmentId)
			}
			if anchor.ancestorClusterId != nil {
				scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ? and %s.cluster_id = ?)", meta, meta), api.SCOPE_CLUSTER, *anchor.ancestorClusterId)
			}
			scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ?) or (%s.scope = ?)", meta, meta), api.SCOPE_TENANT, api.SCOPE_SYSTEM)
		}
		// Projects are leaf scopes: includeDescendants has no effect.
	case api.SCOPE_DEPARTMENT:
		scopeQuery.Add(fmt.Sprintf("(%s.scope = ? and %s.department_id = ?)", meta, meta), api.SCOPE_DEPARTMENT, anchor.id)
		if includeAncestors {
			if anchor.ancestorClusterId != nil {
				scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ? and %s.cluster_id = ?)", meta, meta), api.SCOPE_CLUSTER, *anchor.ancestorClusterId)
			}
			scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ?) or (%s.scope = ?)", meta, meta), api.SCOPE_TENANT, api.SCOPE_SYSTEM)
		}
		if includeDescendants {
			scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ? and %s.department_id = ?)", meta, meta), api.SCOPE_PROJECT, anchor.id)
		}
	case api.SCOPE_CLUSTER:
		scopeQuery.Add(fmt.Sprintf("(%s.scope = ? and %s.cluster_id = ?)", meta, meta), api.SCOPE_CLUSTER, anchor.id)
		if includeAncestors {
			scopeQuery.Add(fmt.Sprintf(" or (%s.scope = ?) or (%s.scope = ?)", meta, meta), api.SCOPE_TENANT, api.SCOPE_SYSTEM)
		}
		if includeDescendants {
			scopeQuery.Add(fmt.Sprintf(" or ((%s.scope = ? or %s.scope = ?) and %s.cluster_id = ?)", meta, meta, meta), api.SCOPE_DEPARTMENT, api.SCOPE_PROJECT, anchor.id)
		}
	}

	return scopeQuery, nil
}

func firstFilterValue(filter queryapi.Filter) string {
	if len(filter.Values) > 0 {
		return filter.Values[0]
	}
	return ""
}
