package policy_compliance

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type ConvEntry struct {
	origValue string
	factor    float64
	expected  string
}

func ConvTest(orig string, factor float64, expected string) TableEntry {
	return Entry(fmt.Sprintf("%s x %v = %s", orig, factor, expected), ConvEntry{
		origValue: orig,
		factor:    factor,
		expected:  expected,
	})
}

var ConvTests = []TableEntry{
	ConvTest("1G", 2, "2G"),
	ConvTest("2G", 2000, "4T"),
	ConvTest("1Mi", 2, "2Mi"),
	ConvTest("2Gi", 2000, "4000Gi"),
	ConvTest("2M", 0.5, "1M"),
	ConvTest("1M", 0.5, "500k"),
	ConvTest("1M", 0.1, "100k"),
	ConvTest("1k", 0.1, "100"),
	ConvTest("1G", 0.5, "500M"),
}

var _ = Describe("Policy compliance validations", func() {

	BeforeEach(func() {
	})

	Context("input VS expected tests", func() {
		DescribeTable("should report compliance errors correctly", func(testDetails ComplianceTest) {
			RunComplianceTest(testDetails, ".")
		}, ComplianceTests)
	})

	Context("conversion of quantities", func() {
		DescribeTable("should return valid values", func(test ConvEntry) {
			result, err := MultiplyResourceQuantity(test.origValue, test.factor)
			Expect(err).To(BeNil())
			Expect(result).To(Equal(test.expected))
		}, ConvTests)
	})

})
