/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/structs"
)

const (
	templatesByIdSelect          = "SELECT * FROM crud_templates WHERE crud_templates.id = ? AND ((tenant_id = 0 or tenant_id = ?)) AND ((scope_type != ? or id not in (select preset_template_id from crud_templates_presets where tenant_id = ?))) AND crud_templates.id = ?"
	templatesByIdScopedSelect    = "SELECT * FROM crud_templates WHERE crud_templates.id = ? AND ((tenant_id = 0 or tenant_id = ?)) AND ((scope_type != ? or id not in (select preset_template_id from crud_templates_presets where tenant_id = ?))) AND ((tenant_id = 0 or tenant_id = ?)) AND ((tenant_id = 0 or (false OR project_id IN ( 'p1'))))"
	templatesListMinimal         = templatesListPrefix + templatesForkedClause + templatesOrderClause
	templatesCountMinimal        = templatesCountPrefix + templatesForkedClause
	templatesNameCountTenant     = "SELECT count(*) FROM crud_templates WHERE name = ? AND ( ( (scope_type = ?) OR (tenant_id = ?) ))"
	templatesNameCountCluster    = "SELECT count(*) FROM crud_templates WHERE name = ? AND ( ( (scope_type = ?) OR (scope_type = ? AND tenant_id = ?) OR (cluster_id = ?) ))"
	templatesNameCountDepartment = "SELECT count(*) FROM crud_templates WHERE name = ? AND ( ( (scope_type = ?) OR (scope_type = ? AND tenant_id = ?) OR (scope_type = ? AND cluster_id = ?) OR (department_id = ?) ))"
	templatesNameCountProject    = "SELECT count(*) FROM crud_templates WHERE name = ? AND ( ( (scope_type = ?) OR (scope_type = ? AND tenant_id = ?) OR (scope_type = ? AND cluster_id = ?) OR (scope_type = ? AND department_id = ?) OR (scope_type = ? AND project_id = ?) ))"
	templatesListAssetsIn        = "SELECT * FROM crud_templates_assets WHERE template_id in (?)"
	templatesPresetsSelect       = "SELECT * FROM crud_templates WHERE crud_templates.scope_type = ?"
	templatesByAssetIdSelect     = "SELECT DISTINCT crud_templates.id,crud_templates.created_at,crud_templates.created_by,crud_templates.updated_at,crud_templates.updated_by,crud_templates.workload_type,crud_templates.name,crud_templates.description,crud_templates.tenant_id,crud_templates.scope_type,crud_templates.scope_id,crud_templates.cluster_id,crud_templates.department_id,crud_templates.project_id,crud_templates.routing_method,crud_templates.spec,crud_templates.master_spec,crud_templates.master_same_as_worker,crud_templates.combined_spec,crud_templates.combined_master_spec,crud_templates.runnable,crud_templates.runnable_issue FROM crud_templates JOIN crud_templates_assets ON crud_templates.tenant_id = crud_templates_assets.tenant_id AND crud_templates.id = crud_templates_assets.template_id WHERE crud_templates_assets.tenant_id = ? AND crud_templates_assets.asset_id = ?"
	templatesByAssetIdPreload    = "SELECT * FROM crud_templates_assets WHERE crud_templates_assets.template_id = ?"
	templatesAssetsInSelect      = "TODO_ASSETS_IN"
	templatesPluckByProject      = "SELECT id FROM crud_templates WHERE tenant_id = ? AND project_id = ?"
	templatesPluckByCluster      = "SELECT id FROM crud_templates WHERE tenant_id = ? AND cluster_id = ?"
	templatesPluckByTenant       = "SELECT id FROM crud_templates WHERE tenant_id = ? AND tenant_id = ?"
	templatesComplianceByIds     = "DELETE FROM policy_compliance WHERE tenant_id = ? AND template_id IN (?)"
	templatesComplianceByProject = "DELETE FROM policy_compliance WHERE tenant_id = ? AND project_id = ?"
	templatesComplianceByTenant  = "DELETE FROM policy_compliance WHERE tenant_id = ?"
	templatesAssetsByIds         = "DELETE FROM crud_templates_assets WHERE tenant_id = ? AND template_id IN (?)"
	templatesScopeDeleteProject  = "DELETE FROM crud_templates WHERE tenant_id = ? AND project_id = ?"
	templatesScopeDeleteCluster  = "DELETE FROM crud_templates WHERE tenant_id = ? AND cluster_id = ?"
	templatesScopeDeleteTenant   = "DELETE FROM crud_templates WHERE tenant_id = ? AND tenant_id = ?"
	templatesInsert              = "INSERT INTO crud_templates (id,created_at,created_by,updated_at,updated_by,workload_type,name,description,tenant_id,scope_type,scope_id,cluster_id,department_id,project_id,routing_method,spec,master_spec,master_same_as_worker,combined_spec,combined_master_spec,runnable,runnable_issue) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	templatesSave                = "UPDATE crud_templates SET id=?,created_at=?,created_by=?,updated_at=?,updated_by=?,workload_type=?,name=?,description=?,tenant_id=?,scope_type=?,scope_id=?,cluster_id=?,department_id=?,project_id=?,routing_method=?,spec=?,master_spec=?,master_same_as_worker=?,combined_spec=?,combined_master_spec=?,runnable=?,runnable_issue=? WHERE crud_templates.id = ? AND id = ?"
	templatesAssetsDelete        = "DELETE FROM crud_templates_assets WHERE crud_templates_assets.tenant_id = ? AND crud_templates_assets.template_id = ?"
	templatesAssetsInsert        = "INSERT INTO crud_templates_assets (id,tenant_id,asset_id,template_id,replica_type,category) VALUES (?,?,?,?,?,?)"
	templatesDeleteById          = "DELETE FROM crud_templates WHERE crud_templates.id = ? AND crud_templates.tenant_id = ?"
	templatesAssetsOfSelect      = "SELECT * FROM crud_templates_assets WHERE crud_templates_assets.tenant_id = ? AND crud_templates_assets.template_id = ?"
	templatesRawByIdSelect       = "SELECT * FROM crud_templates WHERE crud_templates.id = ? AND ((tenant_id = 0 or tenant_id = ?)) AND ((scope_type != ? or id not in (select preset_template_id from crud_templates_presets where tenant_id = ?))) AND crud_templates.id = ?"
	templatesRawByIdScopedSelect = "SELECT * FROM crud_templates WHERE crud_templates.id = ? AND ((tenant_id = 0 or tenant_id = ?)) AND ((scope_type != ? or id not in (select preset_template_id from crud_templates_presets where tenant_id = ?))) AND ((tenant_id = 0 or tenant_id = ?)) AND ((tenant_id = 0 or (false OR project_id IN ( 'p1'))))"
)

