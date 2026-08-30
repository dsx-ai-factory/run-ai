package runaimcp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func textOf(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestRegister_DeclinedRendersAsNonError(t *testing.T) {
	ctx := context.Background()
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	r := New(srv, Options{Logger: discardLogger()})

	Register(r, Tool[inT, outT]{
		Name:        "declines",
		Description: "always declined",
		Handler: func(context.Context, inT) (outT, error) {
			return outT{}, fmt.Errorf("delete of project %q was declined by the user; do not retry or re-prompt: %w", "p", ErrDeclined)
		},
	})
	Register(r, Tool[inT, outT]{
		Name:        "fails",
		Description: "always errors",
		Handler: func(context.Context, inT) (outT, error) {
			return outT{}, fmt.Errorf("boom: %w", ErrUpstream)
		},
	})

	clientT, serverT := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, serverT, nil)
	require.NoError(t, err)
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0"}, nil).Connect(ctx, clientT, nil)
	require.NoError(t, err)
	defer cs.Close()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "declines"})
	require.NoError(t, err)
	assert.False(t, res.IsError, "a user decline must not be an isError result")
	assert.Contains(t, textOf(res), "declined")

	res2, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "fails"})
	require.NoError(t, err)
	assert.True(t, res2.IsError, "genuine errors must still be isError")
}

type inT struct{}

func (inT) Validate() error { return nil }

type outT struct {
	Value string `json:"value"`
}

func newTestRegistry(t *testing.T, readOnly bool) *Registry {
	t.Helper()
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	return New(srv, Options{Logger: discardLogger(), ReadOnly: readOnly})
}

func newTestRegistryWithVersion(t *testing.T, apiVersion string) *Registry {
	t.Helper()
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	return New(srv, Options{Logger: discardLogger(), APIVersion: apiVersion})
}

func TestRegister_AddsToolToServer(t *testing.T) {
	r := newTestRegistry(t, false)
	assert.NotPanics(t, func() {
		Register(r, Tool[inT, outT]{
			Name:        "hello",
			Description: "say hi",
			ReadOnly:    true,
			Handler: func(ctx context.Context, _ inT) (outT, error) {
				return outT{Value: "hi"}, nil
			},
		})
	})
}

func TestRegister_ReadOnlyServerSkipsNonReadOnlyTools(t *testing.T) {
	r := newTestRegistry(t, true)
	// Proof that the gate runs before any handler inspection: a destructive
	// tool with a nil Handler would otherwise panic. Read-only mode must
	// short-circuit before touching it.
	assert.NotPanics(t, func() {
		Register(r, Tool[inT, outT]{
			Name:        "rotate",
			Description: "destructive",
			Destructive: true,
			Handler:     nil,
		})
	})
	// And a read-only tool still registers in read-only mode.
	assert.NotPanics(t, func() {
		Register(r, Tool[inT, outT]{
			Name:        "list",
			Description: "read",
			ReadOnly:    true,
			Handler:     func(ctx context.Context, _ inT) (outT, error) { return outT{}, nil },
		})
	})
}

func TestRegister_PanicsOnNilHandler(t *testing.T) {
	r := newTestRegistry(t, false)
	assert.Panics(t, func() {
		Register(r, Tool[inT, outT]{Name: "bad", Description: "..."})
	})
}

func TestRegister_VersionGating(t *testing.T) {
	// runtime = 2.18.0
	r := newTestRegistryWithVersion(t, "2.18.0")
	// Each of these uses a nil handler on a "skipped" path: if the gate works,
	// Register short-circuits before the nil-handler panic check.
	cases := []Tool[inT, outT]{
		{Name: "min-too-high", MinAPIVersion: "2.20"},
		{Name: "max-too-low", MaxAPIVersion: "2.17"},
	}
	for _, tc := range cases {
		assert.NotPanics(t, func() { Register(r, tc) }, "gate should skip before nil-handler check")
	}
	// Tool in range registers normally.
	assert.NotPanics(t, func() {
		Register(r, Tool[inT, outT]{
			Name:          "in-range",
			MinAPIVersion: "2.10",
			MaxAPIVersion: "2.99",
			Handler:       func(context.Context, inT) (outT, error) { return outT{}, nil },
		})
	})
}

