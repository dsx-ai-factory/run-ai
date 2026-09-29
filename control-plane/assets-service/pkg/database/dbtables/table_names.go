/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbtables

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
)

const (
	TableNameWorkloadSupportedTypes     = "assets_supported_workload_types"
	WorkloadSupportedTypesColumnAssetId = "asset_id"
	TableNameAssetsMeta                 = "assets_metadata"
	AssetMetaColumnId                   = "id"
	AssetMetaColumnName                 = "name"
	AssetMetaColumnKind                 = "kind"
	AssetMetaColumnScope                = "scope"
	AssetMetaColumnClusterId            = "cluster_id"
	AssetMetaColumnProjectId            = "project_id"
	AssetMetaColumnDepartmentId         = "department_id"
	AssetMetaColumnCreationTime         = "created_at"
	AssetMetaColumnTenantId             = "tenant_id"
	AssetMetaColumnUpdateCount          = "update_count"
	AssetMetaColumnCreatedBy            = "created_by"
	AssetMetaColumnDeletedBy            = "deleted_by"
	AssetMetaColumnDeletedAt            = "deleted_at"
	AssetMetaColumnUpdatedAt            = "updated_at"
	AssetMetaColumnPrevName             = "prev_name"
	AssetMetaColumnPrivate              = "private"

	TableNameAssetsCompute = "assets_compute"

	TableNameAssetsEnvironment             = "assets_environment"
	EnvironmentColumnEnvironmentVariables  = "environment_variables"
	EnvironmentColumnRoutingMethod         = "routing_method"
	EnvironmentColumnRoutingMethodHostBase = "host-base"
	EnvironmentColumnRoutingMethodPathBase = "path-base"
	EnvironmentColumnRoutingMethodAgnostic = "agnostic"

	TableNameAssetsS3        = "assets_s3"
	S3ColumnAccessKeyAssetId = "access_key_asset_id"

	TableNameAssetsGit       = "assets_git"
	GitColumnPasswordAssetId = "password_asset_id"
	GitColumnCredentialId    = "credential_id"

	TableNameAssetsAccessKeys     = "assets_access_key"
	TableNameAssetsPassword       = "assets_password"
	TableNameAssetsDockerRegistry = "assets_docker_registry"
	TableNameAssetsRegistry       = "assets_registry"
	TableNameAssetsNFS            = "assets_nfs"

	TableNameAssetsPVC      = "assets_pvc"
	PVCAssetClaimNameColumn = "claim_name"

	TableNameAssetsHostPath            = "assets_host_path"
	TableNameAssetsTemplate            = "assets_template"
	TableNameAssetsTemplateContent     = "assets_template_content"
	AssetsTemplateContentColumnAssetId = "template_asset_id"
	TableNameAssetsConfigMap           = "assets_config_map"
	TableNameAssetsGenericSecret       = "assets_generic_secret"
	TableNameAssetsNgcApiKey           = "assets_ngc_api_key"

	// Common name for asset id column in assets tables
	AssetIdColumnName = "asset_id"

	TableNameAssetsSecretVolume         = "assets_secret_volume"
	SecretVolumeColumnCredentialAssetId = "credential_asset_id" //nolint:gosec

	// TableNameAssetsSpec is the generic spec table. it stores the spec as
	// JSON keyed by asset_id, so many assets can share it.
	TableNameAssetsSpec = "assets_spec"

	TableNameModel = "assets_model"

	TableNameTenantSettings = "tenant_settings"

	TableNamePolicyCompliance        = "policy_compliance"
	PolicyComplianceColumnAssetId    = "asset_id"
	PolicyComplianceColumnPolicyTime = "policy_time"

	TableNameWorkloads              = "workloads"
	TableNameWorkloadsAssets        = "workloads_assets"
	WorkloadsAssetsColumnAssetId    = "asset_id"
	WorkloadsAssetsColumnWorkloadId = "workload_id"
	WorkloadsColumnId               = "id"
	WorkloadsColumnTenantId         = "tenant_id"
	WorkloadsColumnClusterId        = "cluster_id"
	WorkloadsColumnProjectId        = "project_id"
	WorkloadsColumnWorkloadType     = "workload_type"
	WorkloadsColumnDistFramework    = "dist_framework"
	WorkloadsColumnCurrentState     = "current_state"
	WorkloadsColumnUpdatedBy        = "updated_by"

	TableNameLastUsedByWorkload     = "last_used_by_workload"
	LastUsedByWorkloadColumnAssetId = "asset_id"

	// TableNameSyncStatus - this is the new sync status table, used by the agent starting 2.21
	TableNameSyncStatus         = "sync_status"
	SyncStatusColumnTenantId    = "tenant_id"
	SyncStatusColumnAssetId     = "asset_id"
	SyncStatusColumnUpdateCount = "update_count"

	// TableNameAssetsNameLocks this is the table that locks the creation of asset of a specific kind with a specific name. in order to prevent creating 2 assets with the same name in the same hierarchy
	TableNameAssetsNameLocks = "assets_name_locks"

	// TableNamePvcClaimNameLocks this is the table that locks the creation of PVC asset with a specific claim name. in order to prevent creating 2 PVC assets with the same claim name in the same hierarchy
	TableNamePvcClaimNameLocks = "pvc_claim_name_locks"

	// TableNamePvcDatavolume table for pvc-datavolume links
	TableNamePvcDatavolume = "assets_pvc_datavolume"

	// TableNameEvents table for events
	TableNameEvents = "events"

	// TableNameAssetsResourcesIds table for asset resource ids
	TableNameAssetsResourcesIds = "assets_resources_ids"
	ResourceIdColumnName        = "resource_id"
	LastSyncedAtColumnName      = "last_synced_at"
)

