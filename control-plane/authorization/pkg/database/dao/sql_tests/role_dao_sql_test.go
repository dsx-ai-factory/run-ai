/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sql_tests

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Role DAO SQL", func() {
	Describe("DbSetRoleEnabled", func() {
		It("enables role", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE roles SET enabled=? WHERE tenant_id = ? AND id = ?`,
				true, "1", int64(5),
			).Returns(1)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbSetRoleEnabled(context.Background(), "1", 5, true)
			Expect(apiErr).To(BeNil())
		})

		It("disables role", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE roles SET enabled=? WHERE tenant_id = ? AND id = ?`,
				false, "1", int64(5),
			).Returns(1)
			mocker.ExpectTransaction().Commit()

			apiErr := dao.DbSetRoleEnabled(context.Background(), "1", 5, false)
			Expect(apiErr).To(BeNil())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`UPDATE roles SET enabled=? WHERE tenant_id = ? AND id = ?`,
				true, "1", int64(5),
			).ReturnsError(errors.New("update failed"))
			mocker.ExpectTransaction().Rollback()

			apiErr := dao.DbSetRoleEnabled(context.Background(), "1", 5, true)
			Expect(apiErr).NotTo(BeNil())
		})
	})

	Describe("IsRoleInUse", func() {
		It("returns true when role is in use", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules WHERE role_id = ? AND deleted_at IS NULL`,
				int64(sqlTestRoleID),
			).Returns([]sqlCountResult{{Count: 3}})

			inUse, apiErr := dao.IsRoleInUse(context.Background(), sqlTestRoleID)
			Expect(apiErr).To(BeNil())
			Expect(inUse).To(BeTrue())
		})

		It("returns false when role is not in use", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules WHERE role_id = ? AND deleted_at IS NULL`,
				int64(sqlTestRoleID),
			).Returns([]sqlCountResult{{Count: 0}})

			inUse, apiErr := dao.IsRoleInUse(context.Background(), sqlTestRoleID)
			Expect(apiErr).To(BeNil())
			Expect(inUse).To(BeFalse())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM access_rules WHERE role_id = ? AND deleted_at IS NULL`,
				int64(sqlTestRoleID),
			).ReturnsError(errors.New("count failed"))

			inUse, apiErr := dao.IsRoleInUse(context.Background(), sqlTestRoleID)
			Expect(apiErr).NotTo(BeNil())
			Expect(inUse).To(BeFalse())
		})
	})

	Describe("IsRoleInScopeType", func() {
		It("returns true when role is in scope", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM roles WHERE id = ? AND scope_type = ?`,
				int64(sqlTestRoleID), "system",
			).Returns([]sqlCountResult{{Count: 1}})

			inScope, apiErr := dao.IsRoleInScopeType(context.Background(), sqlTestRoleID, "system")
			Expect(apiErr).To(BeNil())
			Expect(inScope).To(BeTrue())
		})

		It("returns false when role is not in scope", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM roles WHERE id = ? AND scope_type = ?`,
				int64(sqlTestRoleID), "tenant",
			).Returns([]sqlCountResult{{Count: 0}})

			inScope, apiErr := dao.IsRoleInScopeType(context.Background(), sqlTestRoleID, "tenant")
			Expect(apiErr).To(BeNil())
			Expect(inScope).To(BeFalse())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				`SELECT count(*) FROM roles WHERE id = ? AND scope_type = ?`,
				int64(sqlTestRoleID), "system",
			).ReturnsError(errors.New("count failed"))

			inScope, apiErr := dao.IsRoleInScopeType(context.Background(), sqlTestRoleID, "system")
			Expect(apiErr).NotTo(BeNil())
			Expect(inScope).To(BeFalse())
		})
	})
})