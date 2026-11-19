/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package yamldoc holds helpers for reading single-document YAML files.
package yamldoc

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// EnsureNoExtraDocuments drains the rest of the YAML stream after the first
// document has been decoded and fails on the first document that carries
// content, or on malformed YAML. Empty documents (a bare "---", a "..." end
// marker or comments only) are allowed; an explicit null ("null" or "~") is
// content and is rejected.
func EnsureNoExtraDocuments(dec *yaml.Decoder) error {
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if !isEmptyDocument(&node) {
			return fmt.Errorf("unexpected extra YAML document at line %d: only one document is allowed", node.Line)
		}
	}
}

// isEmptyDocument reports whether node is a document without content, such as
// one produced by a trailing "---" or comment-only text.
func isEmptyDocument(node *yaml.Node) bool {
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return node.Kind == 0
	}
	value := node.Content[0]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!null" && value.Value == "" && value.Style == 0
}
