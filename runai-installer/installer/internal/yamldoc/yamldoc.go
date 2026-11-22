/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package yamldoc holds helpers shared by the installer's strict YAML loaders.
package yamldoc

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// RejectExtraDocuments drains the rest of the YAML stream after the document the
// caller has already decoded. It fails on the first further document that carries
// content, and on malformed trailing YAML. Empty documents (a bare "---", a "..."
// marker, or comments only) are allowed.
func RejectExtraDocuments(dec *yaml.Decoder) error {
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if !isEmpty(&node) {
			return fmt.Errorf("unexpected extra YAML document at line %d: only one document is allowed", contentLine(&node))
		}
	}
}

// isEmpty reports whether node is a document without content, such as one
// produced by a trailing "---" or comment-only text. An explicit null ("null"
// or "~") counts as content.
func isEmpty(node *yaml.Node) bool {
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return node.Kind == 0
	}
	value := node.Content[0]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!null" && value.Value == "" && value.Style == 0
}

// contentLine returns the line where the document's content starts, which is
// more useful to a reader than the line of the "---" marker above it.
func contentLine(node *yaml.Node) int {
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		return node.Content[0].Line
	}
	return node.Line
}
