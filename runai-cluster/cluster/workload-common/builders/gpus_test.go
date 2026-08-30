package builders

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPU tests", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("when the GPU value is a whole number without a fractional part", func() {
		It("returns false", func() {
			Expect(IsWholeWithFractionalGpu(ctx, 2.0)).To(BeFalse())
		})
	})

	Context("when the GPU value is a whole number with a fractional part", func() {
		It("returns true", func() {
			Expect(IsWholeWithFractionalGpu(ctx, 2.5)).To(BeTrue())
		})
	})

	Context("when the GPU value is less than one", func() {
		It("returns false", func() {
			Expect(IsWholeWithFractionalGpu(ctx, 0.5)).To(BeFalse())
		})
	})

	Context("when the GPU value is equal to one", func() {
		It("returns false", func() {
			Expect(IsWholeWithFractionalGpu(ctx, 1.0)).To(BeFalse())
		})
	})

	Context("when the GPU value is greater than one with a fractional part", func() {
		It("returns true", func() {
			Expect(IsWholeWithFractionalGpu(ctx, 1.5)).To(BeTrue())
		})
	})
})
