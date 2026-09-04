/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package history

import (
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	history "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history"
	configprovider "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"gorm.io/gorm"
)

type EventsRepository struct {
	db     *gorm.DB
	schema string
}

func NewEventsRepository() (*EventsRepository, error) {
	db := configprovider.GetDB()
	if db == nil {
		return nil, fmt.Errorf("default event dao is not initialized")
	}

	return &EventsRepository{db: db, schema: configprovider.GetSchema()}, nil
}

func (er *EventsRepository) GetEvents(ef history.EventFilter) ([]*history.Event, error) {
	events := []*models.Event{}
	query := applyEventFilter(er.db.Table(er.tableName()), ef)
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}

	return toServiceEvents(events), nil
}

func (er *EventsRepository) DeleteEvent(eventId string) error {
	return er.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(er.tableName()).Where(&models.Event{Id: eventId}).Delete(&models.Event{})
		if res.Error != nil {
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return history.ErrEventNotFound
			}
			return res.Error
		}
		if res.RowsAffected == 0 {
			return history.ErrEventNotFound
		}
		return nil
	})
}

func (er *EventsRepository) DeleteEvents(ef history.EventFilter) error {
	return er.db.Transaction(func(tx *gorm.DB) error {
		query := applyEventFilter(tx.Table(er.tableName()), ef)
		res := query.Delete(&models.Event{})
		if res.Error != nil {
			return res.Error
		}
		return nil
	})
}

func (er *EventsRepository) MarkRead(eventId string) error {
	return er.updateReadStatus(eventId, true)
}

func (er *EventsRepository) MarkUnread(eventId string) error {
	return er.updateReadStatus(eventId, false)
}

func (er *EventsRepository) Cleanup(cleanupFilter history.CleanupFilter) error {
	if cleanupFilter.MaxAge == nil && cleanupFilter.MaxEventsPerUser == nil {
		return nil
	}

	return er.db.Transaction(func(tx *gorm.DB) error {
		table := er.tableName()
		if cleanupFilter.MaxAge != nil {
			cutoff := time.Now().UTC().Add(-*cleanupFilter.MaxAge)
			res := tx.Table(table).
				Where("created_at < ?", cutoff).
				Delete(&models.Event{})
			if res.Error != nil {
				return res.Error
			}
		}

		if cleanupFilter.MaxEventsPerUser != nil {
			res := tx.Table(table).
				Where(whereIdInIdsToBeDeletedByLimitPerUser(table), *cleanupFilter.MaxEventsPerUser).
				Delete(&models.Event{})
			if res.Error != nil {
				return res.Error
			}
		}

		return nil
	})
}

func (er *EventsRepository) IsHealthy() (bool, error) {
	var exists bool
	err := er.db.Raw(
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = ? AND table_name = ?
		)`,
		er.schema,
		models.EventsTable,
	).Scan(&exists).Error
	return err == nil && exists, err
}

func (er *EventsRepository) tableName() string {
	return fmt.Sprintf("%s.%s", pq.QuoteIdentifier(er.schema), pq.QuoteIdentifier(models.EventsTable))
}

func (er *EventsRepository) updateReadStatus(eventId string, read bool) error {
	return er.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Event{}).
			Table(er.tableName()).
			Where("id = ?", eventId).
			Updates(map[string]interface{}{"is_read": read})
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return history.ErrEventNotFound
		}

		return nil
	})
}

func applyEventFilter(query *gorm.DB, ef history.EventFilter) *gorm.DB {
	if ef.TenantId != "" {
		query = query.Where("tenant_id = ?", ef.TenantId)
	}
	if ef.UserId != "" {
		query = query.Where("user_id = ?", ef.UserId)
	}
	if ef.Read != nil {
		query = query.Where("is_read = ?", *ef.Read)
	}
	return query
}

func whereIdInIdsToBeDeletedByLimitPerUser(table string) string {
	return fmt.Sprintf(`id IN (
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (
				PARTITION BY tenant_id, user_id
				ORDER BY created_at DESC
			) AS counter
			FROM %s
		) ranked WHERE counter > ?
	)`, table)
}

func toServiceEvents(events []*models.Event) []*history.Event {
	serviceEvents := []*history.Event{}
	for _, event := range events {
		serviceEvents = append(serviceEvents, &history.Event{
			Id:             event.Id,
			SourceListener: event.ListenerId,
			Properties:     event.Properties,
			CreatedAt:      event.CreatedAt,
			Read:           event.IsRead,
		})
	}
	return serviceEvents
}
