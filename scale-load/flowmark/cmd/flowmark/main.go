/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/runner"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/basicflow"
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
	outputDir := flag.String("output", "./out", "directory for results.json")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig of the cluster under test")
	cpURL := flag.String("cp-url", "", "control plane URL (credentials via FLOWMARK_CP_* env vars)")
	flag.Parse()

	env := core.Env{
		KubeconfigPath:  *kubeconfig,
		ControlPlaneURL: *cpURL,
		Username:        os.Getenv("FLOWMARK_CP_USER"),
		Password:        os.Getenv("FLOWMARK_CP_PASSWORD"),
		Token:           os.Getenv("FLOWMARK_CP_TOKEN"),
	}
	registry := buildRegistry(selfcheck.Suite{}, basicflow.Suite{})
	ctx := context.Background()

	results := core.Results{Schema: core.SchemaVersion, Metadata: core.Metadata{StartedAt: time.Now().UTC()}}
	if env.KubeconfigPath != "" {
		offset, err := clockOffset(ctx, env.KubeconfigPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "flowmark:", err)
			results.Metadata.ClockOffsetError = err.Error()
		} else {
			results.Metadata.ClockOffsetMs = &offset
		}
	}

	sections, runErr := runner.Run(ctx, env, *scenariosDir, registry)
	results.Metadata.FinishedAt = time.Now().UTC()
	results.Scenarios = sections
	if len(sections) > 0 {
		if err := writeResults(*outputDir, results); err != nil {
			return errors.Join(runErr, err)
		}
	}
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

func writeResults(dir string, results core.Results) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating output dir %q: %w", dir, err)
	}
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "results.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing %q: %w", path, err)
	}
	return nil
}
