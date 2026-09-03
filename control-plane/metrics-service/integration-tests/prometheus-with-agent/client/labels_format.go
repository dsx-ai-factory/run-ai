/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"
	"strings"
)

func joinLabels(labels [][2]any) string {
	if len(labels) == 0 {
		return ""
	}
	l := make([]string, 0)
	for _, val := range labels {
		l = append(l, fmt.Sprintf("%s=\"%s\"", val[0], val[1]))
	}
	return strings.Join(l, ",")
}

func getPromqlLabels(labels [][2]any) string {
	labelsStr := joinLabels(labels)
	if labelsStr == "" {
		return ""
	}
	return fmt.Sprintf("{%s}", labelsStr)
}
