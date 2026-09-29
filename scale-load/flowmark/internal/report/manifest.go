/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"encoding/json"
	"slices"
)

const ManifestName = "manifest.json"

type Manifest struct {
	Runs []string `json:"runs"`
}

func parseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (m Manifest) with(run string) Manifest {
	if m.contains(run) {
		return m
	}
	return Manifest{Runs: append(slices.Clone(m.Runs), run)}
}

func (m Manifest) contains(run string) bool {
	return slices.Contains(m.Runs, run)
}

func (m Manifest) encode() ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
