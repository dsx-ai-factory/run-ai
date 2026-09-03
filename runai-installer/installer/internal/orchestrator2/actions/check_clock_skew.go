/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

// clockSkewThreshold is the maximum heartbeat lag between nodes tolerated before
// the check warns. Node leases renew roughly every 10s, so this is set well
// above that jitter; placeholder value, tune later.
const clockSkewThreshold = 90 * time.Second

// CheckClockSkew is a read-only preflight check that warns on clock skew between
// nodes. It compares each node's lease renewal time (stamped by the node's own
// clock) against the freshest one; a node whose heartbeat lags far behind is
// either clock-skewed or unreachable/dead. It never fails the install: the
// finding is surfaced as a warning only.
type CheckClockSkew struct{}

func (a CheckClockSkew) Kind() string { return "preflight.clock-skew" }

func (a CheckClockSkew) Args() map[string]any { return map[string]any{} }

func (a CheckClockSkew) Describe() string {
	return "verify node clocks are in sync via lease heartbeats (warning only)"
}

func (a CheckClockSkew) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	beats, err := deps.ClusterProbe.NodeHeartbeats(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not read node heartbeats to check clock skew: %v", err),
			Remediation: "grant the installer permission to list leases in kube-node-lease",
		}, nil
	}
	if len(beats) < 2 {
		// Nothing to compare against (single-node or no leases).
		return orchestrator2.ActionResult{}, nil
	}

	reference := beats[0].RenewTime
	for _, b := range beats {
		if b.RenewTime.After(reference) {
			reference = b.RenewTime
		}
	}

	offenders := make([]map[string]any, 0)
	for _, b := range beats {
		lag := reference.Sub(b.RenewTime)
		if lag > clockSkewThreshold {
			offenders = append(offenders, map[string]any{
				"node":          b.Name,
				"lag":           lag.Round(time.Second).String(),
				"lastHeartbeat": b.RenewTime.UTC().Format(time.RFC3339),
			})
		}
	}
	if len(offenders) == 0 {
		return orchestrator2.ActionResult{}, nil
	}
	sort.Slice(offenders, func(i, j int) bool {
		return offenders[i]["node"].(string) < offenders[j]["node"].(string)
	})

	maxLag := time.Duration(0)
	for _, b := range beats {
		if lag := reference.Sub(b.RenewTime); lag > maxLag {
			maxLag = lag
		}
	}

	return orchestrator2.ActionResult{
		Status: orchestrator2.StatusWarn,
		Reason: fmt.Sprintf("%d node(s) with clock skew or stale heartbeat (max lag %s)", len(offenders), maxLag.Round(time.Second)),
		Data:   map[string]any{"clockSkewNodes": offenders},
	}, nil
}
