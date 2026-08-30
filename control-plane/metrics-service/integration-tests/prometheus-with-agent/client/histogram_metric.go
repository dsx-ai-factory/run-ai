package client

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type HistogramMetric struct {
	name         string
	timestamp    time.Time
	labels       [][2]any
	bucketCounts []float64
	buckets      []float64
	sum          float64
	count        float64
}

func NewHistogramMetric(name string, buckets []float64) *HistogramMetric {
	buckets = append(buckets, math.Inf(1)) // Add "+Inf" bucket

	return &HistogramMetric{
		name:         name,
		timestamp:    time.Now(),
		labels:       make([][2]any, 0),
		buckets:      buckets,
		bucketCounts: make([]float64, len(buckets)),
	}
}

func (m *HistogramMetric) WithTime(t time.Time) *HistogramMetric {
	m.timestamp = t
	return m
}

func (h *HistogramMetric) AddLabel(label, val any) *HistogramMetric {
	h.labels = append(h.labels, [2]any{label, val})
	return h
}

func (h *HistogramMetric) Observe(val float64) *HistogramMetric {
	for i, le := range h.buckets {
		if val <= le {
			h.bucketCounts[i]++
		}
	}

	h.count += 1
	h.sum += val
	return h
}

func (h *HistogramMetric) String() string {
	var sb strings.Builder
	timestamp := h.timestamp.Unix()

	for i, le := range h.buckets {
		leStr := "le=\"+Inf\""
		if !math.IsInf(le, 1) {
			leStr = fmt.Sprintf("le=\"%g\"", le)
		}

		sb.WriteString(fmt.Sprintf(
			"%s_bucket{%s,%s} %f %d\n",
			h.name,
			joinLabels(h.labels),
			leStr,
			h.bucketCounts[i],
			timestamp,
		))
	}

	labelStr := getPromqlLabels(h.labels)
	sb.WriteString(fmt.Sprintf("%s_sum%s %f %d\n", h.name, labelStr, h.sum, timestamp))
	sb.WriteString(fmt.Sprintf("%s_count%s %f %d\n", h.name, labelStr, h.count, timestamp))

	return sb.String()
}
