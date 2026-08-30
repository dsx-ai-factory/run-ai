package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

var _ = Describe("validate assets comparison", func() {
	Context("no value true default", func() {
		It("should return true", func() {
			result := ApplyBoolDefault(nil, true)
			Expect(*result).To(Equal(true))
		})
	})

	Context("no value false default", func() {
		It("should return false", func() {
			result := ApplyBoolDefault(nil, false)
			Expect(*result).To(Equal(false))
		})
	})

	Context("true value false default", func() {
		It("should return true", func() {
			result := ApplyBoolDefault(convutils.Ref(true), false)
			Expect(*result).To(Equal(true))
		})
	})

	Context("false value true default", func() {
		It("should return false", func() {
			result := ApplyBoolDefault(convutils.Ref(false), true)
			Expect(*result).To(Equal(false))
		})
	})
})
