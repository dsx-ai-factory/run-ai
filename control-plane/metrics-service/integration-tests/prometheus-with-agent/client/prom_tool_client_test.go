//go:build manual
// +build manual

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPostMetrics is a manual test, since it is
func TestPostMetrics(t *testing.T) {
	b := NewMetricCollection().
		AddMetric(
			NewMetric("my_metric", 10).AddLabel("code", "500"))

	err := PostMetrics(b.ToOpenMetrics())
	assert.NoError(t, err)
}
