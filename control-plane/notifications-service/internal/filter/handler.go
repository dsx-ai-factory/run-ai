/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

func NewFilteringEventHandler(configProvider FilteringConfigProvider, next events.EventHandler) events.EventHandler {
	return &FilteringEventHandler{
		ConfigProvider: configProvider,
		Next:           next,
		FilterFactory:  NewFilterFactory(),
	}
}

type FilteringEventHandler struct {
	ConfigProvider FilteringConfigProvider
	Next           events.EventHandler
	FilterFactory  Factory
}

func (fh *FilteringEventHandler) Handle(ctx context.Context, event *events.Event) error {
	log.Debug().Msgf("event %s is in filtering handler", event.UID)
	fc := fh.ConfigProvider.GetGlobalFilteringConfig()

	if fc.Policy == PolicyDrop {
		return nil
	}

	for _, filterC := range fc.Filters {
		filter, err := fh.FilterFactory.CreateFilter(filterC)
		if err != nil {
			return fmt.Errorf("failed to create filter %s: %w", filterC.Type, err)
		}

		dropped, err := filter.Filter(ctx, event)
		if err != nil {
			return fmt.Errorf("applying filter on event failed, filter: %s: %w", filterC.Type, err)
		}
		switch fc.Policy {
		case PolicyMatchAll:
			if dropped {
				log.Debug().Msgf("event %s was dropped by filter %s", event.UID, filterC.Type)
				return nil
			}
		case PolicyMatchAny:
			if !dropped {
				log.Debug().Msgf("event %s passed all filters", event.UID)
				return fh.Next.Handle(ctx, event)
			}
		}
	}

	if fc.Policy == PolicyMatchAny && len(fc.Filters) > 0 {
		log.Debug().Msgf("event %s was dropped due to policy match any missing filters", event.UID)
		return nil
	}
	log.Debug().Msgf("event %s passed all filters", event.UID)
	return fh.Next.Handle(ctx, event)
}
