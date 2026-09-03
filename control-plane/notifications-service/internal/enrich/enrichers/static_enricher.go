/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enrichers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

const (
	StaticEnricherName = "static"
)

type StaticEnricherConfig map[string]string

type StaticEnricher struct {
	Config *StaticEnricherConfig
}

func NewStaticEnricher(config *StaticEnricherConfig) (*StaticEnricher, error) {
	return &StaticEnricher{
		Config: config,
	}, nil
}

func (se *StaticEnricher) Enrich(ctx context.Context, event *events.Event) error {
	for key, val := range *se.Config {
		if val != "" {
			event.Properties[key] = val
		}
	}
	return nil
}
