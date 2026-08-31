/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toggles

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyToggler_IsEnabled_AlwaysTrue(t *testing.T) {
	dt := &dummyToggler{}
	ctx := context.Background()

	assert.True(t, dt.IsEnabled(ctx, "any-feature"))
	assert.True(t, dt.IsEnabled(ctx, "another-feature"))
	assert.True(t, dt.IsEnabled(ctx, ""))
}

func TestDummyToggler_Close(t *testing.T) {
	dt := &dummyToggler{}
	assert.NoError(t, dt.Close())
}

func TestFailedToggler_IsEnabled_AlwaysFalse(t *testing.T) {
	ft := &failedToggler{}
	ctx := context.Background()

	assert.False(t, ft.IsEnabled(ctx, "any-feature"))
	assert.False(t, ft.IsEnabled(ctx, "another-feature"))
	assert.False(t, ft.IsEnabled(ctx, ""))
}

func TestFailedToggler_Close(t *testing.T) {
	ft := &failedToggler{}
	assert.NoError(t, ft.Close())
}
