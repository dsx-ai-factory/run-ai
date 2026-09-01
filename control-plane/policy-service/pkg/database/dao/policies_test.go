package dao

import (
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authApi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const policiesInsertCols = `INSERT INTO policies (id,created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,name,workload_type,scope,tenant_id,cluster_id,department_id,project_id,defaults_v2,rules_v2,imposed_assets,gvk_workload,validation_status,validation_issues) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

func newPolicy() *dbmodel.Policy {
	return &dbmodel.Policy{
		BaseModel:    dbmodels.BaseModel{ID: uuid.New()},
		Name:         "p1",
		WorkloadType: "workspace",
		Scope:        api.SCOPETYPE_PROJECT,
		TenantId:     1,
		ProjectId:    convutils.Ref(int32(5)),
	}
}

func newPolicyNoID() *dbmodel.Policy {
	p := newPolicy()
	p.ID = uuid.Nil
	return p
}

var _ = (*authApi.PermittedScopes)(nil)

var _ = Describe("Policies", func() {
	Describe("AddPolicy", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(policiesInsertCols).Returns(1)
			tx.Commit()

			apiErr := AddPolicy(ctxWithTenant(1), runaidb.DB(), newPolicyNoID())
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(policiesInsertCols).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := AddPolicy(ctxWithTenant(1), runaidb.DB(), newPolicyNoID())
			Expect(apiErr).ToNot(BeNil())
		})

		It("returns a conflict error on gvk conflict", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(policiesInsertCols).ReturnsError(errors.New("duplicate key value violates unique constraint"))
			tx.Rollback()

			p := newPolicyNoID()
			p.GvkWorkload = true
			apiErr := AddPolicy(ctxWithTenant(1), runaidb.DB(), p)
			Expect(apiErr).ToNot(BeNil())
			Expect(int(apiErr.Code)).To(Equal(409))
		})
	})

	Describe("UpdatePolicySpec", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET created_at=?,created_by=?,defaults_v2=?,deleted_at=?,gvk_workload=?,imposed_assets=?,name=?,rules_v2=?,updated_at=?,updated_by=?,validation_issues=?,validation_status=? WHERE policies.id = ?`).Returns(1)
			tx.Commit()

			apiErr := UpdatePolicySpec(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).To(BeNil())
		})

		It("updates with validation status and issues", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET created_at=?,created_by=?,defaults_v2=?,deleted_at=?,gvk_workload=?,imposed_assets=?,name=?,rules_v2=?,updated_at=?,updated_by=?,validation_issues=?,validation_status=? WHERE policies.id = ?`).Returns(1)
			tx.Commit()

			p := newPolicy()
			phase := api.POLICYPHASE_READY
			p.ValidationStatus = &phase
			p.ValidationIssues = []string{"issue1"}
			apiErr := UpdatePolicySpec(ctxWithTenant(1), runaidb.DB(), p)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET created_at=?,created_by=?,defaults_v2=?,deleted_at=?,gvk_workload=?,imposed_assets=?,name=?,rules_v2=?,updated_at=?,updated_by=?,validation_issues=?,validation_status=? WHERE policies.id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := UpdatePolicySpec(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("UpdatePoliciesStatus", func() {
		It("succeeds when given empty input", func() {
			setupDaoMock(GinkgoT())
			apiErr := UpdatePoliciesStatus(ctxWithTenant(1), runaidb.DB(), nil, nil)
			Expect(apiErr).To(BeNil())
		})

		It("updates with the given status", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).Returns(1)
			tx.Commit()

			status := &api.PolicyStatus{Phase: api.POLICYPHASE_PENDING_UPDATE, Issues: []string{"x"}}
			apiErr := UpdatePoliciesStatus(ctxWithTenant(1), runaidb.DB(), []string{"id1"}, status)
			Expect(apiErr).To(BeNil())
		})

		It("updates with a nil status", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).Returns(1)
			tx.Commit()

			apiErr := UpdatePoliciesStatus(ctxWithTenant(1), runaidb.DB(), []string{"id1"}, nil)
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("PatchPolicyScopes", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET cluster_id=?,department_id=?,project_id=? WHERE policies.id = ?`).Returns(1)
			tx.Commit()

			apiErr := PatchPolicyScopes(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET cluster_id=?,department_id=?,project_id=? WHERE policies.id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := PatchPolicyScopes(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteObsoletePoliciesByScope", func() {
		It("succeeds when there is nothing to delete", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND project_id = ?`).Returns([]struct{ ID string }{})
			tx.Commit()

			apiErr := DeleteObsoletePoliciesByScope(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, "5")
			Expect(apiErr).To(BeNil())
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			pid := uuid.New().String()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND project_id = ?`).Returns([]struct{ ID string }{{ID: pid}})
			m.ExpectExec(`DELETE FROM policies WHERE id IN (?)`).Returns(1)
			m.ExpectExec(`DELETE FROM effective_policies WHERE policy_id IN (?)`).Returns(1)
			tx.Commit()

			apiErr := DeleteObsoletePoliciesByScope(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, "5")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("SoftDeletePolicyById", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET deleted_at=? WHERE (policies.id = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL AND id = ?`).Returns(1)
			tx.Commit()

			p := newPolicy()
			apiErr := SoftDeletePolicyById(ctxWithTenant(1), runaidb.DB(), p)
			Expect(apiErr).To(BeNil())
			Expect(p.DeletedAt.Valid).To(BeTrue())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET deleted_at=? WHERE (policies.id = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL AND id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := SoftDeletePolicyById(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("HardDeletePolicyById", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			txe := m.ExpectTransaction()
			txe.Begin()
			m.ExpectExec(`DELETE FROM effective_policies WHERE policy_id = ? AND tenant_id = ?`).Returns(1)
			txe.Commit()

			txp := m.ExpectTransaction()
			txp.Begin()
			m.ExpectExec(`DELETE FROM policies WHERE policies.id = ? AND policies.tenant_id = ?`).Returns(1)
			txp.Commit()

			apiErr := HardDeletePolicyById(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).To(BeNil())
		})

		It("returns an error when deleting effective policies fails", func() {
			m := setupDaoMock(GinkgoT())
			txe := m.ExpectTransaction()
			txe.Begin()
			m.ExpectExec(`DELETE FROM effective_policies WHERE policy_id = ? AND tenant_id = ?`).ReturnsError(errors.New("boom"))
			txe.Rollback()

			apiErr := HardDeletePolicyById(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("FindPolicyByID", func() {
		It("returns an error for an invalid id", func() {
			setupDaoMock(GinkgoT())
			res, apiErr := FindPolicyByID(ctxWithTenant(1), runaidb.DB(), "not-a-uuid", FindPolicyOptions{})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(400))
		})

		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.id = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.PolicyWithOptionalEffective{
				{Policy: dbmodel.Policy{BaseModel: dbmodels.BaseModel{ID: id}, TenantId: 1}},
			})

			res, apiErr := FindPolicyByID(ctxWithTenant(1), runaidb.DB(), id.String(), FindPolicyOptions{})
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.id = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.PolicyWithOptionalEffective{})

			res, apiErr := FindPolicyByID(ctxWithTenant(1), runaidb.DB(), id.String(), FindPolicyOptions{})
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("returns the record when including deleted", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) WHERE policies.id = ? AND policies.tenant_id = ?`).Returns([]dbmodel.PolicyWithOptionalEffective{
				{Policy: dbmodel.Policy{BaseModel: dbmodels.BaseModel{ID: id}, TenantId: 1}},
			})

			res, apiErr := FindPolicyByID(ctxWithTenant(1), runaidb.DB(), id.String(),
				FindPolicyOptions{IncludeDeletedGivenPolicies: true, IncludeDeletedEffectivePolicies: true})
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			id := uuid.New()
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.id = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL`).ReturnsError(errors.New("boom"))

			res, apiErr := FindPolicyByID(ctxWithTenant(1), runaidb.DB(), id.String(), FindPolicyOptions{})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("ListPoliciesByOrgUnits", func() {
		It("lists policies by org units", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (false or (policies.scope = ? and policies.project_id = ?)) AND (policies.workload_type = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.PolicyWithOptionalEffective{})

			orgUnits := []scopes.OrgUnit{{Scope: api.SCOPETYPE_PROJECT, Id: "5"}}
			res, apiErr := ListPoliciesByOrgUnits(ctxWithTenant(1), runaidb.DB(), 1, "workspace", orgUnits, ListByOrgUnitOptions{})
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("lists policies including fallback and system", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) WHERE (false or (policies.scope = ? and policies.tenant_id = ?) or (policies.scope = ? and policies.department_id = ?) or (policies.scope = ? and policies.cluster_id = ?)) AND ((policies.workload_type = ? or policies.workload_type = ?))`).Returns([]dbmodel.PolicyWithOptionalEffective{})

			orgUnits := []scopes.OrgUnit{
				{Scope: api.SCOPETYPE_TENANT, Id: "1"},
				{Scope: api.SCOPETYPE_DEPARTMENT, Id: "d1"},
				{Scope: api.SCOPETYPE_CLUSTER, Id: "c1"},
			}
			res, apiErr := ListPoliciesByOrgUnits(ctxWithTenant(1), runaidb.DB(), 1, "workspace", orgUnits,
				ListByOrgUnitOptions{IncludeDeleted: true, IncludeFallback: true, IncludeSystem: true})
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (false or (policies.scope = ? and policies.project_id = ?)) AND (policies.workload_type = ?) AND policies.deleted_at IS NULL`).ReturnsError(errors.New("boom"))

			orgUnits := []scopes.OrgUnit{{Scope: api.SCOPETYPE_PROJECT, Id: "5"}}
			res, apiErr := ListPoliciesByOrgUnits(ctxWithTenant(1), runaidb.DB(), 1, "workspace", orgUnits, ListByOrgUnitOptions{})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("CountGivenPoliciesOfTenant", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT count(*) FROM policies WHERE policies.tenant_id = ?`).Returns([]countRow{{Count: 3}})

			count, apiErr := CountGivenPoliciesOfTenant(ctxWithTenant(1))
			Expect(apiErr).To(BeNil())
			Expect(count).To(Equal(int64(3)))
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT count(*) FROM policies WHERE policies.tenant_id = ?`).ReturnsError(errors.New("boom"))

			count, apiErr := CountGivenPoliciesOfTenant(ctxWithTenant(1))
			Expect(apiErr).ToNot(BeNil())
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("ListGivenPoliciesOfTenant", func() {
		It("returns an error when neither native nor gvk policies are requested", func() {
			setupDaoMock(GinkgoT())
			res, apiErr := ListGivenPoliciesOfTenant(ctxWithTenant(1), "workspace", "", "", 0, "", nil, ListOptions{})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
			Expect(int(apiErr.Code)).To(Equal(400))
		})

		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.update_count, effective_policies.validation_error FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.workload_type = ? AND policies.tenant_id = ?) AND ((policies.defaults_v2 is not null or policies.rules_v2 is not null or policies.imposed_assets is not null)) AND policies.deleted_at IS NULL AND policies.workload_type <> ? AND (policies.gvk_workload = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.PolicyWithPartialEffective{})

			res, apiErr := ListGivenPoliciesOfTenant(ctxWithTenant(1), "workspace", "", "", 0, "", nil,
				ListOptions{IncludeNativePolicies: true})
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.update_count, effective_policies.validation_error FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) WHERE (policies.workload_type = ? AND policies.tenant_id = ?) AND ((policies.defaults_v2 is not null or policies.rules_v2 is not null or policies.imposed_assets is not null))`).ReturnsError(errors.New("boom"))

			res, apiErr := ListGivenPoliciesOfTenant(ctxWithTenant(1), "workspace", "", "", 0, "", nil,
				ListOptions{IncludeNativePolicies: true, IncludeGvkPolicies: true, IncludeDeleted: true, IncludeFallbackPolicies: true})
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("ListGvkPoliciesOfTenant", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.tenant_id = ? AND policies.gvk_workload = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL ORDER BY id DESC`).Returns([]dbmodel.PolicyWithOptionalEffective{})

			res, apiErr := ListGvkPoliciesOfTenant(ctxWithTenant(1), nil, nil, "", "", 0, 0)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("lists with sort, offset and limit", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.tenant_id = ? AND policies.gvk_workload = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL ORDER BY id DESC NULLS LAST,id DESC LIMIT ? OFFSET ?`).Returns([]dbmodel.PolicyWithOptionalEffective{})

			res, apiErr := ListGvkPoliciesOfTenant(ctxWithTenant(1), nil, nil, "id", "desc", 10, 20)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.tenant_id = ? AND policies.gvk_workload = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL ORDER BY id DESC`).ReturnsError(errors.New("boom"))

			res, apiErr := ListGvkPoliciesOfTenant(ctxWithTenant(1), nil, nil, "", "", 0, 0)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("ListPoliciesOfTenant", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.policy_id FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.workload_type = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL`).Returns([]dbmodel.PolicyWithOptionalEffectiveId{})

			res, apiErr := ListPoliciesOfTenant(ctxWithTenant(1), 1, "workspace", false)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("lists including deleted", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.policy_id FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE policies.workload_type = ? AND policies.tenant_id = ?`).Returns([]dbmodel.PolicyWithOptionalEffectiveId{})

			res, apiErr := ListPoliciesOfTenant(ctxWithTenant(1), 1, "workspace", true)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.policy_id FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.workload_type = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL`).ReturnsError(errors.New("boom"))

			res, apiErr := ListPoliciesOfTenant(ctxWithTenant(1), 1, "workspace", false)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("FindGivenPolicyOfOrgUnit", func() {
		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM policies WHERE (policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.Policy{{BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1}})

			res, apiErr := FindGivenPolicyOfOrgUnit(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, nil, convutils.Ref(int32(5)), "workspace", false)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM policies WHERE policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?`).Returns([]dbmodel.Policy{})

			res, apiErr := FindGivenPolicyOfOrgUnit(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, nil, convutils.Ref(int32(5)), "workspace", true)
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM policies WHERE (policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?) AND policies.deleted_at IS NULL`).ReturnsError(errors.New("boom"))

			res, apiErr := FindGivenPolicyOfOrgUnit(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, nil, convutils.Ref(int32(5)), "workspace", false)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("FindGivenPoliciesOfOrgUnitAllTypes", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM policies WHERE (policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.Policy{{BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1}})

			res, apiErr := FindGivenPoliciesOfOrgUnitAllTypes(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, nil, convutils.Ref(int32(5)), false)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM policies WHERE policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?`).ReturnsError(errors.New("boom"))

			res, apiErr := FindGivenPoliciesOfOrgUnitAllTypes(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, nil, convutils.Ref(int32(5)), true)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("FindPolicyOfOrgUnit", func() {
		It("returns the record", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?) AND policies.deleted_at IS NULL`).Returns([]dbmodel.PolicyWithOptionalEffective{
				{Policy: dbmodel.Policy{BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1}},
			})

			res, apiErr := FindPolicyOfOrgUnit(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, convutils.Ref(int32(5)), nil, "workspace", false)
			Expect(apiErr).To(BeNil())
			Expect(res).ToNot(BeNil())
		})

		It("returns not found", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) WHERE policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?`).Returns([]dbmodel.PolicyWithOptionalEffective{})

			res, apiErr := FindPolicyOfOrgUnit(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, convutils.Ref(int32(5)), nil, "workspace", true)
			Expect(apiErr).To(BeNil())
			Expect(res).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?) AND policies.deleted_at IS NULL`).ReturnsError(errors.New("boom"))

			res, apiErr := FindPolicyOfOrgUnit(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, nil, convutils.Ref(int32(5)), nil, "workspace", false)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("FindAllIdsOfOrgUnit", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.id FROM policies WHERE policies.deleted_at IS NULL AND (policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?)`).Returns([]struct{ ID string }{{ID: uuid.New().String()}})

			res, apiErr := FindAllIdsOfOrgUnit(ctxWithTenant(1), runaidb.DB(), newPolicy(), false)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})

		It("returns an error on db failure when including deleted", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.id FROM policies WHERE policies.workload_type = ? AND policies.scope = ? AND policies.tenant_id = ? AND policies.project_id = ?`).ReturnsError(errors.New("boom"))

			res, apiErr := FindAllIdsOfOrgUnit(ctxWithTenant(1), runaidb.DB(), newPolicy(), true)
			Expect(apiErr).ToNot(BeNil())
			Expect(res).To(BeNil())
		})
	})

	Describe("UpdatePoliciesStatus", func() {
		It("updates with empty issues", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).Returns(1)
			tx.Commit()

			status := &api.PolicyStatus{Phase: api.POLICYPHASE_PENDING_UPDATE, Issues: []string{}}
			apiErr := UpdatePoliciesStatus(ctxWithTenant(1), runaidb.DB(), []string{"id1"}, status)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`UPDATE policies SET validation_issues=?,validation_status=? WHERE policies.tenant_id = ? AND policies.id IN (?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			status := &api.PolicyStatus{Phase: api.POLICYPHASE_PENDING_UPDATE, Issues: []string{"x"}}
			apiErr := UpdatePoliciesStatus(ctxWithTenant(1), runaidb.DB(), []string{"id1"}, status)
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteObsoletePoliciesByScope", func() {
		It("succeeds for a system scope", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			tx.Rollback()

			apiErr := DeleteObsoletePoliciesByScope(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_SYSTEM, "x")
			Expect(apiErr).To(BeNil())
		})

		It("documents current behavior: returns nil when deleting policies fails", func() {
			m := setupDaoMock(GinkgoT())
			pid := uuid.New().String()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND project_id = ?`).Returns([]struct{ ID string }{{ID: pid}})
			m.ExpectExec(`DELETE FROM policies WHERE id IN (?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := DeleteObsoletePoliciesByScope(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, "5")
			Expect(apiErr).To(BeNil())
		})

		It("documents current behavior: returns nil when deleting effective policies fails", func() {
			m := setupDaoMock(GinkgoT())
			pid := uuid.New().String()
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND project_id = ?`).Returns([]struct{ ID string }{{ID: pid}})
			m.ExpectExec(`DELETE FROM policies WHERE id IN (?)`).Returns(1)
			m.ExpectExec(`DELETE FROM effective_policies WHERE policy_id IN (?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := DeleteObsoletePoliciesByScope(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, "5")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("HardDeletePolicyById", func() {
		It("returns an error when deleting the policy fails", func() {
			m := setupDaoMock(GinkgoT())
			txe := m.ExpectTransaction()
			txe.Begin()
			m.ExpectExec(`DELETE FROM effective_policies WHERE policy_id = ? AND tenant_id = ?`).Returns(1)
			txe.Commit()

			txp := m.ExpectTransaction()
			txp.Begin()
			m.ExpectExec(`DELETE FROM policies WHERE policies.id = ? AND policies.tenant_id = ?`).ReturnsError(errors.New("boom"))
			txp.Rollback()

			apiErr := HardDeletePolicyById(ctxWithTenant(1), runaidb.DB(), newPolicy())
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ListGvkPoliciesOfTenant", func() {
		It("returns populated records", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.*, policies.deleted_at as policy_deleted, effective_policies.deleted_at as effective_deleted FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.tenant_id = ? AND policies.gvk_workload = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL ORDER BY id DESC`).
				Returns([]dbmodel.PolicyWithOptionalEffective{
					{
						Policy:          dbmodel.Policy{BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1, GvkWorkload: true},
						EffectivePolicy: dbmodel.EffectivePolicy{RulesV2: dbmodels.JsonBMap{"a": 1}, DefaultsV2: dbmodels.JsonBMap{"b": 2}},
					},
				})

			res, apiErr := ListGvkPoliciesOfTenant(ctxWithTenant(1), nil, nil, "", "", 0, 0)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("ListPoliciesOfTenant", func() {
		It("returns populated records", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT policies.*, effective_policies.policy_id FROM policies LEFT JOIN effective_policies ON (effective_policies.policy_id = policies.id) AND (effective_policies.deleted_at IS NULL) WHERE (policies.workload_type = ? AND policies.tenant_id = ?) AND policies.deleted_at IS NULL AND policies.deleted_at IS NULL`).
				Returns([]dbmodel.PolicyWithOptionalEffectiveId{
					{Policy: dbmodel.Policy{BaseModel: dbmodels.BaseModel{ID: uuid.New()}, TenantId: 1, WorkloadType: "workspace", DefaultsV2: dbmodels.JsonBMap{"a": 1}, RulesV2: dbmodels.JsonBMap{"b": 2}}},
				})

			res, apiErr := ListPoliciesOfTenant(ctxWithTenant(1), 1, "workspace", false)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(1))
		})
	})

	Describe("DeleteObsoletePoliciesByScope", func() {
		It("documents current behavior: returns nil when plucking ids fails", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectQuery(`SELECT id FROM policies WHERE tenant_id = ? AND project_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()

			apiErr := DeleteObsoletePoliciesByScope(ctxWithTenant(1), runaidb.DB(), api.SCOPETYPE_PROJECT, "5")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("FindPoliciesForSync", func() {
		It("breaks at the bulk limit", func() {
			config.MockConfig()
			m := setupDaoMock(GinkgoT())

			rows := make([]dbmodel.PolicyAndEffective, 0, 6)
			for i := 1; i <= 6; i++ {
				rows = append(rows, dbmodel.PolicyAndEffective{
					Policy: dbmodel.Policy{TenantId: 1, Scope: api.SCOPETYPE_PROJECT, WorkloadType: "workspace", ProjectId: convutils.Ref(int32(i))},
				})
			}

			m.ExpectQuery(`SELECT * FROM effective_policies JOIN policies ON effective_policies.policy_id = policies.id WHERE effective_policies.tenant_id = ? AND ((policies.defaults_v2 is not null or policies.rules_v2 is not null or policies.imposed_assets is not null)) AND (policies.cluster_id IS NULL OR policies.cluster_id = ?) AND ((workload_type = ? or workload_type = ?)) ORDER BY update_count LIMIT ?`).
				Returns(rows)

			res, apiErr := FindPoliciesForSync(ctxWithTenant(1), "workspace", "c1", 0)
			Expect(apiErr).To(BeNil())
			Expect(res).To(HaveLen(5))
		})
	})
})
