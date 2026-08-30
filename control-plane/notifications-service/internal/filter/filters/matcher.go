package filters

import (
	"context"
	"regexp"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

type MatchType string

const (
	MtEqual           MatchType = "=="
	MtRegex           MatchType = "=~"
	MatcherFilterName           = "matcher"
)

type PropertyMatcher struct {
	MatchType MatchType `mapstructure:"match_type" json:"match_type"`
	Key       string    `mapstructure:"key" json:"key"`
	Value     string    `mapstructure:"value" json:"value"`
}

type MatcherFilterConfig []*PropertyMatcher

type MatcherFilter struct {
	config *MatcherFilterConfig
}

func NewMatcherFilter(config *MatcherFilterConfig) *MatcherFilter {
	return &MatcherFilter{
		config: config,
	}
}

func (mf MatcherFilter) Filter(ctx context.Context, event *events.Event) (bool, error) {
	for _, matcher := range *mf.config {
		val, ok := event.Properties[matcher.Key]
		if !ok {
			return true, nil
		}
		if !matching(val, matcher) {
			return true, nil
		}
	}
	return false, nil
}

func matching(val interface{}, matcher *PropertyMatcher) bool {
	sval, ok := val.(string)
	if !ok {
		return true
	}
	switch matcher.MatchType {
	case MtEqual:
		return sval == matcher.Value
	case MtRegex:
		match, err := regexp.MatchString(matcher.Value, sval)
		return match && err == nil
	}
	return false
}
