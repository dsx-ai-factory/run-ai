/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import "fmt"

const (
	DefaultLimit        int32 = 50
	MaxLimit            int32 = 200
	UpstreamPageSize    int32 = 200
	MaxPaginatedResults int32 = 2000
	MaxAPIFilters             = 10
)

func ValidatePagination(limit, offset int32) error {
	if limit < 0 || limit > MaxLimit {
		return fmt.Errorf("limit must be between 0 and %d", MaxLimit)
	}
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}

func EffectiveLimit(limit int32) int32 {
	if limit == 0 {
		return DefaultLimit
	}
	return limit
}

func Paginate[T any](items []T, offset, limit int32) ([]T, bool, int32) {
	total := int32(len(items))
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	page := items[start:end]
	if end < total {
		return page, true, end
	}
	return page, false, 0
}
