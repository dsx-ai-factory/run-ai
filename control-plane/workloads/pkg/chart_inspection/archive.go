// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	maxCompressedBytes = 32 << 20
	maxExpandedBytes   = 64 << 20
	maxArchiveEntries  = 4096
	maxMetadataBytes   = 1 << 20
	maxYAMLScalarBytes = 4 << 20
)

// Limits apply to expanded values, not only input bytes: aliases can repeat a
// small source node many times before the result is serialized.
type yamlBudget struct {
	nodes int
	bytes int
}

func (b *yamlBudget) charge(value string) error {
	if len(value) > b.bytes {
		return fmt.Errorf("expanded YAML content exceeds 4 MiB")
	}
	b.bytes -= len(value)
	return nil
}

// ParseArchive reads bounded root metadata from a single-root gzip/tar archive.
// It returns errors for unsafe or malformed archives, but reports missing or
// unusable question declarations in Result so callers can distinguish them.
func ParseArchive(data []byte) (*Result, error) {
	if len(data) > maxCompressedBytes {
		return nil, fmt.Errorf("chart archive exceeds 32 MiB")
	}
	source := bytes.NewReader(data)
	gz, err := gzip.NewReader(source)
	if err != nil {
		return nil, fmt.Errorf("read chart gzip: %w", err)
	}
	defer gz.Close()
	gz.Multistream(false)
	limited := &io.LimitedReader{R: gz, N: maxExpandedBytes + 1}
	tr := tar.NewReader(limited)
	files := map[string][]byte{}
	root := ""
	for count := 0; ; count++ {
		header, err := tr.Next()
		if limited.N <= 0 {
			return nil, fmt.Errorf("expanded chart exceeds 64 MiB")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read chart tar: %w", err)
		}
		if count >= maxArchiveEntries {
			return nil, fmt.Errorf("chart archive exceeds 4096 entries")
		}
		name := header.Name
		if strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || path.Clean(name) != strings.TrimSuffix(name, "/") {
			return nil, fmt.Errorf("chart archive contains an unsafe path")
		}
		parts := strings.Split(strings.TrimSuffix(name, "/"), "/")
		for _, part := range parts {
			if part == "" || part == "." || part == ".." || strings.Contains(part, ":") {
				return nil, fmt.Errorf("chart archive contains an unsafe path")
			}
		}
		if root == "" {
			root = parts[0]
		}
		if parts[0] != root {
			return nil, fmt.Errorf("chart archive contains multiple roots")
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return nil, fmt.Errorf("chart archive contains a link or unsupported entry")
		}
		if header.Size < 0 || header.Size > limited.N {
			return nil, fmt.Errorf("expanded chart exceeds 64 MiB")
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if len(parts) != 2 {
			continue
		}
		file := parts[1]
		if file != "Chart.yaml" && file != "questions.yaml" && file != "questions.yml" && file != "values.yaml" {
			continue
		}
		if _, exists := files[file]; exists {
			return nil, fmt.Errorf("duplicate root chart metadata %s", file)
		}
		if header.Size > maxMetadataBytes {
			return nil, fmt.Errorf("%s exceeds 1 MiB", file)
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		files[file] = body
	}
	// Reading through EOF verifies the checksum, including when tar stops at its end marker.
	if err := consumeTarPadding(limited); err != nil {
		return nil, err
	}
	if limited.N <= 0 {
		return nil, fmt.Errorf("expanded chart exceeds 64 MiB")
	}
	if source.Len() != 0 {
		return nil, fmt.Errorf("chart archive contains trailing gzip data")
	}
	if _, yamlExists := files["questions.yaml"]; yamlExists {
		if _, ymlExists := files["questions.yml"]; ymlExists {
			return nil, fmt.Errorf("chart archive contains both root questions.yaml and questions.yml")
		}
	}
	metadata, exists := files["Chart.yaml"]
	if !exists {
		return nil, fmt.Errorf("chart archive has no root Chart.yaml")
	}
	chart, err := yamlObject(metadata)
	if err != nil {
		return nil, fmt.Errorf("parse Chart.yaml: %w", err)
	}
	name, nameOK := chart["name"].(string)
	version, versionOK := chart["version"].(string)
	if !nameOK || strings.TrimSpace(name) == "" || !versionOK || strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("chart metadata requires a chart name and version")
	}
	sum := sha256.Sum256(data)
	out := &Result{ChartName: name, ChartVersion: version, ArchiveSHA256: hex.EncodeToString(sum[:]), QuestionsStatus: QuestionsAbsent, Groups: []map[string]any{}, Questions: []map[string]any{}, Warnings: []Warning{}}
	defaults := map[string]any{}
	if values, ok := files["values.yaml"]; ok {
		defaults, err = yamlObject(values)
		if err != nil {
			// Defaults are advisory; nil suppresses checks that require readable defaults.
			defaults = nil
			out.Warnings = append(out.Warnings, Warning{Code: "UNREADABLE_DEFAULTS", Message: "values.yaml could not be read for default checks", Path: "$"})
		}
	}
	questionsFile := "questions.yaml"
	questions, ok := files[questionsFile]
	if !ok {
		questionsFile = "questions.yml"
		questions, ok = files[questionsFile]
	}
	if !ok {
		return out, nil
	}
	out.HasQuestions = true
	out.QuestionsStatus = QuestionsUnusable
	declaration, err := yamlObject(questions)
	if err != nil {
		out.Warnings = append(out.Warnings, Warning{Code: warningUnusableQuestions, Message: questionsFile + " must contain valid YAML with an object root: " + err.Error(), Path: "$"})
		return out, nil
	}
	parseDeclaration(declaration, defaults, out)
	for index := range out.Warnings {
		if out.Warnings[index].Code == warningUnusableQuestions {
			out.Warnings[index].Message = questionsFile + ": " + out.Warnings[index].Message
		}
	}
	return out, nil
}

// Read through the gzip trailer to validate its checksum and reject payloads
// hidden after the tar end marker; legal zero padding is still accepted.
func consumeTarPadding(reader io.Reader) error {
	buffer := make([]byte, 32<<10)
	for {
		count, err := reader.Read(buffer)
		for _, value := range buffer[:count] {
			if value != 0 {
				return fmt.Errorf("chart tar contains data after its end marker")
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read chart gzip trailer: %w", err)
		}
	}
}

// Decode to nodes before conversion so aliases, duplicate keys and expansion
// budgets are checked before allocating an unrestricted object graph.
func yamlObject(data []byte) (map[string]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		if err == io.EOF {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("decode YAML: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return nil, fmt.Errorf("decode YAML: %w", err)
		}
		return nil, fmt.Errorf("multiple YAML documents are unsupported")
	}
	if len(document.Content) != 1 {
		return nil, fmt.Errorf("YAML must contain an object")
	}
	if document.Content[0].Tag == "!!null" {
		return map[string]any{}, nil
	}
	budget := &yamlBudget{nodes: 50000, bytes: maxYAMLScalarBytes}
	value, err := yamlValue(document.Content[0], map[*yaml.Node]bool{}, 0, budget)
	if err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("YAML must contain an object")
	}
	return object, nil
}

func yamlValue(node *yaml.Node, active map[*yaml.Node]bool, depth int, budget *yamlBudget) (any, error) {
	budget.nodes--
	if depth > 64 || budget.nodes < 0 || active[node] {
		return nil, fmt.Errorf("YAML is too complex or contains an alias cycle")
	}
	active[node] = true
	defer delete(active, node)
	if node.Kind == yaml.AliasNode {
		return yamlValue(node.Alias, active, depth+1, budget)
	}
	switch node.Kind {
	case yaml.MappingNode:
		out := map[string]any{}
		seen := map[string]bool{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || (key.Tag != "!!str" && key.Tag != "!!merge") {
				return nil, fmt.Errorf("YAML object keys must be strings")
			}
			if seen[key.Value] {
				return nil, fmt.Errorf("duplicate YAML key %q", key.Value)
			}
			if err := budget.charge(key.Value); err != nil {
				return nil, err
			}
			seen[key.Value] = true
			value, err := yamlValue(node.Content[i+1], active, depth+1, budget)
			if err != nil {
				return nil, err
			}
			if key.Tag == "!!merge" {
				maps := []any{value}
				if list, ok := value.([]any); ok {
					maps = list
				}
				for _, item := range maps {
					object, ok := item.(map[string]any)
					if !ok {
						return nil, fmt.Errorf("YAML merge requires objects")
					}
					for k, v := range object {
						if _, exists := out[k]; !exists {
							out[k] = v
						}
					}
				}
			} else {
				out[key.Value] = value
			}
		}
		return out, nil
	case yaml.SequenceNode:
		out := make([]any, 0, len(node.Content))
		for _, child := range node.Content {
			value, err := yamlValue(child, active, depth+1, budget)
			if err != nil {
				return nil, err
			}
			out = append(out, value)
		}
		return out, nil
	case yaml.ScalarNode:
		// Aliases revisit scalars, so charge every occurrence before response encoding.
		if err := budget.charge(node.Value); err != nil {
			return nil, err
		}
		switch node.Tag {
		case "!!str", "!!timestamp":
			return node.Value, nil
		case "!!null":
			return nil, nil
		case "!!binary":
			value, err := base64.StdEncoding.DecodeString(node.Value)
			if err != nil {
				return nil, fmt.Errorf("decode YAML binary: %w", err)
			}
			return string(value), nil
		case "!!int", "!!float", "!!bool":
			var value any
			if err := node.Decode(&value); err != nil {
				return nil, fmt.Errorf("decode YAML scalar: %w", err)
			}
			if number, ok := value.(float64); ok && (math.IsNaN(number) || math.IsInf(number, 0)) {
				return nil, fmt.Errorf("YAML numbers must be finite")
			}
			if _, rounded := value.(float64); rounded {
				// YAML falls back to float64 for integers exceeding native integer ranges.
				if integer, ok := new(big.Int).SetString(strings.ReplaceAll(node.Value, "_", ""), 0); ok {
					return json.Number(integer.String()), nil
				}
			}
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported YAML tag %s", node.Tag)
		}
	default:
		return nil, fmt.Errorf("unsupported YAML node")
	}
}
