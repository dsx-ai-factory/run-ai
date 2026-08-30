package sql_tests

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

const (
	sqlGetPermissions = `SELECT resource_type, bool_or(permissions.create) AS create, bool_or(permissions.read) AS read, bool_or(permissions.update) AS update, bool_or(permissions.delete) AS delete FROM permissions INNER JOIN access_rules ON permissions.role_id = access_rules.role_id WHERE access_rules.tenant_id = ? AND access_rules.deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) GROUP BY resource_type`

	sqlGetPermissionSets = `SELECT resource_type, bool_or(role_permissions.create) AS create, bool_or(role_permissions.read) AS read, bool_or(role_permissions.update) AS update, bool_or(role_permissions.delete) AS delete, bool_or(role_permissions.sync) AS sync FROM role_permissions INNER JOIN role_to_permission_set ON role_to_permission_set.permission_set_id = role_permissions.permission_set_id INNER JOIN access_rules ON role_to_permission_set.role_id = access_rules.role_id WHERE access_rules.tenant_id = ? AND access_rules.deleted_at IS NULL AND ((subject_id = ? AND subject_type != ?) OR subject_type = ?) GROUP BY resource_type`
)

var _ = Describe("Permissions DAO SQL", func() {
	Describe("GetPermissions", func() {
		It("returns permissions", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlGetPermissions,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]models.Permission{
				{ResourceType: api.RESOURCETYPE_PROJECT, Read: true, Create: true},
				{ResourceType: api.RESOURCETYPE_CLUSTER, Read: true},
			})

			perms, apiErr := dao.GetPermissions(context.Background(), sqlTestTenantID, "user@example.com", nil)
			Expect(apiErr).To(BeNil())
			Expect(perms).To(HaveLen(2))
			Expect(*perms[0]).To(Equal(models.Permission{
				ResourceType: api.RESOURCETYPE_PROJECT,
				Read:         true,
				Create:       true,
			}))
		})

		It("returns empty", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlGetPermissions,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]models.Permission{})

			perms, apiErr := dao.GetPermissions(context.Background(), sqlTestTenantID, "user@example.com", nil)
			Expect(apiErr).To(BeNil())
			Expect(perms).To(BeEmpty())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlGetPermissions,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).ReturnsError(errors.New("db error"))

			perms, apiErr := dao.GetPermissions(context.Background(), sqlTestTenantID, "user@example.com", nil)
			Expect(apiErr).NotTo(BeNil())
			Expect(perms).To(BeNil())
		})
	})

	Describe("GetPermissionSets", func() {
		It("returns permissions", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlGetPermissionSets,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]models.RolePermission{
				{ResourceType: api.RESOURCETYPE_PROJECT, Read: true, Create: true, Sync: true},
			})

			perms, apiErr := dao.GetPermissionSets(context.Background(), sqlTestTenantID, "user@example.com", nil)
			Expect(apiErr).To(BeNil())
			Expect(perms).To(HaveLen(1))
			Expect(perms[0].ResourceType).To(Equal(api.RESOURCETYPE_PROJECT))
			Expect(perms[0].Read).To(BeTrue())
			Expect(perms[0].Create).To(BeTrue())
		})

		It("returns empty", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlGetPermissionSets,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).Returns([]models.RolePermission{})

			perms, apiErr := dao.GetPermissionSets(context.Background(), sqlTestTenantID, "user@example.com", nil)
			Expect(apiErr).To(BeNil())
			Expect(perms).To(BeEmpty())
		})

		It("returns error on DB failure", func() {
			mocker := setupSQLMock()

			mocker.ExpectQuery(
				sqlGetPermissionSets,
				int64(sqlTestTenantID), "user@example.com", string(api.SUBJECTTYPE_GROUP), string(models.AllAuthenticatedUsers),
			).ReturnsError(errors.New("db error"))

			perms, apiErr := dao.GetPermissionSets(context.Background(), sqlTestTenantID, "user@example.com", nil)
			Expect(apiErr).NotTo(BeNil())
			Expect(perms).To(BeNil())
		})
	})
})
