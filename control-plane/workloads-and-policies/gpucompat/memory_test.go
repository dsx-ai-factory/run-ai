/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpucompat

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeLegacyGPUFields(t *testing.T) {
	for _, pair := range memoryAliases {
		for _, tc := range []struct {
			name     string
			input    map[string]any
			want     any
			conflict bool
		}{
			{"old", map[string]any{pair[0]: .25}, .25, false},
			{"new", map[string]any{pair[1]: .25}, .25, false},
			{"equal", map[string]any{pair[0]: .25, pair[1]: .25}, .25, false},
			{"different numeric encoding", map[string]any{pair[0]: json.Number("0.250"), pair[1]: .25}, .25, false},
			{"conflict", map[string]any{pair[0]: .25, pair[1]: .5}, nil, true},
			{"new null", map[string]any{pair[0]: .25, pair[1]: nil}, .25, false},
			{"old null", map[string]any{pair[0]: nil, pair[1]: .25}, .25, false},
			{"zero preserved", map[string]any{pair[0]: 0.0}, 0.0, false},
			{"null preserved", map[string]any{pair[0]: nil}, nil, false},
		} {
			t.Run(pair[0]+"/"+tc.name, func(t *testing.T) {
				source := map[string]any{"compute": tc.input}
				before := cloneMap(source)
				got, err := NormalizeLegacyGPUFields(source)
				if tc.conflict {
					if err == nil || !strings.Contains(err.Error(), pair[0]) || !strings.Contains(err.Error(), pair[1]) {
						t.Fatalf("expected both aliases in conflict error, got %v", err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					want := map[string]any{"compute": map[string]any{pair[1]: tc.want}}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("got %#v; want %#v", got, want)
					}
				}
				if !reflect.DeepEqual(source, before) {
					t.Fatal("mutated input source")
				}
			})
		}
	}
}

func TestMemoryDBAndResponseRoundTrip(t *testing.T) {
	source := map[string]any{"compute": map[string]any{"gpuPortionRequest": .25, "gpuPortionLimit": .5}}
	canonical, err := NormalizeLegacyGPUFields(source)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := LegacyGPUFields(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored, source) {
		t.Fatalf("DB representation changed: %#v", stored)
	}
	response := LegacyCompatibleGPUFields(stored)
	fields := response["compute"].(map[string]any)
	for _, pair := range memoryAliases {
		if fields[pair[0]] != fields[pair[1]] {
			t.Fatalf("response aliases diverged: %#v", fields)
		}
	}
	fields["gpuPortionRequest"] = .8
	if !reflect.DeepEqual(stored, source) {
		t.Fatal("response changed stored spec")
	}
	if canonical["compute"].(map[string]any)["gpuMemoryPortionRequest"] != .25 {
		t.Fatal("response changed canonical spec")
	}
}

func TestKnownComputeSectionsAndOpaqueFields(t *testing.T) {
	for _, container := range []string{"spec", "masterSpec", "worker", "master", "leader", "defaults", "rules", "combinedSpec", "combinedMasterSpec"} {
		t.Run(container, func(t *testing.T) {
			source := map[string]any{container: map[string]any{"compute": map[string]any{"gpuPortionRequest": .25}}}
			got, err := NormalizeLegacyGPUFields(source)
			if err != nil {
				t.Fatal(err)
			}
			compute := got[container].(map[string]any)["compute"].(map[string]any)
			if compute["gpuMemoryPortionRequest"] != .25 {
				t.Fatal(got)
			}
			response := LegacyCompatibleGPUFields(got)
			fields := response[container].(map[string]any)["compute"].(map[string]any)
			if fields["gpuPortionRequest"] != .25 || fields["gpuMemoryPortionRequest"] != .25 {
				t.Fatalf("response lost aliases in %s: %#v", container, fields)
			}
		})
	}
	source := map[string]any{"annotations": map[string]any{"compute": map[string]any{"gpuPortionRequest": "opaque"}}}
	got, err := NormalizeLegacyGPUFields(source)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, source) {
		t.Fatal("rewrote user-defined fields")
	}
	empty := LegacyCompatibleGPUFields(map[string]any{"compute": map[string]any{}})
	if len(empty["compute"].(map[string]any)) != 0 {
		t.Fatal("fabricated absent values")
	}
}

func TestPolicyAliasesAndDerivedReferences(t *testing.T) {
	rules := map[string]any{"compute": map[string]any{
		"gpuPortionRequest": map[string]any{"canEdit": false},
		"cpuCoreRequest":    map[string]any{"defaultFrom": map[string]any{"field": "compute.gpuPortionRequest"}},
	}}
	canonical, err := NormalizeLegacyGPUPolicy(rules)
	if err != nil {
		t.Fatal(err)
	}
	fields := canonical["compute"].(map[string]any)
	if fields["gpuMemoryPortionRequest"].(map[string]any)["canEdit"] != false {
		t.Fatal("lost lock")
	}
	if fields["cpuCoreRequest"].(map[string]any)["defaultFrom"].(map[string]any)["field"] != "compute.gpuMemoryPortionRequest" {
		t.Fatal("lost dependency")
	}
	stored, err := LegacyGPUPolicy(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored, rules) {
		t.Fatal("stored references do not round trip")
	}
	if CanonicalGPUFieldPath("annotations.gpuPortionRequest") != "annotations.gpuPortionRequest" {
		t.Fatal("rewrote unrelated path")
	}
}

