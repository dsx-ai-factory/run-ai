package runaimcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeElicitor struct{}

func (fakeElicitor) Elicit(context.Context, *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	return &mcp.ElicitResult{Action: "accept"}, nil
}

func (fakeElicitor) InitializeParams() *mcp.InitializeParams { return nil }

func (fakeElicitor) ID() string { return "fake" }

func TestClientCanElicitForm(t *testing.T) {
	form := &mcp.ElicitationCapabilities{Form: &mcp.FormElicitationCapabilities{}}
	urlOnly := &mcp.ElicitationCapabilities{URL: &mcp.URLElicitationCapabilities{}}
	cases := []struct {
		name string
		caps *mcp.ElicitationCapabilities
		want bool
	}{
		{"nil-session-handled-separately", form, true},
		{"empty-object-is-form", &mcp.ElicitationCapabilities{}, true},
		{"explicit-form", form, true},
		{"url-only-not-form", urlOnly, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ClientCanElicitForm(capElicitor{tc.caps}))
		})
	}
	assert.False(t, ClientCanElicitForm(nil), "nil session cannot elicit")
	assert.False(t, ClientCanElicitForm(capElicitor{nil}), "no elicitation capability")
}

type capElicitor struct{ caps *mcp.ElicitationCapabilities }

func (capElicitor) Elicit(context.Context, *mcp.ElicitParams) (*mcp.ElicitResult, error) {
	return &mcp.ElicitResult{Action: "accept"}, nil
}

func (c capElicitor) InitializeParams() *mcp.InitializeParams {
	return &mcp.InitializeParams{Capabilities: &mcp.ClientCapabilities{Elicitation: c.caps}}
}

func (capElicitor) ID() string { return "cap" }

func TestSessionFromContext_RoundTrip(t *testing.T) {
	sess := fakeElicitor{}
	ctx := WithSession(context.Background(), sess)
	require.NotNil(t, SessionFromContext(ctx))
	assert.Equal(t, sess, SessionFromContext(ctx))
}

func TestSessionFromContext_Absent(t *testing.T) {
	assert.Nil(t, SessionFromContext(context.Background()))
}

func TestBeginElicit_SingleFlight(t *testing.T) {
	sess := fakeElicitor{}
	release, ok := BeginElicit(sess)
	require.True(t, ok)
	_, ok2 := BeginElicit(sess)
	require.False(t, ok2, "second claim on a busy session must be rejected")
	release()
	release3, ok3 := BeginElicit(sess)
	require.True(t, ok3, "after release the session is free again")
	release3()
}

func TestBeginElicit_NilSessionAllowed(t *testing.T) {
	release, ok := BeginElicit(nil)
	require.True(t, ok)
	release()
}

func TestWithSession_NilIsNoop(t *testing.T) {
	ctx := WithSession(context.Background(), nil)
	assert.Nil(t, SessionFromContext(ctx))
}