// TableToClean - all the table that we clean as part of cleanup job, with their asset id column
var TableToClean = map[string]string{
	TableNameAssetsAccessKeys:       AssetIdColumnName,
	TableNameAssetsCompute:          AssetIdColumnName,
	TableNameAssetsConfigMap:        AssetIdColumnName,
	TableNameAssetsDockerRegistry:   AssetIdColumnName,
	TableNameAssetsEnvironment:      AssetIdColumnName,
	TableNameAssetsGenericSecret:    AssetIdColumnName,
	TableNameAssetsGit:              AssetIdColumnName,
	TableNameAssetsHostPath:         AssetIdColumnName,
	TableNameModel:                  AssetIdColumnName,
	TableNameAssetsNFS:              AssetIdColumnName,
	TableNameAssetsPassword:         AssetIdColumnName,
	TableNameAssetsPVC:              AssetIdColumnName,
	TableNameAssetsRegistry:         AssetIdColumnName,
	TableNameAssetsS3:               AssetIdColumnName,
	TableNameAssetsSecretVolume:     AssetIdColumnName,
	TableNameAssetsSpec:             AssetIdColumnName,
	TableNameWorkloadSupportedTypes: AssetIdColumnName,
	TableNameLastUsedByWorkload:     AssetIdColumnName,
	TableNamePolicyCompliance:       AssetIdColumnName,
	TableNameSyncStatus:             AssetIdColumnName,

	TableNameAssetsMeta: AssetMetaColumnId,
}

var TablesOfWorkloadAssets = []string{
	TableNameAssetsEnvironment,
	TableNameAssetsCompute,
	TableNameAssetsS3,
	TableNameAssetsGit,
	TableNameAssetsNFS,
	TableNameAssetsPVC,
	TableNameAssetsHostPath,
	TableNameAssetsConfigMap,
	TableNameAssetsSecretVolume,
}

var TablesOfSyncedAssets = []string{
	TableNameAssetsPVC,
	TableNameAssetsAccessKeys,
	TableNameAssetsPassword,
	TableNameAssetsGenericSecret,
	TableNameAssetsDockerRegistry,
	TableNameAssetsNgcApiKey,
}

var TableOfCredentials = []string{
	TableNameAssetsAccessKeys,
	TableNameAssetsPassword,
	TableNameAssetsGenericSecret,
	TableNameAssetsNgcApiKey,
}

// ScopeIdColumnName returns the column name of the scope id for a given scope type
func ScopeIdColumnName(scopeType api.ScopeType) (string, error) {
	switch scopeType {
	case api.SCOPETYPE_TENANT:
		return AssetMetaColumnTenantId, nil
	case api.SCOPETYPE_CLUSTER:
		return AssetMetaColumnClusterId, nil
	case api.SCOPETYPE_DEPARTMENT:
		return AssetMetaColumnDepartmentId, nil
	case api.SCOPETYPE_PROJECT:
		return AssetMetaColumnProjectId, nil
	case api.SCOPETYPE_SYSTEM:
		return "", fmt.Errorf("system scope does not have a scope id column")
	}
	return "", fmt.Errorf("unknown scope type: %s", scopeType)
}
