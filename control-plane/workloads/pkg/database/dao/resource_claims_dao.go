/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
)

const (
	// Filtering rows that have resource_claims that is not empty json
	podsWithResourceClaimsFilter = "resource_claims @> '[{}]'"

	// Must match the json tag of models.PodResourceClaim.ResourceClaimTemplateName.
	resourceClaimTemplateNameKey = "resource_claim_template_name"
)

// DbListPodsWithResourceClaims lists the live pods of the given workloads that reference at least
// one resource claim. When templateName is set, only pods with a claim generated from that template are listed.
func DbListPodsWithResourceClaims(ctx context.Context, tenantId uint, workloadIds []uuid.UUID, templateName string,
	offset int32, limit int32,
) ([]models.Pod, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list pods with resource claims", nil)

	if len(workloadIds) == 0 {
		return []models.Pod{}, nil
	}

	query := runaidb.DB().WithContext(ctx).
		Table(PodsTableName).
		Select("id, name, task, workload_id, resource_claims").
		Where("tenant_id = ? AND workload_id IN ?", tenantId, workloadIds).
		Where(podsWithResourceClaimsFilter)
	if templateName != "" {
		claimFilter, err := resourceClaimTemplateContainment(templateName)
		if err != nil {
			return nil, dbErr.WithDetails(err)
		}
		query = query.Where("resource_claims @> ?::jsonb", claimFilter)
	}
	query = query.Order("name, id")
	if offset != 0 {
		query = query.Offset(int(offset))
	}
	if limit != 0 {
		query = query.Limit(int(limit))
	}

	pods := []models.Pod{}
	if err := query.Find(&pods); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list pods with resource claims of workloads %v, template %q, err: %v",
			workloadIds, templateName, err)
		return nil, dbErr.WithDetails(err)
	}

	return pods, nil
}

// DbListWorkloadsReferencingResourceClaimTemplate lists the workloads whose live pods reference the template by name.
// The scope of the template is also taken into account, as well as permitted scopes the user is allowed to see.
// The pods are matched first, and the workloads table is only probed by id for their workloads.
func DbListWorkloadsReferencingResourceClaimTemplate(ctx context.Context, tenantId uint, templateName string,
	scopeType api.ScopeType, scopeId string, permittedScopes *openapi.PermittedScopes, offset int32, limit int32,
) ([]models.Workload, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to list workloads referencing resource claim template", nil)

	claimFilter, err := resourceClaimTemplateContainment(templateName)
	if err != nil {
		return nil, dbErr.WithDetails(err)
	}

	podsSubQuery := "SELECT p.workload_id FROM pods p WHERE p.tenant_id = ? AND p.deleted_at IS NULL AND p.resource_claims @> ?::jsonb"
	podsArgs := []any{tenantId, claimFilter}
	switch scopeType {
	case api.SCOPETYPE_CLUSTER:
		podsSubQuery += " AND p.cluster_id = ?"
		podsArgs = append(podsArgs, scopeId)
	case api.SCOPETYPE_PROJECT:
		podsSubQuery += " AND p.project_id = ?"
		podsArgs = append(podsArgs, scopeId)
	}

	query := runaidb.DB().WithContext(ctx).
		Table(WorkloadsTableName+" AS w").
		Select("w.id, w.name, w.type, w.project_id").
		Unscoped().
		Where("w.tenant_id = ?", tenantId).
		Where("w.deleted_at IS NULL").
		Where("w.id IN ("+podsSubQuery+")", podsArgs...)
	if scopeType == api.SCOPETYPE_DEPARTMENT {
		query = query.Where("w.department_id = ?", scopeId)
	}
	if permittedScopes != nil && (permittedScopes.System == nil || !*permittedScopes.System) {
		query = query.Where(authorization_utils.GetPermittedScopesDbFilterWithColumnNames(
			*permittedScopes, "w.tenant_id", "w.cluster_id", "w.department_id", "w.project_id"))
	}

	query = query.Order("w.name, w.id")
	if offset != 0 {
		query = query.Offset(int(offset))
	}
	if limit != 0 {
		query = query.Limit(int(limit))
	}

	workloads := []models.Workload{}
	if err := query.Find(&workloads); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list workloads referencing resource claim template %s, err: %v", templateName, err)
		return nil, dbErr.WithDetails(err)
	}

	return workloads, nil
}

func resourceClaimTemplateContainment(templateName string) (string, error) {
	filter, err := json.Marshal([]map[string]string{
		{
			resourceClaimTemplateNameKey: templateName},
	},
	)
	if err != nil {
		return "", fmt.Errorf("marshal resource claim template filter: %w", err)
	}
	return string(filter), nil
}
