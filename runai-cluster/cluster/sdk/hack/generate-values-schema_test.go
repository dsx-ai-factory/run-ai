package main

import (
	"encoding/json"
	"testing"

	sdkschema "github.com/run-ai/runai/runai-cluster/cluster/sdk/schema"
)

func TestGenerateSchema_ExcludesInternalFields(t *testing.T) {
	// Generate schema directly without file I/O
	schema := sdkschema.Generate()

	runaiConfigSpec, exists := schema.Definitions["RunaiConfigSpec"]
	if !exists {
		t.Fatalf("$defs missing RunaiConfigSpec")
	}
	if runaiConfigSpec.Properties == nil {
		t.Fatalf("RunaiConfigSpec missing properties field")
	}

	if _, hasInternal := runaiConfigSpec.Properties.Get("__internal"); hasInternal {
		t.Errorf("RunaiConfigSpec properties contains __internal field, but it should be excluded")
	}

	if runaiConfigSpec.Properties.Len() == 0 {
		t.Errorf("RunaiConfigSpec properties is empty - schema generation may have failed")
	}
}

func TestGenerateSchema_ValidJSON(t *testing.T) {
	// Generate schema directly without file I/O
	schema := sdkschema.Generate()

	// Verify required fields are present using native Schema fields
	if schema.Version == "" {
		t.Errorf("Schema missing $schema field")
	}

	if schema.Ref == "" {
		t.Errorf("Schema missing $ref field")
	}

	if schema.Definitions == nil {
		t.Errorf("Schema missing $defs field")
	}

	// Additional test: Ensure schema can be marshaled to valid JSON
	_, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Errorf("Schema cannot be marshaled to JSON: %v", err)
	}
}
