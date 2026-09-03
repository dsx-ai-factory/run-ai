/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"net/http"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

const (
	wtIdStr       = "11111111-1111-1111-1111-111111111111"
	wtConfigIdStr = "22222222-2222-2222-2222-222222222222"
	wtPriorityStr = "33333333-3333-3333-3333-333333333333"
	wtCategoryStr = "44444444-4444-4444-4444-444444444444"
	wtClusterStr  = "55555555-5555-5555-5555-555555555555"
)

const (
	wtConfigListSelect = "SELECT DISTINCT ON (id) * FROM external_workloads_type_config_view WHERE tenant_id = ? OR tenant_id IS NULL ORDER BY id,CASE WHEN tenant_id = '1' THEN 0 ELSE 1 END"

	wtStatusesSelect = "SELECT * FROM workload_type_statuses WHERE workload_type_id IN (?)"

	wtStatusesScopedSelect = "SELECT * FROM workload_type_statuses WHERE workload_type_id IN (?) AND ((false OR cluster_id IN ( '55555555-5555-5555-5555-555555555555')))"

	wtConfigByIdSelect = "SELECT * FROM external_workloads_type_config_view WHERE id = ? AND ((scope_type = ? OR tenant_id = ?)) ORDER BY CASE WHEN tenant_id = '1' THEN 0 ELSE 1 END,external_workloads_type_config_view.id LIMIT ?"

	wtCountByNameGroup = "SELECT count(*) FROM external_workloads_type_config_view WHERE name = ? AND group = ? AND ((scope_type = ? OR tenant_id = ?))"

	wtConfigForScopeTenantSelect = "SELECT * FROM external_workloads_type_config WHERE workload_type_id = ? AND tenant_id = ? AND scope_type = ? ORDER BY external_workloads_type_config.id LIMIT ?"

	wtTypesListSelect = "SELECT * FROM workload_types"

	wtTypeByIdSelect = "SELECT * FROM workload_types WHERE id = ? ORDER BY workload_types.id LIMIT ?"

	wtConfigByNameGroupSelect      = "SELECT * FROM external_workloads_type_config_view WHERE name = ? AND group = ? AND ((scope_type = ? OR tenant_id = ?)) ORDER BY CASE WHEN tenant_id = '1' THEN 0 ELSE 1 END,external_workloads_type_config_view.id LIMIT ?"
	wtConfigByNameNoGroupSelect    = "SELECT * FROM external_workloads_type_config_view WHERE name = ? AND source != ? AND ((scope_type = ? OR tenant_id = ?)) ORDER BY CASE WHEN tenant_id = '1' THEN 0 ELSE 1 END,external_workloads_type_config_view.id LIMIT ?"
	wtConfigForScopeSystemSelect   = "SELECT * FROM external_workloads_type_config WHERE workload_type_id = ? AND scope_type = ? ORDER BY external_workloads_type_config.id LIMIT ?"
	wtTypesListFilteredSelect      = "SELECT * FROM workload_types WHERE source = ? AND name = ?"
	wtTypeByNameGroupSelect        = "SELECT * FROM workload_types WHERE name = ? AND group = ? ORDER BY workload_types.id LIMIT ?"
	wtUpsertTypeSelect             = "SELECT * FROM workload_types WHERE name = ? AND group = ? ORDER BY workload_types.id LIMIT ?"
	wtTypeInsert                   = "INSERT INTO workload_types (id,name,created_at,created_by,group,source) VALUES (?,?,?,?,?,?)"
	wtTypeDelete                   = "DELETE FROM workload_types WHERE id = ?"
	wtConfigInsert                 = "INSERT INTO external_workloads_type_config (id,tenant_id,scope_type,workload_type_id,category_id,priority_id,preemptibility,created_at,created_by,updated_at,updated_by) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
	wtKartaSelect                  = "SELECT * FROM workload_type_kartas WHERE workload_type_id = ? AND version = ? AND tenant_id = ? AND scope_type = ? ORDER BY workload_type_kartas.id LIMIT ?"
	wtKartaSystemSelect            = "SELECT * FROM workload_type_kartas WHERE workload_type_id = ? AND version = ? AND tenant_id IS NULL AND scope_type = ? ORDER BY workload_type_kartas.id LIMIT ?"
	wtKartaDeleteExcluded          = "DELETE FROM workload_type_kartas WHERE workload_type_id = ? AND id NOT IN (?) AND tenant_id = ? AND scope_type = ?"
	wtKartaDeleteSystem            = "DELETE FROM workload_type_kartas WHERE workload_type_id = ?"
	wtKartaDeleteTenant            = "DELETE FROM workload_type_kartas WHERE workload_type_id = ? AND tenant_id = ? AND scope_type = ?"
	wtConfigUpdate                 = "UPDATE external_workloads_type_config SET category_id=?,preemptibility=?,priority_id=?,updated_at=?,updated_by=? WHERE tenant_id = ? AND scope_type = ? AND workload_type_id = ?"
	wtConfigUpdateSystem           = "UPDATE external_workloads_type_config SET category_id=?,preemptibility=?,priority_id=?,updated_at=?,updated_by=? WHERE tenant_id IS NULL AND scope_type = ? AND workload_type_id = ?"
	wtConfigUpdateMinimal          = "UPDATE external_workloads_type_config SET updated_at=?,updated_by=? WHERE tenant_id = ? AND scope_type = ? AND workload_type_id = ?"
	wtConfigDeleteTenant           = "DELETE FROM external_workloads_type_config WHERE workload_type_id = ? AND tenant_id = ? AND scope_type = ?"
	wtConfigDeleteSystem           = "DELETE FROM external_workloads_type_config WHERE workload_type_id = ?"
	wtConfigCountAllScopes         = "SELECT count(*) FROM external_workloads_type_config WHERE workload_type_id = ?"
	wtConfigCountSystemScope       = "SELECT count(*) FROM external_workloads_type_config WHERE workload_type_id = ? AND scope_type = ?"
	wtPrioritySelect               = "SELECT * FROM priorities WHERE deleted_at IS NULL AND id = ? AND ((scope_type = 'system' OR tenant_id = ?)) AND priorities.deleted_at IS NULL ORDER BY priorities.id LIMIT ?"
	wtSavepoint                    = "SAVEPOINT sp"
	wtRollbackTo                   = "ROLLBACK TO SAVEPOINT sp"
	wtStatusUpsert                 = "INSERT INTO workload_type_statuses (id,workload_type_id,cluster_id,tenant_id,updated_at,phase,conditions,version) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT (workload_type_id,version,cluster_id,tenant_id) DO UPDATE SET conditions=?,phase=?,updated_at=?"
	wtStatusDeleteTargeted         = "DELETE FROM workload_type_statuses WHERE tenant_id = ? AND cluster_id = ? AND workload_type_id = ? AND version = ?"
	wtStatusDeleteFullSync         = "DELETE FROM workload_type_statuses WHERE tenant_id = ? AND cluster_id = ?"
	wtStatusDeleteFullSyncExcluded = "DELETE FROM workload_type_statuses WHERE tenant_id = ? AND cluster_id = ? AND workload_type_id NOT IN (?)"
	wtConfigDeleteByTenant         = "DELETE FROM external_workloads_type_config WHERE tenant_id = ?"
	wtKartaDeleteByTenant          = "DELETE FROM workload_type_kartas WHERE tenant_id = ?"
	wtStatusDeleteByTenant         = "DELETE FROM workload_type_statuses WHERE tenant_id = ?"
	wtConfigListFilteredSelect     = "SELECT DISTINCT ON (id) * FROM external_workloads_type_config_view WHERE (tenant_id = ? OR tenant_id IS NULL) AND name = ? ORDER BY id,CASE WHEN tenant_id = '1' THEN 0 ELSE 1 END"
)

