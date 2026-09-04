/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"gorm.io/gorm"
)

type notificationDao struct {
	db     *gorm.DB
	schema string
}

func newNotificationDao(db *gorm.DB, schema string) *notificationDao {
	return &notificationDao{db: db, schema: schema}
}

func (pnd *notificationDao) CreateNotification(origin, category, typ, description, template string) error {
	err := pnd.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(pnd.notificationsTable()).
			Create(&models.Notification{
				Origin:      origin,
				Category:    category,
				Type:        typ,
				Description: description,
				Template:    template,
			})

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return NewError(fmt.Sprintf("notification %s %s %s already exists", origin, category, typ), ErrCodeDuplicate)
		}

		return nil
	})

	if err == nil {
		return nil
	}

	if isDuplicateKeyError(err) {
		return NewError(fmt.Sprintf("notification %s %s %s already exists", origin, category, typ), ErrCodeDuplicate)
	}
	return err
}

func (pnd *notificationDao) UpdateNotification(origin, category, typ, description, template string) error {
	return pnd.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(pnd.notificationsTable()).
			Where(models.Notification{Category: category, Type: typ}).
			Updates(&models.Notification{Description: description, Template: template})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return NewError(fmt.Sprintf("notification %s %s %s not found", origin, category, typ), ErrCodeNotFound)
		}
		return nil
	})
}

func (pnd *notificationDao) FetchNotifications(origin, category, typ string) ([]*models.Notification, error) {
	notifications := make([]*models.Notification, 0)
	query := pnd.db.
		Table(pnd.notificationsTable())
	if origin != "" {
		query.Where(models.Notification{Origin: origin})
	}
	if category != "" {
		query.Where(models.Notification{Category: category})
	}
	if typ != "" {
		query.Where(models.Notification{Type: typ})
	}

	result := query.Find(&notifications)

	if result.Error != nil {
		log.Err(result.Error).Msg("failed to fetch notifications")
		return nil, result.Error
	}

	return notifications, nil
}

func (pnd *notificationDao) DeleteNotification(category, typ string) error {
	return pnd.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Table(pnd.notificationsTable()).
			Where(models.Notification{Category: category, Type: typ}).
			Unscoped().
			Delete(&models.Notification{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return NewError(fmt.Sprintf("notification category %q type %q not found", category, typ), ErrCodeNotFound)
		}
		return nil
	})
}

func (pnd *notificationDao) notificationsTable() string {
	return fmt.Sprintf("%s.%s", pnd.schema, models.NotificationsTable)
}
