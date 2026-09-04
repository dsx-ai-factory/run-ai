/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sqlmock

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

const (
	tmplCols  = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_template.*`
	tmplJoins = `FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_template on assets_metadata.id = assets_template.asset_id`
	tmplScope = `(assets_metadata.scope = ? OR assets_metadata.tenant_id = ?)`

	tmplGet = tmplCols + ` ` + tmplJoins + ` WHERE (assets_metadata.id = ? AND assets_metadata.kind = ?) AND ` + tmplScope + ` AND assets_metadata.deleted_at IS NULL`

	tmplContentIds = `SELECT contained_asset_id FROM assets_template_content WHERE assets_template_content.template_asset_id = ?`
	tmplContentGet = `SELECT * FROM assets_metadata WHERE id in (?) AND assets_metadata.deleted_at IS NULL`

	tmplEnvAssets = `SELECT assets_template_content.*, assets_metadata.*, assets_supported_workload_types.* FROM assets_template_content left join assets_metadata on assets_template_content.contained_asset_id = assets_metadata.id left join assets_supported_workload_types on assets_template_content.contained_asset_id = assets_supported_workload_types.asset_id WHERE assets_template_content.template_asset_id IN (?) and assets_metadata.kind = ?`

	tmplSpecInsert  = `INSERT INTO assets_template (asset_id,specific_env) VALUES (?,?)`
	tmplContentIns  = `INSERT INTO assets_template_content (template_asset_id,contained_asset_id) VALUES (?,?)`
	tmplSpecFirst   = `SELECT * FROM assets_template WHERE assets_template.asset_id = ? ORDER BY assets_template.asset_id LIMIT ?`
	tmplSpecSave    = `UPDATE assets_template SET specific_env=? WHERE asset_id = ?`
	tmplSpecHardDel = `DELETE FROM assets_template WHERE assets_template.asset_id = ? AND assets_template.asset_id = ?`
	tmplContentDel  = `DELETE FROM assets_template_content WHERE assets_template_content.template_asset_id = ?`

	tmplUsingAssets = `SELECT * FROM assets_template_content join assets_metadata on assets_metadata.id = assets_template_content.template_asset_id WHERE assets_template_content.contained_asset_id = ? AND ((scope = ? or tenant_id = ?)) ORDER BY assets_metadata.name`

	tmplListMeta = `SELECT assets_metadata.*, assets_supported_workload_types.* FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id WHERE COALESCE(assets_metadata.private, FALSE) = false AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND id IN (?) AND assets_metadata.deleted_at IS NULL`

	tmplContainment   = `SELECT * FROM assets_template_content join assets_metadata on assets_metadata.id = assets_template_content.contained_asset_id`
	tmplListTemplates = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_template.* FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_template on assets_metadata.id = assets_template.asset_id WHERE (assets_metadata.auto_delete is null or assets_metadata.auto_delete = ?) AND COALESCE(assets_metadata.private, FALSE) = false AND assets_metadata.kind = ? AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND assets_metadata.deleted_at IS NULL`
	tmplEnvList       = `SELECT assets_metadata.*, assets_supported_workload_types.*,assets_environment.* FROM assets_metadata left join assets_supported_workload_types on assets_supported_workload_types.asset_id = assets_metadata.id left join assets_environment on assets_metadata.id = assets_environment.asset_id WHERE assets_metadata.kind = ? AND (assets_metadata.scope = ? OR assets_metadata.tenant_id = ?) AND assets_metadata.deleted_at IS NULL`
)

func tmplGetRow() []dbmodel.TemplateCombinedRec {
	return []dbmodel.TemplateCombinedRec{{
		AssetMeta: api.AssetMeta{Id: "t-1", Kind: api.ASSETKIND_WORKLOAD_TEMPLATE},
	}}
}

func containedIdRows() []struct{ ContainedAssetId string } {
	return []struct{ ContainedAssetId string }{{ContainedAssetId: "env-1"}}
}

func containedMetaRows() []dbmodel.AssetMetaRec {
	return []dbmodel.AssetMetaRec{{AssetMeta: dbmodel.AssetMeta{Id: "env-1", Kind: api.ASSETKIND_ENVIRONMENT}}}
}

func envAssetRows() []dao.TemplateEnvironmentAsset {
	return []dao.TemplateEnvironmentAsset{{
		AssetMeta:          api.AssetMeta{Id: "env-1"},
		TemplateContentRec: dbmodel.TemplateContentRec{TemplateAssetId: "t-1", ContainedAssetId: "env-1"},
	}}
}

