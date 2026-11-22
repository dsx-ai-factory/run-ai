/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package yamldoc

import (
	"errors"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func RejectExtraDocuments(dec *yaml.Decoder) error {
	for document := 2; ; document++ {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("document %d: %w", document, err)
		}
		if !isEmpty(&node) {
			return fmt.Errorf("unexpected extra YAML document at line %d: only one document is allowed", contentLine(&node))
		}
	}
}

func isEmpty(node *yaml.Node) bool {
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return node.Kind == 0
	}
	value := node.Content[0]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!null" && value.Value == "" && value.Style == 0
}

func contentLine(node *yaml.Node) int {
	if node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		return node.Content[0].Line
	}
	return node.Line
}
