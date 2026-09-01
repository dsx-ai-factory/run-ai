package dao

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

var _ = Describe("SystemPolicy", func() {
	Describe("InjectSystemPolicyForTesting", func() {
		It("handles workspace", func() {
			effective := &api.WorkspacePolicyDefaultsAndRulesV2{}
			InjectSystemPolicyForTesting(effective)

			Expect(effective.Rules).ToNot(BeNil())
			Expect(effective.Rules.Security).ToNot(BeNil())
			Expect(effective.Rules.Security.Privileged).ToNot(BeNil())
		})
	})

	Describe("InjectSystemPolicyForDistributedTesting", func() {
		It("injects the system policy", func() {
			effective := &api.DistributedPolicyDefaultsAndRulesV2{
				Rules: &api.DistributedPolicyRulesV2{
					Master: &api.ReplicaRulesV2{Security: &api.SupersetRulesAllOfSecurity{}},
					Worker: &api.DistributedPolicyRulesV2Worker{Security: &api.SupersetRulesAllOfSecurity{}},
				},
				Defaults: &api.DistributedPolicyDefaultsV2{
					Master: &api.ReplicaDefaultsV2{Security: &api.SupersetDefaultsAllOfSecurity{}},
					Worker: &api.DistributedPolicyDefaultsV2Worker{Security: &api.SupersetDefaultsAllOfSecurity{}},
				},
			}
			InjectSystemPolicyForDistributedTesting(effective)

			Expect(effective.Rules.Master.Security.Privileged).ToNot(BeNil())
			Expect(effective.Rules.Worker.Security.Privileged).ToNot(BeNil())
			Expect(effective.Rules.Master.TerminationGracePeriodSeconds).ToNot(BeNil())
		})
	})

	Describe("InjectSystemPolicyForDistributedInferenceTesting", func() {
		It("injects the system policy", func() {
			effective := &api.DistributedInferencePolicyDefaultsAndRulesV2{
				Rules: &api.DistributedInferenceRulesV2{
					Leader: &api.DistributedInferenceLeaderWorkerRulesV2{Security: &api.DistributedInferenceLeaderWorkerRulesV2Security{}},
					Worker: &api.DistributedInferenceLeaderWorkerRulesV2{Security: &api.DistributedInferenceLeaderWorkerRulesV2Security{}},
				},
				Defaults: &api.DistributedInferenceDefaultsV2{
					Leader: &api.DistributedInferenceLeaderWorkerDefaultsV2{Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{}},
					Worker: &api.DistributedInferenceLeaderWorkerDefaultsV2{Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{}},
				},
			}
			InjectSystemPolicyForDistributedInferenceTesting(effective)

			Expect(effective.Rules.Leader.Security.Privileged).ToNot(BeNil())
			Expect(effective.Rules.Worker.Security.Privileged).ToNot(BeNil())
		})
	})

	Describe("HandleSystemPolicy", func() {
		It("creates a new version", func() {
			m := setupDaoMock(GinkgoT())

			m.ExpectQuery(`SELECT * FROM policies WHERE (policies.workload_type = ? AND policies.scope = ?) AND policies.deleted_at IS NULL`).
				Returns([]struct{ ID string }{})

			m.ExpectQuery(`SELECT id FROM policies WHERE gvk_workload IS NOT TRUE`).
				Returns([]struct{ ID string }{})

			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`INSERT INTO policies (id,created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,name,workload_type,scope,tenant_id,cluster_id,department_id,project_id,defaults_v2,rules_v2,imposed_assets,gvk_workload,validation_status,validation_issues) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()

			HandleSystemPolicy()
		})
	})
})
