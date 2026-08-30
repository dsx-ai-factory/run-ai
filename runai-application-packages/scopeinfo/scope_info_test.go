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
	})

})
