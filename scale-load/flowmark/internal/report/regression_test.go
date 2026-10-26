/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const burstSection = "training-burst"

var sevenNights = RegressionConfig{Window: 7, Threshold: 0.15}

func nights(from, to int, clusterType string, p90 int) []Document {
	var documents []Document
	for night := from; night <= to; night++ {
		documents = append(documents, nightlyDocument(night, clusterType, p90))
	}
	return documents
}

func flagged(t *testing.T, judged Document, metric string) bool {
	t.Helper()
	flag, present := judged.Sections[burstSection].Metrics[metric+regressionSuffix]
	if present && flag != true {
		t.Fatalf("a regression flag is either absent or true, got %v", flag)
	}
	return present
}

func TestRelativeCheckFlagsEighteenPercentAndStaysQuietAtTen(t *testing.T) {
	history := nights(1, 7, kindCluster, 5000)

	slow := DetectRegressions(nightlyDocument(8, kindCluster, 5900), history, sevenNights)
	if !slow.RegressionDetected || !flagged(t, slow, "p90_ms@100") || len(slow.Regressions) != 1 {
		t.Fatalf("+18%% over a median of 5000 must be flagged once: %+v", slow.Regressions)
	}
	culprit := slow.Regressions[0]
	if culprit.Against != AgainstMedian || culprit.Baseline != 5000 || culprit.Value != 5900 || culprit.Section != burstSection || culprit.Metric != "p90_ms@100" {
		t.Fatalf("the culprit must name section, metric, value and the median: %+v", culprit)
	}
	for _, want := range []string{"training-burst p90_ms@100", "5900 ms", "+18%", "median of 5000 ms", "last 7 runs"} {
		if !strings.Contains(culprit.Message, want) {
			t.Fatalf("the Slack line must contain %q, got %q", want, culprit.Message)
		}
	}

	fine := DetectRegressions(nightlyDocument(8, kindCluster, 5500), history, sevenNights)
	if fine.RegressionDetected || flagged(t, fine, "p90_ms@100") || fine.Regressions != nil {
		t.Fatalf("+10%% is inside the 15%% threshold and must stay quiet: %+v", fine.Regressions)
	}
}

func TestSLOBreachIsFlaggedWithoutAnyHistory(t *testing.T) {
	judged := DetectRegressions(nightlyDocument(1, kindCluster, 12000), nil, sevenNights)

	if !judged.RegressionDetected || !flagged(t, judged, "p90_ms@100") || len(judged.Regressions) != 1 {
		t.Fatalf("12000 over an SLO of 10000 must be flagged with no history at all: %+v", judged.Regressions)
	}
	if r := judged.Regressions[0]; r.Against != AgainstSLO || r.Baseline != 10000 || !strings.Contains(r.Message, "over the SLO of 10000 ms") {
		t.Fatalf("the culprit must blame the SLO: %+v", r)
	}
	if judged.Sections[burstSection].SLO["p90_ms"] != 10000 {
		t.Fatal("the slo block must survive judging untouched")
	}
}

func TestRelativeCheckIsSkippedUntilTheWindowIsFull(t *testing.T) {
	judged := DetectRegressions(nightlyDocument(7, kindCluster, 5900), nights(1, 6, kindCluster, 5000), sevenNights)

	if judged.RegressionDetected || flagged(t, judged, "p90_ms@100") {
		t.Fatalf("6 runs of history are not enough for a window of 7, the check must stay silent: %+v", judged.Regressions)
	}
}

func TestMedianUsesTheLastWindowValuesOfTheMetricNotTheLastWindowRuns(t *testing.T) {
	history := nights(1, 7, kindCluster, 5000)
	withoutTheScenario := Build(map[string]core.Section{"workspace-ttj": {Category: core.CategoryLoad, Metrics: map[string]any{"ttj_p90_ms": 1}}}, RunInfo{ClusterType: kindCluster})
	history = append(history, withoutTheScenario)

	judged := DetectRegressions(nightlyDocument(9, kindCluster, 5900), history, sevenNights)

	if !judged.RegressionDetected {
		t.Fatalf("a night where the scenario did not run must not break the 7-value window: %+v", judged.Regressions)
	}
}

