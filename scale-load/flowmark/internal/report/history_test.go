/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const (
	kindCluster    = "kind-fake-gpu"
	vanillaCluster = "vanilla-fake-gpu"
)

func nightlyDocument(night int, clusterType string, p90 int) Document {
	return Build(map[string]core.Section{"training-burst": {Category: core.CategoryScale, Metrics: map[string]any{"p90_ms@100": p90}, SLO: map[string]int{"p90_ms": 10000}}},
		RunInfo{StartedAt: time.Date(2026, 10, night, 1, 0, 0, 0, time.UTC), Provider: "ci", ClusterType: clusterType})
}

func publishAll(t *testing.T, store Store, documents ...Document) {
	t.Helper()
	for _, document := range documents {
		if _, err := Publish(context.Background(), store, document); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadHistoryReturnsTheNewestRunsOfTheSameClusterTypeOldestFirst(t *testing.T) {
	store := newFakeStore()
	var published []Document
	for night := 1; night <= 10; night++ {
		clusterType := kindCluster
		if night%2 == 0 {
			clusterType = vanillaCluster
		}
		published = append(published, nightlyDocument(night, clusterType, 5000+night))
	}
	publishAll(t, store, published...)

	history, _, err := LoadHistory(context.Background(), store, kindCluster, 3)

	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 3 {
		t.Fatalf("window of 3 must return 3 runs, got %d", len(history))
	}
	for i, want := range []int{5, 7, 9} {
		got := history[i]
		if got.Metadata.ClusterType != kindCluster || got.Metadata.Date.Day() != want {
			t.Fatalf("run %d must be night %d of %s, got night %d of %s", i, want, kindCluster, got.Metadata.Date.Day(), got.Metadata.ClusterType)
		}
		if p90, ok := got.Sections["training-burst"].Metrics["p90_ms@100"].(float64); !ok || int(p90) != 5000+want {
			t.Fatalf("metrics must survive the round trip as numbers, got %#v", got.Sections["training-burst"].Metrics["p90_ms@100"])
		}
	}
}

func TestLoadHistoryReturnsEverythingWhenShorterThanTheWindow(t *testing.T) {
	store := newFakeStore()
	publishAll(t, store, nightlyDocument(1, kindCluster, 5000), nightlyDocument(2, vanillaCluster, 5000), nightlyDocument(3, kindCluster, 5000))

	history, _, err := LoadHistory(context.Background(), store, kindCluster, 7)

	if err != nil || len(history) != 2 {
		t.Fatalf("2 matching runs exist, want both, got %d (%v)", len(history), err)
	}
}

func TestLoadHistoryIsEmptyWithoutAManifestAWindowOrAClusterType(t *testing.T) {
	store := newFakeStore()
	if history, _, err := LoadHistory(context.Background(), store, kindCluster, 7); err != nil || history != nil {
		t.Fatalf("no manifest means no history and no error, got %v / %v", history, err)
	}
	publishAll(t, store, nightlyDocument(1, kindCluster, 5000), nightlyDocument(2, "", 5000))
	if history, _, err := LoadHistory(context.Background(), store, kindCluster, 0); err != nil || history != nil {
		t.Fatalf("a zero window disables history, got %v / %v", history, err)
	}
	if history, _, err := LoadHistory(context.Background(), store, "", 7); err != nil || history != nil {
		t.Fatalf("an unknown cluster type must not be matched against runs that also lack one, got %v / %v", history, err)
	}
}

func TestLoadHistoryOrdersRunsByTheirTimestampNotByManifestOrder(t *testing.T) {
	store := newFakeStore()
	publishAll(t, store, nightlyDocument(3, kindCluster, 3), nightlyDocument(1, kindCluster, 1), nightlyDocument(2, kindCluster, 2))

	history, _, err := LoadHistory(context.Background(), store, kindCluster, 2)

	if err != nil || len(history) != 2 {
		t.Fatalf("want the 2 newest runs, got %d (%v)", len(history), err)
	}
	if history[0].Metadata.Date.Day() != 2 || history[1].Metadata.Date.Day() != 3 {
		t.Fatalf("a run that finished late must not hide newer runs appended before it, got days %d and %d", history[0].Metadata.Date.Day(), history[1].Metadata.Date.Day())
	}
}

func TestLoadHistorySkipsAndNamesAMissingOrCorruptRun(t *testing.T) {
	store := newFakeStore()
	publishAll(t, store, nightlyDocument(1, kindCluster, 5000), nightlyDocument(2, kindCluster, 5000), nightlyDocument(3, kindCluster, 5000))
	missing := nightlyDocument(2, kindCluster, 5000).FileName()
	delete(store.objects, missing)
	corrupt := nightlyDocument(1, kindCluster, 5000).FileName()
	store.objects[corrupt] = []byte("{not json")

	history, skipped, err := LoadHistory(context.Background(), store, kindCluster, 7)

	if err != nil || len(history) != 1 || history[0].Metadata.Date.Day() != 3 {
		t.Fatalf("the readable run must still be used, got %d runs (%v)", len(history), err)
	}
	if len(skipped) != 2 || !strings.Contains(skipped[0], "read history "+missing) || !strings.Contains(skipped[1], "parse history "+corrupt) {
		t.Fatalf("every unreadable run must be named: %v", skipped)
	}

	broken := newFakeStore()
	broken.failGet[ManifestName] = errors.New("disk on fire")
	if _, _, err := LoadHistory(context.Background(), broken, kindCluster, 7); err == nil || !strings.Contains(err.Error(), "read manifest") {
		t.Fatalf("a manifest that cannot be read is still fatal: %v", err)
	}
}
