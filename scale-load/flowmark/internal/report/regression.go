/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const (
	HistoryDepth     = 10
	kpiMarker        = "p90_ms"
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
	Message  string  `json:"message"`
	Value    float64 `json:"-"`
	Baseline float64 `json:"-"`
	Against  string  `json:"-"`
	Change   float64 `json:"-"`
}

func Judge(ctx context.Context, store Store, document Document, config RegressionConfig) (Document, string) {
	history, err := LoadHistory(ctx, store, document.Metadata.ClusterType, config.Window*HistoryDepth)
	judged := DetectRegressions(document, history, config)
	note := "no regression detected"
	if judged.RegressionDetected {
		note = fmt.Sprintf("%d regressions detected", len(judged.Regressions))
	}
	switch {
	case err != nil:
		return judged, fmt.Sprintf("comparison with earlier runs skipped: %v; only SLOs were checked: %s", err, note)
	case document.Metadata.ClusterType == "":
		return judged, "comparison with earlier runs skipped: the cluster type is unknown, set --cluster-type or --cp-url; only SLOs were checked: " + note
	case config.Window == 0:
		return judged, "comparison with earlier runs disabled; only SLOs were checked: " + note
	case len(history) < config.Window:
		return judged, fmt.Sprintf("comparison with earlier runs skipped: %d of the %d required %s runs on record; only SLOs were checked: %s", len(history), config.Window, document.Metadata.ClusterType, note)
	default:
		return judged, fmt.Sprintf("compared every p90 with the median of its last %d values among %d %s runs, threshold %.0f%%: %s", config.Window, len(history), document.Metadata.ClusterType, config.Threshold*100, note)
	}
}

func DetectRegressions(document Document, history []Document, config RegressionConfig) Document {
	judged := document
	judged.Sections = make(map[string]core.Section, len(document.Sections))
	judged.Regressions = nil
	for name, section := range document.Sections {
		metrics := maps.Clone(section.Metrics)
		maps.DeleteFunc(metrics, func(key string, _ any) bool { return strings.HasSuffix(key, regressionSuffix) })
		found := map[string][]Regression{}
		for _, metric := range kpis(section.Metrics) {
			value, _ := number(section.Metrics[metric])
			found[metric] = append(found[metric], judgeAgainstMedian(name, metric, value, history, config)...)
		}
		for _, metric := range withSLO(section) {
			value, _ := number(section.Metrics[metric])
			if slo := float64(section.SLO[sloName(metric)]); value > slo {
				found[metric] = append(found[metric], Regression{
					Section: name, Metric: metric, Value: value, Baseline: slo, Against: AgainstSLO,
					Message: fmt.Sprintf("%s %s: %s is over the SLO of %s", name, metric, milliseconds(value), milliseconds(slo)),
				})
			}
		}
		for metric, regressions := range found {
			if len(regressions) > 0 {
				metrics[metric+regressionSuffix] = true
				judged.Regressions = append(judged.Regressions, regressions...)
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
		previousSection, ran := run.Sections[section]
		if !ran || previousSection.Failed {
			continue
		}
		if previous, ok := number(previousSection.Metrics[metric]); ok {
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
	return measured(metrics, func(name string) bool { return strings.Contains(name, kpiMarker) })
}

func withSLO(section core.Section) []string {
	return measured(section.Metrics, func(name string) bool { _, ok := section.SLO[sloName(name)]; return ok })
}

func measured(metrics map[string]any, keep func(name string) bool) []string {
	var names []string
	for name, value := range metrics {
		if _, ok := number(value); !ok || strings.HasSuffix(name, regressionSuffix) || !keep(name) {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func sloName(metric string) string {
	base, _, _ := strings.Cut(metric, "@")
	return base
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
