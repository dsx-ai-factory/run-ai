/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type listenerDao struct {
	db     *gorm.DB
	schema string
}

func newListenerDao(db *gorm.DB, schema string) *listenerDao {
	return &listenerDao{db: db, schema: schema}
}

func (pld *listenerDao) FetchListeners() ([]*models.ListenersConfig, error) {
	listeners, err := fetch[models.ListenersConfig](pld.db, pld.listenersConfigsTable(), nil)
	if err != nil {
		log.Err(err).Msg("failed to fetch listeners")
		return nil, err
	}
	return listeners, nil
}

func (pld *listenerDao) FetchListenersByPort(port int) ([]*models.ListenersConfig, error) {
	listeners := make([]*models.ListenersConfig, 0)
	query := pld.db.
		Table(pld.listenersConfigsTable()).
		Where("(listener_config::jsonb #> '{http_listener,port}')::int = ?", port)
	result := query.Find(&listeners)

	if result.Error != nil {
		log.Err(result.Error).Msgf("failed to fetch listener with port %d", port)
		return listeners, result.Error
	}

	return listeners, nil
}

func (pld *listenerDao) FetchListener(id string) (*models.ListenersConfig, error) {
	listeners := make([]*models.ListenersConfig, 0)
	query := pld.db.
		Table(pld.listenersConfigsTable()).
		Where(models.ListenersConfig{ListenerId: id})
	result := query.Find(&listeners)

	if result.Error != nil {
		log.Err(result.Error).Msgf("failed to fetch listener %s", id)
		return nil, result.Error
	}

	if len(listeners) == 0 {
		return nil, nil
	}

	return listeners[0], nil
}

func (pld *listenerDao) CreateListener(listenerId string, listenerType string, config datatypes.JSON) error {
	err := pld.db.Transaction(func(tx *gorm.DB) error {
		return tx.
			Table(pld.listenersConfigsTable()).
			Create(&models.ListenersConfig{ListenerId: listenerId, ListenerType: listenerType, ListenerConfig: config}).
			Error
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return NewError("listener with same id already exists", ErrCodeDuplicate)
		}
		return err
	}
	return nil
}

func (pld *listenerDao) UpdateListener(listenerId string, listenerType string, config datatypes.JSON) error {
	return pld.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(pld.listenersConfigsTable()).
			Where(models.ListenersConfig{ListenerId: listenerId}).
			Updates(&models.ListenersConfig{ListenerType: listenerType, ListenerConfig: config})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return NewError("listener not found to be updated", ErrCodeNotFound)
		}
		return nil
	})
}

func (pld *listenerDao) DeleteListener(listenerId string) error {
	return pld.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Table(pld.listenersConfigsTable()).
			Where(models.ListenersConfig{ListenerId: listenerId}).
			Unscoped().
			Delete(&models.ListenersConfig{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return NewError("listener not found to be deleted", ErrCodeNotFound)
		}
		return nil
	})
}

func (pld *listenerDao) IsHealthy() (bool, error) {
	res := pld.db.Exec(fmt.Sprintf("select count (*) from %s", pld.listenersConfigsTable()))
	return res.Error == nil, res.Error
}

func (pld *listenerDao) listenersConfigsTable() string {
	return fmt.Sprintf("%s.%s", pld.schema, models.ListenersConfigsTable)
}
