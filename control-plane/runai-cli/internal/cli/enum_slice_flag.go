/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"strings"
)

const (
	typeEnumSlice = "enumSlice"
)

type EnumSliceFlag struct {
	allowed    []string
	allowedSet map[string]struct{}
	value      *[]string
	Changed    bool // tracks if the flag was explicitly set by the user, used by cobra cmd
}

func NewEnumSliceFlag(allowed []string, value *[]string) *EnumSliceFlag {
	return &EnumSliceFlag{
		allowed:    allowed,
		allowedSet: toSet(allowed),
		value:      value,
	}
}

func (e *EnumSliceFlag) Type() string {
	return typeEnumSlice
}

func (e *EnumSliceFlag) String() string {
	return fmt.Sprintf("[%s]", strings.Join(*e.value, ","))
}

func (e *EnumSliceFlag) Set(input string) error {
	values := strings.Split(input, ",")
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if _, ok := e.allowedSet[trimmed]; !ok {
			return fmt.Errorf("must be partial or full match of %s", e.String())
		}
	}
	*e.value = values
	e.Changed = true
	return nil
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}
