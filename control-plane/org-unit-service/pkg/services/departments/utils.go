/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package departments

import (
	"context"
	"net/http"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"
)

const (
	// MaxBatchPatchDepartmentRequests limits the number of department patch operations per request
	MaxBatchPatchDepartmentRequests = 100
)

// validateResetResourcesPermitted
// permittedScopes - the permitted scopes of the API request - scopes in which it is allowed to Update departments
// The func will count all departments in the cluster,
// and compare to the count of departments (in the same scope) that are allowed by the permittedScopes.
// If the counts are equal - then the user is authorized to reset resources for all departments in the scope,
// else - the user is not authorized.
func validateResetResourcesPermitted(ctx context.Context, tenantId, clusterId string, permittedScopes *authorization_api.PermittedScopes) *api.Error {
	clusterIdFilter := []validators.ParameterizedFilter{
		{Field: "cluster_id", Operator: "=", FormattedValues: []string{clusterId}},
	}

	permittedCount, dbErr := dao.CountDepartments(ctx, tenantId, permittedScopes, clusterIdFilter, models.MetaDataVerbosity)
	if dbErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed counting departments for reset resources authorization check with permitted scopes", dbErr)
	}

	allCount, dbErr := dao.CountDepartments(ctx, tenantId, nil, clusterIdFilter, models.MetaDataVerbosity)
	if dbErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed counting departments for reset resources authorization check", dbErr)
	}

	if permittedCount != allCount || allCount == 0 {
		return api.NewError(http.StatusForbidden, "not authorized to reset resources for all departments in the cluster", nil)
	}
	return nil
}
