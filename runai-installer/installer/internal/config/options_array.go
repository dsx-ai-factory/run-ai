/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"regexp"
	"strings"
)

type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *ArrayFlags) Set(value string) error {
	compiledRegExp := regexp.MustCompile(`\s*,\s*`)
	valueRemovedWhiteSpaces := compiledRegExp.ReplaceAllString(value, ",")
	*i = strings.Split(valueRemovedWhiteSpaces, ",")
	return nil
}

func (i *ArrayFlags) Type() string { return "ArrayFlags" }
