//go:build manual
// +build manual

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPostMetrics is a manual test, since it is
func TestPostMetrics(t *testing.T) {
	b := NewMetricCollection().
		AddMetric(
			NewMetric("my_metric", 10).AddLabel("code", "500"))

	err := PostMetrics(b.ToOpenMetrics())
	assert.NoError(t, err)
}
