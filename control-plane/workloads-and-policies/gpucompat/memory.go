/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpucompat

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
)

const computeSection = "compute"

var computeContainers = []string{computeSection, "spec", "masterSpec", "worker", "master", "leader", "defaults", "rules", "combinedSpec", "combinedMasterSpec"}

var memoryAliases = [][2]string{
	{"gpuPortionRequest", "gpuMemoryPortionRequest"},
	{"gpuPortionLimit", "gpuMemoryPortionLimit"},
}

// ValidateMemoryRequestInput runs before adapting submitted specs or defaults.
// Normalized legacy requests may exceed one and must not be validated as new input.
func ValidateMemoryRequestInput(source map[string]any) error {
	return visitCompute(source, "", func(compute map[string]any, path string) error {
		value := compute["gpuMemoryPortionRequest"]
		if value == nil {
			return nil
		}
		number, ok := numericValue(value)
		if !ok || number.Sign() < 0 || number.Cmp(big.NewRat(1, 1)) > 0 {
			return fmt.Errorf("%s.gpuMemoryPortionRequest must be between 0 and 1", path)
		}
		return nil
	})
}

// NormalizeLegacyGPUFields copies one source before resolving deprecated aliases.
// Call it before merging sources so editable defaults cannot conflict with input.
func NormalizeLegacyGPUFields(source map[string]any) (map[string]any, error) {
	result := cloneMap(source)
	if err := visitCompute(result, "", normalizeCompute); err != nil {
		return nil, err
	}
	return result, nil
}

// NormalizeLegacyGPUPolicy also translates references on rules outside compute,
// since any numeric field can derive its default from a GPU memory field.
func NormalizeLegacyGPUPolicy(source map[string]any) (map[string]any, error) {
	result := cloneMap(source)
	normalizeReferences(result)
	if err := visitCompute(result, "", normalizeCompute); err != nil {
		return nil, err
	}
	return result, nil
}

func CanonicalGPUFieldPath(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) < 2 || parts[len(parts)-2] != computeSection {
		return path
	}
	for _, alias := range memoryAliases {
		if parts[len(parts)-1] == alias[0] {
			parts[len(parts)-1] = alias[1]
		}
	}
	return strings.Join(parts, ".")
}

// LegacyGPUFields uses the legacy names required by existing DB records and composers.
func LegacyGPUFields(source map[string]any) (map[string]any, error) {
	result, err := NormalizeLegacyGPUFields(source)
	if err != nil {
		return nil, err
	}
	if err := useLegacyGPUFields(result); err != nil {
		return nil, err
	}
	return result, nil
}

// useLegacyGPUFields operates on the caller's already normalized copy.
func useLegacyGPUFields(result map[string]any) error {
	return visitCompute(result, "", func(compute map[string]any, _ string) error {
		for _, alias := range memoryAliases {
			if value, exists := compute[alias[1]]; exists {
				compute[alias[0]] = value
				delete(compute, alias[1])
			}
		}
		return nil
	})
}

// LegacyGPUPolicy translates policy references as well as field names.
// Ordinary spec conversion must leave user-supplied maps and strings untouched.
func LegacyGPUPolicy(source map[string]any) (map[string]any, error) {
	result, err := NormalizeLegacyGPUPolicy(source)
	if err != nil {
		return nil, err
	}
	if err := useLegacyGPUFields(result); err != nil {
		return nil, err
	}
	useLegacyGPUReferences(result)
	return result, nil
}

func normalizeCompute(compute map[string]any, path string) error {
	for _, alias := range memoryAliases {
		legacy, oldPresent := compute[alias[0]]
		canonical, newPresent := compute[alias[1]]
		if oldPresent && newPresent && legacy != nil && canonical != nil && !equal(legacy, canonical) {
			return fmt.Errorf("%s.%s and %s.%s conflict: provide one alias or equal values", path, alias[0], path, alias[1])
		}
		if oldPresent {
			if !newPresent || canonical == nil {
				compute[alias[1]] = legacy
			}
			delete(compute, alias[0])
		}
	}
	return nil
}

// Only traverse workload/policy containers. User annotation and environment keys
// are opaque even when they happen to be named like a GPU field.
func visitCompute(node map[string]any, path string, visit func(map[string]any, string) error) error {
	for _, key := range computeContainers {
		child, ok := node[key].(map[string]any)
		if !ok {
			continue
		}
		childPath := key
		if path != "" {
			childPath = path + "." + key
		}
		if key == computeSection {
			if err := visit(child, childPath); err != nil {
				return err
			}
		} else if err := visitCompute(child, childPath, visit); err != nil {
			return err
		}
	}
	return nil
}

