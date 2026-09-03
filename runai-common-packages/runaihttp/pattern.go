/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaihttp

import (
	"fmt"
	"regexp"
	"strings"
)

func AssertPattern(optionalValue interface{}, fieldName string, pattern string) error {
	if pattern == "" {
		return nil
	}

	//  pattern received as /.../. for go matching we need to remove the slashes
	pattern = strings.TrimPrefix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %s", pattern)
	}

	valuePtr, err := getValuePtr[string](optionalValue)
	if err != nil || valuePtr == nil {
		return nil
	}

	if !re.MatchString(*valuePtr) {
		return fmt.Errorf("value for %s must comply to the pattern %s", fieldName, pattern)
	}

	return nil
}

func getValuePtr[T any](optionalValue interface{}) (*T, error) {
	switch realValue := optionalValue.(type) {
	case T:
		return &realValue, nil
	case *T:
		return realValue, nil
	default:
		return nil, fmt.Errorf("invalid type %T", optionalValue)
	}
}
