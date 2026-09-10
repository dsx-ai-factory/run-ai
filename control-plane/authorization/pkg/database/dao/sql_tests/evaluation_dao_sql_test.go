/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sql_tests

import (
	"context"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// evaluationRow holds the four columns selected by dao.EvaluatePermission.
// models.AccessRule cannot be used directly because PhaseByCluster.Scan panics
// when given a raw map from convertToMockRows.
type evaluationRow struct {
	ScopeType   api.ScopeType `gorm:"column:scope_type"`
	ScopeId     string        `gorm:"column:scope_id"`
	ClusterUuid string        `gorm:"column:cluster_uuid"`
	TenantId    uint          `gorm:"column:tenant_id"`
}

// sqlEvalPSGlobal — dao.EvaluatePermission (global request)
const sqlEvalPSGlobal = `SELECT scope_type, scope_id, cluster_uuid, tenant_id FROM access_rules JOIN role_to_permission_set ON role_to_permission_set.role_id = access_rules.role_id INNER JOIN role_permissions ON role_permissions.permission_set_id = role_to_permission_set.permission_set_id WHERE tenant_id = ? AND deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) AND access_rules.deleted_at IS NULL`

var _ = Describe("Evaluation DAO SQL", func() {
	Describe("EvaluatePermission", func() {
		It("approves global access without scopes", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalPSGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "1", TenantId: sqlTestTenantID},
			})

			scopes, apiErr := dao.EvaluatePermission(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).To(BeNil())
			Expect(scopes).To(BeNil()) // global + forceCalculatePermittedScopes=false → nil, nil
		})

		It("builds permitted scopes for global access", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalPSGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "42", TenantId: sqlTestTenantID},
			})

			scopes, apiErr := dao.EvaluatePermission(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, true)
			Expect(apiErr).To(BeNil())
			Expect(scopes).NotTo(BeNil())
			Expect(scopes.Tenants).To(ContainElement("42"))
		})

		It("returns forbidden when no matching access rules exist", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalPSGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{})

			_, apiErr := dao.EvaluatePermission(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(403)))
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalPSGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).ReturnsError(errors.New("db error"))

			_, apiErr := dao.EvaluatePermission(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(500)))
		})
	})
})
