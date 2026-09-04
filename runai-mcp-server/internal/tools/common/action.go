/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	authz "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

type WriteAction string

const (
	ActionCreate WriteAction = "create"
	ActionUpdate WriteAction = "update"
	ActionDelete WriteAction = "delete"
)

func (a WriteAction) Valid() bool {
	switch a {
	case ActionCreate, ActionUpdate, ActionDelete:
		return true
	default:
		return false
	}
}

func (a WriteAction) Authz() authz.Action {
	switch a {
	case ActionCreate:
		return authz.ACTION_CREATE
	case ActionUpdate:
		return authz.ACTION_UPDATE
	case ActionDelete:
		return authz.ACTION_DELETE
	default:
		// Fail closed: an unknown action maps to no valid scope, so the
		// permitted-scopes probe rejects it rather than silently probing delete.
		return ""
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, runaimcp.ErrNotFound)
}

// ConfirmRequiredError is the gate a non-dry-run mutation returns without confirm=true: it
// carries the target snapshot so the caller can review before re-issuing with confirm. action
// is the verb shown to the model (e.g. "delete", "update") so it isn't told the wrong operation.
func ConfirmRequiredError(action, kind, name, id string, snapshot map[string]any) error {
	target := fmt.Sprintf("%s of %s %q", action, kind, name)
	if id != "" {
		target = fmt.Sprintf("%s (%s)", target, id)
	}
	return fmt.Errorf("%s requires confirm=true; snapshot: %v: %w", target, snapshot, runaimcp.ErrConfirmationRequired)
}

func WithConfirmWarning(details map[string]any, warning string) map[string]any {
	out := make(map[string]any, len(details)+1)
	for k, v := range details {
		out[k] = v
	}
	out["warning"] = warning
	return out
}

const defaultConfirmTimeout = 5 * time.Minute

type confirmOptions struct {
	message string
	timeout time.Duration
}

type ConfirmOption func(*confirmOptions)

func WithConfirmMessage(message string) ConfirmOption {
	return func(o *confirmOptions) { o.message = message }
}

func WithConfirmTimeout(d time.Duration) ConfirmOption {
	return func(o *confirmOptions) { o.timeout = d }
}

func ConfirmWrite(ctx context.Context, confirmed bool, action, kind, name, id string, details map[string]any, opts ...ConfirmOption) error {
	o := confirmOptions{timeout: defaultConfirmTimeout}
	for _, fn := range opts {
		fn(&o)
	}

	sess := runaimcp.SessionFromContext(ctx)
	if !runaimcp.ClientCanElicitForm(sess) {
		if confirmed {
			return nil
		}
		return ConfirmRequiredError(action, kind, name, id, details)
	}

	release, ok := runaimcp.BeginElicit(sess)
	if !ok {
		return fmt.Errorf("%s of %s %q (%s): another confirmation is already awaiting the user in this session; retry after it completes: %w",
			action, kind, name, id, runaimcp.ErrConfirmationRequired)
	}
	defer release()

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	res, err := sess.Elicit(ctx, &mcp.ElicitParams{
		Mode:            "form",
		Message:         confirmMessage(o.message, action, kind, name, id, details),
		RequestedSchema: &jsonschema.Schema{Type: "object", Properties: map[string]*jsonschema.Schema{}},
	})
	if err != nil {
		return fmt.Errorf("%s of %s %q (%s): confirmation prompt did not complete: %v: %w",
			action, kind, name, id, err, runaimcp.ErrConfirmationRequired)
	}
	if res.Action != "accept" {
		return fmt.Errorf("%s of %s %q (%s) was declined by the user; do not retry or re-prompt: %w",
			action, kind, name, id, runaimcp.ErrDeclined)
	}
	return nil
}

const (
	maxConfirmFields = 8
	maxConfirmLen    = 600
)

func confirmMessage(custom, action, kind, name, id string, details map[string]any) string {
	header := custom
	if header == "" {
		header = fmt.Sprintf("Confirm %s of %s %q?", action, kind, name)
		if id != "" {
			header = fmt.Sprintf("Confirm %s of %s %q (%s)?", action, kind, name, id)
		}
	}
	if len(details) == 0 {
		return header
	}
	lines := summarizeDetails(details)
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")
	shown := 0
	for _, ln := range lines {
		if shown >= maxConfirmFields || b.Len()+len(ln)+1 > maxConfirmLen {
			break
		}
		b.WriteString(ln)
		b.WriteByte('\n')
		shown++
	}
	if shown < len(lines) {
		fmt.Fprintf(&b, "…and %d more.", len(lines)-shown)
	}
	return strings.TrimRight(b.String(), "\n")
}

func summarizeDetails(details map[string]any) []string {
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		if m, ok := details[k].(map[string]any); ok {
			before, hasB := m["before"]
			after, hasA := m["after"]
			if hasB && hasA && len(m) == 2 {
				lines = append(lines, fmt.Sprintf("  %s: %s → %s", k, compactValue(before), compactValue(after)))
				continue
			}
		}
		lines = append(lines, fmt.Sprintf("  %s: %s", k, compactValue(details[k])))
	}
	return lines
}

func compactValue(v any) string {
	const max = 60
	s := fmt.Sprintf("%v", v)
	if b, err := json.Marshal(v); err == nil {
		s = strings.Trim(string(b), `"`)
	}
	if r := []rune(s); len(r) > max {
		s = string(r[:max]) + "…"
	}
	return s
}