func TestGetTemplateAssetById(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(tmplGet, "t-1", "workload-template", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(tmplGetRow())
	m.ExpectQuery(tmplContentIds, dbtesting.AnyArg()).Returns(containedIdRows())
	m.ExpectQuery(tmplContentGet, dbtesting.AnyArg()).Returns(containedMetaRows())
	got, ids, apiErr := dao.GetTemplateAssetById(auth.MockAuthInfo(), "t-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
	require.NotNil(t, ids)
}

func TestCreateTemplateAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectAssetMetaCreated(m, "my-tmpl", "workload-template")
	m.ExpectExec(tmplSpecInsert).Returns(1)
	m.ExpectExec(tmplContentIns).Returns(1)
	m.ExpectQuery(tmplGet, dbtesting.AnyArg(), "workload-template", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(tmplGetRow())
	m.ExpectQuery(tmplContentIds, dbtesting.AnyArg()).Returns(containedIdRows())
	m.ExpectQuery(tmplContentGet, dbtesting.AnyArg()).Returns(containedMetaRows())
	m.ExpectQuery(tmplEnvAssets, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(envAssetRows())
	m.ExpectTransaction().Commit()

	got, apiErr := dao.CreateTemplateAsset(auth.MockAuthInfo(), api.WorkloadTemplateCreationRequest{
		Meta: api.AssetCreationRequest{Name: "my-tmpl", Scope: api.SCOPE_PROJECT},
		Spec: api.SpecificRunCreationFields{Assets: api.AssetsIds{Environment: "env-1"}},
	}, []string{"env-1"}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestUpdateTemplateAsset(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectUpdateAssetMeta(m)
	m.ExpectQuery(tmplListMeta, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(containedMetaRows())
	m.ExpectQuery(tmplSpecFirst, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.TemplateAssetRec{{AssetId: "t-1"}})
	m.ExpectExec(tmplSpecSave).Returns(1)
	m.ExpectQuery(tmplContentIds, dbtesting.AnyArg()).Returns(containedIdRows())
	m.ExpectTransaction().Commit()
	m.ExpectQuery(tmplGet, dbtesting.AnyArg(), "workload-template", dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(tmplGetRow())
	m.ExpectQuery(tmplContentIds, dbtesting.AnyArg()).Returns(containedIdRows())
	m.ExpectQuery(tmplContentGet, dbtesting.AnyArg()).Returns(containedMetaRows())
	m.ExpectQuery(tmplEnvAssets, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(envAssetRows())

	got, apiErr := dao.UpdateTemplateAsset(auth.MockAuthInfo(), "t-1", api.WorkloadTemplateUpdateRequest{
		Meta: api.AssetUpdateRequest{Name: updateAssetMetaName},
		Spec: api.SpecificRunCreationFields{Assets: api.AssetsIds{Environment: "env-1"}},
	}, nil)
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestDeleteTemplateAsset_Hard(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectHardDeleteAssetMeta(m)
	m.ExpectExec(tmplSpecHardDel, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(1)
	m.ExpectExec(tmplContentDel, dbtesting.AnyArg()).Returns(1)
	m.ExpectTransaction().Commit()
	apiErr := dao.DeleteTemplateAsset(auth.MockAuthInfo(), "asset-1", dao.HardDelete)
	require.Nil(t, apiErr)
}

func TestDeleteTemplateAsset_Soft(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectTransaction().Begin()
	expectSoftDeleteUnsupported(m, "workload-template")
	apiErr := dao.DeleteTemplateAsset(auth.MockAuthInfo(), "asset-1", dao.SoftDelete)
	require.NotNil(t, apiErr)
	require.Contains(t, apiErr.Message, "does not support soft deletion")
}

func TestListTemplatesUsingAssets(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(tmplUsingAssets, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.TemplateAssetsCombinedRec{{
		TemplateAssetId:  "t-1",
		ContainedAssetId: "env-1",
		AssetMeta:        dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Id: "t-1"}},
	}})
	got, apiErr := dao.ListTemplatesUsingAssets(auth.MockAuthInfo(), strptr("env-1"))
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListTemplatesUsingAssetId(t *testing.T) {
	m := setupDaoMock(t)
	m.ExpectQuery(tmplUsingAssets, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.TemplateAssetsCombinedRec{{
		TemplateAssetId:  "t-1",
		ContainedAssetId: "env-1",
		AssetMeta:        dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Id: "t-1"}},
	}})
	got, apiErr := dao.ListTemplatesUsingAssetId(auth.MockAuthInfo(), "env-1")
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}

func TestListTemplates(t *testing.T) {
	m := setupDaoMock(t)
	dao.ForceTenantCloneForTesting(0)
	m.ExpectQuery(tmplListTemplates, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(tmplGetRow())
	m.ExpectQuery(tmplEnvAssets, dbtesting.AnyArg(), dbtesting.AnyArg()).Returns(envAssetRows())
	m.ExpectTransaction().Begin()
	m.ExpectQuery(svTenantSettings, dbtesting.AnyArg()).Returns([]dbmodel.TenantSettingsRec{{Value: "5"}})
	m.ExpectTransaction().Rollback()
	m.ExpectQuery(tmplEnvList, dbtesting.AnyArg(), dbtesting.AnyArg(), dbtesting.AnyArg()).Returns([]dbmodel.EnvironmentCombinedRec{})
	m.ExpectQuery(tmplContainment).Returns([]dbmodel.TemplateAssetsCombinedRec{})
	got, apiErr := dao.ListTemplates(auth.MockAuthInfo(), &query_types.QueryParams{})
	require.Nil(t, apiErr)
	require.NotNil(t, got)
}
