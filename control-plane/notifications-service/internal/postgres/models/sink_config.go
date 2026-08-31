/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	SinkTable = "sink_config"
)

type SinkConfig struct {
	gorm.Model
	SinkId string         `gorm:"not null;primaryKey" json:"sink_id"`
	Config datatypes.JSON `gorm:"type:jsonb,not null,serializer:json" json:"config"`
	Type   string         `gorm:"not null" json:"type"`
}

func (s *SinkConfig) TableName() string {
	return SinkTable
}

func (s *SinkConfig) GetConfig() datatypes.JSON {
	return s.Config
}

func (s *SinkConfig) SetConfig(config datatypes.JSON) {
	s.Config = config
}

func NewEmptySinkConfig() *SinkConfig {
	return &SinkConfig{Config: datatypes.JSON{}}
}
