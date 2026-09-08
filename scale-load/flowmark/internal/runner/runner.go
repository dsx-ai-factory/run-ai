/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
	"gopkg.in/yaml.v3"
)

type scenario struct {
	name   string
	config core.Config
}

// Run executes every scenario in dir sequentially, in filename order.
// On failure it returns the collected results together with a non-nil
// error naming each failed scenario; the returned results remain valid.
func Run(ctx context.Context, env core.Env, dir string, registry map[string]core.Scenario) (map[string]core.Section, error) {
	scenarios, err := discover(dir)
	if err != nil {
		return nil, err
	}

	if err := validate(scenarios, registry); err != nil {
		return nil, err
	}

	results := make(map[string]core.Section)
	var errs []error
	for _, s := range scenarios {
		if err := ctx.Err(); err != nil {
			return results, err
		}
		suite := registry[s.config.Suite]
		section, err := suite.Run(ctx, env, s.config)
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
		results[s.name] = section
	}

	if len(errs) > 0 {
		return results, errors.Join(errs...)
	}
	return results, nil
}

func discover(dir string) ([]scenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading scenarios dir %q: %w", dir, err)
	}

	var scenarios []scenario
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		config, err := parse(path)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		scenarios = append(scenarios, scenario{name: name, config: config})
	}

	if len(scenarios) == 0 {
		return nil, fmt.Errorf("no *.yaml scenarios found in %q", dir)
	}
	return scenarios, nil
}

func parse(path string) (core.Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return core.Config{}, fmt.Errorf("reading %q: %w", path, err)
	}

	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return core.Config{}, fmt.Errorf("parsing %q: %w", path, err)
	}

	suite, ok := raw["suite"].(string)
	if !ok || suite == "" {
		return core.Config{}, fmt.Errorf("%q: missing or invalid `suite` field", path)
	}
	delete(raw, "suite")

	return core.Config{Suite: suite, Knobs: raw}, nil
}

func validate(scenarios []scenario, registry map[string]core.Scenario) error {
	for _, s := range scenarios {
		if _, ok := registry[s.config.Suite]; !ok {
			return fmt.Errorf("scenario %q references unknown suite %q", s.name, s.config.Suite)
		}
	}
	return nil
}
