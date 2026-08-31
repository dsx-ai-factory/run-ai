/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enrich

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

type EnrichingEventHandler struct {
	configProvider  EnrichingConfigProvider
	next            events.EventHandler
	enricherFactory Factory
}

func NewEnrichingEventHandler(configProvider EnrichingConfigProvider, enricherFactory Factory, next events.EventHandler) events.EventHandler {
	return &EnrichingEventHandler{
		configProvider:  configProvider,
		next:            next,
		enricherFactory: enricherFactory,
	}
}

func (eh *EnrichingEventHandler) Handle(ctx context.Context, event *events.Event) error {
	log.Debug().Msgf("event %s is in enrich handler", event.UID)
	for _, enricherName := range eh.configProvider.GetGlobalEnrichers() {
		enricher, err := eh.enricherFactory.CreateEnricher(ctx, enricherName, eh.configProvider)
		if err != nil {
			return fmt.Errorf("failed to create enricher %s: %w", enricherName, err)
		}

		err = enricher.Enrich(ctx, event)
		if err != nil {
			return fmt.Errorf("failed to enrich event with enricher %s: %w", enricherName, err)
		}
	}

	log.Debug().Msgf("event %s enriched successfully", event.UID)
	return eh.next.Handle(ctx, event)
}
