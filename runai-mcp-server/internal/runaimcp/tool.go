package runaimcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
)

type ToolInput interface {
	Validate() error
}

type Handler[In ToolInput, Out any] func(ctx context.Context, in In) (Out, error)

type Tool[In ToolInput, Out any] struct {
	Name          string
	Description   string
	ReadOnly      bool
	Destructive   bool
	MinAPIVersion string
	MaxAPIVersion string
	Handler       Handler[In, Out]
}

type Options struct {
	Logger     zerolog.Logger
	ReadOnly   bool
	APIVersion string
}

type Registry struct {
	server *mcp.Server
	opts   Options
	names  []string
}

func New(srv *mcp.Server, opts Options) *Registry {
	srv.AddReceivingMiddleware(declineToNonError())
	return &Registry{server: srv, opts: opts}
}

func (r *Registry) ToolNames() []string {
	return r.names
}

// SupportsAPIVersion reports whether the detected control-plane version is in
// the requested range. Empty and development runtime versions remain compatible,
// matching the registration behavior used by version-gated tools.
func (r *Registry) SupportsAPIVersion(minVersion, maxVersion string) bool {
	if r == nil {
		panic("runaimcp.Registry.SupportsAPIVersion: nil Registry")
	}
	return versionInRange(r.opts.APIVersion, minVersion, maxVersion, "capability check")
}

func Register[In ToolInput, Out any](registry *Registry, tool Tool[In, Out]) {
	if registry == nil {
		panic("runaimcp.Register: nil Registry")
	}
	if registry.opts.ReadOnly && !tool.ReadOnly {
		return
	}
	if !versionInRange(registry.opts.APIVersion, tool.MinAPIVersion, tool.MaxAPIVersion, tool.Name) {
		return
	}
	if tool.Handler == nil {
		panic(fmt.Sprintf("runaimcp.Register: tool %q has nil Handler", tool.Name))
	}

	h := tool.Handler
	h = withValidate(h)
	h = withRecover(h, registry.opts.Logger)
	h = withClassifyErrors(h)
	h = withLogging(h, registry.opts.Logger, tool.Name)

	registry.names = append(registry.names, tool.Name)
	mcp.AddTool(registry.server,
		&mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint:    tool.ReadOnly,
				DestructiveHint: boolPtr(tool.Destructive),
			},
		},
		func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
			if req != nil && req.Session != nil {
				ctx = withSessionID(ctx, req.Session.ID())
				ctx = WithSession(ctx, req.Session)
			}
			out, err := h(ctx, in)
			return nil, out, err
		},
	)
}

func boolPtr(b bool) *bool { return &b }
