package encryption

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Keys Test", func() {
	Describe("notFound function", func() {
		Context("no error", func() {
			It("return false", func() {
				result := notFound(nil)
				Expect(result).To(Equal(false))
			})
		})

		Context("unrelated error", func() {
			It("return false", func() {
				result := notFound(errors.New("just an error"))
				Expect(result).To(Equal(false))
			})
		})
	})
})
