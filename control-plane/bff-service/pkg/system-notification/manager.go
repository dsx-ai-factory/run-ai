package system_notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
)

const (
	SystemNotificationSettingKey = "system-notification"
)

type Message struct {
	ID        string `json:"id"`
	Msg       string `json:"msg"`
	CreatedAt int64  `json:"createdAt"`
}

type SystemNotificationManager struct{}

func NewSystemNotificationManager() *SystemNotificationManager {
	return &SystemNotificationManager{}
}

func (s *SystemNotificationManager) Set(ctx context.Context, message *Message) error {
	message.ID = uuid.New().String()
	return cache.Set(ctx, SystemNotificationSettingKey, *message)
}

func (s *SystemNotificationManager) Invalidate(ctx context.Context) error {
	return cache.Del[Message](ctx, SystemNotificationSettingKey)
}

func (s *SystemNotificationManager) Get(ctx context.Context) (*Message, error) {
	return cache.Get[Message](ctx, SystemNotificationSettingKey)
}
