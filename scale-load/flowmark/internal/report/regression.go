/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const (
	kpiMarker        = "p90_ms"
	sloPrefix        = "slo_"
	regressionSuffix = "_regression"
	AgainstMedian    = "median"
	AgainstSLO       = "slo"
)

type RegressionConfig struct {
	Window    int
	Threshold float64
}

type Regression struct {
	Section  string  `json:"section"`
	Metric   string  `json:"metric"`
	Value    float64 `json:"value"`
	Baseline float64 `json:"baseline"`
	Against  string  `json:"against"`
	Change   float64 `json:"change,omitempty"`
	Message  string  `json:"message"`
}

func DetectRegressions(document Document, history []Document, config RegressionConfig) Document {
	judged := document
	judged.Sections = make(map[string]core.Section, len(document.Sections))
	judged.Regressions = nil
	for name, section := range document.Sections {
		metrics := maps.Clone(section.Metrics)
		for _, metric := range kpis(section.Metrics) {
			value, _ := number(section.Metrics[metric])
			found := judgeAgainstMedian(name, metric, value, history, config)
			if slo, ok := number(section.Metrics[sloKey(metric)]); ok && value > slo {
				found = append(found, Regression{
					Section: name, Metric: metric, Value: value, Baseline: slo, Against: AgainstSLO,
					Message: fmt.Sprintf("%s %s: %s is over the SLO of %s", name, metric, milliseconds(value), milliseconds(slo)),
				})
			}
			if len(found) > 0 {
				metrics[metric+regressionSuffix] = true
				judged.Regressions = append(judged.Regressions, found...)
			}
		}
		section.Metrics = metrics
		judged.Sections[name] = section
	}
	slices.SortFunc(judged.Regressions, func(a, b Regression) int {
		return cmp.Or(cmp.Compare(a.Section, b.Section), cmp.Compare(a.Metric, b.Metric), cmp.Compare(a.Against, b.Against))
	})
	judged.RegressionDetected = len(judged.Regressions) > 0
	return judged
}

func judgeAgainstMedian(section, metric string, value float64, history []Document, config RegressionConfig) []Regression {
	if config.Window <= 0 {
		return nil
	}
	var past []float64
	for _, run := range history {
		if previous, ok := number(run.Sections[section].Metrics[metric]); ok {
			past = append(past, previous)
		}
	}
	if len(past) < config.Window {
		return nil
	}
	baseline := median(past[len(past)-config.Window:])
	if baseline <= 0 || value <= baseline*(1+config.Threshold) {
		return nil
	}
	change := value/baseline - 1
	return []Regression{{
		Section: section, Metric: metric, Value: value, Baseline: baseline, Against: AgainstMedian, Change: change,
		Message: fmt.Sprintf("%s %s: %s is %+.0f%% over the median of %s from the last %d runs", section, metric, milliseconds(value), change*100, milliseconds(baseline), config.Window),
	}}
}

func kpis(metrics map[string]any) []string {
	var names []string
	for name, value := range metrics {
		if _, ok := number(value); !ok || !strings.Contains(name, kpiMarker) || strings.HasPrefix(name, sloPrefix) || strings.HasSuffix(name, regressionSuffix) {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func sloKey(metric string) string {
	base, _, _ := strings.Cut(metric, "@")
	return sloPrefix + base
}

func median(values []float64) float64 {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func number(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func milliseconds(value float64) string {
	return fmt.Sprintf("%.0f ms", value)
}
