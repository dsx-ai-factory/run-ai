/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

// Status is the outcome an action reports beyond a Go error. The empty value
// (StatusOK) means success / "pass" — for an install action that simply did its
// work, and for a check action that found nothing to flag.
type Status string

const (
	StatusOK   Status = ""
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

// ActionResult is the structured outcome of an action's Execute. It carries the
// verdict (Status) plus, for non-OK results, a human reason and remediation, and
// optionally typed Data the action produced (e.g. a helm release or cluster
// UUID) for the step output / status document.
//
// An action returns an error only when it could not run; a ran-but-non-compliant
// outcome (used by preflight checks) is reported as Status=fail with a nil error.
type ActionResult struct {
	Status      Status         `json:"status,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	Remediation string         `json:"remediation,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// IsZero reports whether the result carries nothing worth surfacing: a plain
// success with no data. The runner uses it to keep ordinary install steps free
// of empty result noise on their events.
func (r ActionResult) IsZero() bool {
	return r.Status == StatusOK && r.Reason == "" && r.Remediation == "" && len(r.Data) == 0
}
