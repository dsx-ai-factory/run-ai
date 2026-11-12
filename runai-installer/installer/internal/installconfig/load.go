/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads the installer configuration at path. The file must hold exactly one
// YAML document with no unknown fields; trailing empty or comment-only documents
// are allowed, while any further content or malformed trailing YAML is rejected.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %q: %w", path, err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	if err := ensureNoExtraDocuments(dec); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	if cfg.ControlPlane != nil && cfg.ControlPlane.TLS != nil && cfg.ControlPlane.TLS.GenerateSelfSigned {
		return nil, fmt.Errorf("parse config %q: generateSelfSigned is only allowed under tls", path)
	}
	if cfg.Cluster != nil {
		if cfg.Cluster.TLS != nil && cfg.Cluster.TLS.GenerateSelfSigned {
			return nil, fmt.Errorf("parse config %q: generateSelfSigned is only allowed under tls", path)
		}
		if cfg.Cluster.Inference != nil && cfg.Cluster.Inference.WildcardTLS != nil && cfg.Cluster.Inference.WildcardTLS.GenerateSelfSigned {
			return nil, fmt.Errorf("parse config %q: generateSelfSigned is only allowed under tls", path)
		}
	}
	return &cfg, nil
}

// ensureNoExtraDocuments drains the rest of the YAML stream after the config
// document and fails on the first document that carries content, or on
// malformed YAML. Empty documents (a bare "---" or comments only) are allowed.
func ensureNoExtraDocuments(dec *yaml.Decoder) error {
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
			return fmt.Errorf("unexpected extra YAML document at line %d: only one configuration document is allowed", node.Line)
		}
	}
}

// isEmptyDocument reports whether node is a document without content, such as
// one produced by a trailing "---" or comment-only text. An explicit null
// ("null" or "~") counts as content.
func isEmptyDocument(node *yaml.Node) bool {
	if node.Kind != yaml.DocumentNode || len(node.Content) != 1 {
		return node.Kind == 0
	}
	value := node.Content[0]
	return value.Kind == yaml.ScalarNode && value.Tag == "!!null" && value.Value == "" && value.Style == 0
}