type wtKartaRow struct {
	ID string `gorm:"column:id"`
}

type wtConfigViewRow struct {
	ID    string `gorm:"column:id"`
	Name  string `gorm:"column:name"`
	Group string `gorm:"column:group"`
}

type wtConfigViewKartasRow struct {
	ID     string `gorm:"column:id"`
	Name   string `gorm:"column:name"`
	Group  string `gorm:"column:group"`
	Kartas []byte `gorm:"column:kartas"`
}

type wtStatusRow struct {
	ID             string `gorm:"column:id"`
	WorkloadTypeId string `gorm:"column:workload_type_id"`
	Version        string `gorm:"column:version"`
}

type wtTypeRow struct {
	ID    string `gorm:"column:id"`
	Name  string `gorm:"column:name"`
	Group string `gorm:"column:group"`
}

type wtConfigRow struct {
	ID             string `gorm:"column:id"`
	WorkloadTypeId string `gorm:"column:workload_type_id"`
	ScopeType      string `gorm:"column:scope_type"`
}

type wtPriorityRow struct {
	ID    string `gorm:"column:id"`
	Value int    `gorm:"column:value"`
}

func wtMgmtCtx(tenantId uint) context.Context {
	return context.WithValue(testCtx(tenantId), runaictx.CfIsManagementTenant, "true")
}

func wtScopes() *authorization_api.PermittedScopes {
	return &authorization_api.PermittedScopes{Clusters: []string{wtClusterStr}}
}

func wtKarta(version string) models.Karta {
	return models.Karta{
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "root",
					Kind: &kartav1alpha1.GroupVersionKind{Group: "g", Version: version, Kind: "Kind"},
				},
			},
		},
	}
}

func wtKartasJson(versions ...string) []byte {
	kartas := models.WorkloadTypeKartaArray{}
	for _, v := range versions {
		kartas = append(kartas, models.WorkloadTypeKarta{ID: uuid.MustParse(wtConfigIdStr), Version: v, Karta: wtKarta(v)})
	}
	v, err := kartas.Value()
	Expect(err).ToNot(HaveOccurred())
	return []byte(v.(string))
}