func TestMedianIgnoresRunsWhoseSectionFailed(t *testing.T) {
	history := nights(1, 7, kindCluster, 5000)
	broken := nightlyDocument(8, kindCluster, 50000)
	section := broken.Sections[burstSection]
	section.Failed = true
	section.Error = "20 workloads never reached the cluster"
	broken.Sections[burstSection] = section
	history = append(history, broken)

	judged := DetectRegressions(nightlyDocument(9, kindCluster, 5900), history, sevenNights)

	if !judged.RegressionDetected || judged.Regressions[0].Baseline != 5000 {
		t.Fatalf("a failed night must not enter the baseline, want median 5000 and a flag: %+v", judged.Regressions)
	}
}

func TestBothChecksCanFlagTheSameMetric(t *testing.T) {
	judged := DetectRegressions(nightlyDocument(8, kindCluster, 12000), nights(1, 7, kindCluster, 5000), sevenNights)

	if len(judged.Regressions) != 2 || judged.Regressions[0].Against != AgainstMedian || judged.Regressions[1].Against != AgainstSLO {
		t.Fatalf("a value over both the median and the SLO yields one culprit per check, median first: %+v", judged.Regressions)
	}
	if judged.Sections[burstSection].Metrics["p90_ms@100_regression"] != true {
		t.Fatal("the metric is flagged once")
	}
}

func TestOtherClusterTypesNeverEnterTheWindow(t *testing.T) {
	store := newFakeStore()
	publishAll(t, store, nights(1, 7, vanillaCluster, 1000)...)
	publishAll(t, store, nights(8, 13, kindCluster, 5000)...)

	history, _, err := LoadHistory(context.Background(), store, kindCluster, 7)
	if err != nil {
		t.Fatal(err)
	}
	judged := DetectRegressions(nightlyDocument(14, kindCluster, 5900), history, sevenNights)

	if len(history) != 6 || judged.RegressionDetected {
		t.Fatalf("7 fast vanilla runs must not complete the kind window, got %d runs and detected=%v", len(history), judged.RegressionDetected)
	}
}

func TestDayEightWalkthroughFlagsFourNightsThenTheMedianAdapts(t *testing.T) {
	store := newFakeStore()
	publishAll(t, store, nights(1, 7, kindCluster, 5000)...)

	var flaggedNights []int
	for night := 8; night <= 14; night++ {
		history, _, err := LoadHistory(context.Background(), store, kindCluster, sevenNights.Window)
		if err != nil {
			t.Fatal(err)
		}
		judged := DetectRegressions(nightlyDocument(night, kindCluster, 7000), history, sevenNights)
		if judged.RegressionDetected {
			flaggedNights = append(flaggedNights, night)
		}
		var window []float64
		for _, run := range history {
			value, _ := number(run.Sections[burstSection].Metrics["p90_ms@100"])
			window = append(window, value)
		}
		t.Logf("night %2d: tonight 7000, window %v, median %.0f, flagged=%v", night, window, median(window), judged.RegressionDetected)
		publishAll(t, store, judged)
	}

	if want := []int{8, 9, 10, 11}; !slices.Equal(flaggedNights, want) {
		t.Fatalf("a lasting +40%% must be flagged while the slow nights are the minority of the window, want %v got %v", want, flaggedNights)
	}
}

func TestSLOsApplyToAnyLatencyMetricWhileTheMedianStaysOnP90(t *testing.T) {
	section := core.Section{Category: core.CategoryScale,
		Metrics: map[string]any{"p90_ms@100": 100, "p99_ms@100": 20000, "max_ms@100": 70, "p50_ms@100": 40},
		SLO:     map[string]int{"p90_ms": 1000, "p99_ms": 15000, "max_ms": 50},
	}
	document := Build(map[string]core.Section{burstSection: section}, RunInfo{ClusterType: kindCluster})

	judged := DetectRegressions(document, nights(1, 7, kindCluster, 100), sevenNights)

	if len(judged.Regressions) != 2 || judged.Regressions[0].Metric != "max_ms@100" || judged.Regressions[1].Metric != "p99_ms@100" {
		t.Fatalf("p99 and max are over their SLOs, p90 is not, p50 has none: %+v", judged.Regressions)
	}
	for _, r := range judged.Regressions {
		if r.Against != AgainstSLO {
			t.Fatalf("only the SLO check applies to non-p90 metrics: %+v", r)
		}
	}
	if !flagged(t, judged, "p99_ms@100") || !flagged(t, judged, "max_ms@100") || flagged(t, judged, "p90_ms@100") || flagged(t, judged, "p50_ms@100") {
		t.Fatalf("flags must sit exactly next to the breached metrics: %v", judged.Sections[burstSection].Metrics)
	}
}