const (
	templatesListCols     = "crud_templates.id,crud_templates.created_at,crud_templates.created_by,crud_templates.updated_at,crud_templates.updated_by,crud_templates.workload_type,crud_templates.name,crud_templates.description,crud_templates.tenant_id,crud_templates.scope_type,crud_templates.scope_id,crud_templates.cluster_id,crud_templates.department_id,crud_templates.project_id,crud_templates.routing_method,crud_templates.spec,crud_templates.master_spec,crud_templates.master_same_as_worker,crud_templates.combined_spec,crud_templates.combined_master_spec,crud_templates.runnable,crud_templates.runnable_issue"
	templatesListPrefix   = "SELECT " + templatesListCols + " FROM crud_templates WHERE ((tenant_id = 0 or tenant_id = ?))"
	templatesCountPrefix  = "SELECT count(*) FROM crud_templates WHERE ((tenant_id = 0 or tenant_id = ?))"
	templatesForkedClause = " AND ((scope_type != ? or id not in (select preset_template_id from crud_templates_presets where tenant_id = ?)))"
	templatesOrderClause  = " ORDER BY id DESC"

	templatesScopedClause    = " AND ((tenant_id = 0 or (false OR project_id IN ( 'p1'))))"
	templatesWorkloadClause  = " AND workload_type = ?"
	templatesFilterByClause  = " AND name = ?"
	templatesProjAncClause   = " AND ( ((scope_type = ? or scope_type = ?) or (scope_type = ? and project_id = ?) or (scope_type = ? and department_id = ?) or (scope_type = ? and cluster_id = ?) ))"
	templatesClusterAncClaus = " AND ( ((scope_type = ? or scope_type = ?) or (scope_type IN (?,?,?) and cluster_id = ?) ))"
	templatesRoutingClause   = " AND routing_method IN (?, '', ?)"
	templatesSearchClause    = " AND (search_vector @@ to_tsquery('english', ? || ':*'))"
	templatesSortClause      = " ORDER BY name DESC NULLS LAST,id DESC LIMIT ? OFFSET ?"

	templatesListWorkloadType     = templatesListPrefix + templatesWorkloadClause + templatesForkedClause + templatesOrderClause
	templatesListScoped           = templatesListPrefix + templatesScopedClause + templatesForkedClause + templatesOrderClause
	templatesListFilterBy         = templatesListPrefix + templatesForkedClause + templatesFilterByClause + templatesOrderClause
	templatesListProjectAncestors = templatesListPrefix + templatesProjAncClause + templatesForkedClause + templatesOrderClause
	templatesListClusterAncestors = templatesListPrefix + templatesClusterAncClaus + templatesForkedClause + templatesOrderClause
	templatesListRouting          = templatesListPrefix + templatesForkedClause + templatesRoutingClause + templatesOrderClause
	templatesListSort             = templatesListPrefix + templatesForkedClause + templatesSortClause
	templatesListSearch           = templatesListPrefix + templatesForkedClause + templatesSearchClause + templatesOrderClause

	templatesCountScoped           = templatesCountPrefix + templatesScopedClause + templatesForkedClause
	templatesCountFilterBy         = templatesCountPrefix + templatesForkedClause + templatesFilterByClause
	templatesCountProjectAncestors = templatesCountPrefix + templatesProjAncClause + templatesForkedClause
	templatesCountClusterAncestors = templatesCountPrefix + templatesClusterAncClaus + templatesForkedClause
	templatesCountRouting          = templatesCountPrefix + templatesForkedClause + templatesRoutingClause
	templatesCountSearch           = templatesCountPrefix + templatesForkedClause + templatesSearchClause
)

