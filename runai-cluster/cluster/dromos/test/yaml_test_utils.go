/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// YAMLTestRunner provides utilities for input/output YAML testing
type YAMLTestRunner struct {
	InputsDir   string
	ExpectedDir string
}

// NewYAMLTestRunner creates a new YAMLTestRunner with the specified directories
func NewYAMLTestRunner(inputsDir, expectedDir string) *YAMLTestRunner {
	return &YAMLTestRunner{
		InputsDir:   inputsDir,
		ExpectedDir: expectedDir,
	}
}

// LoadInputFile loads a YAML input file from the inputs directory
func (r *YAMLTestRunner) LoadInputFile(testName string) ([]byte, error) {
	inputPath := fmt.Sprintf("%s/%s.yaml", r.InputsDir, testName)

	var err error
	var inputYaml []byte

	if _, err = os.Stat(inputPath); err != nil {
		return nil, fmt.Errorf("expecting to find input file %s: %w", inputPath, err)
	}

	if inputYaml, err = os.ReadFile(inputPath); err != nil {
		return nil, fmt.Errorf("failed to read input file %s: %w", inputPath, err)
	}

	return inputYaml, nil
}

// LoadExpectedFile loads expected YAML output from the expected directory
// If the expected file doesn't exist, it creates it from the actual output (for bootstrapping new tests)
// Returns: expectedYaml, wasCreated, error
func (r *YAMLTestRunner) LoadExpectedFile(testName string, actualYaml *bytes.Buffer) ([]byte, bool, error) {
	expectedPath := fmt.Sprintf("%s/%s.yaml", r.ExpectedDir, testName)

	if _, err := os.Stat(expectedPath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(r.ExpectedDir, 0755)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create expected directory: %w", err)
		}
		err = os.WriteFile(expectedPath, actualYaml.Bytes(), 0644)
		if err != nil {
			return nil, false, fmt.Errorf("expected file not found and failed to create it: %w", err)
		}
		return actualYaml.Bytes(), true, nil // Created new file
	}

	expectedYaml, err := os.ReadFile(expectedPath)
	if err != nil {
		return nil, false, fmt.Errorf("could not load expected file %s: %w", expectedPath, err)
	}

	return expectedYaml, false, nil
}

// ObjectsToYaml converts a list of Kubernetes objects to a YAML string
func ObjectsToYaml(objects []client.Object) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	for i, obj := range objects {
		if i > 0 {
			buf.WriteString("---\n")
		}

		// Populate TypeMeta (apiVersion and kind) from GVK
		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk.Kind != "" {
			obj.GetObjectKind().SetGroupVersionKind(gvk)
		}

		yamlBytes, err := yaml.Marshal(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal object to yaml: %w", err)
		}
		buf.Write(yamlBytes)
	}

	return &buf, nil
}

// YAMLTestConfig holds configuration for a single test case
type YAMLTestConfig struct {
	TestName              string
	TestExpectedError     string          // If set, expect this error field path (partial match)
	TestExpectedErrorType field.ErrorType // Expected error type (e.g., ErrorTypeRequired, ErrorTypeInvalid)
	ExpectValidation      bool            // If true, run validation and expect error
	Focus                 bool
	EnvVars               map[string]string
}

// FilterYAMLTests checks if there are tests marked as focus, returns only them; otherwise returns all tests
func FilterYAMLTests(tests []YAMLTestConfig) []YAMLTestConfig {
	var filterTests []YAMLTestConfig
	for _, curTest := range tests {
		if curTest.Focus {
			filterTests = append(filterTests, curTest)
		}
	}
	if filterTests == nil {
		return tests
	}
	return filterTests
}
