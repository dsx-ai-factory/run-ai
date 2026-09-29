/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"gopkg.in/yaml.v3"
)

type scenarioFile struct {
	name     string
	folder   string
	scenario core.Scenario
}

type Options struct {
	Only      []string
	Overrides map[string]any
}

func Run(ctx context.Context, target core.Target, dir string, registry map[string]core.Suite, opts Options) (map[string]core.Section, error) {
	scenarios, err := discover(dir, opts.Only)
	if err != nil {
		return nil, err
	}
	if err := validate(scenarios, registry); err != nil {
		return nil, err
	}
	log.Printf("%d scenarios: %s", len(scenarios), strings.Join(scenarioNames(scenarios), ", "))
	if len(opts.Overrides) > 0 {
		log.Printf("knob overrides for every scenario: %s", describeOverrides(opts.Overrides))
		for i := range scenarios {
			scenarios[i].scenario = withOverrides(scenarios[i].scenario, opts.Overrides)
		}
	}

	results := make(map[string]core.Section)
	var errs []error
	for _, s := range scenarios {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		log.Printf("scenario %q: started (suite %s)", s.name, s.scenario.Suite)
		started := time.Now()
		suite := registry[s.scenario.Suite]
		section, err := suite.Run(ctx, target, s.scenario)
		elapsed := time.Since(started).Round(time.Second)
		if section.Category == "" {
			section.Category = suite.Category()
		}
		if err != nil {
			section.Failed = true
			section.Error = err.Error()
			errs = append(errs, fmt.Errorf("scenario %q: %w", s.name, err))
		} else if section.Failed {
			if section.Error == "" {
				section.Error = "suite reported failure"
			}
			errs = append(errs, fmt.Errorf("scenario %q: %s", s.name, section.Error))
		}
		if section.Failed {
			log.Printf("scenario %q: failed in %s: %s", s.name, elapsed, section.Error)
		} else {
			log.Printf("scenario %q: passed in %s", s.name, elapsed)
		}
		results[s.name] = section
	}

	if len(errs) > 0 {
		return results, errors.Join(errs...)
	}
	return results, nil
}

func keepOnly(paths map[string]string, only []string, dir string) ([]string, error) {
	available := slices.Sorted(maps.Keys(paths))
	if len(only) == 0 {
		return available, nil
	}
	seen := make(map[string]bool, len(only))
	kept := make([]string, 0, len(only))
	for _, name := range only {
		if seen[name] {
			return nil, fmt.Errorf("scenario %q is listed twice in --only", name)
		}
		seen[name] = true
		if _, ok := paths[name]; !ok {
			return nil, fmt.Errorf("scenario %q not found in %q; available: %s", name, dir, strings.Join(available, ", "))
		}
		kept = append(kept, name)
	}
	return kept, nil
}

func withOverrides(scenario core.Scenario, overrides map[string]any) core.Scenario {
	knobs := make(map[string]any, len(scenario.Knobs)+len(overrides))
	maps.Copy(knobs, scenario.Knobs)
	maps.Copy(knobs, overrides)
	scenario.Knobs = knobs
	return scenario
}

func describeOverrides(overrides map[string]any) string {
	parts := make([]string, 0, len(overrides))
	for _, key := range slices.Sorted(maps.Keys(overrides)) {
		parts = append(parts, fmt.Sprintf("%s=%v", key, overrides[key]))
	}
	return strings.Join(parts, ", ")
}

func scenarioNames(files []scenarioFile) []string {
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, file.name)
	}
	return names
}

func discover(dir string, only []string) ([]scenarioFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading scenarios dir %q: %w", dir, err)
	}
	paths := map[string]string{}
	folders := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := collectFolder(dir, entry.Name(), paths, folders); err != nil {
				return nil, err
			}
			continue
		}
		if err := collect(dir, entry.Name(), "", paths, folders); err != nil {
			return nil, err
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no *.yaml scenarios found in %q", dir)
	}
	names, err := keepOnly(paths, only, dir)
	if err != nil {
		return nil, err
	}

	scenarios := make([]scenarioFile, 0, len(names))
	for _, name := range names {
		parsed, err := parse(paths[name])
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, scenarioFile{name: name, folder: folders[name], scenario: parsed})
	}
	return scenarios, nil
}

func collectFolder(dir, folder string, paths, folders map[string]string) error {
	entries, err := os.ReadDir(filepath.Join(dir, folder))
	if err != nil {
		return fmt.Errorf("reading scenarios folder %q: %w", folder, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return fmt.Errorf("scenario folder %q must not contain sub-folders, found %q", folder, entry.Name())
		}
		if err := collect(filepath.Join(dir, folder), entry.Name(), folder, paths, folders); err != nil {
			return err
		}
	}
	return nil
}

func collect(dir, file, folder string, paths, folders map[string]string) error {
	if !strings.HasSuffix(file, ".yaml") {
		return nil
	}
	name := strings.TrimSuffix(file, ".yaml")
	if previous, taken := paths[name]; taken {
		return fmt.Errorf("scenario %q exists twice: %s and %s", name, previous, filepath.Join(dir, file))
	}
	paths[name] = filepath.Join(dir, file)
	folders[name] = folder
	return nil
}

func parse(path string) (core.Scenario, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return core.Scenario{}, fmt.Errorf("reading %q: %w", path, err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return core.Scenario{}, fmt.Errorf("parsing %q: %w", path, err)
	}

	suite, ok := raw["suite"].(string)
	if !ok || suite == "" {
		return core.Scenario{}, fmt.Errorf("%q: missing or invalid `suite` field", path)
	}
	delete(raw, "suite")

	return core.Scenario{Suite: suite, Knobs: raw}, nil
}

func validate(scenarios []scenarioFile, registry map[string]core.Suite) error {
	for _, s := range scenarios {
		suite, ok := registry[s.scenario.Suite]
		if !ok {
			return fmt.Errorf("scenario %q references unknown suite %q", s.name, s.scenario.Suite)
		}
		if s.folder != "" && s.folder != string(suite.Category()) {
			return fmt.Errorf("scenario %q sits in folder %q but its suite %q is a %s suite", s.name, s.folder, suite.Name(), suite.Category())
		}
	}
	return nil
}
