package models

import (
	"gorm.io/gorm"
)

const (
	NotificationsTable = "notifications"
)

type Notification struct {
	gorm.Model
	Category    string `gorm:"not null;primaryKey" json:"category_id"`
	Type        string `gorm:"not null;primaryKey" json:"type"`
	Description string `gorm:"not null" json:"description"`
	Template    string `gorm:"not null" json:"template"`
	Origin      string `gorm:"not null" json:"origin"`
}

func (n *Notification) TableName() string {
	return NotificationsTable
}
