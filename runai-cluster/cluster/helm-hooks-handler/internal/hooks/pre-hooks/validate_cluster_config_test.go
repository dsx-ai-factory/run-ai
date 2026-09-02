/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"encoding/json"
	"strings"
	"testing"

	jsonschema "github.com/invopop/jsonschema"
	ra "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	sdkschema "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/schema"
)

// ValuesSchema mirrors the generator's root wrapper for Helm values (reference only)
type ValuesSchema struct {
	ClusterConfig *runaiv1.RunaiConfigSpec `json:"clusterConfig,omitempty" jsonschema:"title=RunAI Cluster Configuration,description=RunAI cluster configuration"`
}

func loadValuesSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	return sdkschema.Generate()
}

func buildAllFieldsObjectVisited(t *testing.T, root *jsonschema.Schema, sch *jsonschema.Schema, visited map[string]bool) map[string]any {
	t.Helper()
	result := map[string]any{}

	// If the schema node uses a JSON Schema "$ref", follow that reference by name to
	// the appropriate definition in the schema's $defs. This resolves the indirection
	// so the function can enumerate object properties even when the structure is defined
	// via "$ref", which is common for reusable object types.
	if sch.Ref != "" {
		ref := strings.TrimPrefix(sch.Ref, "#/$defs/")
		if visited[ref] {
			return result
		}
		def := root.Definitions[ref]
		if def == nil {
			t.Fatalf("missing schema definition for ref %s", sch.Ref)
		}
		visited[ref] = true
		defer delete(visited, ref)
		sch = def
	}

	// Ensure we are working with an object schema. If not, return an empty map.
	if sch.Type != "object" {
		return result
	}
	props := map[string]jsonschema.Schema{}
	if sch.Properties != nil {
		if b, err := json.Marshal(sch.Properties); err == nil {
			json.Unmarshal(b, &props)
		}
	}

	for name, prop := range props {
		val := buildValueForSchemaVisited(t, root, &prop, visited)
		result[name] = val
	}
	return result
}

func buildValueForSchema(t *testing.T, root *jsonschema.Schema, sch *jsonschema.Schema) any {
	return buildValueForSchemaVisited(t, root, sch, map[string]bool{})
}

// durationDefs lists $defs names for Go types that custom-marshal to a JSON string
// but are reflected as objects by the schema generator.
var durationDefs = map[string]string{
	"Duration": "0s",
}

func buildValueForSchemaVisited(t *testing.T, root *jsonschema.Schema, sch *jsonschema.Schema, visited map[string]bool) any {
	// Resolve $ref by ref name before deref
	if sch.Ref != "" {
		ref := strings.TrimPrefix(sch.Ref, "#/$defs/")
		if s, ok := durationDefs[ref]; ok {
			return s
		}
		if visited[ref] {
			return map[string]any{}
		}
		if root.Definitions[ref] == nil {
			t.Fatalf("missing schema definition for ref %s", sch.Ref)
		}
		visited[ref] = true
		defer delete(visited, ref)
		sch = root.Definitions[ref]
	}

	// Handle oneOf/anyOf by picking the first branch
	if len(sch.OneOf) > 0 {
		return buildValueForSchemaVisited(t, root, sch.OneOf[0], visited)
	}
	if len(sch.AnyOf) > 0 {
		return buildValueForSchemaVisited(t, root, sch.AnyOf[0], visited)
	}

	switch sch.Type {
	case "object":
		return buildAllFieldsObjectVisited(t, root, sch, visited)
	case "array":
		return []any{}
	case "integer":
		return 0
	case "number":
		return 0
	case "boolean":
		return false
	case "string":
		if sch.Format == "date-time" {
			return "1970-01-01T00:00:00Z"
		}
		if sch.Pattern != "" {
			// Minimal quantity-like or regex-constrained string: use "0"
			return "0"
		}
		return ""
	default:
		// Fallback: empty object
		return map[string]any{}
	}
}

func Test_ValuesSchema_AllFields_ValidatesAndUnmarshals(t *testing.T) {
	// 1) Generate the schema via the shared generator
	s := loadValuesSchema(t)
	root := s
	if s.Ref != "" {
		ref := strings.TrimPrefix(s.Ref, "#/$defs/")
		if s.Definitions[ref] == nil {
			t.Fatalf("missing root schema definition for ref %s", s.Ref)
		}
		root = s.Definitions[ref]
	}

	// 2) Build a values document containing all clusterConfig fields with minimal valid values
	clusterConfigSchema, ok := root.Properties.Get("clusterConfig")
	if !ok || clusterConfigSchema == nil {
		t.Fatalf("schema missing clusterConfig property")
	}
	allFields := buildValueForSchema(t, s, clusterConfigSchema)

	// 3) Ensure the same values unmarshal via PatchRunaiConfig (validate-cluster-config path)
	patchBytes, err := json.Marshal(allFields)
	if err != nil {
		t.Fatalf("marshal clusterConfig patch: %v", err)
	}
	dummy := &runaiv1.RunaiConfig{}
	if err := ra.PatchRunaiConfig(dummy, patchBytes); err != nil {
		t.Fatalf("PatchRunaiConfig failed: %v", err)
	}
}
