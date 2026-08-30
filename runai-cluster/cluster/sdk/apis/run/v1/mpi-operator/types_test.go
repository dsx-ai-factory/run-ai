package mpi_operator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("MpiOperator", func() {
	Context("Defaulting", func() {
		It("should set default monitor port number when not specified", func() {
			mpiOperator := &MpiOperator{}

			By("verifying MonitorPortNumber is nil before setting defaults")
			Expect(mpiOperator.MonitorPortNumber).To(BeNil())

			mpiOperator.SetDefaultsWhereNeeded()

			By("verifying MonitorPortNumber is set to default value 8081")
			Expect(mpiOperator.MonitorPortNumber).ToNot(BeNil())
			Expect(*mpiOperator.MonitorPortNumber).To(Equal(int32(8081)))
		})

		It("should not override existing monitor port number", func() {
			customPort := int32(9090)
			mpiOperator := &MpiOperator{
				MonitorPortNumber: ptr.To(customPort),
			}

			By("verifying MonitorPortNumber is set to custom value")
			Expect(mpiOperator.MonitorPortNumber).ToNot(BeNil())
			Expect(*mpiOperator.MonitorPortNumber).To(Equal(customPort))

			mpiOperator.SetDefaultsWhereNeeded()

			By("verifying MonitorPortNumber remains unchanged")
			Expect(mpiOperator.MonitorPortNumber).ToNot(BeNil())
			Expect(*mpiOperator.MonitorPortNumber).To(Equal(customPort))
		})

		It("should handle zero value for monitor port number", func() {
			zeroPort := int32(0)
			mpiOperator := &MpiOperator{
				MonitorPortNumber: ptr.To(zeroPort),
			}

			By("verifying MonitorPortNumber is set to zero")
			Expect(mpiOperator.MonitorPortNumber).ToNot(BeNil())
			Expect(*mpiOperator.MonitorPortNumber).To(Equal(zeroPort))

			mpiOperator.SetDefaultsWhereNeeded()

			By("verifying MonitorPortNumber remains zero (not replaced with default)")
			Expect(mpiOperator.MonitorPortNumber).ToNot(BeNil())
			Expect(*mpiOperator.MonitorPortNumber).To(Equal(zeroPort))
		})
	})
})