var _ = Describe("DbListWorkloadTypeConfigs", func() {
	It("returns configs without statuses when statuses are not requested", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListSelect).Returns([]wtConfigViewRow{{ID: wtIdStr, Name: "Training"}})

		configs, apiErr := DbListWorkloadTypeConfigs(testCtx(1), 1, nil, nil, true)

		Expect(apiErr).To(BeNil())
		Expect(configs).To(HaveLen(1))
	})

	It("attaches statuses to the returned configs", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListSelect).Returns([]wtConfigViewRow{{ID: wtIdStr, Name: "Training"}})
		m.ExpectQuery(wtStatusesScopedSelect).Returns([]wtStatusRow{{ID: wtConfigIdStr, WorkloadTypeId: wtIdStr, Version: "v1"}})

		configs, apiErr := DbListWorkloadTypeConfigs(testCtx(1), 1, wtScopes(), nil, false)

		Expect(apiErr).To(BeNil())
		Expect(configs).To(HaveLen(1))
		Expect(configs[0].Statuses).To(HaveLen(1))
	})

	It("applies filters for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListFilteredSelect).Returns([]wtConfigViewRow{{ID: wtIdStr}})
		m.ExpectQuery(wtStatusesSelect).Returns([]wtStatusRow{})

		filters := []validators.ParameterizedFilter{{Field: "name", Operator: "=", FormattedValues: []string{"Training"}}}
		configs, apiErr := DbListWorkloadTypeConfigs(wtMgmtCtx(1), 1, wtScopes(), filters, false)

		Expect(apiErr).To(BeNil())
		Expect(configs).To(HaveLen(1))
	})

	It("queries statuses unscoped when no permitted scopes are given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListSelect).Returns([]wtConfigViewRow{{ID: wtIdStr}})
		m.ExpectQuery(wtStatusesSelect).Returns([]wtStatusRow{})

		configs, apiErr := DbListWorkloadTypeConfigs(testCtx(1), 1, nil, nil, false)

		Expect(apiErr).To(BeNil())
		Expect(configs).To(HaveLen(1))
	})

	It("skips the statuses query when no configs are found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListSelect).Returns([]wtConfigViewRow{})

		configs, apiErr := DbListWorkloadTypeConfigs(testCtx(1), 1, wtScopes(), nil, false)

		Expect(apiErr).To(BeNil())
		Expect(configs).To(BeEmpty())
	})

	It("returns an error when the configs query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListSelect).ReturnsError(errors.New("boom"))

		configs, apiErr := DbListWorkloadTypeConfigs(testCtx(1), 1, nil, nil, true)

		Expect(configs).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the statuses query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigListSelect).Returns([]wtConfigViewRow{{ID: wtIdStr}})
		m.ExpectQuery(wtStatusesScopedSelect).ReturnsError(errors.New("boom"))

		configs, apiErr := DbListWorkloadTypeConfigs(testCtx(1), 1, wtScopes(), nil, false)

		Expect(configs).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadTypeConfigByIdWithStatuses", func() {
	It("returns the config together with its statuses", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByIdSelect).Returns([]wtConfigViewRow{{ID: wtIdStr, Name: "Training"}})
		m.ExpectQuery(wtStatusesScopedSelect).Returns([]wtStatusRow{{ID: wtConfigIdStr, WorkloadTypeId: wtIdStr}})

		config, apiErr := DbGetWorkloadTypeConfigByIdWithStatuses(testCtx(1), 1, wtIdStr, wtScopes())

		Expect(apiErr).To(BeNil())
		Expect(config).ToNot(BeNil())
		Expect(config.Statuses).To(HaveLen(1))
	})

	It("returns an error when the config query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByIdSelect).ReturnsError(errors.New("boom"))

		config, apiErr := DbGetWorkloadTypeConfigByIdWithStatuses(testCtx(1), 1, wtIdStr, wtScopes())

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the statuses query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByIdSelect).Returns([]wtConfigViewRow{{ID: wtIdStr}})
		m.ExpectQuery(wtStatusesScopedSelect).ReturnsError(errors.New("boom"))

		config, apiErr := DbGetWorkloadTypeConfigByIdWithStatuses(testCtx(1), 1, wtIdStr, wtScopes())

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadTypeConfigById", func() {
	It("returns the config", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByIdSelect).Returns([]wtConfigViewRow{{ID: wtIdStr, Name: "Training"}})

		config, apiErr := DbGetWorkloadTypeConfigById(testCtx(1), 1, wtIdStr)

		Expect(apiErr).To(BeNil())
		Expect(config.Name).To(Equal("Training"))
	})
})

