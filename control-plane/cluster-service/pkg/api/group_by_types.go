/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import "fmt"

type GroupBy int32

const (
	Subdomain GroupBy = iota
)

const (
	GroupBySubdomain = "subdomain"
)

// ValidatedGroupByParam validates the groupBy param
func ValidatedGroupByParam(groupBy string) error {
	if groupBy != GroupBySubdomain {
		return fmt.Errorf("invalid groupBy param: %s", groupBy)
	}
	return nil
}
