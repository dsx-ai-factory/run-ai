/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	SettingsTableName = "settings"
)

type TenantSettingInput struct {
	TenantID string
	Value    []byte
}

type Setting struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key"`
	Key               string    `gorm:"not null"`
	Value             []byte
	TenantID          string `gorm:"not null"`
	UpdatedAt         time.Time
	UpdatedByCustomer bool `gorm:"not null"`
}

type PostgresSettingDal struct {
}

func NewSettingDAL() *PostgresSettingDal {
	return &PostgresSettingDal{}
}

func (sd PostgresSettingDal) TableName() string {
	return SettingsTableName
}

func (sd PostgresSettingDal) GetSettingByKey(key, tenantId string) (*Setting, error) {
	setting := &Setting{}
	query := runaidb.DB().Table(sd.TableName()).Where("key = ? and tenant_id = ?", key, tenantId)
	err := query.First(setting)
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (sd PostgresSettingDal) GetSettingsByTenantId(tenantId string) (*[]Setting, error) {
	allSettings := &[]Setting{}
	query := runaidb.DB().Table(sd.TableName()).Where("tenant_id = ?", tenantId)
	err := query.Find(allSettings)
	if err != nil {
		return nil, err
	}
	return allSettings, nil
}

func (sd PostgresSettingDal) SaveSetting(key, tenantId string, updateByUser bool, value []byte) error {
	existingSetting, err := sd.GetSettingByKey(key, tenantId)

	if err == nil && existingSetting != nil {
		existingSetting.Value = value
		existingSetting.UpdatedByCustomer = updateByUser
		existingSetting.UpdatedAt = time.Now()

		_, err := runaidb.DB().Table(sd.TableName()).
			Where("key = ? AND tenant_id = ?", key, tenantId).
			Updates(existingSetting)

		if err != nil {
			return err
		}
	} else {
		settingToSave := &Setting{
			ID:                uuid.New(),
			Key:               key,
			Value:             value,
			TenantID:          tenantId,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: updateByUser,
		}

		_, err := runaidb.DB().Table(sd.TableName()).Create(settingToSave)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sd PostgresSettingDal) ShouldMigrateSetting(key string) (bool, error) {
	setting := &Setting{}
	query := runaidb.DB().Table(sd.TableName()).Where("key = ?", key)
	err := query.First(setting)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (sd PostgresSettingDal) SaveSettingByKeyForTenantsList(key string, tenantSettings []TenantSettingInput) error {
	if len(tenantSettings) == 0 {
		return nil
	}

	settings := make([]Setting, len(tenantSettings))
	now := time.Now()
	for i, ts := range tenantSettings {
		settings[i] = Setting{
			ID:                uuid.New(),
			Key:               key,
			Value:             ts.Value,
			TenantID:          ts.TenantID,
			UpdatedAt:         now,
			UpdatedByCustomer: true,
		}
	}

	_, err := runaidb.DB().Table(sd.TableName()).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&settings)
	return err
}
