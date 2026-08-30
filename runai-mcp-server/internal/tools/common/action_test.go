package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func formCaps() *mcp.ElicitationCapabilities { return &mcp.ElicitationCapabilities{} }

type fakeElicitor struct {
	caps   *mcp.ElicitationCapabilities
	result *mcp.ElicitResult
	err    error
	block  bool
	called bool
	params *mcp.ElicitParams
}

func (f *fakeElicitor) Elicit(ctx context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	f.called = true
	f.params = params
	if f.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return f.result, f.err
}

func (f *fakeElicitor) InitializeParams() *mcp.InitializeParams {
	return &mcp.InitializeParams{Capabilities: &mcp.ClientCapabilities{Elicitation: f.caps}}
}

func (f *fakeElicitor) ID() string { return "fake" }

func ctxWithElicitor(f runaimcp.Elicitor) context.Context {
	return runaimcp.WithSession(context.Background(), f)
}

func TestConfirmWrite_Accept(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), result: &mcp.ElicitResult{Action: "accept"}}
	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "node pool", "gpu-a", "np-1", map[string]any{"name": "gpu-a"})
	require.NoError(t, err)
	require.True(t, f.called)

	assert.Equal(t, "form", f.params.Mode)
	assert.NotNil(t, f.params.RequestedSchema)
	assert.Contains(t, f.params.Message, "delete")
	assert.Contains(t, f.params.Message, "name: gpu-a")
}

func TestConfirmWrite_Decline(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), result: &mcp.ElicitResult{Action: "decline"}}
	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "node pool", "gpu-a", "np-1", nil)
	require.ErrorIs(t, err, runaimcp.ErrDeclined)
	require.NotErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "declined")
}

func TestConfirmWrite_Cancel(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), result: &mcp.ElicitResult{Action: "cancel"}}
	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "node pool", "gpu-a", "np-1", nil)
	require.ErrorIs(t, err, runaimcp.ErrDeclined)
	require.NotErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "declined")
}

func TestConfirmWrite_ConfirmTrueDoesNotBypassElicitation(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), result: &mcp.ElicitResult{Action: "decline"}}
	err := ConfirmWrite(ctxWithElicitor(f), true, "delete", "node pool", "gpu-a", "np-1", nil)
	require.True(t, f.called, "elicitation must run even when confirm=true")
	require.ErrorIs(t, err, runaimcp.ErrDeclined)
}

func TestConfirmWrite_ElicitErrorMidPrompt_AbortsEvenIfConfirmed(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), err: errors.New("connection closed")}
	err := ConfirmWrite(ctxWithElicitor(f), true, "delete", "node pool", "gpu-a", "np-1", nil)
	require.True(t, f.called)
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "did not complete")
}

func TestConfirmWrite_ElicitErrorMidPrompt_NotConfirmedAborts(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), err: errors.New("connection closed")}
	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "node pool", "gpu-a", "np-1", nil)
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "did not complete")
}

func TestConfirmWrite_TimesOutAndAborts(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), block: true}
	err := ConfirmWrite(ctxWithElicitor(f), true, "delete", "node pool", "gpu-a", "np-1", nil, WithConfirmTimeout(20*time.Millisecond))
	require.True(t, f.called)
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "did not complete")
}

func TestConfirmWrite_NoCapability_ConfirmedProceeds(t *testing.T) {
	f := &fakeElicitor{caps: nil}
	err := ConfirmWrite(ctxWithElicitor(f), true, "delete", "node pool", "gpu-a", "np-1", nil)
	require.NoError(t, err)
	assert.False(t, f.called, "must not attempt a prompt a client can't show")
}

func TestConfirmWrite_NoCapability_NotConfirmedFallsBackToGate(t *testing.T) {
	f := &fakeElicitor{caps: nil}
	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "node pool", "gpu-a", "np-1", map[string]any{"name": "gpu-a"})
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "confirm=true")
	assert.False(t, f.called)
}

func TestConfirmWrite_RejectsConcurrentElicitation(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), result: &mcp.ElicitResult{Action: "accept", Content: map[string]any{"confirm": true}}}
	release, ok := runaimcp.BeginElicit(f)
	require.True(t, ok)
	defer release()

	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "node pool", "gpu-a", "np-1", nil)
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "already awaiting")
	assert.False(t, f.called, "must not open a second form while one is in flight")
}

