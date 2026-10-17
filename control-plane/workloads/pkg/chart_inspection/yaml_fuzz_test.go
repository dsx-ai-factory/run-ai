// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"encoding/json"
	"testing"
)

// Keep this target Go-native so go test -fuzz can generate and minimize inputs.
func FuzzYAMLObject(f *testing.F) {
	for _, seed := range []string{
		"questions: [{variable: replicas, type: int, default: 3}]",
		"base: &base {value: example}\ncopy: {<<: *base}",
		"cycle: &cycle [*cycle]",
		"duplicate: 1\nduplicate: 2",
		"value: .nan",
		"value: !!binary aGVsbG8=",
		"value: 18446744073709551617",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, document string) {
		if len(document) > maxMetadataBytes {
			t.Skip()
		}
		value, err := yamlObject([]byte(document))
		if err != nil {
			return
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("accepted YAML could not be encoded: %v", err)
		}

		// JSON escaping can multiply scalar bytes by six; leave room for bounded containers.
		if len(encoded) > 6*maxYAMLScalarBytes+(1<<20) {
			t.Fatalf("accepted YAML exceeded the expanded encoding bound: %d bytes", len(encoded))
		}
	})
}
