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

// ValidateFilterValue rejects the filterBy metacharacters — the comma separator
// and the == operator — so a value can't smuggle an extra predicate into FilterEq.
func ValidateFilterValue(field, value string) error {
	if err := ValidateNoCommas(field, value); err != nil {
		return err
	}
	if strings.Contains(value, "==") {
		return fmt.Errorf("%s must not contain '=='", field)
	}
	return nil
}