func TestMemoryRequestInputBounds(t *testing.T) {
	for _, tc := range []struct {
		name    string
		compute map[string]any
		invalid bool
	}{
		{"omitted", map[string]any{}, false},
		{"null", map[string]any{"gpuMemoryPortionRequest": nil}, false},
		{"zero", map[string]any{"gpuMemoryPortionRequest": 0.0}, false},
		{"one", map[string]any{"gpuMemoryPortionRequest": json.Number("1.0")}, false},
		{"fraction", map[string]any{"gpuMemoryPortionRequest": .25}, false},
		{"negative", map[string]any{"gpuMemoryPortionRequest": -.1}, true},
		{"whole new", map[string]any{"gpuMemoryPortionRequest": 4.0}, true},
		{"equal aliases above one", map[string]any{"gpuPortionRequest": 4, "gpuMemoryPortionRequest": json.Number("4")}, true},
		{"whole legacy", map[string]any{"gpuPortionRequest": 4.0}, false},
		{"legacy with null alias", map[string]any{"gpuPortionRequest": 4.0, "gpuMemoryPortionRequest": nil}, false},
		{"legacy fractional validation stays with composer", map[string]any{"gpuPortionRequest": 4.5}, false},
		{"limit unchanged", map[string]any{"gpuMemoryPortionLimit": 4.0}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, section := range []string{"spec", "masterSpec", "worker", "master", "leader", "defaults"} {
				source := map[string]any{section: map[string]any{"compute": tc.compute}}
				before := cloneMap(source)
				err := ValidateMemoryRequestInput(source)
				if (err != nil) != tc.invalid {
					t.Fatalf("%s: invalid=%v, got %v", section, tc.invalid, err)
				}
				if err != nil && !strings.Contains(err.Error(), section+".compute.gpuMemoryPortionRequest") {
					t.Fatal("missing field path", err)
				}
				if !reflect.DeepEqual(source, before) {
					t.Fatal("mutated input")
				}
			}
		})
	}
}

func TestLegacyWholeRequestResponseRoundTrip(t *testing.T) {
	for _, value := range []any{4.0, json.Number("4"), 4} {
		stored := map[string]any{"compute": map[string]any{"gpuPortionRequest": value, "gpuPortionLimit": value}}
		canonical, err := NormalizeLegacyGPUFields(stored)
		if err != nil {
			t.Fatal(err)
		}
		for _, source := range []map[string]any{stored, canonical} {
			response := LegacyCompatibleGPUFields(source)
			compute := response["compute"].(map[string]any)
			if _, exists := compute["gpuMemoryPortionRequest"]; exists {
				t.Fatal("exposed out-of-range alias", response)
			}
			if compute["gpuPortionRequest"] != value || compute["gpuMemoryPortionLimit"] != value {
				t.Fatal("lost legacy request or limit", response)
			}
			if _, exists := compute["gpuDevicesRequest"]; exists {
				t.Fatal("converted legacy request to devices")
			}
			if err := ValidateMemoryRequestInput(response); err != nil {
				t.Fatal("response cannot be resubmitted", err)
			}
			resubmitted, err := LegacyGPUFields(response)
			if err != nil || !reflect.DeepEqual(resubmitted, stored) {
				t.Fatalf("round trip changed: %#v, %v", resubmitted, err)
			}
		}
	}
}

func TestLegacyWholePolicyResponseRoundTrip(t *testing.T) {
	policy := map[string]any{
		"defaults": map[string]any{"worker": map[string]any{"compute": map[string]any{"gpuPortionRequest": 4.0}}},
		"rules": map[string]any{"worker": map[string]any{"compute": map[string]any{
			"gpuPortionRequest": map[string]any{"canEdit": false, "required": true, "min": 2.0, "max": 8.0},
			"cpuCoreRequest":    map[string]any{"defaultFrom": map[string]any{"field": "worker.compute.gpuPortionRequest"}},
		}}},
	}
	canonical, err := NormalizeLegacyGPUPolicy(policy)
	if err != nil {
		t.Fatal(err)
	}
	response := LegacyCompatibleGPUPolicy(canonical)
	if err := ValidateMemoryRequestInput(response["defaults"].(map[string]any)); err != nil {
		t.Fatal(err)
	}
	stored, err := LegacyGPUPolicy(response)
	if err != nil || !reflect.DeepEqual(stored, policy) {
		t.Fatalf("policy changed on resubmit: %#v, %v", stored, err)
	}
	compute := response["defaults"].(map[string]any)["worker"].(map[string]any)["compute"].(map[string]any)
	if _, exists := compute["gpuMemoryPortionRequest"]; exists {
		t.Fatal("out-of-range default alias")
	}
}
