package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	// Import the main RunaiConfigSpec type
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	sdkschema "github.com/run-ai/runai/runai-cluster/cluster/sdk/schema"
)

const (
	outputFileName = "values.schema.json"
	description    = "JSON Schema for RunAI Helm Chart Values based on RunaiConfigSpec"
)

// ValuesSchema represents the structure for Helm values that maps to RunaiConfigSpec
type ValuesSchema struct {
	ClusterConfig *runaiv1.RunaiConfigSpec `json:"clusterConfig,omitempty" jsonschema:"title=RunAI Cluster Configuration,description=RunAI cluster configuration"`
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		return
	}

	// Determine output path
	outputPath := outputFileName
	if len(os.Args) > 1 {
		outputPath = os.Args[1]
	}

	// Generate the schema
	if err := generateSchema(outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated JSON schema: %s\n", outputPath)
}

func generateSchema(outputPath string) error {
	// Generate the schema
	schema := sdkschema.Generate()

	// Marshal to pretty-printed JSON for file output
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling schema to JSON: %w", err)
	}

	// Create output directory if needed
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Write the schema to file
	if err := os.WriteFile(outputPath, schemaBytes, 0644); err != nil {
		return fmt.Errorf("writing schema file: %w", err)
	}

	return nil
}

func printHelp() {
	fmt.Printf(`Generate values.schema.json for RunAI Helm Chart

Usage:
  %s [output-path]

Arguments:
  output-path    Path where to write the schema file (default: %s)

Description:
  This tool generates a JSON Schema file that can be used to validate
  RunAI Helm chart values. The schema is based on the RunaiConfigSpec
  Go struct and includes all its nested types.

  The generated schema follows the structure:
  {
    "clusterConfig": {
      // RunaiConfigSpec fields here
    }
  }

Examples:
  %s                                    # Generate %s
  %s values.schema.json                 # Generate values.schema.json  
  %s ../runai-chart/values.schema.json  # Generate in chart directory

`, os.Args[0], outputFileName, os.Args[0], outputFileName, os.Args[0], os.Args[0])
}
