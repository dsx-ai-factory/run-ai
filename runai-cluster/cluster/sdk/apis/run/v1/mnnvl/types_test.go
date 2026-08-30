package mnnvl

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MNNVL", func() {
	DescribeTable("MNNVL feature label values are valid Modes",
		func(labelValue string) {
			mode, err := GetMode(labelValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(IsValidMode(mode)).To(BeTrue())
		},
		Entry("None", None),
		Entry("Preferred", Preferred),
		Entry("Required", Required),
	)
})
