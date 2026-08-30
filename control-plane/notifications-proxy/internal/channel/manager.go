package channel

import (
	"context"
)

type Manager struct {
	channelsClient Clienter
}

func NewManager(client Clienter) *Manager {
	return &Manager{channelsClient: client}
}

func (m *Manager) GetNotificationsChannels(ctx context.Context, tenantId string) ([]NotificationChannelMetadata, error) {
	return m.channelsClient.GetTenantChannels(ctx, tenantId)
}

func (m *Manager) DeleteChannel(ctx context.Context, tenantId, sinkId string) error {
	return m.channelsClient.Delete(ctx, tenantId, sinkId)
}

func (m *Manager) ValidateChannel(ctx context.Context, tenantId string, config NotificationChannel) (ValidationResponse, error) {
	return m.channelsClient.Validate(ctx, tenantId, config)
}

func (m *Manager) CreateChannel(ctx context.Context, tenantId string, nc NotificationChannel) error {
	return m.channelsClient.Create(ctx, tenantId, nc)
}

func (m *Manager) GetChannel(ctx context.Context, tenantId, channelId string) (*NotificationChannel, error) {
	return m.channelsClient.Get(ctx, tenantId, channelId)
}

func (m *Manager) UpdateChannel(ctx context.Context, tenantId, channelId string, config NotificationChannel) error {
	return m.channelsClient.Update(ctx, tenantId, channelId, config)
}

func (m *Manager) PatchChannel(ctx context.Context, tenantId, channelId string, config NotificationChannel) error {
	return m.channelsClient.Patch(ctx, tenantId, channelId, config)
}
