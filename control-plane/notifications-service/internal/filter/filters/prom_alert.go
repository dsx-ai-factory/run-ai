package filters

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

const (
	PrometheusAlertsFilterName = "prometheus_alerts"
)

type PrometheusAlertsFilterConfig MatcherFilterConfig

type PrometheusAlertsFilter struct {
	*MatcherFilter
}

func NewPrometheusAlertsFilter(config *PrometheusAlertsFilterConfig) *PrometheusAlertsFilter {
	mc := MatcherFilterConfig(*config)
	return &PrometheusAlertsFilter{
		MatcherFilter: &MatcherFilter{
			config: &mc,
		},
	}
}

func (pmf PrometheusAlertsFilter) Filter(ctx context.Context, event *events.Event) (bool, error) {
	return pmf.MatcherFilter.Filter(ctx, event)
}
