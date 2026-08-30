package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("NullableFloat64", func() {
	var nf *cli.NullableFloat64Flag

	BeforeEach(func() {
		nf = cli.NewNullableFloat64Flag(&openapi.NullableFloat64{})
	})

	Describe("Set", func() {
		It("should set a valid float value", func() {
			err := nf.Set("3.14159265359")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.IsSet()).To(BeTrue())
			Expect(*nf.Value()).To(Equal(float64(3.14159265359)))
		})

		It("should set zero value", func() {
			err := nf.Set("0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.IsSet()).To(BeTrue())
			Expect(*nf.Value()).To(Equal(float64(0)))
		})

		It("should set negative value", func() {
			err := nf.Set("-2.5")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.IsSet()).To(BeTrue())
			Expect(*nf.Value()).To(Equal(float64(-2.5)))
		})

		It("should fail on invalid input", func() {
			err := nf.Set("invalid")
			Expect(err).To(HaveOccurred())
			Expect(nf.IsSet()).To(BeFalse())
		})

		It("should fail on empty string", func() {
			err := nf.Set("")
			Expect(err).To(HaveOccurred())
			Expect(nf.IsSet()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("should return an empty string when not set", func() {
			Expect(nf.String()).To(Equal(""))
		})

		It("should return the float value as a string when set", func() {
			err := nf.Set("4.56")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.String()).To(Equal("4.56"))
		})

		It("should return integer values without decimal", func() {
			err := nf.Set("5.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.String()).To(Equal("5"))
		})

		It("should handle scientific notation", func() {
			err := nf.Set("1.23e-4")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.IsSet()).To(BeTrue())
			Expect(*nf.Value()).To(Equal(float64(0.000123)))
		})
	})

	Describe("IsSet", func() {
		It("should return false when value is not set", func() {
			Expect(nf.IsSet()).To(BeFalse())
		})

		It("should return true when value is set", func() {
			err := nf.Set("2.718281828")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.IsSet()).To(BeTrue())
		})
	})

	Describe("Value", func() {
		It("should return nil when not set", func() {
			Expect(nf.Value()).To(BeNil())
		})

		It("should return the correct value when set", func() {
			err := nf.Set("1.414213562")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.Value()).NotTo(BeNil())
			Expect(*nf.Value()).To(Equal(float64(1.414213562)))
		})
	})

	Describe("Type", func() {
		It("should return 'float64'", func() {
			Expect(nf.Type()).To(Equal("float64"))
		})
	})
})
