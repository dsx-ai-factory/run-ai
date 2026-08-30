package subscription

import (
	"context"
	"errors"
)

var (
	ErrClient                    = errors.New("client error")
	ErrFailedToParse             = errors.New("failed to parse subscription")
	ErrNotFound                  = errors.New("not found")
	ErrCreatingForbidden         = errors.New("creating additional subscription is forbidden")
	ErrSubscriptionAlreadyExists = errors.New("subscription with same id already exist")
)

type UserRoutesClienter interface {
	Create(ctx context.Context, tenantId, userId string, route Subscription) (*Subscription, error)
	Update(ctx context.Context, tenantId, userId string, route Subscription) error
	Delete(ctx context.Context, tenantId, userId, routeId string) error
	GetUserSubscriptions(ctx context.Context, tenantId, userId string) ([]Subscription, error)
}

type SubscriptionEvents []SubscriptionEvent
type Subscription struct {
	Events   SubscriptionEvents
	Channels []string
	Id       string
}

type SubscriptionEvent struct {
	Category string
	Types    []string
}