func TestARoundSLOOverridesTheBaseSLO(t *testing.T) {
	section := core.Section{Category: core.CategoryScale,
		Metrics: map[string]any{"p90_ms@100": 7000, "p90_ms@500": 50000, "max_ms@500": 70000},
		SLO:     map[string]int{"p90_ms": 5000, "p90_ms@500": 60000, "max_ms": 10000},
	}
	judged := DetectRegressions(Build(map[string]core.Section{burstSection: section}, RunInfo{ClusterType: kindCluster}), nil, sevenNights)

	if len(judged.Regressions) != 2 || judged.Regressions[0].Metric != "max_ms@500" || judged.Regressions[1].Metric != "p90_ms@100" {
		t.Fatalf("round 100 breaks the base p90, round 500 is under its own p90 limit but over the base max: %+v", judged.Regressions)
	}
	if judged.Regressions[0].Baseline != 10000 || judged.Regressions[1].Baseline != 5000 {
		t.Fatalf("each breach must be judged against the limit that applied: %+v", judged.Regressions)
	}
}

func TestJudgingAgainClearsStaleFlags(t *testing.T) {
	flagged := DetectRegressions(nightlyDocument(1, kindCluster, 12000), nil, sevenNights)
	if flagged.Sections[burstSection].Metrics["p90_ms@100_regression"] != true {
		t.Fatal("precondition: the first judgement flags the SLO breach")
	}
	flagged.Sections[burstSection].Metrics["p90_ms@100"] = 4000

	rejudged := DetectRegressions(flagged, nil, sevenNights)

	if rejudged.RegressionDetected || len(rejudged.Regressions) != 0 {
		t.Fatalf("4000 is under the SLO, nothing may be reported: %+v", rejudged.Regressions)
	}
	if _, stale := rejudged.Sections[burstSection].Metrics["p90_ms@100_regression"]; stale {
		t.Fatal("a flag from an earlier judgement must not survive a clean check")
	}
}

func TestKPIsAreTheP90MetricsOnly(t *testing.T) {
	metrics := map[string]any{
		"p90_ms@100": 1, "ttj_p90_ms": 2, "cluster_synced_p90_ms": 3,
		"p50_ms@100": 4, "p90_ms@100_regression": true, "flow": "x", "max_ms": 6,
	}
	if got := kpis(metrics); strings.Join(got, ",") != "cluster_synced_p90_ms,p90_ms@100,ttj_p90_ms" {
		t.Fatalf("only p90 metrics are KPIs, sorted, got %v", got)
	}
	section := core.Section{Metrics: metrics, SLO: map[string]int{"p90_ms": 5}}
	if got := withSLO(section); strings.Join(got, ",") != "p90_ms@100" {
		t.Fatalf("only metrics with a limit in the slo block are checked against an SLO, got %v", got)
	}
	if sloName("p90_ms@250") != "p90_ms" || sloName("ttj_p90_ms") != "ttj_p90_ms" {
		t.Fatal("the SLO of a round metric is the SLO of its base name")
	}
	if median([]float64{5, 1, 3}) != 3 || median([]float64{4, 1, 3, 2}) != 2.5 {
		t.Fatal("median must sort and average the two middle values for even counts")
	}
}

