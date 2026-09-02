/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	static_legacy_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-legacy-roles"
	static_roles "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func Migrate(sourceClient *sql.DB, tenantsClient *sql.DB, targetClient orm.DB) {
	logging.Infof("Starting roles assignments + users migration")

	tenantsManagerTenants := getTenantsManagerTenants(tenantsClient)

	rows, err := sourceClient.QueryContext(
		context.Background(),
		"SELECT DISTINCT au.user_id AS user_id,"+
			" au.client_id AS client_id,"+
			" au.tenant_id AS tenant_id,"+
			" au.entity_type AS entity_type,"+
			" au.created_by AS created_by,"+
			" au.created_at AS created_at,"+
			" au.updated_at AS updated_at,"+
			" au.all_clusters_permitted AS all_clusters_permitted,"+
			" r.role AS role,"+
			" cp.cluster_uuid AS cluster_uuid,"+
			" da.department_id AS department_id,"+
			" pup.project_id AS project_id,"+
			" c.cluster_name AS cluster_name,"+
			" d.name AS department_name,"+
			" d.cluster_uuid AS department_cluster_uuid,"+
			" p.project_name AS project_name,"+
			" p.cluster_uuid AS project_cluster_uuid,"+
			" t.tenant_name AS tenant_name,"+
			" t.auth_realm AS auth_realm"+
			" FROM auth_user As au"+
			" LEFT JOIN roles As r ON au.user_id = r.user_id AND au.tenant_id = r.tenant_id"+
			" LEFT JOIN department_admins As da ON au.user_id = da.user_id AND au.tenant_id = da.tenant_id AND r.role = 'department_admin'"+
			" LEFT JOIN departments As d ON d.id = da.department_id AND d.tenant_id = da.tenant_id AND r.role = 'department_admin'"+
			" LEFT JOIN projects_users_permissions As pup ON au.user_id = pup.user_id AND au.tenant_id = pup.tenant_id AND r.role = 'researcher'"+
			" LEFT JOIN projects As p ON p.project_id = pup.project_id AND p.tenant_id = pup.tenant_id AND r.role = 'researcher'"+
			" LEFT JOIN cluster_permission As cp ON au.user_id = cp.user_id AND au.tenant_id = cp.tenant_id"+
			" LEFT JOIN clusters As c ON c.cluster_uuid = cp.cluster_uuid AND c.tenant_id = cp.tenant_id"+
			" LEFT JOIN tenants As t ON t.tenant_id = au.tenant_id")
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "42P01" {
			logging.Infof("Tables are not set, this means that migration is not necessary")
			return
		}

		logging.Errorf("Failed to query data. err: %+v", err)
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var row RoleAssignmentRow
		err := rows.Scan(
			&row.UserId,
			&row.ClientId,
			&row.TenantId,
			&row.EntityType,
			&row.CreatedBy,
			&row.CreatedAt,
			&row.UpdatedAt,
			&row.AllClustersPermitted,
			&row.Role,
			&row.ClusterUuid,
			&row.DepartmentId,
			&row.ProjectId,
			&row.ClusterName,
			&row.DepartmentName,
			&row.DepartmentClusterUuid,
			&row.ProjectName,
			&row.ProjectClusterUuid,
			&row.TenantName,
			&row.AuthRealm,
		)
		if err != nil {
			logging.Errorf("Failed to scan data. err: %+v", err)
			continue
		}
		logging.Infof("Migrating row: %+v", row)

		authRealm := getAuthRealm(row, tenantsManagerTenants)
		tenantName := getTenantName(row, tenantsManagerTenants)

		var subjectId string
		switch row.EntityType {
		case EntityTypeRegularUser:
			user, err := MigrateLocalUser(row, authRealm)
			if err != nil {
				continue
			}

			if user.Email != nil {
				subjectId = *user.Email
			} else {
				subjectId = *user.Username
			}
		case EntityTypeSSOUser:
			MigrateSsoUser(row, authRealm)
			subjectId = row.UserId
		case EntityTypeGroup:
			subjectId = row.UserId
		case EntityTypeApp:
			application, err := MigrateApplication(row, authRealm)
			if err != nil {
				continue
			}

			subjectId = *application.ClientID
			if isInternalClient(subjectId) {
				CreateAccessRule(targetClient, row.TenantId, subjectId, row.EntityType, static_roles.ClusterApp.Name, fmt.Sprintf("%d", row.TenantId), api.SCOPETYPE_TENANT, tenantName, "")
				continue
			}
		}

		if row.Role == nil {
			continue
		}

		var roleName string
		var scopeId string
		var scopeType api.ScopeType
		var scopeName string
		var clusterUuid string
		if row.AllClustersPermitted != nil && *row.AllClustersPermitted {
			scopeId = fmt.Sprintf("%d", row.TenantId)
			scopeType = api.SCOPETYPE_TENANT
			scopeName = tenantName
		} else {
			if row.ClusterUuid != nil && row.ClusterName != nil {
				scopeId = *row.ClusterUuid
				scopeType = api.SCOPETYPE_CLUSTER
				scopeName = *row.ClusterName
				clusterUuid = *row.ClusterUuid
			} else {
				logging.Warnf("User is not all clusters permitted, but does not have cluster assigned. %+v", row)
				continue
			}
		}
		switch *row.Role {
		case "admin":
			roleName = static_roles.SystemAdmin.Name
		case "department_admin":
			if row.DepartmentId != nil && row.DepartmentName != nil {
				roleName = static_roles.DepartmentAdmin.Name
				scopeId = fmt.Sprintf("%d", *row.DepartmentId)
				scopeType = api.SCOPETYPE_DEPARTMENT
				scopeName = *row.DepartmentName
				clusterUuid = *row.DepartmentClusterUuid
			} else {
				logging.Debugf("User is department admin, but does not assigned to department. %+v", row)
				continue
			}
		case "editor":
			roleName = static_roles.Editor.Name
		case "ml_engineer":
			roleName = static_legacy_roles.MLEngineer.Name
		case "research_manager":
			roleName = static_legacy_roles.ResearchManager.Name
		case "researcher":
			if row.ProjectId != nil && row.ProjectName != nil {
				roleName = static_legacy_roles.Researcher.Name
				scopeId = fmt.Sprintf("%d", *row.ProjectId)
				scopeType = api.SCOPETYPE_PROJECT
				scopeName = *row.ProjectName
				clusterUuid = *row.ProjectClusterUuid
			} else {
				logging.Debugf("User is researcher, but does not assigned to project. %+v", row)
				continue
			}
		case "viewer":
			roleName = static_roles.Viewer.Name
		}
		CreateAccessRule(targetClient, row.TenantId, subjectId, row.EntityType, roleName, scopeId, scopeType, scopeName, clusterUuid)
	}
	if err := rows.Err(); err != nil {
		// Until all teams fix their code, at the moment, preventing unsafe queries can cause catastrophic behavior, so we just log it for now, but in the near future, the entire function will be removed.
		logging.Errorf("Error iterating migration rows: %+v", err)
	}

	logging.Infof("Roles assignments migration completed successfully")
}

