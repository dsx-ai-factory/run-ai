/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package containermetrics derives per-container resource usage from Prometheus for
// components VPA cannot target. Static pods such as kube-apiserver and etcd are owned by
// the Node rather than a controller, so they have no targetRef and VPA produces nothing
// for them — yet they are among the largest consumers on a cluster.
package containermetrics

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Row matches the VPA recommendation schema so both sources merge into one table.
// Source distinguishes them: bounds here are observed usage, not VPA's confidence
// interval.
type Row struct {
	Namespace     string `json:"namespace"`
	Workload      string `json:"workload"`
	Container     string `json:"container"`
	Source        string `json:"source"`
	CPULowerBound string `json:"cpu_lower_bound"`
	CPUTarget     string `json:"cpu_target"`
	CPUUpperBound string `json:"cpu_upper_bound"`
	MemLowerBound string `json:"memory_lower_bound"`
	MemTarget     string `json:"memory_target"`
	MemUpperBound string `json:"memory_upper_bound"`
}

const source = "prometheus"

// Extract returns one row per container in the given namespaces, aggregating across pods.
// lower/target/upper are min, p95 and max observed over the window.
func Extract(ctx context.Context, api v1.API, namespaces []string, window time.Duration, endTime time.Time) ([]Row, error) {
	sel := fmt.Sprintf(`namespace=~%q, container!="", container!="POD"`, nsRegex(namespaces))
	dur := fmt.Sprintf("%ds", int(window.Seconds()))

	cpu := fmt.Sprintf(`sum by (namespace, container) (rate(container_cpu_usage_seconds_total{%s}[5m]))`, sel)
	mem := fmt.Sprintf(`sum by (namespace, container) (container_memory_working_set_bytes{%s})`, sel)

	type stat struct{ lo, mid, hi float64 }
	gather := func(base string) (map[string]stat, error) {
		out := map[string]stat{}
		for name, fn := range map[string]string{
			"lo":  fmt.Sprintf(`min_over_time((%s)[%s:1m])`, base, dur),
			"mid": fmt.Sprintf(`quantile_over_time(0.95, (%s)[%s:1m])`, base, dur),
			"hi":  fmt.Sprintf(`max_over_time((%s)[%s:1m])`, base, dur),
		} {
			res, _, _, err := api.Query(ctx, fn, endTime)
			if err != nil {
				return nil, fmt.Errorf("query failed: %w", err)
			}
			vec, ok := res.(model.Vector)
			if !ok {
				continue
			}
			for _, s := range vec {
				k := fmt.Sprintf("%s/%s", s.Metric["namespace"], s.Metric["container"])
				e := out[k]
				switch name {
				case "lo":
					e.lo = float64(s.Value)
				case "mid":
					e.mid = float64(s.Value)
				case "hi":
					e.hi = float64(s.Value)
				}
				out[k] = e
			}
		}
		return out, nil
	}

	cpuStats, err := gather(cpu)
	if err != nil {
		return nil, err
	}
	memStats, err := gather(mem)
	if err != nil {
		return nil, err
	}

	var rows []Row
	for k, c := range cpuStats {
		ns, container, ok := strings.Cut(k, "/")
		if !ok {
			continue
		}
		row := Row{
			Namespace:     ns,
			Workload:      container,
			Container:     container,
			Source:        source,
			CPULowerBound: cores(c.lo),
			CPUTarget:     cores(c.mid),
			CPUUpperBound: cores(c.hi),
		}
		// Left blank when the memory series is missing. Zero-filling from the map would
		// publish a 0-byte recommendation that reads as a measurement.
		if m, ok := memStats[k]; ok {
			row.MemLowerBound = bytes(m.lo)
			row.MemTarget = bytes(m.mid)
			row.MemUpperBound = bytes(m.hi)
		} else {
			log.Printf("no memory series for %s — leaving memory blank", k)
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Container < rows[j].Container })
	return rows, nil
}

func nsRegex(namespaces []string) string { return strings.Join(namespaces, "|") }

func cores(v float64) string { return fmt.Sprintf("%dm", int(v*1000)) }
func bytes(v float64) string { return fmt.Sprintf("%d", int64(v)) }

// Write emits JSON and CSV alongside the VPA output, in the same schema.
func Write(rows []Row, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", outputDir, err)
	}

	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "container-metrics.json"), data, 0o644); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(outputDir, "container-metrics.csv"))
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{
		"namespace", "workload", "container", "source",
		"cpu_lower_bound", "cpu_target", "cpu_upper_bound",
		"memory_lower_bound", "memory_target", "memory_upper_bound",
	}); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{
			r.Namespace, r.Workload, r.Container, r.Source,
			r.CPULowerBound, r.CPUTarget, r.CPUUpperBound,
			r.MemLowerBound, r.MemTarget, r.MemUpperBound,
		}); err != nil {
			return err
		}
	}
	return nil
}
