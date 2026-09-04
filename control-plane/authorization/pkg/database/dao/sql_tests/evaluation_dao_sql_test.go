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

// evaluationRow holds the four columns selected by dao.EvaluatePermission /
// dao.EvaluatePermissionLegacy. models.AccessRule cannot be used directly because
// PhaseByCluster.Scan panics when given a raw map from convertToMockRows.
type evaluationRow struct {
	ScopeType   api.ScopeType `gorm:"column:scope_type"`
	ScopeId     string        `gorm:"column:scope_id"`
	ClusterUuid string        `gorm:"column:cluster_uuid"`
	TenantId    uint          `gorm:"column:tenant_id"`
}

const (
	// sqlEvalPSGlobal — dao.EvaluatePermission (UsePermissionSets=true, global request)
	sqlEvalPSGlobal = `SELECT scope_type, scope_id, cluster_uuid, tenant_id FROM access_rules JOIN role_to_permission_set ON role_to_permission_set.role_id = access_rules.role_id INNER JOIN role_permissions ON role_permissions.permission_set_id = role_to_permission_set.permission_set_id WHERE tenant_id = ? AND deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) AND access_rules.deleted_at IS NULL`

	// sqlEvalLegacyGlobal — dao.EvaluatePermissionLegacy (always legacy path, global request)
	sqlEvalLegacyGlobal = `SELECT scope_type, scope_id, cluster_uuid, tenant_id FROM access_rules INNER JOIN permissions ON permissions.role_id = access_rules.role_id WHERE tenant_id = ? AND deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) AND access_rules.deleted_at IS NULL`

	sqlEvalLegacyWithType = `SELECT scope_type, scope_id, cluster_uuid, tenant_id FROM access_rules INNER JOIN permissions ON permissions.role_id = access_rules.role_id WHERE tenant_id = ? AND deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) AND permissions.resource_type = ? AND access_rules.deleted_at IS NULL`

	sqlEvalLegacyWithRead = `SELECT scope_type, scope_id, cluster_uuid, tenant_id FROM access_rules INNER JOIN permissions ON permissions.role_id = access_rules.role_id WHERE tenant_id = ? AND deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) AND permissions.resource_type = ? AND permissions.read = true AND access_rules.deleted_at IS NULL`

	sqlEvalLegacyWithGroups = `SELECT scope_type, scope_id, cluster_uuid, tenant_id FROM access_rules INNER JOIN permissions ON permissions.role_id = access_rules.role_id WHERE tenant_id = ? AND deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ? OR (subject_id IN (?) AND subject_type = ?)) AND access_rules.deleted_at IS NULL`
)

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

	Describe("EvaluatePermissionLegacy", func() {
		It("approves global access without scopes", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "1", TenantId: sqlTestTenantID},
			})

			scopes, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).To(BeNil())
			Expect(scopes).To(BeNil())
		})

		It("builds scopes and deduplicates tenants", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "10", TenantId: sqlTestTenantID},
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "20", TenantId: sqlTestTenantID},
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "10", TenantId: sqlTestTenantID}, // duplicate
			})

			scopes, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, true)
			Expect(apiErr).To(BeNil())
			Expect(scopes).NotTo(BeNil())
			Expect(scopes.Tenants).To(HaveLen(2))
			Expect(scopes.Tenants).To(ContainElement("10"))
			Expect(scopes.Tenants).To(ContainElement("20"))
		})

		It("sets system true for system scope", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_SYSTEM, ScopeId: "0", TenantId: sqlTestTenantID},
			})

			scopes, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, true)
			Expect(apiErr).To(BeNil())
			Expect(scopes).NotTo(BeNil())
			Expect(scopes.System).To(BeTrue())
		})

		It("returns forbidden when no matching rules exist", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]evaluationRow{})

			_, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(403)))
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyGlobal,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).ReturnsError(errors.New("db error"))

			_, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).NotTo(BeNil())
			Expect(apiErr.Code).To(Equal(int32(500)))
		})

		It("filters permissions by resource type", func() {
			mocker := setupSQLMock()

			// ResourceId=nil → !isGlobal path but no GetAncestors cache call.
			// With isGlobal=false, PermittedScopes are always built from access rules.
			mocker.ExpectQuery(
				sqlEvalLegacyWithType,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
				string(api.RESOURCETYPE_PROJECT),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "tenant-1", TenantId: sqlTestTenantID},
			})

			rt := api.RESOURCETYPE_PROJECT
			scopes, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{ResourceType: &rt}, false)
			Expect(apiErr).To(BeNil())
			Expect(scopes).NotTo(BeNil())
			Expect(scopes.Tenants).To(ContainElement("tenant-1"))
		})

		It("filters permissions by resource type and read action", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyWithRead,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
				string(api.RESOURCETYPE_PROJECT),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "tenant-2", TenantId: sqlTestTenantID},
			})

			rt := api.RESOURCETYPE_PROJECT
			action := api.ACTION_READ
			scopes, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com", nil,
				api.BackendEvaluationRequest{ResourceType: &rt, RequestAction: &action}, false)
			Expect(apiErr).To(BeNil())
			Expect(scopes).NotTo(BeNil())
			Expect(scopes.Tenants).To(ContainElement("tenant-2"))
		})

		It("expands subject clause when groups are provided", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlEvalLegacyWithGroups,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
				"group-a", string(api.SUBJECTTYPE_GROUP),
			).Returns([]evaluationRow{
				{ScopeType: api.SCOPETYPE_TENANT, ScopeId: "1", TenantId: sqlTestTenantID},
			})

			scopes, apiErr := dao.EvaluatePermissionLegacy(context.Background(), sqlTestTenantID, "user@example.com",
				[]string{"group-a"},
				api.BackendEvaluationRequest{}, false)
			Expect(apiErr).To(BeNil())
			Expect(scopes).To(BeNil())
		})
	})
})
