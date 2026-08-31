/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package history

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type EventsCleanerer interface {
	Cleanup(cf CleanupFilter) error
}

type EventsCleaner struct {
	repository    EventsCleanerer
	cleanupConfig CleanupConfig
}

func NewEventsCleaner(repository EventsCleanerer, cleanupConfig CleanupConfig) *EventsCleaner {
	return &EventsCleaner{repository: repository, cleanupConfig: cleanupConfig}
}

func (ec *EventsCleaner) Schedule(ctx context.Context) {
	cf := ec.constructCleanupFilter()
	log.Info().Msgf("scheduling history cleanup with filters: %v", cf)

	go func() {
		ticker := time.NewTicker(ec.cleanupConfig.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ec.cleanup(cf)
			}
		}
	}()
}

func (ec *EventsCleaner) cleanup(cf *CleanupFilter) {
	if err := ec.repository.Cleanup(*cf); err != nil {
		log.Err(err).Msg("history cleanup failed")
	}
}

func (ec *EventsCleaner) constructCleanupFilter() *CleanupFilter {
	cf := &CleanupFilter{}
	if ec.cleanupConfig.MaxEventsPerUser > 0 {
		cf.WithMaxEventsPerUser(ec.cleanupConfig.MaxEventsPerUser)
	}
	if ec.cleanupConfig.MaxAge > 0 {
		cf.WithMaxAge(ec.cleanupConfig.MaxAge)
	}
	return cf
}
