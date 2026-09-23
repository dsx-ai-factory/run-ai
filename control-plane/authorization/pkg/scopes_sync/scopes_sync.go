/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopes_sync

import (
	"context"
	"fmt"
	"net/http"
	"time"

	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	org_unit_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	tenants_api "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
	tenants_openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

const (
	// Use a direct UPDATE with GORM placeholders; the driver will expand the IN ? slice properly
	GhostAccessRules = "UPDATE access_rules SET deleted_at = now() " +
		"WHERE scope_type = ? AND scope_id IN ? AND tenant_id = ? AND deleted_at IS NULL"

	ProjectScope    = "project"
	DepartmentScope = "department"
	ClusterScope    = "cluster"
)

func InitGoRoutines(ctx context.Context, stopCh <-chan struct{}, tenantsClient tenants_api.ITenantsManager, orgUnitClient org_unit_client.IOrgUnitClient) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("scopes sync job: Init go routine for deleted scopes sync job")

	jobConfig := config.Get[*authorization_config.SyncJob]()
	if jobConfig == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("scopes sync job: failed to get config for sync")
		return
	}
	intervalHours := jobConfig.SyncJobIntervalHours
	tenantInterval := jobConfig.SyncJobTenantIntervalSeconds
	if intervalHours <= 0 || tenantInterval <= 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("scopes sync job: failed to get sync job config")
		return
	}

	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)

	go func() {
		syncScopeAccessRules(ctx, tenantsClient, orgUnitClient, tenantInterval, intervalHours)
		for {
			select {
			case <-ticker.C:
				syncScopeAccessRules(ctx, tenantsClient, orgUnitClient, tenantInterval, intervalHours)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func syncScopeAccessRules(ctx context.Context, tenantsClient tenants_api.ITenantsManager, orgUnitClient org_unit_client.IOrgUnitClient, tenantIntervalSeconds int, intervalHours int) {
	tenants, err := getTenants(ctx, tenantsClient)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to get tenants, err %v", err)
		return
	}

	ticker := time.NewTicker(time.Duration(tenantIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for _, tenant := range tenants {
		<-ticker.C
		syncScopeAccessRulesTenant(ctx, tenant, orgUnitClient, intervalHours)
		ticker.Reset(time.Duration(tenantIntervalSeconds) * time.Second)
	}
}

func syncScopeAccessRulesTenant(ctx context.Context, tenant tenants_openapi.Tenant, orgUnitClient org_unit_client.IOrgUnitClient, intervalHours int) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("syncScopeAccessRulesTenant for tenant %d", tenant.Id)
	deletedAfter := time.Now().Add(time.Duration(-intervalHours-1) * time.Hour)

	tenantId := fmt.Sprint(tenant.Id)
	ctxWithTenantId := context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	deletedOrgUnits, err := orgUnitClient.GetInternalDeletedOrgUnits(ctxWithTenantId, deletedAfter)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get deleted org units for tenant %v: %+v", tenant.Id, err)
		return
	}
	clusters, departments, projects := sortOrgUnits(ctx, deletedOrgUnits)
	deleteGhostClustersAccessRules(ctx, tenantId, clusters, deletedAfter)
	deleteGhostDepartmentsAccessRules(ctx, tenantId, departments, deletedAfter)
	deleteGhostProjectsAccessRules(ctx, tenantId, projects, deletedAfter)
}

func executeCleanup(ctx context.Context, tenantId string, ids []string, scope string, deletedAfter time.Time) {
	rowsAffected, err := runaidb.DB().WithContext(ctx).Exec(GhostAccessRules, scope, ids, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to execute cleanup for tenant %s, scope type: %s, ids: %v, from time: %v, error: %+v", tenantId, scope, ids, deletedAfter, err)
		return
	}
	if rowsAffected > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Execute cleanup for tenant %s, scope type: %s, ids: %v, from time: %v, deleted %d access rules", tenantId, scope, ids, deletedAfter, rowsAffected)
	}
}

func sortOrgUnits(ctx context.Context, deletedOrgUnits []openapi.DeletedOrgUnit) ([]string, []string, []string) {
	deletedProjects := []string{}
	deletedDepartments := []string{}
	deletedClusters := []string{}

	for _, deletedOrgUnit := range deletedOrgUnits {
		if deletedOrgUnit.DeletedAt == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Recieved org unit with deleted_at null: %s", *deletedOrgUnit.Id)
			return nil, nil, nil
		}
		if deletedOrgUnit.Id == nil || deletedOrgUnit.Type == nil {
			continue
		}
		switch *deletedOrgUnit.Type {
		case "project":
			deletedProjects = append(deletedProjects, *deletedOrgUnit.Id)
		case "department":
			deletedDepartments = append(deletedDepartments, *deletedOrgUnit.Id)
		case "cluster":
			deletedClusters = append(deletedClusters, *deletedOrgUnit.Id)
		}
	}
	return deletedClusters, deletedDepartments, deletedProjects
}

func deleteGhostProjectsAccessRules(ctx context.Context, tenantId string, deletedProjects []string, deletedAfter time.Time) {
	if len(deletedProjects) == 0 {
		return
	}
	executeCleanup(ctx, tenantId, deletedProjects, ProjectScope, deletedAfter)
}

func deleteGhostDepartmentsAccessRules(ctx context.Context, tenantId string, deletedDepartments []string, deletedAfter time.Time) {
	if len(deletedDepartments) == 0 {
		return
	}
	executeCleanup(ctx, tenantId, deletedDepartments, DepartmentScope, deletedAfter)
}

func deleteGhostClustersAccessRules(ctx context.Context, tenantId string, deletedClusters []string, deletedAfter time.Time) {
	if len(deletedClusters) == 0 {
		return
	}
	executeCleanup(ctx, tenantId, deletedClusters, ClusterScope, deletedAfter)
}

func getTenants(ctx context.Context, tenantsClient tenants_api.ITenantsManager) ([]tenants_openapi.Tenant, error) {
	tenants, response, err := tenantsClient.ListAllTenantsInternal(ctx)
	if err != nil {
		return []tenants_openapi.Tenant{}, err
	}
	if response.StatusCode != http.StatusOK {
		return []tenants_openapi.Tenant{}, fmt.Errorf("failed to get tenants, response code: %v", response.StatusCode)
	}
	return tenants, nil
}
