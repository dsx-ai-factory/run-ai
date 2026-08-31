package dao

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"k8s.io/utils/ptr"
)

var _ = Describe("DAO utils", func() {
	Describe("IsEligibleForCachedData", func() {
		It("succeeds", func() {
			Expect(IsEligibleForCachedData(nil, testTenantID, "cluster-1")).To(BeTrue())
			Expect(IsEligibleForCachedData(&openapi.PermittedScopes{System: ptr.To(true)}, testTenantID, "cluster-1")).To(BeTrue())
			Expect(IsEligibleForCachedData(&openapi.PermittedScopes{Tenants: []string{testTenantID}}, testTenantID, "cluster-1")).To(BeTrue())
			Expect(IsEligibleForCachedData(&openapi.PermittedScopes{Clusters: []string{"cluster-1"}}, testTenantID, "cluster-1")).To(BeTrue())
			Expect(IsEligibleForCachedData(&openapi.PermittedScopes{}, testTenantID, "cluster-1")).To(BeFalse())
			Expect(IsEligibleForCachedData(&openapi.PermittedScopes{System: ptr.To(false)}, testTenantID, "cluster-1")).To(BeFalse())
		})
	})
	Describe("GetCacheKey", func() {
		It("succeeds", func() {
			Expect(GetCacheKey("departments", testTenantID, "cluster-1")).To(Equal("departments-tenant-1-cluster-1"))
		})
	})
	Describe("IsTenantPermitted", func() {
		It("succeeds", func() {
			Expect(IsTenantPermitted(testTenantID, &openapi.PermittedScopes{Tenants: []string{testTenantID}})).To(BeTrue())
			Expect(IsTenantPermitted(testTenantID, &openapi.PermittedScopes{Tenants: []string{"other"}})).To(BeFalse())
		})
	})
	Describe("IsClusterPermitted", func() {
		It("succeeds", func() {
			Expect(IsClusterPermitted("cluster-1", &openapi.PermittedScopes{Clusters: []string{"cluster-1"}})).To(BeTrue())
			Expect(IsClusterPermitted("cluster-1", &openapi.PermittedScopes{Clusters: []string{"other"}})).To(BeFalse())
		})
	})
	Describe("GetPermittedScopesForProjectsWithColumn", func() {
		It("succeeds", func() {
			got := GetPermittedScopesForProjectsWithColumn(openapi.PermittedScopes{
				Clusters:    []string{"c1"},
				Departments: []string{"11"},
				Projects:    []string{"22"},
			}, "id")
			Expect(got).To(ContainSubstring("cluster_id IN"))
			Expect(got).To(ContainSubstring("parent_id IN"))
			Expect(got).To(ContainSubstring("id IN"))
			Expect(got).To(ContainSubstring("'c1'"))
		})
	})
})
