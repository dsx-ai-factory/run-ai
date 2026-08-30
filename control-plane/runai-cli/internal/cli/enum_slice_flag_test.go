package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
)

var _ = Describe("EnumSliceFlag", func() {
	var ef *cli.EnumSliceFlag
	var values []string

	BeforeEach(func() {
		values = []string{}
		ef = cli.NewEnumSliceFlag([]string{"value1", "value2", "value3"}, &values)
	})

	Describe("Type", func() {
		It("should return 'enumSlice'", func() {
			Expect(ef.Type()).To(Equal("enumSlice"))
		})
	})

	Describe("String", func() {
		It("should return empty string when not set", func() {
			Expect(ef.String()).To(Equal("[]"))
		})

		It("should return formatted string when values are set", func() {
			err := ef.Set("value1,value3")
			Expect(err).NotTo(HaveOccurred())
			Expect(ef.String()).To(Equal("[value1,value3]"))
		})
	})

	Describe("Set", func() {
		Context("with valid values", func() {
			It("should set a single valid value", func() {
				err := ef.Set("value1")
				Expect(err).NotTo(HaveOccurred())
				Expect(values).To(Equal([]string{"value1"}))
			})

			It("should set multiple valid values", func() {
				err := ef.Set("value2,value3")
				Expect(err).NotTo(HaveOccurred())
				Expect(values).To(Equal([]string{"value2", "value3"}))
			})
		})

		Context("with invalid values", func() {
			It("should fail on single invalid value", func() {
				err := ef.Set("invalid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must be partial or full match of"))
				Expect(values).To(Equal([]string{}))
			})

			It("should be case sensitive", func() {
				err := ef.Set("VALUE1")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must be partial or full match of"))
			})
		})
	})
})
