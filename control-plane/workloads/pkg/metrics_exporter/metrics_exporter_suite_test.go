package metrics_exporter

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMetricsExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics exporter suite")
}
