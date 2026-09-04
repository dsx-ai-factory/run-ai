/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"fmt"
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

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var lock ArtifactLock
	if err := dec.Decode(&lock); err != nil {
		return nil, fmt.Errorf("parse artifact lock %q: %w", path, err)
	}
	if lock.APIVersion != APIVersion || lock.Kind != Kind {
		return nil, fmt.Errorf("artifact lock %q is not a %s %s", path, APIVersion, Kind)
	}
	if lock.Metadata.Version == "" {
		return nil, fmt.Errorf("artifact lock %q has no metadata.version", path)
	}
	if err := lock.ValidateImages(); err != nil {
		return nil, fmt.Errorf("artifact lock %q: %w", path, err)
	}
	if err := lock.ValidateCLIs(); err != nil {
		return nil, fmt.Errorf("artifact lock %q: %w", path, err)
	}
	return &lock, nil
}
