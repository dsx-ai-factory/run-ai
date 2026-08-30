package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	ListenersConfigsTable = "listeners_config"
)

type ListenersConfig struct {
	gorm.Model
	Id             uint64         `gorm:"primaryKey;autoIncrement:true" json:"id"`
	ListenerId     string         `gorm:"not null;primaryKey" json:"listener_id"`
	ListenerType   string         `gorm:"not null" json:"listener_type"`
	ListenerConfig datatypes.JSON `gorm:"type:jsonb,not null,serializer:json" json:"listener_config"`
}

func (r *ListenersConfig) TableName() string {
	return ListenersConfigsTable
}

func (r *ListenersConfig) GetConfig() datatypes.JSON {
	return r.ListenerConfig
}

func (r *ListenersConfig) SetConfig(config datatypes.JSON) {
	r.ListenerConfig = config
}
