package metrics

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	TimeZeroValueStr = "0001-01-01T00:00:00Z"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Metrics")
}

var _ = Describe("CLI Metrics Base", func() {
	var metrics Metrics

	BeforeEach(func() {
		metrics = *NewMetrics()
	})

	Context("Test metric cmd_timer", func() {
		It("Metrics default struct", func() {
			metricDefault := Metrics{}
			mar, err := json.Marshal(metricDefault) //nolint:staticcheck // unmarshall does create object
			Expect(err).To(BeNil())
			Expect(string(mar)).To(Equal("{}"))
		})

		It("Metrics new struct", func() {
			metrics.start = 1712088115
			mar, err := json.Marshal(&metrics)
			Expect(err).To(BeNil())
			Expect(string(mar)).To(ContainSubstring("cmd_start"))
			Expect(string(mar)).To(ContainSubstring("cmd_end"))
			Expect(string(mar)).To(ContainSubstring("timers"))
			Expect(string(mar)).To(ContainSubstring("counter"))
			Expect(string(mar)).To(ContainSubstring("\"cmd_start\":1712088115"))
			Expect(string(mar)).To(ContainSubstring("\"cmd_end\":0"))
		})
	})
})