var _ = Describe("DbGetWorkloadTypeConfigByNameGroup", func() {
	It("looks up by name and group when a group is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByNameGroupSelect).Returns([]wtConfigViewRow{{ID: wtIdStr, Name: "Training", Group: "g"}})

		config, apiErr := DbGetWorkloadTypeConfigByNameGroup(testCtx(1), 1, "Training", "g")

		Expect(apiErr).To(BeNil())
		Expect(config.Name).To(Equal("Training"))
	})

	It("looks up by name only when the group is empty", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByNameNoGroupSelect).Returns([]wtConfigViewRow{{ID: wtIdStr, Name: "Training"}})

		config, apiErr := DbGetWorkloadTypeConfigByNameGroup(testCtx(1), 1, "Training", "")

		Expect(apiErr).To(BeNil())
		Expect(config.Name).To(Equal("Training"))
	})
})

var _ = Describe("DbCountWorkloadTypeConfigsByNameGroup", func() {
	It("returns the count", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtCountByNameGroup).Returns([]countRow{{Count: 3}})

		count, apiErr := DbCountWorkloadTypeConfigsByNameGroup(testCtx(1), 1, "Training", "g")

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(3)))
	})

	It("returns a zero count and an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtCountByNameGroup).ReturnsError(errors.New("boom"))

		count, apiErr := DbCountWorkloadTypeConfigsByNameGroup(testCtx(1), 1, "Training", "g")

		Expect(apiErr).ToNot(BeNil())
		Expect(count).To(Equal(int64(0)))
	})
})

var _ = Describe("dbGetWorkloadTypeConfigByFields", func() {
	It("returns an internal server error when no fields are given", func() {
		setupDaoMock(GinkgoT())

		config, apiErr := dbGetWorkloadTypeConfigByFields(testCtx(1), 1, nil, nil)

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(http.StatusInternalServerError)))
	})
})

var _ = Describe("DbGetWorkloadTypeConfigByIdForScope", func() {
	It("queries the tenant scope for a regular tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr, WorkloadTypeId: wtIdStr}})

		config, apiErr := DbGetWorkloadTypeConfigByIdForScope(testCtx(1), 1, wtIdStr)

		Expect(apiErr).To(BeNil())
		Expect(config).ToNot(BeNil())
	})

	It("queries the system scope for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigForScopeSystemSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr, WorkloadTypeId: wtIdStr}})

		config, apiErr := DbGetWorkloadTypeConfigByIdForScope(wtMgmtCtx(1), 1, wtIdStr)

		Expect(apiErr).To(BeNil())
		Expect(config).ToNot(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigForScopeTenantSelect).ReturnsError(errors.New("boom"))

		config, apiErr := DbGetWorkloadTypeConfigByIdForScope(testCtx(1), 1, wtIdStr)

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbListWorkloadTypes", func() {
	It("returns the workload types", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtTypesListSelect).Returns([]wtTypeRow{{ID: wtIdStr, Name: "Training"}})

		types, apiErr := DbListWorkloadTypes(testCtx(1), nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(types).To(HaveLen(1))
	})

	It("applies the source and the given filters", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtTypesListFilteredSelect).Returns([]wtTypeRow{{ID: wtIdStr}})

		source := models.WorkloadTypeSourceCustom
		filters := []validators.ParameterizedFilter{{Field: "name", Operator: "=", FormattedValues: []string{"Training"}}}
		types, apiErr := DbListWorkloadTypes(testCtx(1), &source, filters)

		Expect(apiErr).To(BeNil())
		Expect(types).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtTypesListSelect).ReturnsError(errors.New("boom"))

		types, apiErr := DbListWorkloadTypes(testCtx(1), nil, nil)

		Expect(types).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadTypeById", func() {
	It("returns the workload type", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtTypeByIdSelect).Returns([]wtTypeRow{{ID: wtIdStr, Name: "Training"}})

		workloadType, apiErr := DbGetWorkloadTypeById(testCtx(1), uuid.MustParse(wtIdStr))

		Expect(apiErr).To(BeNil())
		Expect(workloadType.Name).To(Equal("Training"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtTypeByIdSelect).ReturnsError(errors.New("boom"))

		workloadType, apiErr := DbGetWorkloadTypeById(testCtx(1), uuid.MustParse(wtIdStr))

		Expect(workloadType).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadTypeByNameGroup", func() {
	It("returns the workload type", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtTypeByNameGroupSelect).Returns([]wtTypeRow{{ID: wtIdStr, Name: "Training", Group: "g"}})

		workloadType, apiErr := DbGetWorkloadTypeByNameGroup(testCtx(1), "Training", "g")

		Expect(apiErr).To(BeNil())
		Expect(workloadType.Group).To(Equal("g"))
	})
})