const templatesTestId = "22222222-2222-2222-2222-222222222222"

type templateRow struct {
	ID           uuid.UUID `gorm:"column:id"`
	Name         string
	WorkloadType string
	ScopeType    string
	TenantId     uint
}

type templateAssetRow struct {
	ID         uuid.UUID `gorm:"column:id"`
	TenantId   uint
	AssetId    string
	TemplateId string
}

func templatesPermittedScopes() *openapi.PermittedScopes {
	return &openapi.PermittedScopes{
		Projects: []string{"p1"},
	}
}

var _ = Describe("GetTemplateById", func() {
	It("returns an error when the id is not a valid uuid", func() {
		setupDaoMock(GinkgoT())

		template, apiErr := GetTemplateById(testCtx(7), 7, "not-a-uuid", nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns the template with its assets", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByIdSelect).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId), Name: "tpl"}})
		m.ExpectQuery(templatesAssetsOfSelect).Returns([]templateAssetRow{{ID: uuid.New(), TenantId: 7, AssetId: "a1", TemplateId: templatesTestId}})

		template, apiErr := GetTemplateById(testCtx(7), 7, templatesTestId, nil)

		Expect(apiErr).To(BeNil())
		Expect(template).ToNot(BeNil())
		Expect(template.Assets).To(HaveLen(1))
	})

	It("scopes the query when permitted scopes are given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByIdScopedSelect).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId), Name: "tpl"}})
		m.ExpectQuery(templatesAssetsOfSelect).Returns([]templateAssetRow{})

		template, apiErr := GetTemplateById(testCtx(7), 7, templatesTestId, templatesPermittedScopes())

		Expect(apiErr).To(BeNil())
		Expect(template).ToNot(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByIdSelect).ReturnsError(errors.New("boom"))

		template, apiErr := GetTemplateById(testCtx(7), 7, templatesTestId, nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the template does not exist", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByIdSelect).Returns([]templateRow{})

		template, apiErr := GetTemplateById(testCtx(7), 7, templatesTestId, nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when loading the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByIdSelect).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId)}})
		m.ExpectQuery(templatesAssetsOfSelect).ReturnsError(errors.New("boom"))

		template, apiErr := GetTemplateById(testCtx(7), 7, templatesTestId, nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("GetRawTemplateById", func() {
	It("returns an error when the id is not a valid uuid", func() {
		setupDaoMock(GinkgoT())

		template, apiErr := GetRawTemplateById(testCtx(7), 7, "nope", nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns the raw template", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesRawByIdScopedSelect).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId), Name: "tpl"}})

		template, apiErr := GetRawTemplateById(testCtx(7), 7, templatesTestId, templatesPermittedScopes())

		Expect(apiErr).To(BeNil())
		Expect(template).ToNot(BeNil())
		Expect(template.Name).To(Equal("tpl"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesRawByIdSelect).ReturnsError(errors.New("boom"))

		template, apiErr := GetRawTemplateById(testCtx(7), 7, templatesTestId, nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the template does not exist", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesRawByIdSelect).Returns([]templateRow{})

		template, apiErr := GetRawTemplateById(testCtx(7), 7, templatesTestId, nil)

		Expect(template).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("MakeTemplateNameAndType", func() {
	It("builds the name and type key", func() {
		key := MakeTemplateNameAndType("name", "workspace")

		Expect(key).To(Equal(TemplateNameAndType{Name: "name", WorkloadType: "workspace"}))
	})
})

func templatesFixture() *models_crud.Template {
	return &models_crud.Template{
		ID: uuid.MustParse(templatesTestId),
		TemplateMeta: models_crud.TemplateMeta{
			TenantId:     7,
			Name:         "tpl",
			WorkloadType: "workspace",
			ScopeType:    string(api.SCOPETYPE_TENANT),
			ScopeId:      "7",
		},
	}
}

func templatesAssetsFixture() []models_crud.TemplateAssets {
	return []models_crud.TemplateAssets{
		{ID: uuid.MustParse("33333333-3333-3333-3333-333333333333"), TenantId: 7, AssetId: "a1", Category: "compute", ReplicaType: "worker"},
	}
}

var _ = Describe("CreateTemplate", func() {
	It("creates the template and its assets", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesInsert).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(0)
		m.ExpectExec(templatesAssetsInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := CreateTemplate(testCtx(7), templatesFixture(), templatesAssetsFixture())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := CreateTemplate(testCtx(7), templatesFixture(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := CreateTemplate(testCtx(7), templatesFixture(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when clearing the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesInsert).Returns(1)
		m.ExpectExec(templatesAssetsDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := CreateTemplate(testCtx(7), templatesFixture(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when inserting the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesInsert).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(0)
		m.ExpectExec(templatesAssetsInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := CreateTemplate(testCtx(7), templatesFixture(), templatesAssetsFixture())

		Expect(apiErr).ToNot(BeNil())
	})

	It("panics when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesInsert).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(0)

		Expect(func() {
			_ = CreateTemplate(testCtx(7), templatesFixture(), nil)
		}).To(Panic())
	})
})

var _ = Describe("UpdateTemplate", func() {
	It("updates the template and its assets", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesSave).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(1)
		m.ExpectExec(templatesAssetsInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := UpdateTemplate(testCtx(7), templatesFixture(), templatesAssetsFixture())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := UpdateTemplate(testCtx(7), templatesFixture(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the save fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesSave).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := UpdateTemplate(testCtx(7), templatesFixture(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when clearing the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesSave).Returns(1)
		m.ExpectExec(templatesAssetsDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := UpdateTemplate(testCtx(7), templatesFixture(), nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("panics when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesSave).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(1)

		Expect(func() {
			_ = UpdateTemplate(testCtx(7), templatesFixture(), nil)
		}).To(Panic())
	})
})

var _ = Describe("DeleteTemplateById", func() {
	It("deletes the template, its assets and its compliance", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesDeleteById).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(1)
		m.ExpectExec(complianceDeleteOfTemplate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesDeleteById).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns a 404 when no template was deleted", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesDeleteById).Returns(0)
		m.ExpectTransaction().Rollback()

		apiErr := DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(404)))
	})

	It("returns an error when deleting the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesDeleteById).Returns(1)
		m.ExpectExec(templatesAssetsDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when deleting the compliance fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesDeleteById).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(1)
		m.ExpectExec(complianceDeleteOfTemplate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).ToNot(BeNil())
	})

	It("panics when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(templatesDeleteById).Returns(1)
		m.ExpectExec(templatesAssetsDelete).Returns(1)
		m.ExpectExec(complianceDeleteOfTemplate).Returns(1)

		Expect(func() {
			_ = DeleteTemplateById(testCtx(7), 7, uuid.MustParse(templatesTestId))
		}).To(Panic())
	})
})

type templatesIdRow struct {
	ID string `gorm:"column:id"`
}

var _ = Describe("DeleteTemplatesOfProjectId", func() {
	It("deletes the templates of the project", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByProject).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).Returns(1)
		m.ExpectExec(templatesComplianceByProject).Returns(1)
		m.ExpectExec(templatesAssetsByIds).Returns(1)
		m.ExpectExec(templatesScopeDeleteProject).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteTemplatesOfProjectId(testCtx(7), "7", "proj-1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("succeeds when the project has no templates", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByProject).Returns([]templatesIdRow{})
		m.ExpectExec(templatesComplianceByProject).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteTemplatesOfProjectId(testCtx(7), "7", "proj-1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when deleting the compliance of the project fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByProject).Returns([]templatesIdRow{})
		m.ExpectExec(templatesComplianceByProject).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfProjectId(testCtx(7), "7", "proj-1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteTemplatesOfClusterId", func() {
	It("deletes the templates of the cluster", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).Returns(1)
		m.ExpectExec(templatesAssetsByIds).Returns(1)
		m.ExpectExec(templatesScopeDeleteCluster).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("succeeds when the cluster has no templates", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).Returns([]templatesIdRow{})
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).To(HaveOccurred())
	})

	It("returns an error when plucking the template ids fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).To(HaveOccurred())
	})

	It("returns an error when deleting the compliance by ids fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).To(HaveOccurred())
	})

	It("returns an error when deleting the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).Returns(1)
		m.ExpectExec(templatesAssetsByIds).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).To(HaveOccurred())
	})

	It("returns an error when deleting the templates fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).Returns(1)
		m.ExpectExec(templatesAssetsByIds).Returns(1)
		m.ExpectExec(templatesScopeDeleteCluster).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).To(HaveOccurred())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByCluster).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).Returns(1)
		m.ExpectExec(templatesAssetsByIds).Returns(1)
		m.ExpectExec(templatesScopeDeleteCluster).Returns(1)

		err := DeleteTemplatesOfClusterId(testCtx(7), "7", "cluster-1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteTemplatesOfTenantId", func() {
	It("deletes the templates of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByTenant).Returns([]templatesIdRow{{ID: templatesTestId}})
		m.ExpectExec(templatesComplianceByIds).Returns(1)
		m.ExpectExec(templatesComplianceByTenant).Returns(1)
		m.ExpectExec(templatesAssetsByIds).Returns(1)
		m.ExpectExec(templatesScopeDeleteTenant).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteTemplatesOfTenantId(testCtx(7), "7")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when deleting the compliance of the tenant fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(templatesPluckByTenant).Returns([]templatesIdRow{})
		m.ExpectExec(templatesComplianceByTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteTemplatesOfTenantId(testCtx(7), "7")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeletePresetForTenant", func() {
	It("marks the preset as deleted for the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(presetsSaveUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DeletePresetForTenant(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(presetsSaveUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DeletePresetForTenant(testCtx(7), 7, uuid.MustParse(templatesTestId))

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("ListPresetTemplates", func() {
	It("returns the preset templates keyed by name and type", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesPresetsSelect).Returns([]templateRow{
			{ID: uuid.MustParse(templatesTestId), Name: "tpl", WorkloadType: "workspace"},
		})

		presets, apiErr := ListPresetTemplates(testCtx(7))

		Expect(apiErr).To(BeNil())
		Expect(presets).To(HaveLen(1))
		Expect(presets).To(HaveKey(MakeTemplateNameAndType("tpl", "workspace")))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesPresetsSelect).ReturnsError(errors.New("boom"))

		presets, apiErr := ListPresetTemplates(testCtx(7))

		Expect(presets).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("ListTemplatesByAssetId", func() {
	It("returns the templates that use the asset", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByAssetIdSelect).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId)}})
		m.ExpectQuery(templatesByAssetIdPreload).Returns([]templateAssetRow{{ID: uuid.New(), TenantId: 7, AssetId: "a1", TemplateId: templatesTestId}})

		templates, apiErr := ListTemplatesByAssetId(testCtx(7), 7, "a1")

		Expect(apiErr).To(BeNil())
		Expect(templates).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesByAssetIdSelect).ReturnsError(errors.New("boom"))

		templates, apiErr := ListTemplatesByAssetId(testCtx(7), 7, "a1")

		Expect(templates).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

func templatesSystemScopes() *openapi.PermittedScopes {
	system := true
	return &openapi.PermittedScopes{System: &system}
}

func templatesAncestorsOfProject() *authorization_utils.Ancestors {
	clusterId := "cluster-1"
	departmentId := "dep-1"
	projectId := "proj-1"
	return &authorization_utils.Ancestors{TenantId: "7", ClusterUuid: &clusterId, DepartmentId: &departmentId, ProjectId: &projectId}
}

func templatesAncestorsOfCluster() *authorization_utils.Ancestors {
	clusterId := "cluster-1"
	return &authorization_utils.Ancestors{TenantId: "7", ClusterUuid: &clusterId}
}

func templatesFilterBy() []queryapi.Filter {
	return []queryapi.Filter{{Field: "name", Operator: "=", Values: []string{"tpl"}}}
}

type templatesListCase struct {
	name           string
	sql            string
	permittedScope *openapi.PermittedScopes
	offset         int32
	limit          int32
	workloadType   *api.WorkloadType
	sortBy         string
	sortOrder      string
	filterBy       []queryapi.Filter
	search         string
	ancestors      *authorization_utils.Ancestors
	routingMethod  string
}

func templatesListCases() []templatesListCase {
	workloadType := api.WORKLOADTYPE_WORKSPACE
	return []templatesListCase{
		{name: "Minimal", sql: templatesListMinimal},
		{name: "WorkloadType", sql: templatesListWorkloadType, workloadType: &workloadType},
		{name: "SystemScopes", sql: templatesListMinimal, permittedScope: templatesSystemScopes()},
		{name: "LimitedScopes", sql: templatesListScoped, permittedScope: templatesPermittedScopes()},
		{name: "FilterBy", sql: templatesListFilterBy, filterBy: templatesFilterBy()},
		{name: "ProjectAncestors", sql: templatesListProjectAncestors, ancestors: templatesAncestorsOfProject()},
		{name: "ClusterAncestors", sql: templatesListClusterAncestors, ancestors: templatesAncestorsOfCluster()},
		{name: "RoutingMethod", sql: templatesListRouting, routingMethod: "host-base"},
		{name: "SortAndPaging", sql: templatesListSort, offset: 5, limit: 10, sortBy: "name", sortOrder: "DESC"},
		{name: "Search", sql: templatesListSearch, search: "abc"},
	}
}

func templatesListTableArgs(body func(templatesListCase)) []interface{} {
	args := []interface{}{body}
	for _, tc := range templatesListCases() {
		args = append(args, Entry(tc.name, tc))
	}
	return args
}

var _ = Describe("ListTemplates", func() {
	DescribeTable("builds the expected query for each filter permutation", templatesListTableArgs(
		func(tc templatesListCase) {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(tc.sql).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId)}})
			m.ExpectQuery(templatesListAssetsIn).Returns([]templateAssetRow{{ID: uuid.New(), TenantId: 7, AssetId: "a1", TemplateId: templatesTestId}})

			templates, apiErr := ListTemplates(testCtx(7), 7, tc.permittedScope, tc.offset, tc.limit, tc.workloadType,
				tc.sortOrder, tc.sortBy, tc.filterBy, tc.search, tc.ancestors, tc.routingMethod)

			Expect(apiErr).To(BeNil())
			Expect(templates).To(HaveLen(1))
			Expect(templates[0].Assets).To(HaveLen(1))
		})...,
	)

	It("ignores assets that belong to an unknown template", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesListMinimal).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId)}})
		m.ExpectQuery(templatesListAssetsIn).Returns([]templateAssetRow{{ID: uuid.New(), TenantId: 7, AssetId: "a1", TemplateId: "44444444-4444-4444-4444-444444444444"}})

		templates, apiErr := ListTemplates(testCtx(7), 7, nil, 0, 0, nil, "", "", nil, "", nil, "")

		Expect(apiErr).To(BeNil())
		Expect(templates).To(HaveLen(1))
		Expect(templates[0].Assets).To(BeEmpty())
	})

	It("returns an error when the list query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesListMinimal).ReturnsError(errors.New("boom"))

		templates, apiErr := ListTemplates(testCtx(7), 7, nil, 0, 0, nil, "", "", nil, "", nil, "")

		Expect(templates).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when loading the assets fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesListMinimal).Returns([]templateRow{{ID: uuid.MustParse(templatesTestId)}})
		m.ExpectQuery(templatesListAssetsIn).ReturnsError(errors.New("boom"))

		templates, apiErr := ListTemplates(testCtx(7), 7, nil, 0, 0, nil, "", "", nil, "", nil, "")

		Expect(templates).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

type templatesCountCase struct {
	name           string
	sql            string
	permittedScope *openapi.PermittedScopes
	filterBy       []queryapi.Filter
	search         string
	ancestors      *authorization_utils.Ancestors
	routingMethod  string
}

func templatesCountCases() []templatesCountCase {
	return []templatesCountCase{
		{name: "Minimal", sql: templatesCountMinimal},
		{name: "SystemScopes", sql: templatesCountMinimal, permittedScope: templatesSystemScopes()},
		{name: "LimitedScopes", sql: templatesCountScoped, permittedScope: templatesPermittedScopes()},
		{name: "FilterBy", sql: templatesCountFilterBy, filterBy: templatesFilterBy()},
		{name: "ProjectAncestors", sql: templatesCountProjectAncestors, ancestors: templatesAncestorsOfProject()},
		{name: "ClusterAncestors", sql: templatesCountClusterAncestors, ancestors: templatesAncestorsOfCluster()},
		{name: "RoutingMethod", sql: templatesCountRouting, routingMethod: "host-base"},
		{name: "Search", sql: templatesCountSearch, search: "abc"},
	}
}

func templatesCountTableArgs(body func(templatesCountCase)) []interface{} {
	args := []interface{}{body}
	for _, tc := range templatesCountCases() {
		args = append(args, Entry(tc.name, tc))
	}
	return args
}

var _ = Describe("CountTemplates", func() {
	DescribeTable("builds the expected query for each filter permutation", templatesCountTableArgs(
		func(tc templatesCountCase) {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(tc.sql).Returns([]countRow{{Count: 3}})

			count, apiErr := CountTemplates(testCtx(7), 7, tc.permittedScope, tc.filterBy, tc.search, tc.ancestors, tc.routingMethod)

			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(3)))
		})...,
	)

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesCountMinimal).ReturnsError(errors.New("boom"))

		count, apiErr := CountTemplates(testCtx(7), 7, nil, nil, "", nil, "")

		Expect(count).To(BeZero())
		Expect(apiErr).ToNot(BeNil())
	})
})

