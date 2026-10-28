/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/google/uuid"
)

var (
	MapStaticPermissionSetNameToId = map[string]uuid.UUID{}
	PermissionSets                 = []models.PermissionSet{
		static_permission_sets.WorkspaceEdit,
		static_permission_sets.TrainingEdit,
		static_permission_sets.InferenceEdit,
		static_permission_sets.YamlWorkloadEdit,
		static_permission_sets.WorkloadRead,
		static_permission_sets.WorkloadPropertiesEdit,
		static_permission_sets.WorkloadPropertiesRead,
		static_permission_sets.TemplatesEdit,
		static_permission_sets.TemplatesRead,
		static_permission_sets.TemplatesOnlyEdit,
		static_permission_sets.TemplatesOnlyRead,
		static_permission_sets.EnvironmentsOnlyEdit,
		static_permission_sets.EnvironmentsOnlyRead,
		static_permission_sets.ComputeResourcesOnlyEdit,
		static_permission_sets.ComputeResourcesOnlyRead,
		static_permission_sets.CredentialsOnlyEdit,
		static_permission_sets.CredentialsOnlyRead,
		static_permission_sets.DataStorageEdit,
		static_permission_sets.DataStorageRead,
		static_permission_sets.PoliciesEdit,
		static_permission_sets.PoliciesRead,
		static_permission_sets.ProjectEdit,
		static_permission_sets.ProjectRead,
		static_permission_sets.DepartmentEdit,
		static_permission_sets.DepartmentRead,
		static_permission_sets.ClusterEdit,
		static_permission_sets.ClusterRead,
		static_permission_sets.ClusterConfigEdit,
		static_permission_sets.ClusterConfigRead,
		static_permission_sets.NetworkTopologyEdit,
		static_permission_sets.NetworkTopologyRead,
		static_permission_sets.NodesRead,
		static_permission_sets.NodePoolsEdit,
		static_permission_sets.NodePoolsRead,
		static_permission_sets.OverviewDashboardRead,
		static_permission_sets.QuotaManagementDashboardRead,
		static_permission_sets.UsersEdit,
		static_permission_sets.UsersRead,
		static_permission_sets.UsersMinimalRead,
		static_permission_sets.ServiceAccountsEdit,
		static_permission_sets.ServiceAccountsRead,
		static_permission_sets.AccessRulesEdit,
		static_permission_sets.AccessRulesRead,
		static_permission_sets.RolesEdit,
		static_permission_sets.RolesRead,
		static_permission_sets.SettingsEdit,
		static_permission_sets.SettingsRead,
		static_permission_sets.SecuritySettingsEdit,
		static_permission_sets.SecuritySettingsRead,
		static_permission_sets.NotificationSettingsEdit,
		static_permission_sets.NotificationSettingsRead,
		static_permission_sets.BrandingSettingsEdit,
		static_permission_sets.BrandingSettingsRead,
		static_permission_sets.EventHistoryRead,
		static_permission_sets.AccountEdit,
		static_permission_sets.AccountRead,
		static_permission_sets.IntegrationMetricsRead,
		static_permission_sets.AccessKeysEdit,
		static_permission_sets.ApplicationEdit,
		static_permission_sets.ApplicationRead,
		static_permission_sets.StorageClassConfigEdit,
		static_permission_sets.InstallerServiceAccountPS,
		static_permission_sets.ClusterServiceAccountPS,
	}
)

func InitStaticPermissionSets() {
	logging.Infof("Initializing static permission sets")

	initializerDao := PermissionSetDaoInitializer{}
	for _, permissionSet := range PermissionSets {
		id, err := initializerDao.DbUpsertPermissionSet(context.Background(), permissionSet)
		if err != nil {
			logging.Warnf("Error in initializing permission set %s. err: %+v", permissionSet.Name, err)
		}
		MapStaticPermissionSetNameToId[permissionSet.Name] = id
	}

	logging.Infof("Initializing static permission sets completed")
}

func InitPermissionSetsMap(psMap *map[string]uuid.UUID) {
	logging.Infof("Initializing permission sets map")
	permissionSets, _ := DbGetPermissionSets(context.Background(), true)

	for _, ps := range *permissionSets {
		(*psMap)[ps.Name] = ps.ID
	}
}
