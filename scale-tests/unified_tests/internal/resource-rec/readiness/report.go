/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package readiness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	StatusPass = "PASS"
	StatusFail = "FAIL"
	// StatusWarn marks an observation that is worth surfacing but must not block an
	// environment — e.g. a scrape target that is down because its dependency (kubeflow,
	// knative) is legitimately not installed.
	StatusWarn = "WARN"
)

// CheckResult is one readiness check.
type CheckResult struct {
	Check   string   `json:"check"`
	Status  string   `json:"status"`
	Summary string   `json:"summary"`
	Detail  []string `json:"detail,omitempty"`
}

// Report is the top-level output. Ready is false if any check FAILed; WARN never
// affects Ready.
type Report struct {
	Ready  bool          `json:"ready"`
	Checks []CheckResult `json:"checks"`
}

func (r *Report) Add(c CheckResult) {
	r.Checks = append(r.Checks, c)
}

func (r *Report) Finalize() {
	r.Ready = true
	for _, c := range r.Checks {
		if c.Status == StatusFail {
			r.Ready = false
		}
	}
}

// Print writes a human-readable summary to stdout. Detail lines are capped so a
// cluster with thousands of unready pods does not bury the result.
func (r *Report) Print() {
	const maxDetail = 20

	fmt.Println("=== environment readiness ===")
	for _, c := range r.Checks {
		fmt.Printf("  [%-4s] %-28s %s\n", c.Status, c.Check, c.Summary)
		for i, d := range c.Detail {
			if i == maxDetail {
				fmt.Printf("           ... and %d more\n", len(c.Detail)-maxDetail)
				break
			}
			fmt.Printf("           - %s\n", d)
		}
	}
	if r.Ready {
		fmt.Println("RESULT: READY")
	} else {
		fmt.Println("RESULT: NOT READY")
	}
}

func (r *Report) Write(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir %s: %w", outputDir, err)
	}
	path := filepath.Join(outputDir, "env-readiness-report.json")
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling readiness report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Printf("wrote %s\n", path)
	return nil
}
