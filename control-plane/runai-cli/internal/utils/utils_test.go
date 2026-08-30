package utils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Helper Utils Tests")
}

var _ = Describe("IntSliceToString", func() {
	Context("when given a slice of integers", func() {
		It("should convert it to a comma-separated string", func() {
			input := []int{1, 2, 3, 4, 5}
			expected := "1,2,3,4,5"

			result := IntSliceToString(input)

			Expect(result).To(Equal(expected))
		})

		It("should handle an empty slice", func() {
			var input []int
			expected := ""

			result := IntSliceToString(input)

			Expect(result).To(Equal(expected))
		})

		It("should handle a slice with one element", func() {
			input := []int{42}
			expected := "42"

			result := IntSliceToString(input)

			Expect(result).To(Equal(expected))
		})

		It("should handle negative numbers", func() {
			input := []int{-1, 0, 1}
			expected := "-1,0,1"

			result := IntSliceToString(input)

			Expect(result).To(Equal(expected))
		})

		It("should handle large numbers", func() {
			input := []int{1000000, 2000000, 3000000}
			expected := "1000000,2000000,3000000"

			result := IntSliceToString(input)

			Expect(result).To(Equal(expected))
		})
	})
})
