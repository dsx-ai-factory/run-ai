package notification_state

import (
	"context"
	"errors"
)

const (
	EmailEnabled       = "email_enabled"
	SlackEnabled       = "slack_enabled"
	SlackWorkspaceName = "slack_workspace_name"
)

var (
	ErrClientFailure                    = errors.New("client failure")
	ErrFailedToRemoveSinkFromUserRoutes = errors.New("failed to remove sink from user routes")
	ErrFailedToUpdateNotificationState  = errors.New("failed to update notifications state")
	ErrNoSinkConfigured                 = errors.New("no sinks configured")
	ErrIncompatibleSinkType             = errors.New("incompatible sink type")
)

type StateClienter interface {
	Get(ctx context.Context, tenantId string) ([]Stater, error)
	GetDetails(ctx context.Context, tenantId string) ([]Stater, error)
	UpdateState(ctx context.Context, tenantId string, notificationState []Stater) ([]Stater, error)
	GetSlack(ctx context.Context, tenantId string) (Stater, error)
	GetEmail(ctx context.Context, tenantId string) (Stater, error)
	GetSlackDetails(ctx context.Context, tenantId string) (Stater, error)
	UpdateEmail(ctx context.Context, tenantId string, notificationState Stater) (Stater, error)
	UpdateSlack(ctx context.Context, tenantId string, notificationState Stater) (Stater, error)
}

type StateMigrationClient interface {
	Migrate(ctx context.Context, tenantId string) error
}

type ChannelClienter interface {
	IsExists(ctx context.Context, tenantId string, state Stater) (bool, error)
}

type SubUserClienter interface {
	DisableChannel(ctx context.Context, tenantId string, channelId string) error
}

type Stater interface {
	IsEnabled() bool
	GetType() string
	GetChanName() string
	GetMetadata() any
}
