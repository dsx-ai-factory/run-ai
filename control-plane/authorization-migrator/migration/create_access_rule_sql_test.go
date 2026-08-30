package migration

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	static_roles "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-roles"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
)

type roleIDRow struct {
	ID uint `gorm:"column:id"`
}

type accessRuleInsertReturningRow struct {
	ID uint `gorm:"column:id"`
}

var _ = Describe("CreateAccessRule SQL", func() {
	It("looks up the role and inserts an access rule with ON CONFLICT DO NOTHING", func() {
		mocker := dbtesting.SetSQLMock(GinkgoT())

		mocker.ExpectQuery(
			`SELECT id FROM roles WHERE name = ? ORDER BY roles.id LIMIT ?`,
			static_roles.Viewer.Name, int64(1),
		).Returns([]roleIDRow{{ID: 10}})

		mocker.ExpectTransaction().Begin()
		mocker.ExpectQuery(
			`INSERT INTO access_rules (created_at,updated_at,created_by,subject_id,subject_type,role_id,scope_id,scope_type,scope_name,tenant_id,cluster_uuid,deleted_at,phase,phase_message,phase_by_cluster) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING RETURNING id`,
			dbtesting.AnyTime(), dbtesting.AnyTime(), "",
			"user@example.com", string(api.SUBJECTTYPE_USER), int64(10),
			"1", string(api.SCOPETYPE_TENANT), "tenant",
			int64(1), "", nil, "", "", "null",
		).Returns([]accessRuleInsertReturningRow{{ID: 100}})
		mocker.ExpectTransaction().Commit()

		Expect(func() {
			CreateAccessRule(runaidb.DB(), 1, "user@example.com", EntityTypeRegularUser, static_roles.Viewer.Name, "1", api.SCOPETYPE_TENANT, "tenant", "")
		}).NotTo(Panic())
	})

	It("panics when role lookup fails", func() {
		mocker := dbtesting.SetSQLMock(GinkgoT())

		mocker.ExpectQuery(
			`SELECT id FROM roles WHERE name = ? ORDER BY roles.id LIMIT ?`,
			static_roles.Viewer.Name, int64(1),
		).ReturnsError(errors.New("role not found"))

		Expect(func() {
			CreateAccessRule(runaidb.DB(), 1, "user@example.com", EntityTypeRegularUser, static_roles.Viewer.Name, "1", api.SCOPETYPE_TENANT, "tenant", "")
		}).To(Panic())
	})

	It("panics when insert fails", func() {
		mocker := dbtesting.SetSQLMock(GinkgoT())

		mocker.ExpectQuery(
			`SELECT id FROM roles WHERE name = ? ORDER BY roles.id LIMIT ?`,
			static_roles.Viewer.Name, int64(1),
		).Returns([]roleIDRow{{ID: 10}})

		mocker.ExpectTransaction().Begin()
		mocker.ExpectQuery(
			`INSERT INTO access_rules (created_at,updated_at,created_by,subject_id,subject_type,role_id,scope_id,scope_type,scope_name,tenant_id,cluster_uuid,deleted_at,phase,phase_message,phase_by_cluster) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING RETURNING id`,
			dbtesting.AnyTime(), dbtesting.AnyTime(), "",
			"user@example.com", string(api.SUBJECTTYPE_USER), int64(10),
			"1", string(api.SCOPETYPE_TENANT), "tenant",
			int64(1), "", nil, "", "", "null",
		).ReturnsError(errors.New("insert failed"))
		mocker.ExpectTransaction().Rollback()

		Expect(func() {
			CreateAccessRule(runaidb.DB(), 1, "user@example.com", EntityTypeRegularUser, static_roles.Viewer.Name, "1", api.SCOPETYPE_TENANT, "tenant", "")
		}).To(Panic())
	})
})

var _ = Describe("getRoleId SQL", func() {
	It("returns the role id", func() {
		mocker := dbtesting.SetSQLMock(GinkgoT())

		mocker.ExpectQuery(
			`SELECT id FROM roles WHERE name = ? ORDER BY roles.id LIMIT ?`,
			"Editor", int64(1),
		).Returns([]roleIDRow{{ID: 42}})

		Expect(getRoleId(runaidb.DB(), "Editor")).To(Equal(uint(42)))
	})
})
