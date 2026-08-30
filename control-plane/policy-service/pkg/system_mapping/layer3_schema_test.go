package system_mapping_test

// OpenAPI v3 schema loading and validation for Layer 3+4 of the system-mapping
// test plan. Used to validate that the manifest produced by running the
// "set-everything" policy over a skeleton actually satisfies the Kubernetes
// API server's schema (enums, formats, required fields, etc.).
//
// The K8s OpenAPI v3 documents under testing/crd/ are full API-group docs
// containing many inter-referencing schemas (a Deployment reference pulls in
// PodSpec, which pulls in Container, Volume, etc.). The kube-openapi
// validator does not resolve `$ref` itself, so we pre-flatten the target
// schema by recursively inlining every reference under `components.schemas`
// before handing it to the validator. Cycles are broken by leaving a
// permissive empty object behind on the second visit.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	openapierrors "k8s.io/kube-openapi/pkg/validation/errors"
	openapispec "k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

const openAPIRefPrefix = "#/components/schemas/"

// loadOpenAPISchema reads an OpenAPI v3 document at path, looks up the schema
// keyed by name (e.g. "io.k8s.api.apps.v1.Deployment"), inlines every `$ref`
// reachable from it, and returns the resulting self-contained schema ready
// for validation.
func loadOpenAPISchema(path, name string) (*openapispec.Schema, error) {
	return loadOpenAPISchemaWithSynthetic(path, name, nil)
}

// loadOpenAPISchemaWithSynthetic is loadOpenAPISchema with support for fictive
// GVKs: when syntheticSpecRefs is non-empty the manifest wrapper schema is
// synthesized in code (see resolveTargetSchema) rather than looked up by name.
func loadOpenAPISchemaWithSynthetic(path, name string, syntheticSpecRefs map[string]string) (*openapispec.Schema, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	if schemas == nil {
		return nil, fmt.Errorf("%s has no components.schemas", path)
	}

	target, err := resolveTargetSchema(path, name, schemas, syntheticSpecRefs)
	if err != nil {
		return nil, err
	}

	inlined := inlineOpenAPIRefs(target, schemas, map[string]bool{})

	jsonBytes, err := json.Marshal(inlined)
	if err != nil {
		return nil, fmt.Errorf("marshal inlined schema: %w", err)
	}

	var schema openapispec.Schema
	if err := json.Unmarshal(jsonBytes, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal into spec.Schema: %w", err)
	}
	return &schema, nil
}

// resolveTargetSchema returns the schema to validate against. Normally it is
// the schema keyed by name in the OpenAPI document. For fictive GVKs (the
// run.ai/* wrappers) whose manifest schema does not exist in the read-only
// swagger doc, syntheticSpecRefs is non-empty and we synthesize the manifest
// wrapper here in code: a plain object with apiVersion/kind/metadata plus one
// $ref per mapped field (e.g. spec -> WorkspaceSpecSpec) pointing at the real
// component schema. inlineOpenAPIRefs then resolves those refs against the
// document's schemas just like any other reference.
func resolveTargetSchema(path, name string, schemas map[string]any, syntheticSpecRefs map[string]string) (any, error) {
	if len(syntheticSpecRefs) == 0 {
		target, exists := schemas[name]
		if !exists {
			return nil, fmt.Errorf("%s: schema %q not found in components.schemas", path, name)
		}
		return target, nil
	}

	props := map[string]any{
		"apiVersion": map[string]any{"type": "string"},
		"kind":       map[string]any{"type": "string"},
		"metadata":   map[string]any{"type": "object"},
	}
	// Each key is a dot-path (e.g. "spec" or "spec.worker") into the manifest;
	// walk/create the intermediate object schemas and place the $ref at the leaf
	// so a nested field validates against its real component schema.
	for fieldPath, refName := range syntheticSpecRefs {
		if _, ok := schemas[refName]; !ok {
			return nil, fmt.Errorf("%s: synthetic ref target %q for field %q not found in components.schemas", path, refName, fieldPath)
		}
		segments := strings.Split(fieldPath, ".")
		cursor := props
		for i, seg := range segments {
			if i == len(segments)-1 {
				cursor[seg] = map[string]any{"$ref": openAPIRefPrefix + refName}
				break
			}
			child, ok := cursor[seg].(map[string]any)
			if !ok {
				child = map[string]any{"type": "object", "properties": map[string]any{}}
				cursor[seg] = child
			}
			childProps, ok := child["properties"].(map[string]any)
			if !ok {
				childProps = map[string]any{}
				child["properties"] = childProps
			}
			cursor = childProps
		}
	}
	return map[string]any{"type": "object", "properties": props}, nil
}

// inlineOpenAPIRefs walks v and replaces every `$ref` whose target lives in
// schemas with the referenced schema (recursively, so the result has no refs
// left as long as the input is closed under schemas). A ref currently being
// expanded on the same path is left as the empty object {}: this prevents
// runaway recursion through self-referential types (e.g. JSONSchemaProps)
// at the cost of treating them permissively during validation.
func inlineOpenAPIRefs(v any, schemas map[string]any, expanding map[string]bool) any {
	switch x := v.(type) {
	case map[string]any:
		if ref, ok := x["$ref"].(string); ok && strings.HasPrefix(ref, openAPIRefPrefix) {
			name := strings.TrimPrefix(ref, openAPIRefPrefix)
			if expanding[name] {
				return map[string]any{}
			}
			target, exists := schemas[name]
			if !exists {
				return x
			}
			expanding[name] = true
			result := inlineOpenAPIRefs(target, schemas, expanding)
			delete(expanding, name)
			return result
		}
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = inlineOpenAPIRefs(val, schemas, expanding)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = inlineOpenAPIRefs(val, schemas, expanding)
		}
		return out
	default:
		return v
	}
}

// validateAgainstOpenAPI runs the kube-openapi validator against schema and
// returns its raw error slice (or nil when the object is valid). Any non-empty
// result means the object is rejected by the API-server-side schema.
func validateAgainstOpenAPI(schema *openapispec.Schema, obj any) []error {
	validator := validate.NewSchemaValidator(schema, nil, "", strfmt.Default)
	if validator == nil {
		return nil
	}
	res := validator.Validate(obj)
	if res == nil || res.IsValid() {
		return nil
	}
	return res.Errors
}

// formatOpenAPIErrors renders the raw validator errors into a single
// human-readable block with one line per error, marking validation-typed
// errors with their kube-openapi error code. Returns "" when errs is empty.
func formatOpenAPIErrors(errs []error) string {
	if len(errs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("OpenAPI schema validation errors:\n")
	for _, e := range errs {
		switch v := e.(type) {
		case *openapierrors.Validation:
			fmt.Fprintf(&b, "  - [%d] %s\n", v.Code(), v.Error())
		default:
			fmt.Fprintf(&b, "  - %s\n", e.Error())
		}
	}
	return b.String()
}