func getTenantName(row RoleAssignmentRow, tenants map[int]string) string {
	if row.TenantName != nil && *row.TenantName != "" {
		return *row.TenantName
	}

	if tenantName, found := tenants[row.TenantId]; found {
		return tenantName
	}

	return "-"
}

func getAuthRealm(row RoleAssignmentRow, tenants map[int]string) string {
	if row.AuthRealm != nil && *row.AuthRealm != "" {
		return *row.AuthRealm
	}

	if tenantName, found := tenants[row.TenantId]; found {
		return tenantName
	}

	return ""
}

func getTenantsManagerTenants(client *sql.DB) map[int]string {
	tenantsManagerTenantsRows, err := client.QueryContext(context.Background(), "SELECT id AS tenant_id, name AS tenant_name FROM tenants")
	if err != nil {
		logging.Warnf("Failed to query tenants manager tenants. err: %+v", err)
		return map[int]string{}
	}
	defer tenantsManagerTenantsRows.Close()

	tenantIdToNameMap := map[int]string{}
	for tenantsManagerTenantsRows.Next() {
		var row TenantRow
		err := tenantsManagerTenantsRows.Scan(
			&row.TenantId,
			&row.TenantName,
		)
		if err != nil {
			logging.Errorf("Failed to scan tenant. err: %+v", err)
			continue
		}

		tenantIdToNameMap[row.TenantId] = row.TenantName
	}
	if err := tenantsManagerTenantsRows.Err(); err != nil {
		logging.Warnf("Error iterating tenants manager tenants. err: %+v", err)
		return map[int]string{}
	}

	return tenantIdToNameMap
}