func normalizeReferences(node map[string]any) {
	mapReferences(node, CanonicalGPUFieldPath)
}

func useLegacyGPUReferences(node map[string]any) {
	mapReferences(node, LegacyGPUFieldPath)
}

func mapReferences(node map[string]any, transform func(string) string) {
	for key, value := range node {
		child, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if key == "defaultFrom" {
			if field, ok := child["field"].(string); ok {
				child["field"] = transform(field)
			}
		} else {
			mapReferences(child, transform)
		}
	}
}

func equal(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	if left, ok := a.(map[string]any); ok {
		right, ok := b.(map[string]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for key, value := range left {
			other, exists := right[key]
			if !exists || !equal(value, other) {
				return false
			}
		}
		return true
	}
	if left, ok := a.([]any); ok {
		right, ok := b.([]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for i, value := range left {
			if !equal(value, right[i]) {
				return false
			}
		}
		return true
	}
	x, xOK := numericValue(a)
	y, yOK := numericValue(b)
	return xOK && yOK && x.Cmp(y) == 0
}

func numericValue(value any) (*big.Rat, bool) {
	switch value.(type) {
	case json.Number, float64, float32, int, int32, int64, uint, uint32, uint64:
		return new(big.Rat).SetString(fmt.Sprint(value))
	default:
		return nil, false
	}
}

func cloneMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = clone(value)
	}
	return result
}

func clone(value any) any {
	switch value := value.(type) {
	case map[string]any:
		return cloneMap(value)
	case []any:
		result := make([]any, len(value))
		for i, item := range value {
			result[i] = clone(item)
		}
		return result
	default:
		return value
	}
}

// LegacyGPUFieldPath maps the canonical name to its deprecated alias.
func LegacyGPUFieldPath(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) >= 2 && parts[len(parts)-2] == computeSection {
		for _, alias := range memoryAliases {
			if parts[len(parts)-1] == alias[1] {
				parts[len(parts)-1] = alias[0]
			}
		}
	}
	return strings.Join(parts, ".")
}

func LegacyGPUDefaultFromOrder(source map[string][]string) map[string][]string {
	if source == nil {
		return nil
	}
	result := make(map[string][]string, len(source))
	for field, dependencies := range source {
		var values []string
		if dependencies != nil {
			values = make([]string, len(dependencies))
		}
		for i, dependency := range dependencies {
			values[i] = LegacyGPUFieldPath(dependency)
		}
		result[LegacyGPUFieldPath(field)] = values
	}
	return result
}

// LegacyCompatibleGPUFields projects validated canonical or stored data. It does
// not modify the input or introduce fields that were omitted in the source.
func LegacyCompatibleGPUFields(source map[string]any) map[string]any {
	result := cloneMap(source)
	projectResponseFields(result)
	return result
}

func projectResponseFields(node map[string]any) {
	for _, key := range computeContainers {
		child, ok := node[key].(map[string]any)
		if !ok {
			continue
		}
		if key != computeSection {
			projectResponseFields(child)
			continue
		}
		for _, alias := range memoryAliases {
			old, oldSet := child[alias[0]]
			value, newSet := child[alias[1]]
			if !newSet || value == nil {
				if oldSet {
					value = old
				}
			}
			if oldSet || newSet {
				child[alias[0]] = clone(value)
				if number, ok := numericValue(value); alias[1] == "gpuMemoryPortionRequest" && ok && number.Cmp(big.NewRat(1, 1)) > 0 {
					delete(child, alias[1])
				} else {
					child[alias[1]] = clone(value)
				}
			}
		}
	}
}

func LegacyCompatibleGPUPolicy(source map[string]any) map[string]any {
	result := LegacyCompatibleGPUFields(source)
	// Old clients still resolve references against the deprecated response fields.
	useLegacyGPUReferences(result)
	return result
}

func ResolveLegacyGPUMemory(kind string, legacy, canonical *float64) (*float64, error) {
	if legacy != nil && canonical != nil && *legacy != *canonical {
		return nil, fmt.Errorf("gpuPortion%s and gpuMemoryPortion%s conflict: provide one alias or equal values", kind, kind)
	}
	if canonical != nil {
		return canonical, nil
	}
	return legacy, nil
}

func CanonicalGPUDefaultFromOrder(source map[string][]string) map[string][]string {
	if source == nil {
		return nil
	}
	result := make(map[string][]string, len(source))
	for field, dependencies := range source {
		var values []string
		if dependencies != nil {
			values = make([]string, len(dependencies))
		}
		for i, dependency := range dependencies {
			values[i] = CanonicalGPUFieldPath(dependency)
		}
		result[CanonicalGPUFieldPath(field)] = values
	}
	return result
}
