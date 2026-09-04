/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources_sync

const NodePoolTable = "node_pool_list AS (SELECT unnest(ARRAY%s) AS node_pool_id)"

const NodePoolNameTable = "node_pool_list AS (SELECT unnest(ARRAY%s) AS node_pool_id, unnest(ARRAY%s) AS node_pool_name)"

const FullProjectResourcesTable = "full_project AS " +
	"(SELECT project.id as project_id,project.org_unit_data_id as project_org_unit_data_id,project.cluster_id,project.tenant_id, project.deleted_at as project_deleted_at, resources.* " +
	"FROM project LEFT JOIN resources ON project.org_unit_data_id = resources.org_unit_data_id)"

const FullDepartmentsResourcesTable = "full_department AS " +
	"(SELECT department.id as department_id,department.org_unit_data_id as department_org_unit_data_id, department.cluster_id,department.tenant_id, department.deleted_at as department_deleted_at, resources.* " +
	"FROM department LEFT JOIN resources ON department.org_unit_data_id = resources.org_unit_data_id)"

const MissingProjectNodePoolsOrgUnits = "WITH " + NodePoolTable + ", " + FullProjectResourcesTable +
	" SELECT count(project_org_unit_data_id), project_org_unit_data_id, project_id FROM full_project " +
	"LEFT JOIN node_pool_list ON full_project.node_pool_id = node_pool_list.node_pool_id " +
	"WHERE project_deleted_at is null and deleted_at is null AND cluster_id='%s' AND tenant_id='%s' GROUP BY project_org_unit_data_id, project_id HAVING COUNT(project_org_unit_data_id) != %d;"

const MissingDepartmentNodePoolsOrgUnits = "WITH " + NodePoolTable + ", " + FullDepartmentsResourcesTable +
	" SELECT count(department_org_unit_data_id), department_org_unit_data_id, department_id FROM full_department " +
	"LEFT JOIN node_pool_list ON full_department.node_pool_id = node_pool_list.node_pool_id " +
	"WHERE department_deleted_at is null and deleted_at is null AND cluster_id='%s' AND tenant_id='%s' GROUP BY department_org_unit_data_id, department_id HAVING COUNT(department_org_unit_data_id) != %d;"

const DeletedProjectNodePoolsOrgUnits = "WITH " + NodePoolTable + ", " + FullProjectResourcesTable +
	" SELECT DISTINCT project_org_unit_data_id, project_id FROM full_project " +
	"LEFT JOIN node_pool_list ON full_project.node_pool_id = node_pool_list.node_pool_id " +
	"WHERE project_deleted_at is null and deleted_at is null AND cluster_id='%s' AND tenant_id='%s' AND node_pool_list.node_pool_id is null;"

const DeletedDepartmentNodePoolsOrgUnits = "WITH " + NodePoolTable + ", " + FullDepartmentsResourcesTable +
	" SELECT DISTINCT department_org_unit_data_id, department_id FROM full_department " +
	"LEFT JOIN node_pool_list ON full_department.node_pool_id = node_pool_list.node_pool_id " +
	"WHERE department_deleted_at is null and deleted_at is null AND cluster_id='%s' AND tenant_id='%s' AND node_pool_list.node_pool_id is null;"

const EmptyProjects = "SELECT id as project_id, org_unit_data_id as project_org_unit_data_id FROM %s WHERE resources_json IS NULL AND cluster_id='%s' AND tenant_id='%s'"

const EmptyDepartments = "SELECT id as department_id, org_unit_data_id as department_org_unit_data_id FROM %s WHERE resources_json IS NULL AND cluster_id='%s' AND tenant_id='%s'"

const DeletedProjectDefaultNodePoolsOrgUnits = "WITH " + NodePoolNameTable +
	" SELECT org_unit_data.id, project.id as project_id, project.cluster_id, project.tenant_id FROM org_unit_data JOIN project ON org_unit_data.id = project.org_unit_data_id " +
	"CROSS JOIN unnest(org_unit_data.default_node_pool_list) AS default_node_pools " +
	"WHERE NOT EXISTS ( " +
	"SELECT 1 " +
	"FROM node_pool_list " +
	"WHERE node_pool_list.node_pool_name = default_node_pools " +
	") AND project.deleted_at is null AND cluster_id='%s' AND tenant_id='%s';"

const DeletedProjectDefaultNodePoolsEffectiveOrgUnits = "WITH " + NodePoolNameTable +
	" SELECT org_unit_data.id, project.id as project_id, project.cluster_id, project.tenant_id FROM org_unit_data JOIN project ON org_unit_data.id = project.org_unit_data_id " +
	"CROSS JOIN unnest(org_unit_data.default_node_pool_list_effective) AS default_node_pools " +
	"WHERE NOT EXISTS ( " +
	"SELECT 1 " +
	"FROM node_pool_list " +
	"WHERE node_pool_list.node_pool_name = default_node_pools " +
	") AND org_unit_data.default_node_pool_list_effective_origin is not null AND project.deleted_at is null AND cluster_id='%s' AND tenant_id='%s';"

const DeletedDepartmentDefaultNodePoolsOrgUnits = "WITH " + NodePoolNameTable +
	" SELECT org_unit_data.id, department.id as department_id, department.cluster_id, department.tenant_id FROM org_unit_data JOIN department ON org_unit_data.id = department.org_unit_data_id " +
	"CROSS JOIN unnest(org_unit_data.default_node_pool_list) AS default_node_pools " +
	"WHERE NOT EXISTS ( " +
	"SELECT 1 " +
	"FROM node_pool_list " +
	"WHERE node_pool_list.node_pool_name = default_node_pools " +
	") AND department.deleted_at is null AND cluster_id='%s' AND tenant_id='%s';"