func TestRegister_EmptyRuntimeVersionDisablesGating(t *testing.T) {
	r := newTestRegistryWithVersion(t, "")
	// MinAPIVersion 99.0 would normally skip, but empty runtime disables gating.
	assert.NotPanics(t, func() {
		Register(r, Tool[inT, outT]{
			Name:          "future",
			MinAPIVersion: "99.0",
			Handler:       func(context.Context, inT) (outT, error) { return outT{}, nil },
		})
	})
}

func TestRegister_DevBuildRuntimeVersionDisablesGating(t *testing.T) {
	// Run:ai master/dev clusters report /_version like "0.0.0-<sha>".
	// The gate must treat that as unknown so local tests against master
	// aren't blocked by MinAPIVersion declarations.
	for _, devVersion := range []string{"0.0.0", "0.0.0-dev", "0.0.0-20002916.master-e494a34"} {
		r := newTestRegistryWithVersion(t, devVersion)
		assert.NotPanicsf(t, func() {
			Register(r, Tool[inT, outT]{
				Name:          "future",
				MinAPIVersion: "99.0",
				Handler:       func(context.Context, inT) (outT, error) { return outT{}, nil },
			})
		}, "runtime version %q should disable gating", devVersion)
	}
}

func TestRegister_PanicsOnInvalidMinAPIVersion(t *testing.T) {
	r := newTestRegistryWithVersion(t, "2.18.0")
	assert.Panics(t, func() {
		Register(r, Tool[inT, outT]{
			Name:          "bad",
			MinAPIVersion: "not-a-version",
			Handler:       func(context.Context, inT) (outT, error) { return outT{}, nil },
		})
	})
}

func TestRegister_PanicsOnNilRegistry(t *testing.T) {
	assert.Panics(t, func() {
		Register(nil, Tool[inT, outT]{
			Name:    "x",
			Handler: func(ctx context.Context, _ inT) (outT, error) { return outT{}, nil },
		})
	})
}

func TestToolNames_RegisteredToolsAppear(t *testing.T) {
	r := newTestRegistry(t, false)
	assert.Empty(t, r.ToolNames())

	Register(r, Tool[inT, outT]{Name: "alpha", ReadOnly: true, Handler: func(context.Context, inT) (outT, error) { return outT{}, nil }})
	Register(r, Tool[inT, outT]{Name: "beta", ReadOnly: true, Handler: func(context.Context, inT) (outT, error) { return outT{}, nil }})

	assert.Equal(t, []string{"alpha", "beta"}, r.ToolNames())
}

func TestToolNames_SkippedToolsExcluded(t *testing.T) {
	r := newTestRegistry(t, true) // read-only registry

	// write tool — skipped
	Register(r, Tool[inT, outT]{Name: "delete", ReadOnly: false, Handler: nil})
	// read-only tool — registered
	Register(r, Tool[inT, outT]{Name: "list", ReadOnly: true, Handler: func(context.Context, inT) (outT, error) { return outT{}, nil }})

	assert.Equal(t, []string{"list"}, r.ToolNames())
}

func TestToolNames_VersionSkippedToolsExcluded(t *testing.T) {
	r := newTestRegistryWithVersion(t, "2.18.0")

	// out-of-range tool — skipped, nil Handler is safe because skip happens first
	Register(r, Tool[inT, outT]{Name: "future", MinAPIVersion: "99.0", Handler: nil})
	// in-range tool — registered
	Register(r, Tool[inT, outT]{Name: "current", Handler: func(context.Context, inT) (outT, error) { return outT{}, nil }})

	assert.Equal(t, []string{"current"}, r.ToolNames())
}
