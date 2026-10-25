/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"strings"
)

// FilterEq builds a "field==value"
func FilterEq(field, value string) string {
	return field + "==" + value
}

func ValidateNoCommas(field, value string) error {
	if strings.ContainsRune(value, ',') {
		return fmt.Errorf("%s must not contain commas", field)
	}
	return nil
}

// ValidateFilterValue rejects filterBy syntax so FilterEq produces one exact value.
func ValidateFilterValue(field, value string) error {
	if err := ValidateNoCommas(field, value); err != nil {
		return err
	}
	for _, token := range []string{"==", "!=", "<=", ">=", "=@", "!@", "=^", "=$", "||"} {
		if strings.Contains(value, token) {
			return fmt.Errorf("%s must not contain %q", field, token)
		}
	}
	return nil
}
