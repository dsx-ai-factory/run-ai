/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var projectRequired = errors.New(errorProjectRequired) //nolint:staticcheck // user facing message

// ManifestData holds the parsed manifest and metadata
type ManifestData struct {
	RawYAML           []byte
	ParsedManifest    map[string]interface{}
	NamespaceFromYAML string
	NameFromYAML      string
	KindFromYAML      string
}

// ReadAndParseManifest reads a YAML file and parses it
func ReadAndParseManifest(filePath string) (*ManifestData, error) {
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(errorFailedToReadManifest, err)
	}

	var manifestMap map[string]interface{}
	err = yaml.Unmarshal(yamlData, &manifestMap)
	if err != nil {
		return nil, fmt.Errorf(errorFailedToParseYAML, err)
	}

	data := &ManifestData{
		RawYAML:        yamlData,
		ParsedManifest: manifestMap,
	}

	// Extract namespace from metadata if present
	if metadata, ok := manifestMap["metadata"].(map[string]interface{}); ok {
		if namespace, ok := metadata["namespace"].(string); ok {
			data.NamespaceFromYAML = namespace
		}
		if name, ok := metadata["name"].(string); ok {
			data.NameFromYAML = name
		}
	}
	if kind, ok := manifestMap["kind"].(string); ok {
		data.KindFromYAML = kind
	}

	return data, nil
}

// resolveProject implements the project resolution hierarchy
func resolveProject(projectFlag string, namespaceFromYAML string, defaultProject string) (string, error) {
	if projectFlag != "" {
		return projectFlag, nil
	}
	// If we have a namespace in the YAML, and the project flag isn't set,
	// the project will be resolved from the yaml in the api.
	if namespaceFromYAML != "" {
		return "", nil
	}

	if defaultProject != "" {
		return defaultProject, nil
	}

	return "", projectRequired
}
