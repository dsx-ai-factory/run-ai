/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sql_tests

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// permissionSetRow excludes Permissions ([]RolePermission slice) because
// collectFields includes slice fields and convertToMockRows would store a raw
// slice that the SQL driver can't handle.
type permissionSetRow struct {
	ID          string `gorm:"column:id"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Internal    bool   `gorm:"column:internal"`
	CreatedBy   string `gorm:"column:created_by"`
	UpdatedBy   string `gorm:"column:updated_by"`
}

// dao.DbGetPermissionSetsByRoleId uses .Table("permission_sets") so GORM generates
// an explicit column list instead of SELECT *.
const sqlGetPermissionSetsByRoleId = `SELECT permission_sets.id,permission_sets.created_at,permission_sets.updated_at,permission_sets.created_by,permission_sets.updated_by,permission_sets.name,permission_sets.description,permission_sets.internal FROM permission_sets JOIN role_to_permission_set rps ON rps.permission_set_id = permission_sets.id WHERE rps.role_id = ?`

var _ = Describe("PermissionSets DAO SQL", func() {
	Describe("DbGetPermissionSets", func() {
		It("returns all public permission sets", func() {
			mocker := setupSQLMock()

			psID := uuid.New().String()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE internal = false ORDER BY id DESC`,
			).Returns([]permissionSetRow{
				{ID: psID, Name: "Viewer", Internal: false},
			})
			mocker.ExpectQuery(
				`SELECT * FROM role_permissions WHERE role_permissions.permission_set_id = ?`,
				dbtesting.AnyArg(),
			).Returns([]models.RolePermission{})

			result, apiErr := dao.DbGetPermissionSets(context.Background(), false)
			Expect(apiErr).To(BeNil())
			Expect(*result).To(HaveLen(1))
			Expect((*result)[0].Name).To(Equal("Viewer"))
		})

		It("returns empty", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE internal = false ORDER BY id DESC`,
			).Returns([]permissionSetRow{})

			result, apiErr := dao.DbGetPermissionSets(context.Background(), false)
			Expect(apiErr).To(BeNil())
			Expect(*result).To(BeEmpty())
		})

		It("includes internal permission sets when requested", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets ORDER BY id DESC`,
			).Returns([]permissionSetRow{})

			result, apiErr := dao.DbGetPermissionSets(context.Background(), true)
			Expect(apiErr).To(BeNil())
			Expect(*result).To(BeEmpty())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE internal = false ORDER BY id DESC`,
			).ReturnsError(errors.New("db error"))

			result, apiErr := dao.DbGetPermissionSets(context.Background(), false)
			Expect(apiErr).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Describe("DbGetPermissionSetsByIds", func() {
		It("returns results", func() {
			mocker := setupSQLMock()

			psID := uuid.New().String()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE id in (?)`,
				psID,
			).Returns([]permissionSetRow{
				{ID: psID, Name: "Editor", Internal: false},
			})
			mocker.ExpectQuery(
				`SELECT * FROM role_permissions WHERE role_permissions.permission_set_id = ?`,
				dbtesting.AnyArg(),
			).Returns([]models.RolePermission{})

			result, apiErr := dao.DbGetPermissionSetsByIds(context.Background(), []string{psID})
			Expect(apiErr).To(BeNil())
			Expect(*result).To(HaveLen(1))
			Expect((*result)[0].Name).To(Equal("Editor"))
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			psID := uuid.New().String()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE id in (?)`,
				psID,
			).ReturnsError(errors.New("db error"))

			result, apiErr := dao.DbGetPermissionSetsByIds(context.Background(), []string{psID})
			Expect(apiErr).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Describe("DbGetPermissionSetsByRoleId", func() {
		It("returns results", func() {
			mocker := setupSQLMock()

			psID := uuid.New().String()

			mocker.ExpectQuery(sqlGetPermissionSetsByRoleId, int64(sqlTestRoleID)).
				Returns([]permissionSetRow{{ID: psID, Name: "Admin", Internal: false}})
			mocker.ExpectQuery(
				`SELECT * FROM role_permissions WHERE role_permissions.permission_set_id = ?`,
				dbtesting.AnyArg(),
			).Returns([]models.RolePermission{})

			result, apiErr := dao.DbGetPermissionSetsByRoleId(context.Background(), sqlTestRoleID)
			Expect(apiErr).To(BeNil())
			Expect(*result).To(HaveLen(1))
			Expect((*result)[0].Name).To(Equal("Admin"))
		})

		It("returns empty", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(sqlGetPermissionSetsByRoleId, int64(sqlTestRoleID)).
				Returns([]permissionSetRow{})

			result, apiErr := dao.DbGetPermissionSetsByRoleId(context.Background(), sqlTestRoleID)
			Expect(apiErr).To(BeNil())
			Expect(*result).To(BeEmpty())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(sqlGetPermissionSetsByRoleId, int64(sqlTestRoleID)).
				ReturnsError(errors.New("db error"))

			result, apiErr := dao.DbGetPermissionSetsByRoleId(context.Background(), sqlTestRoleID)
			Expect(apiErr).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Describe("DbGetPermissionSetById", func() {
		It("returns result", func() {
			mocker := setupSQLMock()

			psID := uuid.New()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE internal = false AND id = ? ORDER BY permission_sets.id LIMIT ?`,
				dbtesting.AnyArg(), int64(1),
			).Returns([]permissionSetRow{
				{ID: psID.String(), Name: "Viewer", Internal: false},
			})
			mocker.ExpectQuery(
				`SELECT * FROM role_permissions WHERE role_permissions.permission_set_id = ?`,
				dbtesting.AnyArg(),
			).Returns([]models.RolePermission{})

			result, apiErr := dao.DbGetPermissionSetById(context.Background(), psID)
			Expect(apiErr).To(BeNil())
			Expect(result.Name).To(Equal("Viewer"))
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			psID := uuid.New()

			mocker.ExpectQuery(
				`SELECT * FROM permission_sets WHERE internal = false AND id = ? ORDER BY permission_sets.id LIMIT ?`,
				dbtesting.AnyArg(), int64(1),
			).ReturnsError(errors.New("not found"))

			result, apiErr := dao.DbGetPermissionSetById(context.Background(), psID)
			Expect(apiErr).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})
})
