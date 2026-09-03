/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const BytesToMB = 1000000

// MapMe convert an interface (hopefully) containing a generic map to generic map
func MapMe(src interface{}) map[string]interface{} {
	switch realType := src.(type) {
	case map[string]interface{}:
		return realType
	default:
		return make(map[string]interface{})
	}
}

// BoolMe convert to boolean
func BoolMe(src interface{}, defaultValue bool) bool {
	switch realType := src.(type) {
	case bool:
		return realType
	default:
		return defaultValue
	}
}

// DotUs ->
//
//	DotUs(a, b, c) -> a.b.c
//	DotUs(a, "", b) -> a.b
func DotUs(args ...string) string {
	emptyRemoved := make([]string, 0, len(args))
	for _, cur := range args {
		if cur != "" {
			emptyRemoved = append(emptyRemoved, cur)
		}
	}
	return strings.Join(emptyRemoved, ".")
}

// DecodeInt64List decode 1;23;3;4;55 string to []int64
func DecodeInt64List(value string, delimiter string) ([]int64, error) {
	value = strings.ReplaceAll(value, ",", ";") // we also support comma
	parts := strings.Split(value, delimiter)
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		intPart, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("non-numeric value '%s'", part)
		}
		result = append(result, intPart)
	}
	return result, nil
}

func JsonNumberToInt(rawNumber interface{}) (*int, error) {
	if rawNumber == nil {
		return nil, nil
	}
	asString := fmt.Sprintf("%v", rawNumber)
	result, err := strconv.Atoi(asString)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func ArgToMap(arg string) map[string]string {
	result := make(map[string]string)
	splitArg := strings.Split(arg, ",")
	for _, arg := range splitArg {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			continue
		}
		result[kv[0]] = kv[1]
	}
	return result
}
