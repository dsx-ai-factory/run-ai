package report_generator

import (
	"github.com/run-ai/runai/scale-tests/sdk/reports"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

type Generator interface {
	GenerateReport(testResults *scale.SamplesReport) (*reports.SuiteReport, error)
}
