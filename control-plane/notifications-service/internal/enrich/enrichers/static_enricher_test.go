/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enrichers_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/enrich/enrichers"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
)

func TestStaticEnricher_Adds(t *testing.T) {
	event := &events.Event{
		Properties: make(map[string]interface{}),
	}
	config := &enrichers.StaticEnricherConfig{"test": "test"}

	enricher := enrichers.StaticEnricher{
		Config: config,
	}
	err := enricher.Enrich(context.Background(), event)
	assert.Nil(t, err)
	assert.Contains(t, event.Properties, "test")
	assert.Equal(t, event.Properties["test"], "test")
}

func TestStaticEnricher_IgnoresEmpty(t *testing.T) {
	event := &events.Event{
		Properties: make(map[string]interface{}),
	}
	config := &enrichers.StaticEnricherConfig{
		"test": "",
	}

	enricher := enrichers.StaticEnricher{
		Config: config,
	}
	err := enricher.Enrich(context.Background(), event)
	assert.Nil(t, err)
	assert.NotContains(t, event.Properties, "test")
}