func TestConfirmWrite_NoSession_ConfirmedProceeds(t *testing.T) {
	err := ConfirmWrite(context.Background(), true, "delete", "node pool", "gpu-a", "np-1", nil)
	require.NoError(t, err)
}

func TestConfirmWrite_NoSession_NotConfirmedFallsBackToGate(t *testing.T) {
	err := ConfirmWrite(context.Background(), false, "delete", "node pool", "gpu-a", "np-1", map[string]any{"name": "gpu-a"})
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), "confirm=true")
}

func TestConfirmRequiredErrorOmitsEmptyID(t *testing.T) {
	err := ConfirmRequiredError("create", "workload", "training", "", nil)
	require.ErrorIs(t, err, runaimcp.ErrConfirmationRequired)
	assert.Contains(t, err.Error(), `create of workload "training" requires confirm=true`)
	assert.NotContains(t, err.Error(), "()")
}

func TestConfirmMessage_OmitsEmptyDetails(t *testing.T) {
	msg := confirmMessage("", "delete", "node pool", "gpu-a", "np-1", nil)
	assert.Equal(t, `Confirm delete of node pool "gpu-a" (np-1)?`, msg)
}

func TestConfirmMessage_OmitsEmptyID(t *testing.T) {
	msg := confirmMessage("", "create", "project", "ml-team", "", nil)
	assert.Equal(t, `Confirm create of project "ml-team"?`, msg)
}

func TestConfirmMessage_RendersDiffCompactly(t *testing.T) {
	details := map[string]any{
		"description": map[string]any{"before": "", "after": "new"},
		"gpuDeserved": map[string]any{"before": 2.0, "after": 4.0},
	}
	msg := confirmMessage("", "update", "project", "seed-proj-1", "4500003", details)
	assert.Contains(t, msg, "description:  → new")
	assert.Contains(t, msg, "gpuDeserved: 2 → 4")
	assert.NotContains(t, msg, "\n  \"", "must not be indented JSON")
}

func TestConfirmMessage_BoundsLargeDiff(t *testing.T) {
	details := map[string]any{}
	for i := 0; i < 40; i++ {
		details[fmt.Sprintf("field_%02d", i)] = map[string]any{"before": i, "after": i + 1}
	}
	msg := confirmMessage("", "update", "project", "p", "1", details)
	assert.Less(t, len(msg), 700, "message must stay short enough to fit the prompt")
	assert.Contains(t, msg, "more")
	assert.NotContains(t, msg, "dryRun")
}

func TestConfirmMessage_UsesCustomHeader(t *testing.T) {
	msg := confirmMessage("Are you sure you want to delete demo?", "delete", "project", "demo", "123", map[string]any{"name": "demo"})
	assert.Contains(t, msg, "Are you sure you want to delete demo?")
	assert.NotContains(t, msg, "Confirm delete")
	assert.Contains(t, msg, "name: demo")
}

func TestConfirmWrite_WithCustomMessage(t *testing.T) {
	f := &fakeElicitor{caps: formCaps(), result: &mcp.ElicitResult{Action: "accept", Content: map[string]any{"confirm": true}}}
	err := ConfirmWrite(ctxWithElicitor(f), false, "delete", "project", "demo", "123", nil, WithConfirmMessage("Are you sure?"))
	require.NoError(t, err)
	assert.Equal(t, "Are you sure?", f.params.Message)
}

func TestCompactValue_TruncatesOnRuneBoundary(t *testing.T) {
	out := compactValue(strings.Repeat("é", 100))
	assert.True(t, utf8.ValidString(out), "must not cut a multibyte rune")
	assert.Equal(t, 61, len([]rune(out)), "60 runes plus the ellipsis")
}

func TestWithConfirmWarning_CopiesAndAddsWarning(t *testing.T) {
	before := map[string]any{"name": "ml-team"}
	got := WithConfirmWarning(before, "deletes all workloads")
	assert.Equal(t, "deletes all workloads", got["warning"])
	assert.Equal(t, "ml-team", got["name"])
	_, mutated := before["warning"]
	assert.False(t, mutated, "must not mutate the caller's map")
}
