/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package options_test

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/stretchr/testify/assert"
)

func TestOptionsWithTTL(t *testing.T) {
	options := options.NewStoreOptions(options.Ttl(1 * time.Second))
	assert.Equal(t, 1*time.Second, options.Ttl)
}

func TestOptionsWithStoreSize(t *testing.T) {
	options := options.NewStoreOptions(options.MaxStoreSizeMB(44))
	assert.Equal(t, 44, options.MaxStoreSizeMB)
}

func TestOptionsWithStoreSize_AboveLimit(t *testing.T) {
	options := options.NewStoreOptions(options.MaxStoreSizeMB(99999))
	assert.Equal(t, 4000, options.MaxStoreSizeMB)
}

func TestOptionsWithStoreSize_BelowLimit(t *testing.T) {
	options := options.NewStoreOptions(options.MaxStoreSizeMB(1))
	assert.Equal(t, 20, options.MaxStoreSizeMB)
}

func TestOptionsWithMaxRecordSize(t *testing.T) {
	options := options.NewStoreOptions(options.MaxRecordSizeKB(1))
	assert.Equal(t, 1, options.MaxRecordSizeKB)
}

func TestOptionsWithReadOnly(t *testing.T) {
	options := options.NewStoreOptions(options.ReadOnly())
	assert.Equal(t, true, options.ReadOnly)
}

func TestOptionsWithNotReadOnly(t *testing.T) {
	options := options.NewStoreOptions()
	assert.Equal(t, false, options.ReadOnly)
}
