/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

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