func TestJudgeExplainsWhatItCompared(t *testing.T) {
	ctx := context.Background()
	empty := newFakeStore()
	tonight := nightlyDocument(8, kindCluster, 12000)

	judged, note := Judge(ctx, empty, tonight, sevenNights)
	if !judged.RegressionDetected || !strings.Contains(note, "0 of the 7 required kind-fake-gpu runs") || !strings.Contains(note, "1 regressions detected") {
		t.Fatalf("without history only the SLO fires and the note says so: %q", note)
	}

	full := newFakeStore()
	publishAll(t, full, nights(1, 7, kindCluster, 5000)...)
	if _, note := Judge(ctx, full, nightlyDocument(8, kindCluster, 5500), sevenNights); !strings.Contains(note, "compared 1 p90 metrics with the median of their last 7 values among 7 kind-fake-gpu runs") || !strings.HasSuffix(note, "no regression detected") {
		t.Fatalf("with history the note names the window: %q", note)
	}
	otherScenario := Build(map[string]core.Section{"workspace-ttj": {Category: core.CategoryLoad, Metrics: map[string]any{"ttj_p90_ms": 9000}}}, RunInfo{StartedAt: time.Date(2026, 10, 8, 1, 0, 0, 0, time.UTC), ClusterType: kindCluster})
	if _, note := Judge(ctx, full, otherScenario, sevenNights); !strings.Contains(note, "none of the 1 p90 metrics has 7 earlier values yet among the 7 kind-fake-gpu runs") {
		t.Fatalf("seven runs that never ran this scenario must not be reported as a comparison: %q", note)
	}

	broken := newFakeStore()
	broken.failGet[ManifestName] = errors.New("disk on fire")
	if judged, note := Judge(ctx, broken, tonight, sevenNights); !judged.RegressionDetected || !strings.Contains(note, "skipped: read manifest: disk on fire") {
		t.Fatalf("a broken history must not hide the SLO verdict: %q", note)
	}

	if _, note := Judge(ctx, full, nightlyDocument(8, "", 5500), sevenNights); !strings.Contains(note, "cluster type is unknown") {
		t.Fatalf("an unknown cluster type must be called out: %q", note)
	}
	if _, note := Judge(ctx, full, tonight, RegressionConfig{Window: 0, Threshold: 0.15}); !strings.Contains(note, "disabled") {
		t.Fatalf("a zero window must be reported as disabled: %q", note)
	}
}

func TestJudgeStillComparesWhenAnOldRunIsUnreadable(t *testing.T) {
	store := newFakeStore()
	publishAll(t, store, nights(1, 8, kindCluster, 5000)...)
	oldest := nightlyDocument(1, kindCluster, 5000).FileName()
	store.objects[oldest] = []byte("{broken")

	judged, note := Judge(context.Background(), store, nightlyDocument(9, kindCluster, 5900), sevenNights)

	if !judged.RegressionDetected || len(judged.Regressions) != 1 || judged.Regressions[0].Against != AgainstMedian {
		t.Fatalf("seven readable runs are a full window even when an older one is broken: %+v", judged.Regressions)
	}
	if !strings.Contains(note, "1 earlier runs could not be read and were skipped") || !strings.Contains(note, oldest) {
		t.Fatalf("the skipped file must be named in the note: %q", note)
	}
}

func TestThroughputSLOIsAFloorNotACeiling(t *testing.T) {
	section := core.Section{
		Category: core.CategoryScale,
		Metrics:  map[string]any{"throughput_per_sec@100": 6.3, "throughput_per_sec@500": 4.2, "p90_ms@100": int64(900)},
		SLO:      map[string]int{"throughput_per_sec": 5, "p90_ms": 1000},
	}
	judged := DetectRegressions(Build(map[string]core.Section{burstSection: section}, RunInfo{}), nil, sevenNights)
	if len(judged.Regressions) != 1 {
		t.Fatalf("only the round under 5 per second may be flagged: %+v", judged.Regressions)
	}
	r := judged.Regressions[0]
	if r.Metric != "throughput_per_sec@500" || r.Against != AgainstSLO || !strings.Contains(r.Message, "4.2/s is under the SLO of 5.0/s") {
		t.Fatalf("the slow round must blame the floor: %+v", r)
	}
	if judged.Sections[burstSection].Metrics["throughput_per_sec@500_regression"] != true {
		t.Fatal("the flagged metric must get its sibling flag")
	}
}

func TestCountSLOIsReportedAsAPlainNumber(t *testing.T) {
	section := core.Section{Category: core.CategoryLoad, Metrics: map[string]any{"failed_count": 3}, SLO: map[string]int{"failed_count": 2}}
	judged := DetectRegressions(Build(map[string]core.Section{"workspace-ttj": section}, RunInfo{}), nil, sevenNights)
	if len(judged.Regressions) != 1 || judged.Regressions[0].Message != "workspace-ttj failed_count: 3 is over the SLO of 2" {
		t.Fatalf("a count is judged as a ceiling and printed without a unit: %+v", judged.Regressions)
	}
}
