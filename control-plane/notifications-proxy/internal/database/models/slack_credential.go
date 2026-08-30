package models

import (
	"time"
)

type SlackCredential struct {
	Id           int16     `gorm:"type:smallint;primaryKey;default:1"`
	AppId        string    `gorm:"column:app_id"`
	ClientId     string    `gorm:"column:client_id;not null"`
	ClientSecret string    `gorm:"column:client_secret;not null;serializer:encrypted"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