var _ = Describe("dbGetWorkloadTypeByFields", func() {
	It("returns an error when no fields are given", func() {
		setupDaoMock(GinkgoT())

		workloadType, apiErr := dbGetWorkloadTypeByFields(testCtx(1), nil)

		Expect(workloadType).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetKartaByGVK", func() {
	It("returns the karta matching the requested version", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByNameGroupSelect).Returns([]wtConfigViewKartasRow{{ID: wtIdStr, Name: "Kind", Group: "g", Kartas: wtKartasJson("v1", "v2")}})

		karta, apiErr := DbGetKartaByGVK(testCtx(1), 1, "Kind", "g", "v2")

		Expect(apiErr).To(BeNil())
		Expect(karta).ToNot(BeNil())
		Expect(karta.Version).To(Equal("v2"))
	})

	It("returns nothing when the requested version is not present", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByNameGroupSelect).Returns([]wtConfigViewKartasRow{{ID: wtIdStr, Name: "Kind", Group: "g", Kartas: wtKartasJson("v1")}})

		karta, apiErr := DbGetKartaByGVK(testCtx(1), 1, "Kind", "g", "v9")

		Expect(apiErr).To(BeNil())
		Expect(karta).To(BeNil())
	})

	It("returns an error when the config query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigByNameGroupSelect).ReturnsError(errors.New("boom"))

		karta, apiErr := DbGetKartaByGVK(testCtx(1), 1, "Kind", "g", "v1")

		Expect(karta).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

func wtCreateFields(withKartas bool) WorkloadTypeCreateFieldsParsed {
	fields := WorkloadTypeCreateFieldsParsed{
		WorkloadTypeCreateUpdateFieldsParsed: WorkloadTypeCreateUpdateFieldsParsed{
			WorkloadTypeName: "Training",
			WorkloadTypeId:   uuid.MustParse(wtIdStr),
			CategoryId:       ptr.To(uuid.MustParse(wtCategoryStr)),
			Preemptibility:   ptr.To("preemptible"),
		},
		Group: "g",
	}
	if withKartas {
		fields.Kartas = ptr.To([]models.Karta{wtKarta("v1")})
	}
	return fields
}

var _ = Describe("DbCreateWorkloadTypeWithConfig", func() {
	It("creates the type, its config and its kartas", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtUpsertTypeSelect).Returns([]wtTypeRow{{ID: wtIdStr, Name: "Training", Group: "g"}})
		m.ExpectExec(wtConfigInsert).Returns(1)
		m.ExpectQuery(wtKartaSelect).Returns([]wtKartaRow{{ID: wtConfigIdStr}})
		m.ExpectTransaction().Commit()

		config, apiErr := DbCreateWorkloadTypeWithConfig(testCtx(1), 1, wtCreateFields(true))

		Expect(apiErr).To(BeNil())
		Expect(config).ToNot(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		config, apiErr := DbCreateWorkloadTypeWithConfig(testCtx(1), 1, wtCreateFields(false))

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when upserting the type fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtUpsertTypeSelect).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		config, apiErr := DbCreateWorkloadTypeWithConfig(testCtx(1), 1, wtCreateFields(false))

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when creating the config fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtUpsertTypeSelect).Returns([]wtTypeRow{{ID: wtIdStr}})
		m.ExpectExec(wtConfigInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		config, apiErr := DbCreateWorkloadTypeWithConfig(testCtx(1), 1, wtCreateFields(false))

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when creating the kartas fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtUpsertTypeSelect).Returns([]wtTypeRow{{ID: wtIdStr}})
		m.ExpectExec(wtConfigInsert).Returns(1)
		m.ExpectQuery(wtKartaSelect).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		config, apiErr := DbCreateWorkloadTypeWithConfig(testCtx(1), 1, wtCreateFields(true))

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtUpsertTypeSelect).Returns([]wtTypeRow{{ID: wtIdStr}})
		m.ExpectExec(wtConfigInsert).Returns(1)

		config, apiErr := DbCreateWorkloadTypeWithConfig(testCtx(1), 1, wtCreateFields(false))

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertWorkloadType", func() {
	It("returns the existing type when one is found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtUpsertTypeSelect).Returns([]wtTypeRow{{ID: wtIdStr, Name: "Training", Group: "g"}})

		workloadType, apiErr := DbUpsertWorkloadType(testCtx(1), dao_utils.DB(), wtCreateFields(false))

		Expect(apiErr).To(BeNil())
		Expect(workloadType.ID).To(Equal(uuid.MustParse(wtIdStr)))
	})

	It("creates a new type when none exists", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtUpsertTypeSelect).Returns([]wtTypeRow{})
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtTypeInsert).Returns(1)
		m.ExpectTransaction().Commit()

		workloadType, apiErr := DbUpsertWorkloadType(testCtx(1), dao_utils.DB(), wtCreateFields(false))

		Expect(apiErr).To(BeNil())
		Expect(workloadType).ToNot(BeNil())
	})

	It("returns an error when the lookup fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtUpsertTypeSelect).ReturnsError(errors.New("boom"))

		workloadType, apiErr := DbUpsertWorkloadType(testCtx(1), dao_utils.DB(), wtCreateFields(false))

		Expect(workloadType).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("dbDeleteWorkloadType", func() {
	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtTypeDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := dbDeleteWorkloadType(testCtx(1), dao_utils.DB(), wtIdStr)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("dbCreateExternalWorkloadTypeConfig", func() {
	It("creates a system scoped config for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigInsert).Returns(1)
		m.ExpectTransaction().Commit()

		config, apiErr := dbCreateExternalWorkloadTypeConfig(wtMgmtCtx(1), dao_utils.DB(), 1, wtCreateFields(false))

		Expect(apiErr).To(BeNil())
		Expect(config.ScopeType).To(Equal("system"))
	})

	It("derives the preemptibility from the priority when it is not given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtPrioritySelect).Returns([]wtPriorityRow{{ID: wtPriorityStr, Value: 100}})
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigInsert).Returns(1)
		m.ExpectTransaction().Commit()

		fields := wtCreateFields(false)
		fields.Preemptibility = nil
		fields.PriorityId = ptr.To(uuid.MustParse(wtPriorityStr))

		config, apiErr := dbCreateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, fields)

		Expect(apiErr).To(BeNil())
		Expect(*config.Preemptibility).To(Equal("non-preemptible"))
	})

	It("returns an error when deriving the preemptibility fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtPrioritySelect).ReturnsError(errors.New("boom"))

		fields := wtCreateFields(false)
		fields.Preemptibility = nil
		fields.PriorityId = ptr.To(uuid.MustParse(wtPriorityStr))

		config, apiErr := dbCreateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, fields)

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		config, apiErr := dbCreateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, wtCreateFields(false))

		Expect(config).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("dbCreateWorkloadKartas", func() {
	It("returns no ids when there are no kartas", func() {
		setupDaoMock(GinkgoT())

		ids, apiErr := dbCreateWorkloadKartas(testCtx(1), dao_utils.DB(), 1, wtCreateFields(false).WorkloadTypeCreateUpdateFieldsParsed)

		Expect(apiErr).To(BeNil())
		Expect(ids).To(BeEmpty())
	})

	It("creates system scoped kartas for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtKartaSystemSelect).Returns([]wtKartaRow{{ID: wtConfigIdStr}})

		ids, apiErr := dbCreateWorkloadKartas(wtMgmtCtx(1), dao_utils.DB(), 1, wtCreateFields(true).WorkloadTypeCreateUpdateFieldsParsed)

		Expect(apiErr).To(BeNil())
		Expect(ids).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtKartaSelect).ReturnsError(errors.New("boom"))

		ids, apiErr := dbCreateWorkloadKartas(testCtx(1), dao_utils.DB(), 1, wtCreateFields(true).WorkloadTypeCreateUpdateFieldsParsed)

		Expect(ids).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("dbDeleteWorkloadKartasOfWorkloadTypeInScope", func() {
	It("excludes the given karta ids from the delete", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteExcluded).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := dbDeleteWorkloadKartasOfWorkloadTypeInScope(testCtx(1), dao_utils.DB(), 1, uuid.MustParse(wtIdStr), []uuid.UUID{uuid.MustParse(wtConfigIdStr)})

		Expect(apiErr).To(BeNil())
	})

	It("deletes all kartas of the type for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteSystem).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := dbDeleteWorkloadKartasOfWorkloadTypeInScope(wtMgmtCtx(1), dao_utils.DB(), 1, uuid.MustParse(wtIdStr), nil)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := dbDeleteWorkloadKartasOfWorkloadTypeInScope(testCtx(1), dao_utils.DB(), 1, uuid.MustParse(wtIdStr), nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

func wtUpdateFields(withKartas bool) WorkloadTypeUpdateFieldsParsed {
	fields := WorkloadTypeUpdateFieldsParsed{
		WorkloadTypeCreateUpdateFieldsParsed: WorkloadTypeCreateUpdateFieldsParsed{
			WorkloadTypeName: "Training",
			WorkloadTypeId:   uuid.MustParse(wtIdStr),
			CategoryId:       ptr.To(uuid.MustParse(wtCategoryStr)),
			PriorityId:       ptr.To(uuid.MustParse(wtPriorityStr)),
			Preemptibility:   ptr.To("preemptible"),
		},
	}
	if withKartas {
		fields.Kartas = ptr.To([]models.Karta{wtKarta("v1")})
	}
	return fields
}

var _ = Describe("DbUpdateWorkloadTypeWithConfig", func() {
	It("updates the config and reconciles the kartas", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtConfigUpdate).Returns(1)
		m.ExpectQuery(wtKartaSelect).Returns([]wtKartaRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtKartaDeleteExcluded).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(true))

		Expect(apiErr).To(BeNil())
	})

	It("skips the karta reconciliation when no kartas are given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtConfigUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(false))

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(false))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the config lookup fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtConfigForScopeTenantSelect).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(false))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when creating the kartas fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtConfigUpdate).Returns(1)
		m.ExpectQuery(wtKartaSelect).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(true))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when deleting the stale kartas fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtConfigUpdate).Returns(1)
		m.ExpectQuery(wtKartaSelect).Returns([]wtKartaRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtKartaDeleteExcluded).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(true))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{{ID: wtConfigIdStr}})
		m.ExpectExec(wtConfigUpdate).Returns(1)

		apiErr := DbUpdateWorkloadTypeWithConfig(testCtx(1), 1, wtUpdateFields(false))

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("dbAddOrUpdateExternalWorkloadTypeConfig", func() {
	It("creates the config when none exists for the scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigForScopeTenantSelect).Returns([]wtConfigRow{})
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := dbAddOrUpdateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, wtUpdateFields(false))

		Expect(apiErr).To(BeNil())
	})
})

