/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"errors"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
)

const (
	complianceFirstOrCreateSelect = "SELECT * FROM policy_compliance WHERE policy_compliance.tenant_id = ? AND policy_compliance.project_id = ? AND policy_compliance.template_id = ? AND policy_compliance.workload_type = ? ORDER BY policy_compliance.tenant_id LIMIT ?"
	complianceUpdate              = "UPDATE policy_compliance SET aligned=?,policy_time=?,project_id=?,template_id=?,tenant_id=?,workload_type=? WHERE policy_compliance.tenant_id = ? AND policy_compliance.project_id = ? AND policy_compliance.template_id = ? AND policy_compliance.workload_type = ?"
	complianceInsert              = "INSERT INTO policy_compliance (tenant_id,project_id,template_id,workload_type,policy_time,aligned,reason,imposed_asset) VALUES (?,?,?,?,?,?,?,?)"
	complianceListSelect          = "SELECT * FROM policy_compliance WHERE (policy_compliance.tenant_id = ? AND policy_compliance.project_id = ? AND policy_compliance.workload_type = ?) AND template_id IN (?)"
	complianceDeleteOfTemplate    = "DELETE FROM policy_compliance WHERE policy_compliance.tenant_id = ? AND policy_compliance.template_id = ?"
	complianceDeleteNonRelevant   = "DELETE FROM policy_compliance WHERE (policy_compliance.tenant_id = ? AND policy_compliance.project_id = ? AND policy_compliance.workload_type = ?) AND policy_time != (?)"
)

var compliancePolicyTime = time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

type complianceRow struct {
	TenantId     int32
	ProjectId    int32
	TemplateId   string
	WorkloadType string
	Aligned      bool
}

func compliancePolicyInfo() PolicyInfo {
	return PolicyInfo{
		TenantId:     7,
		ProjectId:    11,
		WorkloadType: "workspace",
		ReplicaType:  "worker",
		PolicyTime:   compliancePolicyTime,
	}
}

var _ = Describe("PolicyInfo.WorkloadAndReplicaType", func() {
	It("joins the workload and replica types when a replica type is set", func() {
		pi := PolicyInfo{WorkloadType: "workspace", ReplicaType: "worker"}

		Expect(pi.WorkloadAndReplicaType()).To(Equal("workspace:worker"))
	})

	It("returns just the workload type when no replica type is set", func() {
		pi := PolicyInfo{WorkloadType: "workspace"}

		Expect(pi.WorkloadAndReplicaType()).To(Equal("workspace"))
	})
})

var _ = Describe("UpsertPolicyComplianceOfTemplate", func() {
	It("updates an existing record", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(complianceFirstOrCreateSelect).Returns([]complianceRow{{TenantId: 7, ProjectId: 11, TemplateId: "tpl-1", WorkloadType: "workspace:worker"}})
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := UpsertPolicyComplianceOfTemplate(testCtx(7), "tpl-1", compliancePolicyInfo(), nil)

		Expect(apiErr).To(BeNil())
	})

	It("inserts a new record", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(complianceFirstOrCreateSelect).Returns([]complianceRow{})
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceInsert).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := UpsertPolicyComplianceOfTemplate(testCtx(7), "tpl-1", compliancePolicyInfo(), []api.ComplianceInfoReason{{}})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the lookup fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(complianceFirstOrCreateSelect).ReturnsError(errors.New("boom"))

		apiErr := UpsertPolicyComplianceOfTemplate(testCtx(7), "tpl-1", compliancePolicyInfo(), nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("ListPolicyComplianceOfTemplates", func() {
	It("returns the matching compliance records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteNonRelevant).Returns(1)
		m.ExpectTransaction().Commit()
		m.ExpectQuery(complianceListSelect).Returns([]complianceRow{{TenantId: 7, TemplateId: "tpl-1"}})

		records, apiErr := ListPolicyComplianceOfTemplates(testCtx(7), []string{"tpl-1"}, compliancePolicyInfo())

		Expect(apiErr).To(BeNil())
		Expect(records).To(HaveLen(1))
	})

	It("returns an error when the cleanup delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteNonRelevant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		records, apiErr := ListPolicyComplianceOfTemplates(testCtx(7), []string{"tpl-1"}, compliancePolicyInfo())

		Expect(records).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the list query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteNonRelevant).Returns(1)
		m.ExpectTransaction().Commit()
		m.ExpectQuery(complianceListSelect).ReturnsError(errors.New("boom"))

		records, apiErr := ListPolicyComplianceOfTemplates(testCtx(7), []string{"tpl-1"}, compliancePolicyInfo())

		Expect(records).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("deletePolicyComplianceOfTemplate", func() {
	It("deletes the compliance records of the template", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteOfTemplate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := deletePolicyComplianceOfTemplate(testCtx(7), runaidb.DB(), "tpl-1")

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteOfTemplate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := deletePolicyComplianceOfTemplate(testCtx(7), runaidb.DB(), "tpl-1")

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("deleteNonRelevantCompliance", func() {
	It("deletes the stale compliance records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteNonRelevant).Returns(2)
		m.ExpectTransaction().Commit()

		apiErr := deleteNonRelevantCompliance(testCtx(7), runaidb.DB(), compliancePolicyInfo())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(complianceDeleteNonRelevant).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := deleteNonRelevantCompliance(testCtx(7), runaidb.DB(), compliancePolicyInfo())

		Expect(apiErr).ToNot(BeNil())
	})
})