type templatesOrgUnitMock struct {
	ancestors *orgUnitApi.Ancestors
	err       *orgUnitApi.Error
}

func (m *templatesOrgUnitMock) GetProjectById(context.Context, string) (*orgUnitApi.Project, *orgUnitApi.Error) {
	return nil, nil
}

func (m *templatesOrgUnitMock) GetProjectsMetadataByDepartment(context.Context, string) ([]orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	return nil, nil
}

func (m *templatesOrgUnitMock) GetAncestorsOfScope(context.Context, orgUnitApi.ScopeType, string) (*orgUnitApi.Ancestors, *orgUnitApi.Error) {
	return m.ancestors, m.err
}

func (m *templatesOrgUnitMock) GetOrgTreeMap(context.Context) (structs.OrgUnitMap, map[structs.OrgUnitKey]string, *orgUnitApi.Error) {
	return structs.OrgUnitMap{}, nil, nil
}

func (m *templatesOrgUnitMock) GetDepartmentName(context.Context, string) (string, *orgUnitApi.Error) {
	return "", nil
}

func templatesInstallOrgUnitMock(mock *templatesOrgUnitMock) {
	org_unit.SetMock(mock)
	DeferCleanup(func() { org_unit.SetMock(nil) })
}

func templatesAncestors(clusterId *string, departmentId *string) *orgUnitApi.Ancestors {
	return &orgUnitApi.Ancestors{
		TenantId:     "7",
		ClusterUuid:  *orgUnitApi.NewNullableString(clusterId),
		DepartmentId: *orgUnitApi.NewNullableString(departmentId),
	}
}

