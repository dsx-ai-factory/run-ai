package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("split and join lines", func() {
	Context("split empty line", func() {
		It("should return 1 item array", func() {
			result := SplitLines("")
			Expect(len(result)).To(Equal(1))
		})
	})

	Context("split single line", func() {
		It("should return 1 item array", func() {
			result := SplitLines("xxx")
			Expect(len(result)).To(Equal(1))
		})
	})

	Context("split multiple lines", func() {
		It("should return correct array", func() {
			result := SplitLines("\nxxx\n\n233\n")
			Expect(len(result)).To(Equal(5))
		})
	})

	Context("join empty array", func() {
		It("should return empty line", func() {
			result := JoinLines([]string{})
			Expect(result).To(Equal(""))
		})
	})

	Context("join array", func() {
		It("should return correct line", func() {
			result := JoinLines([]string{"", "aa", "", "bb", ""})
			Expect(result).To(Equal("\naa\n\nbb\n"))
		})
	})
})
