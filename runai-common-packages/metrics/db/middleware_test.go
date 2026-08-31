/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBMiddleware(t *testing.T) {
	errCount := 0
	mid := DBMiddleware(func(e error) { errCount++ })
	assert.Equal(t, 0, errCount)
	assert.Equal(t, "RunaiGormPrometheus", mid.Name())
}

func TestDBMiddleware_Err(t *testing.T) {
	errCount := 0
	mid := DBMiddleware(func(e error) { errCount++ })
	mid.Initialize(nil)
	assert.Equal(t, 1, errCount)
}
