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

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/runner"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/suites/selfcheck"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "flowmark:", err)
		os.Exit(1)
	}
}

func run() error {
	scenariosDir := flag.String("scenarios", "./scenarios", "directory with scenario yaml files")
	outputDir := flag.String("output", "./out", "directory for results.json")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (reserved for future suites)")
	cpURL := flag.String("cp-url", "", "control plane URL (reserved for future suites)")
	flag.Parse()

	env := core.Env{KubeconfigPath: *kubeconfig, ControlPlaneURL: *cpURL}
	registry := buildRegistry(selfcheck.Suite{})

	results, runErr := runner.Run(context.Background(), env, *scenariosDir, registry)
	if len(results) > 0 {
		if err := writeResults(*outputDir, results); err != nil {
			return errors.Join(runErr, err)
		}
	}
	return runErr
}

func buildRegistry(suites ...core.Scenario) map[string]core.Scenario {
	registry := make(map[string]core.Scenario, len(suites))
	for _, s := range suites {
		registry[s.Name()] = s
	}
	return registry
}

func writeResults(dir string, results map[string]core.Section) error {
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