var _ = Describe("dbUpdateExternalWorkloadTypeConfig", func() {
	It("updates the system scoped config for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigUpdateSystem).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := dbUpdateExternalWorkloadTypeConfig(wtMgmtCtx(1), dao_utils.DB(), 1, wtUpdateFields(false))

		Expect(apiErr).To(BeNil())
	})

	It("derives the preemptibility from the priority when it is not given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtPrioritySelect).Returns([]wtPriorityRow{{ID: wtPriorityStr, Value: 1}})
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		fields := wtUpdateFields(false)
		fields.Preemptibility = nil

		apiErr := dbUpdateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, fields)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when deriving the preemptibility fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtPrioritySelect).ReturnsError(errors.New("boom"))

		fields := wtUpdateFields(false)
		fields.Preemptibility = nil

		apiErr := dbUpdateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, fields)

		Expect(apiErr).ToNot(BeNil())
	})

	It("updates only the audit columns when no optional fields are given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigUpdateMinimal).Returns(1)
		m.ExpectTransaction().Commit()

		fields := WorkloadTypeUpdateFieldsParsed{
			WorkloadTypeCreateUpdateFieldsParsed: WorkloadTypeCreateUpdateFieldsParsed{WorkloadTypeId: uuid.MustParse(wtIdStr)},
		}
		apiErr := dbUpdateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, fields)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := dbUpdateExternalWorkloadTypeConfig(testCtx(1), dao_utils.DB(), 1, wtUpdateFields(false))

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbDeleteWorkloadTypeWithConfig", func() {
	It("deletes the type when the last config is removed", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).Returns(1)
		m.ExpectExec(wtConfigDeleteTenant).Returns(1)
		m.ExpectQuery(wtConfigCountAllScopes).Returns([]countRow{{Count: 0}})
		m.ExpectExec(wtTypeDelete).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).To(BeNil())
	})

	It("keeps the type when other configs remain", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).Returns(1)
		m.ExpectExec(wtConfigDeleteTenant).Returns(1)
		m.ExpectQuery(wtConfigCountAllScopes).Returns([]countRow{{Count: 2}})
		m.ExpectTransaction().Commit()

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when deleting the kartas fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when deleting the config fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).Returns(1)
		m.ExpectExec(wtConfigDeleteTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when counting the remaining configs fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).Returns(1)
		m.ExpectExec(wtConfigDeleteTenant).Returns(1)
		m.ExpectQuery(wtConfigCountAllScopes).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when deleting the type fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).Returns(1)
		m.ExpectExec(wtConfigDeleteTenant).Returns(1)
		m.ExpectQuery(wtConfigCountAllScopes).Returns([]countRow{{Count: 0}})
		m.ExpectExec(wtTypeDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteTenant).Returns(1)
		m.ExpectExec(wtConfigDeleteTenant).Returns(1)
		m.ExpectQuery(wtConfigCountAllScopes).Returns([]countRow{{Count: 1}})

		apiErr := DbDeleteWorkloadTypeWithConfig(testCtx(1), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("dbDeleteExternalWorkloadTypeConfig", func() {
	It("deletes every config of the type for a management tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigDeleteSystem).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := dbDeleteExternalWorkloadTypeConfig(wtMgmtCtx(1), dao_utils.DB(), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the system scoped delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigDeleteSystem).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := dbDeleteExternalWorkloadTypeConfig(wtMgmtCtx(1), dao_utils.DB(), 1, uuid.MustParse(wtIdStr))

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbDoesExternalWorkloadTypeConfigExistForIdForSystemScope", func() {
	It("reports that the config exists", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigCountSystemScope).Returns([]countRow{{Count: 1}})

		exists, apiErr := DbDoesExternalWorkloadTypeConfigExistForIdForSystemScope(testCtx(1), wtIdStr)

		Expect(apiErr).To(BeNil())
		Expect(exists).To(BeTrue())
	})

	It("reports no existence and an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtConfigCountSystemScope).ReturnsError(errors.New("boom"))

		exists, apiErr := DbDoesExternalWorkloadTypeConfigExistForIdForSystemScope(testCtx(1), wtIdStr)

		Expect(apiErr).ToNot(BeNil())
		Expect(exists).To(BeFalse())
	})
})

