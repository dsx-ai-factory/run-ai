package models

import (
	"time"
)

const (
	EventsTable = "events"
)

type Event struct {
	Id         string `gorm:"not null;primaryKey;unique" json:"id"`
	TenantId   string `gorm:"not null;primaryKey;index:idx_events_tenant_user_created_at,priority:1" json:"tenant_id"`
	UserId     string `gorm:"not null;primaryKey;index:idx_events_tenant_user_created_at,priority:2" json:"user_id"`
	ListenerId string `gorm:"not null" json:"listener_id"`

	Description string                 `gorm:"not null" json:"description"`
	Properties  map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"properties"`

	IsRead    bool      `gorm:"column:is_read" json:"is_read"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_events_created_at;index:idx_events_tenant_user_created_at,priority:3,sort:desc" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (e *Event) TableName() string {
	return EventsTable
}
