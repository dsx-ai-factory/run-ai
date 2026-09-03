/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// extract-container-metrics derives per-container CPU and memory from Prometheus for
// namespaces VPA cannot cover — notably kube-system, whose apiserver and etcd are static
// pods with no controller for VPA to target.
package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/resource-rec/containermetrics"
	validity_gate_monitor "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/validity-gate-monitor"
	"github.com/kelseyhightower/envconfig"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type config struct {
	Namespaces []string `envconfig:"CONTAINER_METRICS_NAMESPACES" default:"kube-system"`

	// Bound the query to the run. Anchoring at RunEnd matters because this runs after the
	// VPA settle wait, so "now" would include idle time.
	RunStart string `envconfig:"RUN_START" required:"true"`
	RunEnd   string `envconfig:"RUN_END" required:"true"`

	OutputDir           string `envconfig:"OUTPUT_DIR" default:"output"`
	PrometheusPodName   string `envconfig:"PROMETHEUS_POD_NAME" default:"prometheus-runai-0"`
	PrometheusNamespace string `envconfig:"PROMETHEUS_NAMESPACE" default:"runai"`
	PrometheusPort      int    `envconfig:"PROMETHEUS_PORT" default:"9090"`
}

func main() {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	start, err := time.Parse(time.RFC3339, cfg.RunStart)
	if err != nil {
		log.Fatalf("invalid RUN_START %q: %v", cfg.RunStart, err)
	}
	end, err := time.Parse(time.RFC3339, cfg.RunEnd)
	if err != nil {
		log.Fatalf("invalid RUN_END %q: %v", cfg.RunEnd, err)
	}
	if !end.After(start) {
		log.Fatalf("RUN_END %s is not after RUN_START %s", cfg.RunEnd, cfg.RunStart)
	}
	window := end.Sub(start)

	ctx := context.Background()
	restConfig := ctrl.GetConfigOrDie()

	var rows []containermetrics.Row
	var etcdRows []containermetrics.EtcdDBSizeRow
	err = validity_gate_monitor.WithPrometheusAPI(ctx,
		cfg.PrometheusNamespace, cfg.PrometheusPodName, cfg.PrometheusPort, restConfig,
		func(api v1.API) error {
			var qErr error
			if rows, qErr = containermetrics.Extract(ctx, api, cfg.Namespaces, window, end); qErr != nil {
				return qErr
			}
			etcdRows, qErr = containermetrics.ExtractEtcdDBSize(ctx, api, window, end)
			return qErr
		})
	if err != nil {
		log.Fatalf("failed to extract container metrics: %v", err)
	}

	log.Printf("namespaces: %v, window: %v ending %s", cfg.Namespaces, window, end.Format(time.RFC3339))
	log.Printf("extracted %d container rows", len(rows))
	for _, r := range rows {
		log.Printf("  %s/%s cpu=%s mem=%s", r.Namespace, r.Container, r.CPUTarget, r.MemTarget)
	}

	if len(etcdRows) > 0 {
		log.Printf("  etcd db size: min=%s p95=%s max=%s",
			etcdRows[0].MinBytes, etcdRows[0].P95Bytes, etcdRows[0].MaxBytes)
		if err := containermetrics.WriteEtcdDBSize(etcdRows, cfg.OutputDir); err != nil {
			log.Fatalf("failed to write etcd db size: %v", err)
		}
		log.Printf("wrote etcd-db-size.{json,csv} to %s", cfg.OutputDir)
	} else {
		log.Printf("etcd db size metric not scraped, skipping")
	}

	// Fail closed. container-metrics.csv is optional to the report builder, so exiting 0
	// without writing it turns a broken extraction into a silently smaller report.
	if len(rows) == 0 {
		log.Fatalf("no container metrics found for %s", strings.Join(cfg.Namespaces, ","))
	}
	if err := containermetrics.Write(rows, cfg.OutputDir); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
	log.Printf("wrote container-metrics.{json,csv} to %s", cfg.OutputDir)
}
