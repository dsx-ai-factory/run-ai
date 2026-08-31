/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package httputils

import "errors"

const (
	AscendingStr = "asc"
	CreatedAtStr = "createdAt"
)

var (
	ErrRequestIsNil = errors.New("request is nil")
)

func SetOrderFunc[T comparable](request *T, sortOrder func(sortOrder string) T) {
	if request == nil {
		return
	}
	*request = sortOrder(AscendingStr)
}

func SetSortByFunc[T, K comparable](request *T, sortBy func(sort K) T, setter func(v string) (*K, error)) error {
	if request == nil {
		return ErrRequestIsNil
	}
	value, err := setter(CreatedAtStr)
	if err != nil {
		return err
	}
	*request = sortBy(*value)
	return nil
}
