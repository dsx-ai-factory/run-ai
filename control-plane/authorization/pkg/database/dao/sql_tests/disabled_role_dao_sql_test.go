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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DisabledRole DAO SQL", func() {
	Describe("DbGetRoleDisabledPerTenant", func() {
		It("returns a map of disabled roles", func() {
			mocker, ormDB := setupSQLMockWithDB()

			mocker.ExpectQuery(
				`SELECT * FROM disabled_roles WHERE scope_type = ? AND scope_id = ?`,
				string(models.ROLE_SCOPE_TENANT), "1",
			).Returns([]models.DisabledRole{
				{RoleId: 10, ScopeType: models.ROLE_SCOPE_TENANT, ScopeId: "1"},
				{RoleId: 20, ScopeType: models.ROLE_SCOPE_TENANT, ScopeId: "1"},
			})

			result, apiErr := dao.DbGetRoleDisabledPerTenant(context.Background(), ormDB, "1")
			Expect(apiErr).To(BeNil())
			Expect(result).To(HaveLen(2))
			Expect(result).To(HaveKey(int64(10)))
			Expect(result).To(HaveKey(int64(20)))
		})

		It("returns empty", func() {
			mocker, ormDB := setupSQLMockWithDB()

			mocker.ExpectQuery(
				`SELECT * FROM disabled_roles WHERE scope_type = ? AND scope_id = ?`,
				string(models.ROLE_SCOPE_TENANT), "1",
			).Returns([]models.DisabledRole{})

			result, apiErr := dao.DbGetRoleDisabledPerTenant(context.Background(), ormDB, "1")
			Expect(apiErr).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("returns error on DB failure", func() {
			mocker, ormDB := setupSQLMockWithDB()

			mocker.ExpectQuery(
				`SELECT * FROM disabled_roles WHERE scope_type = ? AND scope_id = ?`,
				string(models.ROLE_SCOPE_TENANT), "1",
			).ReturnsError(errors.New("db error"))

			result, apiErr := dao.DbGetRoleDisabledPerTenant(context.Background(), ormDB, "1")
			Expect(apiErr).NotTo(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Describe("DbGetRoleDisabledById", func() {
		It("returns disabled role when found", func() {
			mocker, ormDB := setupSQLMockWithDB()

			mocker.ExpectQuery(
				`SELECT * FROM disabled_roles WHERE (scope_type = ? AND scope_id = ?) AND role_id = ? ORDER BY disabled_roles.role_id LIMIT ?`,
				string(models.ROLE_SCOPE_TENANT), "1", int64(10), int64(1),
			).Returns([]models.DisabledRole{
				{RoleId: 10, ScopeType: models.ROLE_SCOPE_TENANT, ScopeId: "1"},
			})

			result, apiErr := dao.DbGetRoleDisabledById(context.Background(), ormDB, "1", uint(10))
			Expect(apiErr).To(BeNil())
			Expect(result.RoleId).To(Equal(int64(10)))
		})

		It("returns zero value when not found", func() {
			mocker, ormDB := setupSQLMockWithDB()

			mocker.ExpectQuery(
				`SELECT * FROM disabled_roles WHERE (scope_type = ? AND scope_id = ?) AND role_id = ? ORDER BY disabled_roles.role_id LIMIT ?`,
				string(models.ROLE_SCOPE_TENANT), "1", int64(10), int64(1),
			).Returns([]models.DisabledRole{})

			result, apiErr := dao.DbGetRoleDisabledById(context.Background(), ormDB, "1", uint(10))
			Expect(apiErr).To(BeNil())
			Expect(result.RoleId).To(BeZero())
		})
	})
})
