package network_test

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Network Package", func() {
	Describe("ValidAddress", func() {
		Context("when the address is valid", func() {
			It("should return nil", func() {
				validAddress := "localhost:8080"
				err := network.ValidAddress(validAddress)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the address is invalid", func() {
			It("should return an error for missing port", func() {
				invalidAddress := "localhost"
				err := network.ValidAddress(invalidAddress)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not in the format of host:port"))
			})

			It("should return an error for invalid port number", func() {
				invalidAddress := "localhost:invalid"
				err := network.ValidAddress(invalidAddress)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not a valid port number"))
			})
		})
	})

	Describe("AddressValidator", func() {
		var (
			validator *network.AddressValidator
		)

		Context("when the address is valid", func() {
			BeforeEach(func() {
				validAddress := "localhost:8080"
				validator = network.NewAddressValidator(validAddress)
			})

			It("should return an empty string and no error", func() {
				msg, err := validator.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(msg).To(BeEmpty())
			})
		})

		Context("when the address is invalid", func() {
			BeforeEach(func() {
				invalidAddress := "localhost:invalid"
				validator = network.NewAddressValidator(invalidAddress)
			})

			It("should return the fail message and an error", func() {
				msg, err := validator.Execute()
				Expect(err).To(HaveOccurred())
				Expect(msg).To(Equal(consts.InvalidAddress))
			})
		})
	})
})
