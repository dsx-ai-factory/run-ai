// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var pathIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*`)

// Accept a bounded values-path grammar: dotted keys, bracketed array indices
// and JSON-quoted keys. Reserved keys are rejected because consumers may apply
// these bindings to JavaScript objects.
func parsePath(value string) ([]any, error) {
	if value == "" || len(value) > 512 {
		return nil, fmt.Errorf("invalid values path")
	}
	segments := []any{}
	pos := 0
	identifier := func() error {
		match := pathIdentifier.FindString(value[pos:])
		if match == "" {
			return fmt.Errorf("invalid values path")
		}
		segments = append(segments, match)
		pos += len(match)
		return nil
	}
	if err := identifier(); err != nil {
		return nil, err
	}
	for pos < len(value) {
		switch value[pos] {
		case '.':
			pos++
			if err := identifier(); err != nil {
				return nil, err
			}
		case '[':
			pos++
			if pos >= len(value) {
				return nil, fmt.Errorf("invalid values path")
			}
			if value[pos] == '"' {
				start := pos
				pos++
				for pos < len(value) && value[pos] != '"' {
					if value[pos] == '\\' {
						pos++
					}
					pos++
				}
				if pos >= len(value) {
					return nil, fmt.Errorf("invalid values path")
				}
				pos++
				var key string
				if err := json.Unmarshal([]byte(value[start:pos]), &key); err != nil {
					return nil, fmt.Errorf("decode values key: %w", err)
				}
				segments = append(segments, key)
			} else {
				start := pos
				for pos < len(value) && value[pos] >= '0' && value[pos] <= '9' {
					pos++
				}
				if start == pos {
					return nil, fmt.Errorf("invalid array index")
				}
				index, err := strconv.Atoi(value[start:pos])
				if err != nil || index > 10000 {
					return nil, fmt.Errorf("invalid array index")
				}
				segments = append(segments, index)
			}
			if pos >= len(value) || value[pos] != ']' {
				return nil, fmt.Errorf("invalid values path")
			}
			pos++
		default:
			return nil, fmt.Errorf("invalid values path")
		}
	}
	if len(segments) > 32 {
		return nil, fmt.Errorf("values path is too deep")
	}
	for _, segment := range segments {
		if key, ok := segment.(string); ok && (key == "__proto__" || key == "prototype" || key == "constructor") {
			return nil, fmt.Errorf("values path uses a reserved key")
		}
	}
	return segments, nil
}

func readPath(values map[string]any, segments []any) (any, bool) {
	var value any = values
	for _, segment := range segments {
		switch key := segment.(type) {
		case string:
			object, ok := value.(map[string]any)
			if !ok {
				return nil, false
			}
			value, ok = object[key]
			if !ok {
				return nil, false
			}
		case int:
			list, ok := value.([]any)
			if !ok || key >= len(list) {
				return nil, false
			}
			value = list[key]
		}
	}
	return value, true
}

// Equivalent spellings such as service.port and service["port"] must identify
// the same binding when checking duplicates and visibility dependencies.
func canonicalPath(segments []any) string {
	var out strings.Builder
	for _, segment := range segments {
		out.WriteByte('[')
		switch key := segment.(type) {
		case string:
			out.WriteString(strconv.Quote(key))
		case int:
			out.WriteString(strconv.Itoa(key))
		}
		out.WriteByte(']')
	}
	return out.String()
}
