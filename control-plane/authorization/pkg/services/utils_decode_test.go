package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Group name decoding", func() {
	Context("decodeGroups", func() {
		It("decodes an escaped comma back into a literal comma", func() {
			Expect(decodeGroups([]string{"x%2Crunai-admins", "eng", "ops"})).To(Equal([]string{"x,runai-admins", "eng", "ops"}))
		})

		It("leaves a malformed percent-sequence unchanged", func() {
			Expect(decodeGroups([]string{"100% done", "x%2Crunai-admins"})).To(Equal([]string{"100% done", "x,runai-admins"}))
		})
	})

	Context("parseAuthData", func() {
		It("decodes escaped group names read from context", func() {
			ctx := setContext("42", "alice@run.ai", []string{"x%2Crunai-admins"})

			tenantId, subjectId, groups, apiErr := parseAuthData(ctx)

			Expect(apiErr).To(BeNil())
			Expect(tenantId).To(Equal(uint(42)))
			Expect(subjectId).To(Equal("alice@run.ai"))
			Expect(groups).To(Equal([]string{"x,runai-admins"}))
		})
	})
})
