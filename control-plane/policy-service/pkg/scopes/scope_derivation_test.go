/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("scopes", func() {
	contextOfTenant1 := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	contextOfTenant2 := context.WithValue(context.Background(), runaictx.CfTenantId, "2")

	BeforeEach(func() {
		//
		// create a rich hierarchy
		// tenants: 1 2, departments: 11, 12, 21, 22, projects: 111, 112, ..., 221, 222
		//
		orgUnitMock := NewOrgUnitBasedBackendMockClient()
		for _, tenant := range []string{"1", "2"} {
			tenantOrgUnit := OrgUnit{Scope: api.SCOPETYPE_TENANT, Id: tenant}

			clusterOrgUnit := OrgUnit{Scope: api.SCOPETYPE_CLUSTER, Id: fmt.Sprintf("cluster-of-tenant-%s", tenant)}
			orgUnitMock.Add(clusterOrgUnit, tenantOrgUnit)

			for _, department := range []string{"1", "2"} {
				deptId := fmt.Sprintf("%s%s", tenant, department)
				departmentOrgUnit := OrgUnit{Scope: api.SCOPETYPE_DEPARTMENT, Id: deptId}
				orgUnitMock.Add(departmentOrgUnit, clusterOrgUnit)

				for _, project := range []string{"1", "2"} {
					projectId := fmt.Sprintf("%s%s", deptId, project)
					projectOrgUnit := OrgUnit{Scope: api.SCOPETYPE_PROJECT, Id: projectId}
					orgUnitMock.Add(projectOrgUnit, departmentOrgUnit)
				}
			}
		}

		orgunit.SetMock(orgUnitMock)
	})

	Context("malformed unit key string", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromString(convutils.Ref("xxx"))
			Expect(err.Message).To(Equal("malformed org unit key xxx, expecting scope:id"))
		})
	})

	Context("unsupported scope", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromString(convutils.Ref("xxx:123"))
			Expect(err.Message).To(Equal("unsupported scope xxx in xxx:123"))
		})
	})

	Context("invalid project id", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromString(convutils.Ref("project:xxx"))
			Expect(err.Message).To(ContainSubstring("is not numeric"))
		})
	})

	Context("invalid project id for json", func() {
		It("should return error", func() {
			orgUnit := OrgUnit{
				Scope: api.SCOPETYPE_PROJECT,
				Id:    "xyx",
			}
			_, err := orgUnit.Json()
			Expect(err.Message).To(Equal("non numeric project id xyx"))
		})
	})

	Context("nil org unit", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromString(nil)
			Expect(err.Message).To(Equal("nil org unit key receives"))
		})
	})

	Context("parents of project", func() {
		It("should return the correct ancestors", func() {
			orgUnit := OrgUnit{
				Scope: api.SCOPETYPE_PROJECT,
				Id:    "212",
			}

			result, err := orgUnit.Ancestors(contextOfTenant2)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]OrgUnit{
				{
					Scope: api.SCOPETYPE_PROJECT,
					Id:    "212",
				},
				{
					Scope: api.SCOPETYPE_DEPARTMENT,
					Id:    "21",
				},
				{
					Scope: api.SCOPETYPE_CLUSTER,
					Id:    "cluster-of-tenant-2",
				},
				{
					Scope: api.SCOPETYPE_TENANT,
					Id:    "2",
				},
				{
					Scope: api.SCOPETYPE_SYSTEM,
				},
			}))
		})
	})

	Context("parents of department", func() {
		It("should return the correct ancestors", func() {
			orgUnit := OrgUnit{
				Scope: api.SCOPETYPE_DEPARTMENT,
				Id:    "12",
			}

			result, err := orgUnit.Ancestors(contextOfTenant1)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]OrgUnit{
				{
					Scope: api.SCOPETYPE_DEPARTMENT,
					Id:    "12",
				},
				{
					Scope: api.SCOPETYPE_CLUSTER,
					Id:    "cluster-of-tenant-1",
				},
				{
					Scope: api.SCOPETYPE_TENANT,
					Id:    "1",
				},
				{
					Scope: api.SCOPETYPE_SYSTEM,
				},
			}))
		})
	})

	Context("parents of tenant", func() {
		It("should return the correct ancestors", func() {
			orgUnit := OrgUnit{
				Scope: api.SCOPETYPE_TENANT,
				Id:    "2",
			}

			result, err := orgUnit.Ancestors(contextOfTenant2)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]OrgUnit{
				{
					Scope: api.SCOPETYPE_TENANT,
					Id:    "2",
				},
				{
					Scope: api.SCOPETYPE_SYSTEM,
				},
			}))
		})
	})

	Context("parents of cluster", func() {
		It("should return the correct ancestors", func() {
			orgUnit := OrgUnit{
				Scope: api.SCOPETYPE_CLUSTER,
				Id:    "cluster-of-tenant-2",
			}

			result, err := orgUnit.Ancestors(contextOfTenant2)
			Expect(err).To(BeNil())
			Expect(result).To(Equal([]OrgUnit{
				{
					Scope: api.SCOPETYPE_CLUSTER,
					Id:    "cluster-of-tenant-2",
				},
				{
					Scope: api.SCOPETYPE_TENANT,
					Id:    "2",
				},
				{
					Scope: api.SCOPETYPE_SYSTEM,
				},
			}))
		})
	})

	Context("parents of non existent unit", func() {
		It("should return error", func() {
			orgUnit := OrgUnit{
				Scope: api.SCOPETYPE_DEPARTMENT,
				Id:    "18",
			}

			_, err := orgUnit.Ancestors(contextOfTenant1)
			Expect(int(err.Code)).To(Equal(int(http.StatusNotFound)))
		})
	})

	Context("department scope with no department id", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromCreationFields(&api.PolicyCreationFields{
				Scope: api.SCOPETYPE_DEPARTMENT,
			}, 1)
			Expect(err.Message).To(Equal("departmentId is mandatory for department scoped policy"))
		})
	})

	Context("project scope with no project id", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromCreationFields(&api.PolicyCreationFields{
				Scope: api.SCOPETYPE_PROJECT,
			}, 1)
			Expect(err.Message).To(Equal("projectId is mandatory for project scoped policy"))
		})
	})

	Context("unsupported scope", func() {
		It("should return error", func() {
			_, err := NewOrgUnitFromCreationFields(&api.PolicyCreationFields{
				Scope: api.ScopeType("xyz"),
			}, 1)
			Expect(err.Message).To(Equal("unsupported scope 'xyz'"))
		})
	})
})
