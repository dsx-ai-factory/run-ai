/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
)

var snakeCaseRE = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

func Lint(register func(*Registry)) []error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv := mcp.NewServer(&mcp.Implementation{Name: "lint", Version: "0"}, nil)
	r := New(srv, Options{Logger: zerolog.New(io.Discard)})
	register(r)

	clientT, serverT := mcp.NewInMemoryTransports()
	go func() { _ = srv.Run(ctx, serverT) }()

	client := mcp.NewClient(&mcp.Implementation{Name: "lint-client", Version: "0"}, nil)
	session, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		return []error{fmt.Errorf("lint: connect in-memory client: %w", err)}
	}
	defer session.Close()

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		return []error{fmt.Errorf("lint: tools/list: %w", err)}
	}

	var errs []error
	for _, t := range listed.Tools {
		errs = append(errs, lintName(t.Name)...)
		errs = append(errs, lintInputDescriptions(t)...)
		errs = append(errs, lintOutputDescriptions(t)...)
	}
	return errs
}

func lintName(name string) []error {
	if !snakeCaseRE.MatchString(name) {
		return []error{fmt.Errorf(
			"tool name %q is not snake_case (expected match against %s)",
			name, snakeCaseRE.String(),
		)}
	}
	return nil
}

func lintInputDescriptions(t *mcp.Tool) []error {
	schema, err := decodeToolSchema(t.Name, "input", t.InputSchema)
	if err != nil {
		return []error{err}
	}
	if schema == nil {
		return nil
	}
	return lintSchemaProperties(t.Name, "input", "", schema)
}

func lintOutputDescriptions(t *mcp.Tool) []error {
	schema, err := decodeToolSchema(t.Name, "output", t.OutputSchema)
	if err != nil {
		return []error{err}
	}
	if schema == nil {
		return nil
	}
	return lintSchemaProperties(t.Name, "output", "", schema)
}

func lintSchemaProperties(toolName, side, path string, schema *jsonschema.Schema) []error {
	var errs []error
	for propName, prop := range schema.Properties {
		fullPath := propName
		if path != "" {
			fullPath = path + "." + propName
		}
		if prop == nil || strings.TrimSpace(prop.Description) == "" {
			errs = append(errs, fmt.Errorf(
				"tool %q %s property %q is missing a non-empty description (set via `jsonschema:\"...\"` on the struct field)",
				toolName, side, fullPath,
			))
		}
		if prop != nil {
			if len(prop.Properties) > 0 {
				errs = append(errs, lintSchemaProperties(toolName, side, fullPath, prop)...)
			}
			if prop.Items != nil && len(prop.Items.Properties) > 0 {
				errs = append(errs, lintSchemaProperties(toolName, side, fullPath+"[]", prop.Items)...)
			}
		}
	}
	return errs
}

func decodeToolSchema(toolName, kind string, raw any) (*jsonschema.Schema, error) {
	switch schema := raw.(type) {
	case nil:
		return nil, nil
	case *jsonschema.Schema:
		return schema, nil
	default:
		b, err := json.Marshal(schema)
		if err != nil {
			return nil, fmt.Errorf("tool %q %s schema: %w", toolName, kind, err)
		}
		var decoded jsonschema.Schema
		if err := json.Unmarshal(b, &decoded); err != nil {
			return nil, fmt.Errorf("tool %q %s schema: %w", toolName, kind, err)
		}
		return &decoded, nil
	}
}
