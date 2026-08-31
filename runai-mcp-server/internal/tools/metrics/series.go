/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"time"

	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

const (
	defaultSamples int32 = 20
	maxSamples     int32 = 1000

	// Token guard: requested samples × metrics beyond this drops the series and
	// asks the caller to narrow numberOfSamples.
	maxResultPoints = 5000
)

var windowDurations = map[string]time.Duration{
	"last_1h":  time.Hour,
	"last_24h": 24 * time.Hour,
	"last_7d":  7 * 24 * time.Hour,
	"last_30d": 30 * 24 * time.Hour,
}

type point struct {
	T string  `json:"t" jsonschema:"Sample timestamp (RFC3339, UTC)."`
	V float64 `json:"v" jsonschema:"Sample value; units depend on the series (device count, cores, GiB, percent 0-100, seconds, or count)."`
}

type timeRange struct {
	From            string `json:"from"            jsonschema:"Start of the returned range (RFC3339, UTC)."`
	To              string `json:"to"              jsonschema:"End of the returned range (RFC3339, UTC)."`
	NumberOfSamples int32  `json:"numberOfSamples" jsonschema:"Number of points requested per series across the range."`
}

func resolveTimeRange(window, start, end string) (time.Time, time.Time) {
	if startT, startSet, _ := common.ParseRFC3339IfSet("start", start); startSet {
		endT, _, _ := common.ParseRFC3339IfSet("end", end)
		return startT.UTC(), endT.UTC()
	}
	to := time.Now().UTC()
	return to.Add(-windowDurations[window]), to
}

func dedupe(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}
