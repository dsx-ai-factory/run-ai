// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package helmvalues handles single destinations in structured Helm values.
package helmvalues

import (
	"cmp"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const MaxBindings = 128

type segment struct {
	key     string
	index   int
	isIndex bool
}

type Path struct {
	segments []segment
}

// Compare orders paths by their keys and numeric indexes, so contiguous array writes can be applied in order.
func (p Path) Compare(other Path) int {
	for i := range min(len(p.segments), len(other.segments)) {
		a, b := p.segments[i], other.segments[i]
		if a.isIndex != b.isIndex {
			if a.isIndex {
				return 1
			}
			return -1
		}
		if a.isIndex {
			if order := cmp.Compare(a.index, b.index); order != 0 {
				return order
			}
		} else if order := strings.Compare(a.key, b.key); order != 0 {
			return order
		}
	}
	return cmp.Compare(len(p.segments), len(other.segments))
}

// ArrayParent identifies the outermost array that Helm would replace as a whole.
func (p Path) ArrayParent() (Path, bool) {
	for i, part := range p.segments {
		if part.isIndex {
			return Path{segments: p.segments[:i]}, true
		}
	}
	return Path{}, false
}

func (p Path) Get(values map[string]interface{}) (interface{}, error) {
	var current interface{} = values
	for _, part := range p.segments {
		if part.isIndex {
			items, ok := current.([]interface{})
			if !ok || part.index >= len(items) {
				return nil, fmt.Errorf("array destination does not exist")
			}
			current = items[part.index]
		} else {
			object, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("object destination does not exist")
			}
			current, ok = object[part.key]
			if !ok {
				return nil, fmt.Errorf("key %q does not exist", part.key)
			}
		}
	}
	return current, nil
}

// Parse accepts dotted keys, backslash-escaped characters, numeric indexes, and JSON-quoted bracket keys.
// It deliberately does not accept Helm's multi-assignment or value-expression syntax.
func Parse(value string) (Path, error) {
	var parts []segment
	if len(value) == 0 || len(value) > 1024 {
		return Path{}, fmt.Errorf("path must contain between 1 and 1024 characters")
	}
	for i := 0; i < len(value); {
		var part segment
		if value[i] == '[' {
			i++
			start := i
			if i < len(value) && value[i] == '"' {
				i++
				for i < len(value) && value[i] != '"' {
					if value[i] == '\\' {
						i++
					}
					i++
				}
				if i >= len(value) || json.Unmarshal([]byte(value[start:i+1]), &part.key) != nil || part.key == "" {
					return Path{}, fmt.Errorf("invalid quoted path key")
				}
				i++
			} else {
				for i < len(value) && value[i] >= '0' && value[i] <= '9' {
					i++
				}
				index, err := strconv.Atoi(value[start:i])
				if err != nil || index > 1024 || len(parts) == 0 {
					return Path{}, fmt.Errorf("array indexes must be between 0 and 1024, beneath a key")
				}
				part = segment{index: index, isIndex: true}
			}
			if i >= len(value) || value[i] != ']' {
				return Path{}, fmt.Errorf("unclosed path bracket")
			}
			i++
		} else {
			var key strings.Builder
			for i < len(value) && value[i] != '.' && value[i] != '[' {
				if value[i] == '\\' {
					i++
					if i == len(value) {
						return Path{}, fmt.Errorf("incomplete path escape")
					}
				} else if strings.ContainsRune("]\n\r\t", rune(value[i])) {
					return Path{}, fmt.Errorf("invalid path character")
				}
				key.WriteByte(value[i])
				i++
			}
			part.key = key.String()
			if part.key == "" {
				return Path{}, fmt.Errorf("empty path key")
			}
		}
		parts = append(parts, part)
		if len(parts) > 32 {
			return Path{}, fmt.Errorf("path exceeds 32 segments")
		}
		if i < len(value) && value[i] == '.' {
			i++
			if i == len(value) || value[i] == '[' || value[i] == '.' {
				return Path{}, fmt.Errorf("empty path key")
			}
		} else if i < len(value) && value[i] != '[' {
			return Path{}, fmt.Errorf("expected a path separator")
		}
	}
	return Path{segments: parts}, nil
}

// CreateFlag returns only the companion setting required by the two Blueprint compatibility paths.
func (p Path) CreateFlag() string {
	if len(p.segments) != 3 || p.segments[0] != (segment{key: "global"}) || p.segments[2] != (segment{key: "name"}) {
		return ""
	}
	for _, key := range []string{"imagePullSecret", "ngcApiSecret"} {
		if p.segments[1] == (segment{key: key}) {
			return "global." + key + ".create"
		}
	}
	return ""
}

// ValidatePaths rejects duplicate/overlapping destinations, including automatically managed create flags.
func ValidatePaths(variables []string) error {
	if len(variables) > MaxBindings {
		return fmt.Errorf("at most %d credential bindings are supported", MaxBindings)
	}
	var paths []Path
	var owners []string
	for _, variable := range variables {
		path, err := Parse(variable)
		if err != nil {
			return fmt.Errorf("invalid credential destination %q: %w", variable, err)
		}
		additions := []Path{path}
		if flag := path.CreateFlag(); flag != "" {
			companion, err := Parse(flag)
			if err != nil {
				return err
			}
			additions = append(additions, companion)
		}
		for _, added := range additions {
			for i, existing := range paths {
				common := min(len(added.segments), len(existing.segments))
				overlaps := true
				for j := range common {
					if added.segments[j].isIndex != existing.segments[j].isIndex {
						return fmt.Errorf("credential destinations %q and %q require incompatible containers", owners[i], variable)
					}
					if added.segments[j] != existing.segments[j] {
						overlaps = false
						break
					}
				}
				if overlaps {
					return fmt.Errorf("credential destinations %q and %q overlap", owners[i], variable)
				}
			}
			paths = append(paths, added)
			owners = append(owners, variable)
		}
	}
	return nil
}

// Set changes one destination in place, preserving sibling keys and array entries.
// Missing containers are created; sparse arrays and replacement of scalar parents are rejected.
func (p Path) Set(values map[string]interface{}, value interface{}) error {
	if len(p.segments) == 0 || values == nil {
		return fmt.Errorf("a parsed path and initialized values object are required")
	}
	_, err := set(values, p.segments, value)
	return err
}

func set(current interface{}, parts []segment, value interface{}) (interface{}, error) {
	if len(parts) == 0 {
		return value, nil
	}
	part := parts[0]
	if part.isIndex {
		if current == nil {
			current = []interface{}{}
		}
		items, ok := current.([]interface{})
		if !ok {
			return nil, fmt.Errorf("array index applied to a non-array value")
		}
		if part.index > len(items) {
			return nil, fmt.Errorf("array index %d would leave gaps in an array of length %d", part.index, len(items))
		}
		if part.index == len(items) {
			items = append(items, nil)
		}
		child, err := set(items[part.index], parts[1:], value)
		if err != nil {
			return nil, err
		}
		items[part.index] = child
		return items, nil
	}
	if current == nil {
		current = map[string]interface{}{}
	}
	object, ok := current.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %q applied to a non-object value", part.key)
	}
	child, err := set(object[part.key], parts[1:], value)
	if err != nil {
		return nil, err
	}
	object[part.key] = child
	return object, nil
}
