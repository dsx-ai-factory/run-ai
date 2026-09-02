package dao

import (
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

var _ = Describe("EffectivePolicies", func() {
	Describe("GetEffectivePolicyById", func() {
		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM effective_policies WHERE (effective_policies.policy_id = ? AND effective_policies.tenant_id = ?) AND effective_policies.deleted_at IS NULL`).
				Returns([]dbmodel.EffectivePolicy{{PolicyId: uuid.New(), TenantId: 1}})

			res, apiErr := GetEffectivePolicyById(ctxWithTenant(1), uuid.New(), false)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM effective_policies WHERE (effective_policies.policy_id = ? AND effective_policies.tenant_id = ?) AND effective_policies.deleted_at IS NULL`).
				Returns([]dbmodel.EffectivePolicy{})

			res, apiErr := GetEffectivePolicyById(ctxWithTenant(1), uuid.New(), false)
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("includes deleted records", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? AND effective_policies.tenant_id = ?`).
				Returns([]dbmodel.EffectivePolicy{{PolicyId: uuid.New(), TenantId: 1}})

			res, apiErr := GetEffectivePolicyById(ctxWithTenant(1), uuid.New(), true)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM effective_policies WHERE (effective_policies.policy_id = ? AND effective_policies.tenant_id = ?) AND effective_policies.deleted_at IS NULL`).
				ReturnsError(errors.New("boom"))

			res, apiErr := GetEffectivePolicyById(ctxWithTenant(1), uuid.New(), false)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("UpsertEffectivePolicy", func() {
		It("creates when missing", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? ORDER BY effective_policies.policy_id LIMIT ?`).
				Returns([]dbmodel.EffectivePolicy{})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO effective_policies (policy_id,tenant_id,defaults_v2,rules_v2,imposed_assets,updated_at,update_count,default_from_order,validation_error,deleted_at) VALUES (?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()

			apiErr := UpsertEffectivePolicy(ctxWithTenant(1), runaidb.DB(), &dbmodel.EffectivePolicy{PolicyId: uuid.New()})
			Expect(apiErr).To(BeNil())
		})

		It("updates when exists", func() {
			m := setupDaoMock(GinkgoT())

			existing := uuid.New()
			m.ExpectQuery(`SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? ORDER BY effective_policies.policy_id LIMIT ?`).
				Returns([]dbmodel.EffectivePolicy{{PolicyId: existing}})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET default_from_order=?,defaults_v2=?,deleted_at=?,imposed_assets=?,rules_v2=?,validation_error=? WHERE effective_policies.policy_id = ?`).Returns(1)
			tx.Commit()

			apiErr := UpsertEffectivePolicy(ctxWithTenant(1), runaidb.DB(), &dbmodel.EffectivePolicy{PolicyId: existing})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on first db failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? ORDER BY effective_policies.policy_id LIMIT ?`).
				ReturnsError(errors.New("boom"))

			apiErr := UpsertEffectivePolicy(ctxWithTenant(1), runaidb.DB(), &dbmodel.EffectivePolicy{PolicyId: uuid.New()})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("SoftDeleteEffectivePoliciesByOrgUnit", func() {
		It("deletes for project scope", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE effective_policies.policy_id in (select id from policies where policies.id = effective_policies.policy_id and policies.workload_type = ? and policies.scope = ? and policies.project_id = ?)`).Returns(1)
			tx.Commit()

			apiErr := SoftDeleteEffectivePoliciesByOrgUnit(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_PROJECT, Id: "5"}, "workspace")
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE effective_policies.policy_id in (select id from policies where policies.id = effective_policies.policy_id and policies.workload_type = ? and policies.scope = ? and policies.tenant_id = ?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := SoftDeleteEffectivePoliciesByOrgUnit(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_TENANT, Id: "1"}, "workspace")
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("SoftDeleteEffectivePoliciesOfSubtree", func() {
		It("returns nil for project scope with no subtree", func() {
			setupDaoMock(GinkgoT())
			ids, apiErr := SoftDeleteEffectivePoliciesOfSubtree(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_PROJECT, Id: "5"}, "workspace")
			Expect(apiErr).To(BeNil())
			Expect(ids).To(BeNil())
		})

		It("returns nil when there are no descendants", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND workload_type = ? AND scope = ? AND department_id = ?`).
				Returns([]struct{ ID string }{})

			ids, apiErr := SoftDeleteEffectivePoliciesOfSubtree(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"}, "workspace")
			Expect(apiErr).To(BeNil())
			Expect(ids).To(BeNil())
		})

		It("succeeds for department scope", func() {
			m := setupDaoMock(GinkgoT())
			pid := uuid.New().String()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND workload_type = ? AND scope = ? AND department_id = ?`).
				Returns([]struct{ ID string }{{ID: pid}})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE policy_id IN (?)`).Returns(1)
			tx.Commit()

			ids, apiErr := SoftDeleteEffectivePoliciesOfSubtree(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"}, "workspace")
			Expect(apiErr).To(BeNil())
			Expect(ids).To(Equal([]string{pid}))
		})

		It("returns an error when listing fails", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND workload_type = ? AND scope = ? AND department_id = ?`).
				ReturnsError(errors.New("boom"))

			ids, apiErr := SoftDeleteEffectivePoliciesOfSubtree(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"}, "workspace")
			Expect(apiErr).ToNot(BeNil())
			Expect(ids).To(BeNil())
		})
	})

	Describe("listSubtreePolicyIds", func() {
		It("lists ids for system scope", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT id FROM policies WHERE gvk_workload IS NOT TRUE`).
				Returns([]struct{ ID string }{{ID: uuid.New().String()}})

			ids, apiErr := listSubtreePolicyIds(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_SYSTEM}, FallbackPolicyAsType)
			Expect(apiErr).To(BeNil())
			Expect(ids).To(HaveLen(1))
		})

		It("lists ids for cluster scope", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND workload_type = ? AND scope IN (?,?) AND cluster_id = ?`).
				Returns([]struct{ ID string }{{ID: uuid.New().String()}})

			ids, apiErr := listSubtreePolicyIds(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_CLUSTER, Id: "c1"}, "workspace")
			Expect(apiErr).To(BeNil())
			Expect(ids).To(HaveLen(1))
		})

		It("returns empty for tenant fallback", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND scope IN (?,?,?)`).
				Returns([]struct{ ID string }{})

			ids, apiErr := listSubtreePolicyIds(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_TENANT, Id: "1"}, FallbackPolicyAsType)
			Expect(apiErr).To(BeNil())
			Expect(ids).To(BeEmpty())
		})
	})

	Describe("FindPoliciesForSync", func() {
		It("succeeds", func() {
			config.MockConfig()
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM effective_policies JOIN policies ON effective_policies.policy_id = policies.id WHERE effective_policies.tenant_id = ? AND ((policies.defaults_v2 is not null or policies.rules_v2 is not null or policies.imposed_assets is not null)) AND (policies.cluster_id IS NULL OR policies.cluster_id = ?) AND ((workload_type = ? or workload_type = ?)) ORDER BY update_count LIMIT ?`).Returns([]dbmodel.PolicyAndEffective{})

			res, apiErr := FindPoliciesForSync(ctxWithTenant(1), "workspace", "c1", 0)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			config.MockConfig()
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM effective_policies JOIN policies ON effective_policies.policy_id = policies.id WHERE effective_policies.tenant_id = ? AND ((policies.defaults_v2 is not null or policies.rules_v2 is not null or policies.imposed_assets is not null)) AND (policies.cluster_id IS NULL OR policies.cluster_id = ?) AND ((workload_type = ? or workload_type = ?)) AND update_count > ? ORDER BY update_count LIMIT ?`).ReturnsError(errors.New("boom"))

			res, apiErr := FindPoliciesForSync(ctxWithTenant(1), "workspace", "c1", 10)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("UpsertEffectivePolicy", func() {
		It("returns an error on create failure", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? ORDER BY effective_policies.policy_id LIMIT ?`).
				Returns([]dbmodel.EffectivePolicy{})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO effective_policies (policy_id,tenant_id,defaults_v2,rules_v2,imposed_assets,updated_at,update_count,default_from_order,validation_error,deleted_at) VALUES (?,?,?,?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := UpsertEffectivePolicy(ctxWithTenant(1), runaidb.DB(), &dbmodel.EffectivePolicy{PolicyId: uuid.New()})
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns an error on update failure", func() {
			m := setupDaoMock(GinkgoT())

			existing := uuid.New()
			m.ExpectQuery(`SELECT * FROM effective_policies WHERE effective_policies.policy_id = ? ORDER BY effective_policies.policy_id LIMIT ?`).
				Returns([]dbmodel.EffectivePolicy{{PolicyId: existing}})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET default_from_order=?,defaults_v2=?,deleted_at=?,imposed_assets=?,rules_v2=?,validation_error=? WHERE effective_policies.policy_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := UpsertEffectivePolicy(ctxWithTenant(1), runaidb.DB(), &dbmodel.EffectivePolicy{PolicyId: existing})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("SoftDeleteEffectivePoliciesByOrgUnit", func() {
		It("deletes for department scope", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE effective_policies.policy_id in (select id from policies where policies.id = effective_policies.policy_id and policies.workload_type = ? and policies.scope = ? and policies.department_id = ?)`).Returns(1)
			tx.Commit()

			apiErr := SoftDeleteEffectivePoliciesByOrgUnit(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"}, "workspace")
			Expect(apiErr).To(BeNil())
		})

		It("deletes for cluster scope", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE effective_policies.policy_id in (select id from policies where policies.id = effective_policies.policy_id and policies.workload_type = ? and policies.scope = ? and policies.cluster_id = ?)`).Returns(1)
			tx.Commit()

			apiErr := SoftDeleteEffectivePoliciesByOrgUnit(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_CLUSTER, Id: "c1"}, "workspace")
			Expect(apiErr).To(BeNil())
		})

		It("deletes for fallback policy type", func() {
			m := setupDaoMock(GinkgoT())

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE effective_policies.policy_id in (select id from policies where policies.id = effective_policies.policy_id and policies.scope = ? and policies.tenant_id = ?)`).Returns(1)
			tx.Commit()

			apiErr := SoftDeleteEffectivePoliciesByOrgUnit(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_TENANT, Id: "1"}, FallbackPolicyAsType)
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("SoftDeleteEffectivePoliciesOfSubtree", func() {
		It("returns an error on update failure", func() {
			m := setupDaoMock(GinkgoT())
			pid := uuid.New().String()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND workload_type = ? AND scope = ? AND department_id = ?`).
				Returns([]struct{ ID string }{{ID: pid}})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE effective_policies SET deleted_at=? WHERE policy_id IN (?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			ids, apiErr := SoftDeleteEffectivePoliciesOfSubtree(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"}, "workspace")
			Expect(apiErr).ToNot(BeNil())
			Expect(ids).To(BeNil())
		})
	})

	Describe("listSubtreePolicyIds", func() {
		It("returns nil for default scope", func() {
			setupDaoMock(GinkgoT())
			ids, apiErr := listSubtreePolicyIds(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_PROJECT, Id: "5"}, "workspace")
			Expect(apiErr).To(BeNil())
			Expect(ids).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND workload_type = ? AND scope = ? AND department_id = ?`).
				ReturnsError(errors.New("boom"))

			ids, apiErr := listSubtreePolicyIds(ctxWithTenant(1), runaidb.DB(),
				scopes.OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"}, "workspace")
			Expect(apiErr).ToNot(BeNil())
			Expect(ids).To(BeNil())
		})
	})

	Describe("FindPoliciesForSync", func() {
		It("chooses type specific over fallback", func() {
			config.MockConfig()
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM effective_policies JOIN policies ON effective_policies.policy_id = policies.id WHERE effective_policies.tenant_id = ? AND ((policies.defaults_v2 is not null or policies.rules_v2 is not null or policies.imposed_assets is not null)) AND (policies.cluster_id IS NULL OR policies.cluster_id = ?) AND ((workload_type = ? or workload_type = ?)) ORDER BY update_count LIMIT ?`).
				Returns([]dbmodel.PolicyAndEffective{
					{Policy: dbmodel.Policy{TenantId: 1, Scope: api.SCOPETYPE_TENANT, WorkloadType: FallbackPolicyAsType}},
					{Policy: dbmodel.Policy{TenantId: 1, Scope: api.SCOPETYPE_TENANT, WorkloadType: "workspace"}},
				})

			res, apiErr := FindPoliciesForSync(ctxWithTenant(1), "workspace", "c1", 0)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
			Expect(res[0].WorkloadType).To(Equal("workspace"))
		})
	})
})
