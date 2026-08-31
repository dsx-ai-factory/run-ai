/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimcp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
)

func declineToNonError() mcp.Middleware {
	prefix := string(catDeclined) + ":"
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			res, err := next(ctx, method, req)
			if err != nil || method != "tools/call" {
				return res, err
			}
			ctr, ok := res.(*mcp.CallToolResult)
			if !ok || !ctr.IsError {
				return res, err
			}
			for _, c := range ctr.Content {
				if tc, ok := c.(*mcp.TextContent); ok && strings.HasPrefix(tc.Text, prefix) {
					ctr.IsError = false
					break
				}
			}
			return res, err
		}
	}
}

func withRecover[In ToolInput, Out any](h Handler[In, Out], log zerolog.Logger) Handler[In, Out] {
	return func(ctx context.Context, in In) (out Out, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Str("panic_type", fmt.Sprintf("%T", r)).
					Bytes("stack", debug.Stack()).
					Msg("runaimcp: panic in tool handler")
				var zero Out
				out = zero
				err = ErrInternal
			}
		}()
		return h(ctx, in)
	}
}

func withValidate[In ToolInput, Out any](h Handler[In, Out]) Handler[In, Out] {
	return func(ctx context.Context, in In) (Out, error) {
		if err := in.Validate(); err != nil {
			var zero Out
			return zero, validationError{inner: err}
		}
		return h(ctx, in)
	}
}

func withClassifyErrors[In ToolInput, Out any](h Handler[In, Out]) Handler[In, Out] {
	return func(ctx context.Context, in In) (Out, error) {
		out, err := h(ctx, in)
		if err == nil {
			return out, nil
		}
		cat := classify(err)
		if cat == catInternal {
			var zero Out
			return zero, fmt.Errorf("%s: %w", cat, ErrInternal)
		}
		return out, fmt.Errorf("%s: %w", cat, err)
	}
}

// withLogging logs one "tool call" line per call and attaches a tool-tagged
// logger to ctx for handlers (via zerolog.Ctx(ctx)). Both auto-include "tool". E.g.:
//
//	{"level":..., "tool":..., "duration":..., "category":..., "output_bytes"|"error":..., "message":...}
func withLogging[In ToolInput, Out any](h Handler[In, Out], log zerolog.Logger, toolName string) Handler[In, Out] {
	return func(ctx context.Context, in In) (Out, error) {
		toolLog := log.With().Str("tool", toolName).Logger()
		ctx = toolLog.WithContext(ctx)

		start := time.Now()
		out, err := h(ctx, in)
		dur := time.Since(start)

		cat := classify(err)
		var evt *zerolog.Event
		switch {
		case err == nil:
			evt = toolLog.Info()
		case cat == catInternal:
			evt = toolLog.Error()
		default:
			evt = toolLog.Warn()
		}

		evt = evt.
			Dur("duration", dur).
			Str("category", string(cat))
		if sid := sessionIDFromContext(ctx); sid != "" {
			evt = evt.Str("session_id", sid)
		}
		if err == nil {
			if b, mErr := json.Marshal(out); mErr == nil {
				evt = evt.Int("output_bytes", len(b))
			}
		} else {
			evt = evt.Err(err)
		}
		evt.Msg("tool call")
		return out, err
	}
}

type sessionIDCtxKey struct{}

func withSessionID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionIDCtxKey{}, id)
}

func sessionIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionIDCtxKey{}).(string)
	return v
}

type Elicitor interface {
	Elicit(ctx context.Context, params *mcp.ElicitParams) (*mcp.ElicitResult, error)
	InitializeParams() *mcp.InitializeParams
	ID() string
}

func ClientCanElicitForm(sess Elicitor) bool {
	if sess == nil {
		return false
	}
	ip := sess.InitializeParams()
	if ip == nil || ip.Capabilities == nil || ip.Capabilities.Elicitation == nil {
		return false
	}
	e := ip.Capabilities.Elicitation
	return e.Form != nil || e.URL == nil
}

var (
	elicitInFlightMu sync.Mutex
	elicitInFlight   = map[string]struct{}{}
)

func BeginElicit(sess Elicitor) (release func(), ok bool) {
	if sess == nil {
		return func() {}, true
	}
	id := sess.ID()
	if id == "" {
		return func() {}, true
	}
	elicitInFlightMu.Lock()
	defer elicitInFlightMu.Unlock()
	if _, busy := elicitInFlight[id]; busy {
		return nil, false
	}
	elicitInFlight[id] = struct{}{}
	return func() {
		elicitInFlightMu.Lock()
		delete(elicitInFlight, id)
		elicitInFlightMu.Unlock()
	}, true
}

type sessionCtxKey struct{}

func WithSession(ctx context.Context, s Elicitor) context.Context {
	if s == nil {
		return ctx
	}
	return context.WithValue(ctx, sessionCtxKey{}, s)
}

func SessionFromContext(ctx context.Context) Elicitor {
	v, _ := ctx.Value(sessionCtxKey{}).(Elicitor)
	return v
}
