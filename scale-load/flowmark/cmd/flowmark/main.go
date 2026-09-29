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
	"os"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/report"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/runner"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/burst"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/selfcheck"
)

const clockProbeNamespace = "default"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "flowmark:", err)
		os.Exit(1)
	}
}

func run() error {
	scenariosDir := flag.String("scenarios", "./scenarios", "directory with scenario yaml files")
	outputDir := flag.String("output", "./out", "directory that receives results-<timestamp>.json and manifest.json")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig of the cluster under test")
	cpURL := flag.String("cp-url", "", "control plane URL (credentials via FLOWMARK_CP_* env vars)")
	provider := flag.String("provider", "local", "who provided the environment, e.g. ci or kind")
	clusterType := flag.String("cluster-type", "local", "cluster class recorded with the results so runs are compared like with like")
	cpVersion := flag.String("cp-version", "", "control plane version to record, if known")
	clusterVersion := flag.String("cluster-version", "", "cluster version to record, if known")
	flag.Parse()

	env := core.Env{
		KubeconfigPath:  *kubeconfig,
		ControlPlaneURL: *cpURL,
		Username:        os.Getenv("FLOWMARK_CP_USER"),
		Password:        os.Getenv("FLOWMARK_CP_PASSWORD"),
		Token:           os.Getenv("FLOWMARK_CP_TOKEN"),
	}
	registry := buildRegistry(selfcheck.Suite{}, burst.Suite{})
	ctx := context.Background()

	info := report.RunInfo{
		StartedAt:           time.Now(),
		Provider:            *provider,
		ClusterType:         *clusterType,
		ControlPlaneVersion: *cpVersion,
		ClusterVersion:      *clusterVersion,
	}
	if env.KubeconfigPath != "" {
		offset, err := clockOffset(ctx, env.KubeconfigPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "flowmark:", err)
			info.ClockOffsetError = err.Error()
		} else {
			info.ClockOffsetMs = &offset
		}
	}

	sections, runErr := runner.Run(ctx, env, *scenariosDir, registry)
	info.FinishedAt = time.Now()
	if len(sections) == 0 {
		return runErr
	}

	store, err := report.NewLocal(*outputDir)
	if err != nil {
		return errors.Join(runErr, fmt.Errorf("report: %w", err))
	}
	receipt, err := report.Publish(ctx, store, report.Build(sections, info))
	if err != nil {
		return errors.Join(runErr, fmt.Errorf("report: %w", err))
	}
	fmt.Fprintf(os.Stderr, "flowmark: wrote %s to %s, %d runs in %s\n", receipt.File, store, receipt.Runs, report.ManifestName)
	return runErr
}

func clockOffset(ctx context.Context, kubeconfig string) (int64, error) {
	kube, err := client.Kube(kubeconfig)
	if err != nil {
		return 0, err
	}
	offset, err := measure.ClockOffset(ctx, kube, clockProbeNamespace)
	if err != nil {
		return 0, fmt.Errorf("measuring the cluster clock offset: %w", err)
	}
	return offset.Milliseconds(), nil
}

func buildRegistry(suites ...core.Scenario) map[string]core.Scenario {
	registry := make(map[string]core.Scenario, len(suites))
	for _, s := range suites {
		registry[s.Name()] = s
	}
	return registry
}
