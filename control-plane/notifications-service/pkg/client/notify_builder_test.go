/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotify_ShouldSetAllValues(t *testing.T) {
	nb := NewNotifyBuilder("category", "type")
	nb.WithDescription("description").WithTitle("title").WithProperty("label", "val").WithTenant("tenant1")

	assert.Equal(t, "description", nb.Notification.Description)
	assert.Equal(t, "title", nb.Notification.Title)
	assert.Equal(t, map[string]interface{}{"label": "val"}, nb.Notification.AdditionalProperties)
	assert.Equal(t, "tenant1", nb.Notification.Tenant)
}

func TestNotify_ShouldNotPanic(t *testing.T) {
	BuildNotification("category", "type").Notify()
}
