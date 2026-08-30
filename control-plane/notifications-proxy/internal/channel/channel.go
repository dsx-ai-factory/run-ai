package channel

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/email"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
)

var (
	ErrClientFailure       = fmt.Errorf("client error")
	ErrAlreadyExists       = fmt.Errorf("channel already exist")
	ErrNotFound            = fmt.Errorf("channel not found")
	ErrFailedToParse       = fmt.Errorf("failed to parse server response")
	ErrNotSupportedType    = fmt.Errorf("channel not supported type")
	ErrValidationFailed    = fmt.Errorf("validation failed")
	ErrChannelNameNotValid = fmt.Errorf("channel name not valid")
)

type Clienter interface {
	Create(ctx context.Context, tenantId string, channel NotificationChannel) error
	Validate(ctx context.Context, tenantId string, channel NotificationChannel) (ValidationResponse, error)
	Update(ctx context.Context, tenantId, channelId string, channel NotificationChannel) error
	Patch(ctx context.Context, tenantId, channelId string, config NotificationChannel) error
	GetTenantChannels(ctx context.Context, tenantId string) ([]NotificationChannelMetadata, error)
	Get(ctx context.Context, tenantId, channelId string) (*NotificationChannel, error)
	Delete(ctx context.Context, tenantId, channelId string) error
}

type NotificationChannelMetadata struct {
	Id   string
	Type string
}

type NotificationChannel struct {
	Id    string
	Type  string
	Slack slack.ChannelConfig
	Email email.ChannelConfig
}

type ValidationResponse struct {
	Status  bool
	Message string
}