var _ = Describe("CheckTemplateNameAvailability", func() {
	It("reports the name as available in tenant scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesNameCountTenant).Returns([]countRow{{Count: 0}})

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_TENANT), "7", nil)

		Expect(apiErr).To(BeNil())
		Expect(available).To(BeTrue())
	})

	It("reports the name as taken in tenant scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesNameCountTenant).Returns([]countRow{{Count: 1}})

		templateId := templatesTestId
		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_TENANT), "7", &templateId)

		Expect(apiErr).To(BeNil())
		Expect(available).To(BeFalse())
	})

	It("reports the name as available in cluster scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesNameCountCluster).Returns([]countRow{{Count: 0}})

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_CLUSTER), "cluster-1", nil)

		Expect(apiErr).To(BeNil())
		Expect(available).To(BeTrue())
	})

	It("reports the name as available in department scope", func() {
		clusterId := "cluster-1"
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{ancestors: templatesAncestors(&clusterId, nil)})
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesNameCountDepartment).Returns([]countRow{{Count: 0}})

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_DEPARTMENT), "dep-1", nil)

		Expect(apiErr).To(BeNil())
		Expect(available).To(BeTrue())
	})

	It("returns an error when the department ancestors lookup fails", func() {
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{err: orgUnitApi.NewError(500, "nope")})
		setupDaoMock(GinkgoT())

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_DEPARTMENT), "dep-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})

	It("returns an error when the department has no cluster ancestor", func() {
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{ancestors: templatesAncestors(nil, nil)})
		setupDaoMock(GinkgoT())

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_DEPARTMENT), "dep-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})

	It("reports the name as available in project scope", func() {
		clusterId := "cluster-1"
		departmentId := "dep-1"
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{ancestors: templatesAncestors(&clusterId, &departmentId)})
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesNameCountProject).Returns([]countRow{{Count: 0}})

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_PROJECT), "proj-1", nil)

		Expect(apiErr).To(BeNil())
		Expect(available).To(BeTrue())
	})

	It("returns an error when the project ancestors lookup fails", func() {
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{err: orgUnitApi.NewError(404, "nope")})
		setupDaoMock(GinkgoT())

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_PROJECT), "proj-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})

	It("returns an error when the project has no cluster ancestor", func() {
		departmentId := "dep-1"
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{ancestors: templatesAncestors(nil, &departmentId)})
		setupDaoMock(GinkgoT())

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_PROJECT), "proj-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})

	It("returns an error when the project has no department ancestor", func() {
		clusterId := "cluster-1"
		templatesInstallOrgUnitMock(&templatesOrgUnitMock{ancestors: templatesAncestors(&clusterId, nil)})
		setupDaoMock(GinkgoT())

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_PROJECT), "proj-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})

	It("returns an error for an unsupported scope type", func() {
		setupDaoMock(GinkgoT())

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", "galaxy", "g-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(templatesNameCountTenant).ReturnsError(errors.New("boom"))

		available, apiErr := CheckTemplateNameAvailability(testCtx(7), 7, "tpl", string(api.SCOPETYPE_TENANT), "7", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(available).To(BeFalse())
	})
})
