/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads an artifact lock, rejecting unknown fields and a lock that is
// not of the expected apiVersion and kind.
func ParseFile(path string) (*ArtifactLock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open artifact lock %q: %w", path, err)
	}
	defer f.Close()
	return Parse(f, path)
}

func Parse(r io.Reader, name string) (*ArtifactLock, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	var lock ArtifactLock
	if err := dec.Decode(&lock); err != nil {
		return nil, fmt.Errorf("parse artifact lock %q: %w", name, err)
	}
	if lock.APIVersion != APIVersion || lock.Kind != Kind {
		return nil, fmt.Errorf("artifact lock %q is not a %s %s", name, APIVersion, Kind)
	}
	if lock.Metadata.Version == "" {
		return nil, fmt.Errorf("artifact lock %q has no metadata.version", name)
	}
	if err := lock.ValidateImages(); err != nil {
		return nil, fmt.Errorf("artifact lock %q: %w", name, err)
	}
	if err := lock.ValidateCLIs(); err != nil {
		return nil, fmt.Errorf("artifact lock %q: %w", name, err)
	}
	return &lock, nil
}
