package actions

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// clockDriftThreshold is the maximum difference between a node's clock and this
// installer machine's clock tolerated before the check warns. Placeholder value,
// tune later.
const clockDriftThreshold = 120 * time.Second

// CheckClockDrift is a read-only preflight check that warns when a node's clock
// differs from the machine running the installer by more than the threshold.
// Unlike the relative clock-skew check (which compares nodes to each other), this
// catches whole-cluster drift (e.g. an air-gapped cluster with no NTP where all
// nodes share the same wrong time) by comparing against an external reference.
//
// The reference is this machine's clock, so a warning may equally mean the
// machine's clock is wrong — the message states the reference explicitly. The
// comparison is done on absolute instants, so the machine and the nodes being in
// different timezones does not matter; only the real elapsed difference does.
type CheckClockDrift struct {
	// Now returns the reference (installer machine) time; defaults to time.Now.
	// Injectable for tests.
	Now func() time.Time
}

func (a CheckClockDrift) Kind() string { return "preflight.clock-drift" }

func (a CheckClockDrift) Args() map[string]any { return map[string]any{} }

func (a CheckClockDrift) Describe() string {
	return "verify node clocks match this machine's clock via lease heartbeats (warning only)"
}

func (a CheckClockDrift) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a CheckClockDrift) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	beats, err := deps.ClusterProbe.NodeHeartbeats(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not read node heartbeats to check clock drift: %v", err),
			Remediation: "grant the installer permission to list leases in kube-node-lease",
		}, nil
	}
	if len(beats) == 0 {
		return orchestrator2.ActionResult{}, nil
	}

	now := a.now()
	offenders := make([]map[string]any, 0)
	for _, b := range beats {
		offset := now.Sub(b.RenewTime)
		if offset > clockDriftThreshold || offset < -clockDriftThreshold {
			offenders = append(offenders, map[string]any{
				"node":          b.Name,
				"offset":        driftDescription(offset),
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

	// Show the reference in both the machine's local zone and UTC so a timezone
	// difference between this machine and the nodes is obvious and not mistaken
	// for drift.
	reference := fmt.Sprintf("%s (%s)", now.Format(time.RFC3339), now.UTC().Format(time.RFC3339))

	return orchestrator2.ActionResult{
		Status: orchestrator2.StatusWarn,
		Reason: fmt.Sprintf("%d node(s) differ from this machine's clock by more than %s", len(offenders), clockDriftThreshold),
		Data:   map[string]any{"driftNodes": offenders, "reference": reference},
	}, nil
}

// driftDescription renders a signed offset (machine time minus node heartbeat)
// as a human direction relative to this machine.
func driftDescription(offset time.Duration) string {
	if offset >= 0 {
		return offset.Round(time.Second).String() + " behind this machine"
	}
	return (-offset).Round(time.Second).String() + " ahead of this machine"
}
