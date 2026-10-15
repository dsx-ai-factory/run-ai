/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopeinfo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("scope info", func() {

	Describe("scope info constructors", func() {
		Context("tenant scope", func() {
			It("should return valid scope info", func() {
				scopeInfo := NewTenantScope()
				Expect(scopeInfo).NotTo(BeNil())
				Expect(*scopeInfo).To(Equal(ScopeInfo{
					Scope: "tenant",
				}))
			})
		})

		Context("department scope", func() {
			It("should return valid scope info", func() {
				scopeInfo := NewDepartmentScope("123")
				Expect(scopeInfo).NotTo(BeNil())
				Expect(*scopeInfo).To(Equal(ScopeInfo{
					Scope:   "department",
					ScopeId: "123",
				}))
			})
		})

		Context("project scope", func() {
			It("should return valid scope info", func() {
				scopeInfo := NewProjectScope("321")
				Expect(scopeInfo).NotTo(BeNil())
				Expect(*scopeInfo).To(Equal(ScopeInfo{
					Scope:   "project",
					ScopeId: "321",
				}))
			})
		})

		Context("scope info of an org unit id", func() {
			It("should resolve the id belonging to the scope", func() {
				projectId := int32(7)
				departmentId := "dept-1"
				clusterId := "cluster-1"
				tenantId := int32(3)

				Expect(*ScopeInfoOf(Project, &projectId, &departmentId, &clusterId, &tenantId)).To(Equal(ScopeInfo{
					Scope:   Project,
					ScopeId: "7",
				}))
				Expect(*ScopeInfoOf(Department, &projectId, &departmentId, &clusterId, &tenantId)).To(Equal(ScopeInfo{
					Scope:   Department,
					ScopeId: "dept-1",
				}))
				Expect(*ScopeInfoOf(Cluster, &projectId, &departmentId, &clusterId, &tenantId)).To(Equal(ScopeInfo{
					Scope:   Cluster,
					ScopeId: "cluster-1",
				}))
				Expect(*ScopeInfoOf(Tenant, &projectId, &departmentId, &clusterId, &tenantId)).To(Equal(ScopeInfo{
					Scope:   Tenant,
					ScopeId: "3",
				}))
			})

			// The API models keep every org unit id in a nullable column, so a
			// caller that scopes an asset by project passes nil for the others.
			It("should not panic when the id for the scope is unset", func() {
				Expect(func() { ScopeInfoOf(Project, nil, nil, nil, nil) }).NotTo(Panic())
				Expect(func() { ScopeInfoOf(Department, nil, nil, nil, nil) }).NotTo(Panic())
				Expect(func() { ScopeInfoOf(Cluster, nil, nil, nil, nil) }).NotTo(Panic())
				Expect(func() { ScopeInfoOf(Tenant, nil, nil, nil, nil) }).NotTo(Panic())
			})

			It("should leave the scope id empty when the id for the scope is unset", func() {
				Expect(*ScopeInfoOf(Project, nil, nil, nil, nil)).To(Equal(ScopeInfo{Scope: Project}))
				Expect(*ScopeInfoOf(Department, nil, nil, nil, nil)).To(Equal(ScopeInfo{Scope: Department}))
				Expect(*ScopeInfoOf(Cluster, nil, nil, nil, nil)).To(Equal(ScopeInfo{Scope: Cluster}))
				Expect(*ScopeInfoOf(Tenant, nil, nil, nil, nil)).To(Equal(ScopeInfo{Scope: Tenant}))
			})

			It("should ignore ids that do not belong to the scope", func() {
				projectId := int32(7)
				Expect(*ScopeInfoOf(Project, &projectId, nil, nil, nil)).To(Equal(ScopeInfo{
					Scope:   Project,
					ScopeId: "7",
				}))
			})
		})
	})

})
