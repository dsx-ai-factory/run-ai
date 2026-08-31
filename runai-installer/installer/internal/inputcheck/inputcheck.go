/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package inputcheck validates a resolved installer spec against the local
// machine before any installation work begins. Every check is offline: it
// relies only on the spec (in memory) and the filesystem (certs, keys,
// bundles, token files). Anything network-related (DNS, registry auth) is out
// of scope and handled later by the preflight checks.
//
// Validate aggregates all findings rather than failing on the first, so the
// caller can surface every problem at once.
package inputcheck

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

// Finding is a single input-validation problem.
type Finding struct {
	Check   string
	Field   string
	Message string
}

// Report is the aggregated outcome of all checks.
type Report struct {
	Findings []Finding
}

// Err returns a concise summary error when there are findings, or nil. The
// per-finding detail is carried in Findings (rendered by the events layer), so
// this stays short to avoid duplicating that detail in CLI error output.
func (r Report) Err() error {
	if len(r.Findings) == 0 {
		return nil
	}
	return fmt.Errorf("input validation failed: %d problem(s)", len(r.Findings))
}

// check inspects the spec and returns any findings it discovers. A check must
// be nil-safe: the spec may be partially populated.
type check func(*spec.Spec) []Finding

// checks is the registry every Validate run iterates. Each check file appends
// its check in an init function.
var checks []check

// Validate runs every registered check and aggregates their findings.
func Validate(s *spec.Spec) Report {
	var r Report
	for _, c := range checks {
		r.Findings = append(r.Findings, c(s)...)
	}
	return r
}

// fromResult adapts a leaf validation.ValidationResult into findings, reusing
// the existing validators. A passing result yields no findings.
func fromResult(res validation.ValidationResult, check, field string) []Finding {
	if res.Status {
		return nil
	}
	return []Finding{{Check: check, Field: field, Message: res.Message}}
}
