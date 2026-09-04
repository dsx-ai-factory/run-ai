/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
)

var _ = Describe("db model", func() {
	Context("compare policies", func() {
		It("should return correct result", func() {
			pol1 := Policy{
				BaseModel: dbmodels.BaseModel{
					ID: uuid.New(),
				},
			}
			pol2 := Policy{
				BaseModel: dbmodels.BaseModel{
					ID: uuid.New(),
				},
			}
			Expect(pol1.Compare(pol1)).To(Equal(0))
			if pol1.BaseModel.ID.String() < pol2.BaseModel.ID.String() {
				Expect(pol1.Compare(pol2)).To(Equal(-1))
				Expect(pol2.Compare(pol1)).To(Equal(1))
			} else {
				Expect(pol1.Compare(pol2)).To(Equal(1))
				Expect(pol2.Compare(pol1)).To(Equal(-1))
			}
		})
	})

	Context("compare policies", func() {
		It("should return correct result", func() {
			pol1 := Policy{
				BaseModel: dbmodels.BaseModel{
					ID: uuid.New(),
				},
			}
			pol2 := Policy{
				BaseModel: dbmodels.BaseModel{
					ID: uuid.New(),
				},
			}
			Expect(pol1.Compare(pol1)).To(Equal(0))
			if pol1.BaseModel.ID.String() < pol2.BaseModel.ID.String() {
				Expect(pol1.Compare(pol2)).To(Equal(-1))
				Expect(pol2.Compare(pol1)).To(Equal(1))
			} else {
				Expect(pol1.Compare(pol2)).To(Equal(1))
				Expect(pol2.Compare(pol1)).To(Equal(-1))
			}
		})
	})

	Context("org unitization", func() {
		It("should return correct org unit string", func() {
			Expect(Policy{
				Scope:    api.SCOPETYPE_TENANT,
				TenantId: 4,
			}.OrgUnit().String()).To(Equal("tenant:4"))

			Expect(Policy{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				TenantId:     4,
				DepartmentId: convutils.Ref("123"),
			}.OrgUnit().String()).To(Equal("department:123"))

			Expect(Policy{
				Scope:     api.SCOPETYPE_PROJECT,
				TenantId:  4,
				ProjectId: convutils.Ref(int32(456)),
			}.OrgUnit().String()).To(Equal("project:456"))

			Expect(Policy{
				Scope:     api.SCOPETYPE_CLUSTER,
				TenantId:  4,
				ClusterId: convutils.Ref("789"),
			}.OrgUnit().String()).To(Equal("cluster:789"))
		})
	})

	Context("ApplyScopeQueryParams", func() {
		It("should set the policy scope details correctly", func() {
			policy := Policy{
				TenantId: 1,
			}

			policy.ApplyScopeQueryParams(string(api.POLICYTYPE_WORKSPACE), string(api.SCOPETYPE_TENANT), "5", 2, "uuidofcluster")

			Expect(policy.WorkloadType).To(Equal(string(api.POLICYTYPE_WORKSPACE)))
			Expect(policy.Scope).To(Equal(api.SCOPETYPE_TENANT))
			Expect(*policy.DepartmentId).To(Equal("5"))
			Expect(*policy.ProjectId).To(Equal(int32(2)))
			Expect(*policy.ClusterId).To(Equal("uuidofcluster"))
		})
	})

	Context("ToGvkPolicyApi", func() {
		It("preserves the complete organizational-unit lineage", func() {
			clusterId := "cluster-uuid"
			departmentId := "department-id"
			projectId := int32(123)
			stored := PolicyWithOptionalEffective{
				Policy: Policy{
					BaseModel:    dbmodels.BaseModel{ID: uuid.New()},
					Name:         "apps deployment policy",
					WorkloadType: "apps/Deployment/v1",
					Scope:        api.SCOPETYPE_PROJECT,
					ClusterId:    &clusterId,
					DepartmentId: &departmentId,
					ProjectId:    &projectId,
					GvkWorkload:  true,
				},
			}

			result, err := stored.ToGvkPolicyApi()

			Expect(err).NotTo(HaveOccurred())
			Expect(result.ClusterId).To(HaveValue(Equal(clusterId)))
			Expect(result.DepartmentId).To(HaveValue(Equal(departmentId)))
			Expect(result.ProjectId).To(HaveValue(Equal("123")))
		})
	})
})
