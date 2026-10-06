/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/report"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/runner"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/load"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/scale"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/selfcheck"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmsgprefix)
	log.SetPrefix("flowmark: ")
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	scenariosDir := flag.String("scenarios", "./scenarios", "directory with scenario yaml files")
	only := flag.String("only", "", "comma-separated scenario names to run, e.g. training-burst (default: every scenario in --scenarios)")
	workloads := flag.String("workloads", "", "override the workloads knob for this run, e.g. 100 or 100,250")
	outputDir := flag.String("output", "./reports", "directory that receives results-<timestamp>.json and manifest.json")
	kubeconfig := flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "path to kubeconfig of the cluster under test (default $KUBECONFIG)")
	cpURL := flag.String("cp-url", os.Getenv("FLOWMARK_CP_URL"), "control plane URL (default $FLOWMARK_CP_URL; credentials via FLOWMARK_CP_USER and FLOWMARK_CP_PASSWORD, or FLOWMARK_CP_TOKEN)")
	cluster := flag.String("cluster", os.Getenv("FLOWMARK_CLUSTER"), "cluster name on the control plane whose versions are recorded, needed when it has several (default $FLOWMARK_CLUSTER)")
	provider := flag.String("provider", "local", "who provided the environment, e.g. ci or kind")
	clusterType := flag.String("cluster-type", "", "cluster class recorded with the results so runs are compared like with like (default: detected, e.g. vanilla-fake-gpu)")
	cpVersion := flag.String("cp-version", "", "control plane version to record (default: read from the control plane)")
	clusterVersion := flag.String("cluster-version", "", "cluster version to record (default: read from the control plane)")
	regressionWindow := flag.Int("regression-window", 7, "number of earlier runs of the same cluster type whose median each p90 is compared with; 0 disables the comparison")
	regressionThreshold := flag.Float64("regression-threshold", 0.15, "relative increase over the median that counts as a regression, e.g. 0.15 for 15%")
	flag.Parse()
	if flag.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q; scenarios are selected with --only", flag.Arg(0))
	}
	if *regressionWindow < 0 || *regressionThreshold <= 0 {
		return fmt.Errorf("--regression-window must be 0 or more and --regression-threshold above 0, got %d and %g", *regressionWindow, *regressionThreshold)
	}
	log.Printf("started, scenarios from %s, results to %s", *scenariosDir, *outputDir)
	options, err := runOptions(*only, *workloads)
	if err != nil {
		return err
	}

	target := core.Target{
		KubeconfigPath:  *kubeconfig,
		ControlPlaneURL: *cpURL,
		Username:        os.Getenv("FLOWMARK_CP_USER"),
		Password:        os.Getenv("FLOWMARK_CP_PASSWORD"),
		Token:           os.Getenv("FLOWMARK_CP_TOKEN"),
	}
	registry := buildRegistry(selfcheck.Suite{}, scale.Suite{}, load.Suite{})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	context.AfterFunc(ctx, func() {
		log.Print("interrupted: finishing cleanup and writing the rounds that completed; a second Ctrl-C or SIGTERM quits at once without cleanup")
		stop()
	})

	info := report.RunInfo{
		StartedAt:           time.Now(),
		Provider:            *provider,
		ClusterType:         *clusterType,
		ControlPlaneVersion: *cpVersion,
		ClusterVersion:      *clusterVersion,
	}
	if target.ControlPlaneURL != "" {
		describeEnvironment(ctx, target, *cluster, &info)
	}

	sections, runErr := runner.Run(ctx, target, *scenariosDir, registry, options)
	info.FinishedAt = time.Now()
	if len(sections) == 0 {
		return runErr
	}
	log.Printf("%d scenarios finished in %s, %d failed", len(sections), info.FinishedAt.Sub(info.StartedAt).Round(time.Second), failedCount(sections))

	store, err := report.NewLocal(*outputDir)
	if err != nil {
		return errors.Join(runErr, fmt.Errorf("report: %w", err))
	}
	document := judge(ctx, store, report.Build(sections, info), report.RegressionConfig{Window: *regressionWindow, Threshold: *regressionThreshold})
	receipt, err := report.Publish(ctx, store, document)
	if err != nil {
		return errors.Join(runErr, fmt.Errorf("report: %w", err))
	}
	log.Printf("wrote %s to %s, %d runs in %s", receipt.File, store, receipt.Runs, report.ManifestName)
	return runErr
}

func judge(ctx context.Context, store report.Store, document report.Document, config report.RegressionConfig) report.Document {
	history, err := report.LoadHistory(ctx, store, document.Metadata.ClusterType, config.Window)
	switch {
	case err != nil:
		log.Printf("comparison with earlier runs skipped: %v", err)
	case config.Window == 0:
		log.Print("comparison with earlier runs disabled, only SLOs are checked")
	case len(history) < config.Window:
		log.Printf("comparison with earlier runs skipped: %d of the %d required %s runs on record, only SLOs are checked", len(history), config.Window, document.Metadata.ClusterType)
	default:
		log.Printf("comparing every p90 with the median of the last %d %s runs, threshold %.0f%%", config.Window, document.Metadata.ClusterType, config.Threshold*100)
	}
	judged := report.DetectRegressions(document, history, config)
	for _, regression := range judged.Regressions {
		log.Printf("regression: %s", regression.Message) //nolint:gosec
	}
	if !judged.RegressionDetected {
		log.Print("no regression detected")
	}
	return judged
}

func describeEnvironment(ctx context.Context, target core.Target, cluster string, info *report.RunInfo) {
	clients, err := client.New(ctx, target)
	if err != nil {
		log.Printf("environment not described: %v", err)
		return
	}
	environment, err := clients.Environment(ctx, cluster)
	if err != nil {
		log.Printf("environment not described: %v", err)
		return
	}
	log.Printf("environment: %s", environment) //nolint:gosec
	if info.ControlPlaneVersion == "" {
		info.ControlPlaneVersion = environment.ControlPlaneVersion
	}
	if info.ClusterVersion == "" {
		info.ClusterVersion = environment.ClusterVersion
	}
	if info.ClusterType == "" {
		info.ClusterType = environment.ClusterType()
	}
}

func failedCount(sections map[string]core.Section) int {
	failed := 0
	for _, section := range sections {
		if section.Failed {
			failed++
		}
	}
	return failed
}

func buildRegistry(suites ...core.Suite) map[string]core.Suite {
	registry := make(map[string]core.Suite, len(suites))
	for _, s := range suites {
		registry[s.Name()] = s
	}
	return registry
}
