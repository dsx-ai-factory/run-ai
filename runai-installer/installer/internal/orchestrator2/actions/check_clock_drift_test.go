package actions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckClockDrift_InjectedNow exercises the action against a fixed reference
// clock so the offset classification (behind/ahead) is deterministic.
func TestCheckClockDrift_InjectedNow(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	probe := &harness.FakeClusterProbe{Heartbeats: []orchestrator2.NodeHeartbeat{
		{Name: "in-sync", RenewTime: base.Add(-5 * time.Second)},
		{Name: "behind", RenewTime: base.Add(-5 * time.Minute)},
		{Name: "ahead", RenewTime: base.Add(5 * time.Minute)},
	}}
	a := CheckClockDrift{Now: func() time.Time { return base }}

	res, err := a.Execute(context.Background(), orchestrator2.Deps{ClusterProbe: probe})
	require.NoError(t, err)
	assert.Equal(t, orchestrator2.StatusWarn, res.Status)

	nodes, ok := res.Data["driftNodes"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, nodes, 2, "only the out-of-sync nodes are reported")

	offsets := map[string]string{}
	for _, n := range nodes {
		offsets[n["node"].(string)] = n["offset"].(string)
	}
	assert.Contains(t, offsets["behind"], "behind this machine")
	assert.Contains(t, offsets["ahead"], "ahead of this machine")
	assert.NotContains(t, offsets, "in-sync")
	assert.Contains(t, res.Data["reference"].(string), "2026-01-01T12:00:00Z")
}

func TestCheckClockDrift_HeartbeatError(t *testing.T) {
	probe := &harness.FakeClusterProbe{NodeHeartbeatsErr: errors.New("boom")}
	res, err := CheckClockDrift{}.Execute(context.Background(), orchestrator2.Deps{ClusterProbe: probe})
	require.NoError(t, err)
	assert.Equal(t, orchestrator2.StatusWarn, res.Status)
	assert.Contains(t, res.Reason, "could not read node heartbeats")
}
