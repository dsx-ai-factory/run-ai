/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import "strings"

func SplitLines(text string) []string {
	return strings.Split(text, "\n")
}

func JoinLines(arr []string) string {
	res := ""
	if len(arr) > 0 {
		res = strings.Join(arr, "\n")
	}
	return res
}

func StringHasValue(str *string) bool {
	return str != nil && *str != ""
}
