package authz_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz"
)

var _ = Describe("UnauthorizedError", func() {
	var (
		err     *authz.UnauthorizedError
		subject string
	)

	BeforeEach(func() {
		subject = "test-subject"
		err = authz.NewUnauthorizedError(subject)
	})

	It("should return the correct error message", func() {
		Expect(err.Error()).To(Equal("test-subject"))
	})

	Context("when creating a new UnauthorizedError", func() {
		It("should return a new instance with the correct subject", func() {
			Expect(err).To(BeAssignableToTypeOf(&authz.UnauthorizedError{}))
			Expect(err.Error()).To(ContainSubstring(subject))
		})
	})
	Context("isUnauthorizedError", func() {
		Context("when the error is an UnauthorizedError", func() {
			It("should return true", func() {
				Expect(authz.IsUnauthorizedError(err)).To(BeTrue())
			})
		})
		Context("when the error is nil", func() {
			It("should return false", func() {
				Expect(authz.IsUnauthorizedError(nil)).To(BeFalse())
			})
		})
		Context("when the error is not an UnauthorizedError", func() {
			It("should return false", func() {
				Expect(authz.IsUnauthorizedError(errors.New("some error"))).To(BeFalse())
			})
		})
	})
})
