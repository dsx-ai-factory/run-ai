/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"strconv"
	"strings"
)

func IntSliceToString(nums []int) string {
	strNums := make([]string, len(nums))

	for i, num := range nums {
		strNums[i] = strconv.Itoa(num)
	}

	return strings.Join(strNums, ",")
}
