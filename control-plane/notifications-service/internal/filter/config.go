package filter

import "github.com/run-ai/runai/control-plane/notifications-service/internal/filter/filters"

const (
	PolicyMatchAll = "all"
	PolicyMatchAny = "any"
	PolicyDrop     = "drop"
)

type FilteringConfigProvider interface {
	GetGlobalFilteringConfig() *FilteringConfig
}

type FilterConfig struct {
	Type               string                                `mapstructure:"type" json:"type"`
	PrometheusMatchers *filters.PrometheusAlertsFilterConfig `mapstructure:"prometheus_matchers" json:"prometheus_matchers"`
	Matcher            *filters.MatcherFilterConfig          `mapstructure:"matchers" json:"matchers"`
}

type FilteringConfig struct {
	Policy  string          `mapstructure:"policy,omitempty" json:"policy,omitempty"`
	Filters []*FilterConfig `mapstructure:"filters" json:"filters"`
}
