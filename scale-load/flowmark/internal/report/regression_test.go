/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"slices"
	"strings"
	"testing"
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
	if judged.Sections[burstSection].Metrics["slo_p90_ms_regression"] != nil {
		t.Fatal("the SLO key itself is not a KPI and must never be flagged")
	}
}

func TestRelativeCheckIsSkippedUntilTheWindowIsFull(t *testing.T) {
	judged := DetectRegressions(nightlyDocument(7, kindCluster, 5900), nights(1, 6, kindCluster, 5000), sevenNights)

	if judged.RegressionDetected || flagged(t, judged, "p90_ms@100") {
		t.Fatalf("6 runs of history are not enough for a window of 7, the check must stay silent: %+v", judged.Regressions)
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

	history, err := LoadHistory(context.Background(), store, kindCluster, 7)
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
		history, err := LoadHistory(context.Background(), store, kindCluster, sevenNights.Window)
		if err != nil {
			t.Fatal(err)
		}
		judged := DetectRegressions(nightlyDocument(night, kindCluster, 7000), history, sevenNights)
		if judged.RegressionDetected {
			flaggedNights = append(flaggedNights, night)
		}
		publishAll(t, store, judged)
	}

	if want := []int{8, 9, 10, 11}; !slices.Equal(flaggedNights, want) {
		t.Fatalf("a lasting +40%% must be flagged while the slow nights are the minority of the window, want %v got %v", want, flaggedNights)
	}
}

func TestKPIsAreTheP90MetricsOnly(t *testing.T) {
	metrics := map[string]any{
		"p90_ms@100": 1, "ttj_p90_ms": 2, "cluster_synced_p90_ms": 3,
		"p50_ms@100": 4, "slo_p90_ms": 5, "p90_ms@100_regression": true, "flow": "x", "max_ms": 6,
	}
	if got := kpis(metrics); strings.Join(got, ",") != "cluster_synced_p90_ms,p90_ms@100,ttj_p90_ms" {
		t.Fatalf("only p90 metrics are KPIs, sorted, got %v", got)
	}
	if sloKey("p90_ms@250") != "slo_p90_ms" || sloKey("ttj_p90_ms") != "slo_ttj_p90_ms" {
		t.Fatal("the SLO of a round metric is the SLO of its base name")
	}
	if median([]float64{5, 1, 3}) != 3 || median([]float64{4, 1, 3, 2}) != 2.5 {
		t.Fatal("median must sort and average the two middle values for even counts")
	}
}
