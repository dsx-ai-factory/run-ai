package cluster_seeder

import (
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

type SamplerRunner interface {
	InitSampledEntities(clients *sampler_actions.Clients) ([]scale.Entity, error)
	UpdateRawResultReportMetaData(report *scale.SamplesReport)
	GetMetricLabels() map[string]string
}
