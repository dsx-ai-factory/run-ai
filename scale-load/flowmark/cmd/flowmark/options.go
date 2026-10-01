/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/runner"
)

func runOptions(only, workloads string) (runner.Options, error) {
	options := runner.Options{Only: splitList(only)}
	if workloads == "" {
		return options, nil
	}
	sizes, err := parseWorkloads(workloads)
	if err != nil {
		return runner.Options{}, err
	}
	options.Overrides = map[string]any{"workloads": sizes}
	return options, nil
}

func splitList(text string) []string {
	var items []string
	for _, item := range strings.Split(text, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func parseWorkloads(text string) ([]any, error) {
	items := splitList(text)
	if len(items) == 0 {
		return nil, fmt.Errorf("--workloads %q: expected one or more whole numbers like 100 or 100,250", text)
	}
	sizes := make([]any, 0, len(items))
	for _, item := range items {
		size, err := strconv.Atoi(item)
		if err != nil || size <= 0 {
			return nil, fmt.Errorf("--workloads %q: %q is not a positive whole number", text, item)
		}
		sizes = append(sizes, size)
	}
	return sizes, nil
}
