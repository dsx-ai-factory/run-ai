/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sql_tests

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

// accessRuleRow excludes PhaseByCluster (map with a custom Scan expecting []byte)
// and non-anonymous struct fields (Role, DeletedAt) which collectFields already skips.
// Without it, convertToMockRows stores a raw map that causes PhaseByCluster.Scan to panic.
type accessRuleRow struct {
	ID           uint            `gorm:"column:id"`
	SubjectId    string          `gorm:"column:subject_id"`
	SubjectType  api.SubjectType `gorm:"column:subject_type"`
	RoleID       uint            `gorm:"column:role_id"`
	ScopeId      string          `gorm:"column:scope_id"`
	ScopeType    api.ScopeType   `gorm:"column:scope_type"`
	ScopeName    string          `gorm:"column:scope_name"`
	TenantId     uint            `gorm:"column:tenant_id"`
	ClusterUuid  string          `gorm:"column:cluster_uuid"`
	Phase        string          `gorm:"column:phase"`
	PhaseMessage string          `gorm:"column:phase_message"`
	CreatedBy    string          `gorm:"column:created_by"`
}

var _ = Describe("AccessRule DAO SQL", func() {
	Describe("DbGetDeprecatedAccessRulesForMigration", func() {
		It("returns a single rule", func() {
			mocker := setupSQLMock()

			row := accessRuleRow{
				ID:          1,
				SubjectId:   "user@example.com",
				SubjectType: api.SUBJECTTYPE_USER,
				RoleID:      sqlTestRoleID,
				ScopeId:     "scope-1",
				ScopeType:   api.SCOPETYPE_TENANT,
				TenantId:    sqlTestTenantID,
			}

			mocker.ExpectQuery(
				`SELECT * FROM access_rules WHERE tenant_id = ? AND role_id = ? AND scope_type != ? AND access_rules.deleted_at IS NULL LIMIT ?`,
				int64(sqlTestTenantID), int64(sqlTestRoleID), string(api.SCOPETYPE_SYSTEM), int64(10),
			).Returns([]accessRuleRow{row})

			rules, apiErr := dao.DbGetDeprecatedAccessRulesForMigration(context.Background(), sqlTestTenantID, sqlTestRoleID, 10)
			Expect(apiErr).To(BeNil())
			Expect(rules).To(HaveLen(1))
			Expect(rules[0].SubjectId).To(Equal(row.SubjectId))
			Expect(rules[0].RoleID).To(Equal(row.RoleID))
			Expect(rules[0].ScopeType).To(Equal(row.ScopeType))
		})

		It("returns empty", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT * FROM access_rules WHERE tenant_id = ? AND role_id = ? AND scope_type != ? AND access_rules.deleted_at IS NULL LIMIT ?`,
				int64(sqlTestTenantID), int64(sqlTestRoleID), string(api.SCOPETYPE_SYSTEM), int64(5),
			).Returns([]accessRuleRow{})

			rules, apiErr := dao.DbGetDeprecatedAccessRulesForMigration(context.Background(), sqlTestTenantID, sqlTestRoleID, 5)
			Expect(apiErr).To(BeNil())
			Expect(rules).To(BeEmpty())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT * FROM access_rules WHERE tenant_id = ? AND role_id = ? AND scope_type != ? AND access_rules.deleted_at IS NULL LIMIT ?`,
				int64(sqlTestTenantID), int64(sqlTestRoleID), string(api.SCOPETYPE_SYSTEM), int64(10),
			).ReturnsError(errors.New("db error"))

			rules, apiErr := dao.DbGetDeprecatedAccessRulesForMigration(context.Background(), sqlTestTenantID, sqlTestRoleID, 10)
			Expect(apiErr).NotTo(BeNil())
			Expect(rules).To(BeNil())
		})
	})

	Describe("DbBatchDeleteAccessRules", func() {
		It("deletes a single rule", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND id IN (?) AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), int64(1),
			).Returns(1)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbBatchDeleteAccessRules(context.Background(), sqlTestTenantID, []uint{1})
			Expect(apiErr).To(BeNil())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND id IN (?) AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), int64(1),
			).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			apiErr := dao.DbBatchDeleteAccessRules(context.Background(), sqlTestTenantID, []uint{1})
			Expect(apiErr).NotTo(BeNil())
		})

		It("returns error when rows affected mismatch", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND id IN (?,?) AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), int64(1), int64(2),
			).Returns(1) // only 1 row affected, but 2 were expected
			mocker.ExpectTransaction().Rollback()

			apiErr := dao.DbBatchDeleteAccessRules(context.Background(), sqlTestTenantID, []uint{1, 2})
			Expect(apiErr).NotTo(BeNil())
		})
	})

	Describe("DbDeleteSubjectAccessRules", func() {
		It("deletes all subject rules", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND subject_type = ? AND subject_id = ? AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), string(api.SUBJECTTYPE_USER), "user@example.com",
			).Returns(2)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbDeleteSubjectAccessRules(context.Background(), sqlTestTenantID, string(api.SUBJECTTYPE_USER), "user@example.com")
			Expect(apiErr).To(BeNil())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND subject_type = ? AND subject_id = ? AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), string(api.SUBJECTTYPE_USER), "user@example.com",
			).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			apiErr := dao.DbDeleteSubjectAccessRules(context.Background(), sqlTestTenantID, string(api.SUBJECTTYPE_USER), "user@example.com")
			Expect(apiErr).NotTo(BeNil())
		})
	})

	Describe("HardDeleteTenantAccessRules", func() {
		It("deletes all rows", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`DELETE FROM access_rules WHERE tenant_id = ?`,
				"1",
			).Returns(5)
			mocker.ExpectTransaction().Commit()

			err := dao.HardDeleteTenantAccessRules(context.Background(), "1")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`DELETE FROM access_rules WHERE tenant_id = ?`,
				"1",
			).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			err := dao.HardDeleteTenantAccessRules(context.Background(), "1")
			Expect(err).To(MatchError("delete failed"))
		})
	})

	Describe("DbCountAccessRules", func() {
		It("returns count", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules left join roles on roles.id = access_rules.role_id WHERE roles.internal_only = false AND access_rules.tenant_id = ? AND deleted_at IS NULL`,
				int64(sqlTestTenantID),
			).Returns([]sqlCountResult{{Count: 7}})

			count, apiErr := dao.DbCountAccessRules(context.Background(), sqlTestTenantID, false, nil, "")
			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(7)))
		})

		It("returns zero", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules left join roles on roles.id = access_rules.role_id WHERE roles.internal_only = false AND access_rules.tenant_id = ? AND deleted_at IS NULL`,
				int64(sqlTestTenantID),
			).Returns([]sqlCountResult{{Count: 0}})

			count, apiErr := dao.DbCountAccessRules(context.Background(), sqlTestTenantID, false, nil, "")
			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(0)))
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules left join roles on roles.id = access_rules.role_id WHERE roles.internal_only = false AND access_rules.tenant_id = ? AND deleted_at IS NULL`,
				int64(sqlTestTenantID),
			).ReturnsError(errors.New("count failed"))

			count, apiErr := dao.DbCountAccessRules(context.Background(), sqlTestTenantID, false, nil, "")
			Expect(apiErr).NotTo(BeNil())
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("DbSetRoleEnabledPerTenant", func() {
		It("deletes disabled role", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`DELETE FROM disabled_roles WHERE role_id = ? AND scope_type = ? AND scope_id = ?`,
				int64(sqlTestRoleID), string(models.ROLE_SCOPE_TENANT), "1",
			).Returns(1)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbSetRoleEnabledPerTenant(context.Background(), "1", sqlTestRoleID)
			Expect(apiErr).To(BeNil())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`DELETE FROM disabled_roles WHERE role_id = ? AND scope_type = ? AND scope_id = ?`,
				int64(sqlTestRoleID), string(models.ROLE_SCOPE_TENANT), "1",
			).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			apiErr := dao.DbSetRoleEnabledPerTenant(context.Background(), "1", sqlTestRoleID)
			Expect(apiErr).NotTo(BeNil())
		})
	})

	Describe("DbUpdateScopeAccessRules", func() {
		It("updates scope name", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET scope_name=?,updated_at=? WHERE tenant_id = ? AND scope_type = ? AND scope_id = ? AND access_rules.deleted_at IS NULL`,
				"new-cluster-name", dbtesting.AnyTime(), int64(sqlTestTenantID), string(api.SCOPETYPE_CLUSTER), "cluster-uuid-1",
			).Returns(3)
			mocker.ExpectTransaction().Commit()

			req := api.UpdateScopeAccessRulesRequest{ScopeName: "new-cluster-name"}
			apiErr := dao.DbUpdateScopeAccessRules(context.Background(), sqlTestTenantID, string(api.SCOPETYPE_CLUSTER), "cluster-uuid-1", req)
			Expect(apiErr).To(BeNil())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET scope_name=?,updated_at=? WHERE tenant_id = ? AND scope_type = ? AND scope_id = ? AND access_rules.deleted_at IS NULL`,
				"new-cluster-name", dbtesting.AnyTime(), int64(sqlTestTenantID), string(api.SCOPETYPE_CLUSTER), "cluster-uuid-1",
			).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			req := api.UpdateScopeAccessRulesRequest{ScopeName: "new-cluster-name"}
			apiErr := dao.DbUpdateScopeAccessRules(context.Background(), sqlTestTenantID, string(api.SCOPETYPE_CLUSTER), "cluster-uuid-1", req)
			Expect(apiErr).NotTo(BeNil())
		})
	})

	Describe("DbDeleteScopeAccessRules", func() {
		It("deletes cluster scope rules", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND cluster_uuid = ? AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), "cluster-uuid-1",
			).Returns(4)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbDeleteScopeAccessRules(context.Background(), false, sqlTestTenantID, string(api.SCOPETYPE_CLUSTER), "cluster-uuid-1", "", nil)
			Expect(apiErr).To(BeNil())
		})

		It("deletes project scope rules", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND access_rules.scope_type = ? AND access_rules.scope_id = ? AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), string(api.SCOPETYPE_PROJECT), "proj-1",
			).Returns(2)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbDeleteScopeAccessRules(context.Background(), false, sqlTestTenantID, string(api.SCOPETYPE_PROJECT), "proj-1", "", nil)
			Expect(apiErr).To(BeNil())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE access_rules SET deleted_at=? WHERE tenant_id = ? AND cluster_uuid = ? AND access_rules.deleted_at IS NULL`,
				dbtesting.AnyTime(), int64(sqlTestTenantID), "cluster-uuid-1",
			).ReturnsError(errors.New("delete failed"))
			mocker.ExpectTransaction().Rollback()

			apiErr := dao.DbDeleteScopeAccessRules(context.Background(), false, sqlTestTenantID, string(api.SCOPETYPE_CLUSTER), "cluster-uuid-1", "", nil)
			Expect(apiErr).NotTo(BeNil())
		})
	})

	Describe("DbCountMeAccessRules", func() {
		It("returns count", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules LEFT JOIN roles ON roles.id = access_rules.role_id WHERE roles.internal_only = false AND access_rules.tenant_id = ? AND access_rules.deleted_at is null AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?)`,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]sqlCountResult{{Count: 3}})

			count, apiErr := dao.DbCountMeAccessRules(context.Background(), sqlTestTenantID, "user@example.com", nil, "")
			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(3)))
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules LEFT JOIN roles ON roles.id = access_rules.role_id WHERE roles.internal_only = false AND access_rules.tenant_id = ? AND access_rules.deleted_at is null AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?)`,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).ReturnsError(errors.New("count failed"))

			count, apiErr := dao.DbCountMeAccessRules(context.Background(), sqlTestTenantID, "user@example.com", nil, "")
			Expect(apiErr).NotTo(BeNil())
			Expect(count).To(Equal(int64(0)))
		})
	})
})