func wtStatus(version string) models.WorkloadTypeStatus {
	return models.WorkloadTypeStatus{
		ID:             uuid.MustParse(wtConfigIdStr),
		WorkloadTypeId: uuid.MustParse(wtIdStr),
		ClusterId:      uuid.MustParse(wtClusterStr),
		TenantId:       "1",
		Phase:          "Ready",
		Version:        version,
	}
}

var _ = Describe("DbSyncWorkloadTypeStatuses", func() {
	It("upserts the given statuses and deletes the removed ones", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusUpsert).Returns(1)
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusDeleteTargeted).Returns(1)
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr),
			[]models.WorkloadTypeStatus{wtStatus("v1")}, []models.WorkloadTypeStatus{wtStatus("v2")}, false)

		Expect(apiErr).To(BeNil())
	})

	It("rolls back to the savepoint and succeeds when an upsert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusUpsert).ReturnsError(errors.New("boom"))
		m.ExpectExec(wtRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr),
			[]models.WorkloadTypeStatus{wtStatus("v1")}, nil, false)

		Expect(apiErr).To(BeNil())
	})

	It("rolls back to the savepoint and succeeds when a targeted delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusDeleteTargeted).ReturnsError(errors.New("boom"))
		m.ExpectExec(wtRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr),
			nil, []models.WorkloadTypeStatus{wtStatus("v2")}, false)

		Expect(apiErr).To(BeNil())
	})

	It("excludes the upserted types from the full sync delete", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusUpsert).Returns(1)
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusDeleteFullSyncExcluded).Returns(1)
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr),
			[]models.WorkloadTypeStatus{wtStatus("v1")}, nil, true)

		Expect(apiErr).To(BeNil())
	})

	It("deletes every status of the cluster on a full sync with nothing to keep", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusDeleteFullSync).Returns(1)
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr), nil, nil, true)

		Expect(apiErr).To(BeNil())
	})

	It("rolls back to the savepoint and succeeds when the full sync delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)
		m.ExpectExec(wtStatusDeleteFullSync).ReturnsError(errors.New("boom"))
		m.ExpectExec(wtRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr), nil, nil, true)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr), nil, nil, false)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtSavepoint).Returns(0)

		apiErr := DbSyncWorkloadTypeStatuses(testCtx(1), "1", uuid.MustParse(wtClusterStr), nil, nil, false)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DeleteExternalWorkloadTypeConfigsOfTenantId", func() {
	It("deletes the configs of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigDeleteByTenant).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteExternalWorkloadTypeConfigsOfTenantId(testCtx(1), "1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtConfigDeleteByTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteExternalWorkloadTypeConfigsOfTenantId(testCtx(1), "1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteExternalWorkloadTypeKartasOfTenantId", func() {
	It("deletes the kartas of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteByTenant).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteExternalWorkloadTypeKartasOfTenantId(testCtx(1), "1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtKartaDeleteByTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteExternalWorkloadTypeKartasOfTenantId(testCtx(1), "1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DeleteWorkloadTypeStatusesOfTenantId", func() {
	It("deletes the statuses of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtStatusDeleteByTenant).Returns(1)
		m.ExpectTransaction().Commit()

		err := DeleteWorkloadTypeStatusesOfTenantId(testCtx(1), "1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(wtStatusDeleteByTenant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DeleteWorkloadTypeStatusesOfTenantId(testCtx(1), "1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DerivePreemptibilityFromPriority", func() {
	DescribeTable("maps a priority value to a preemptibility",
		func(value int, expected string) {
			Expect(DerivePreemptibilityFromPriority(&models.Priority{Value: value})).To(Equal(expected))
		},
		Entry("high value is non preemptible", NonPreemptibleFromPriorityValue, "non-preemptible"),
		Entry("low value is preemptible", 1, "preemptible"),
	)
})

var _ = Describe("DerivePreemptibilityFromPriorityID", func() {
	It("returns nothing when no priority id is given", func() {
		setupDaoMock(GinkgoT())

		preemptibility, apiErr := DerivePreemptibilityFromPriorityID(testCtx(1), 1, nil)

		Expect(apiErr).To(BeNil())
		Expect(preemptibility).To(BeNil())
	})

	It("returns the derived preemptibility", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtPrioritySelect).Returns([]wtPriorityRow{{ID: wtPriorityStr, Value: 5}})

		preemptibility, apiErr := DerivePreemptibilityFromPriorityID(testCtx(1), 1, ptr.To(uuid.MustParse(wtPriorityStr)))

		Expect(apiErr).To(BeNil())
		Expect(*preemptibility).To(Equal("preemptible"))
	})

	It("returns an error when the priority lookup fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wtPrioritySelect).ReturnsError(errors.New("boom"))

		preemptibility, apiErr := DerivePreemptibilityFromPriorityID(testCtx(1), 1, ptr.To(uuid.MustParse(wtPriorityStr)))

		Expect(apiErr).ToNot(BeNil())
		Expect(preemptibility).To(BeNil())
	})
})
