package subscription

import (
	"context"
)

type Manager struct {
	userRoutesClient UserRoutesClienter
}

func NewManager(userRoutesClient UserRoutesClienter) *Manager {
	return &Manager{
		userRoutesClient: userRoutesClient,
	}
}

func (m *Manager) DeleteSubscription(ctx context.Context, tenantId, userId, routeId string) error {
	return m.userRoutesClient.Delete(ctx, tenantId, userId, routeId)
}

func (m *Manager) CreateSubscription(ctx context.Context, tenantId, userId string, subscription *Subscription) (*Subscription, error) {
	subscriptions, err := m.userRoutesClient.GetUserSubscriptions(ctx, tenantId, userId)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) > 0 {
		return nil, ErrCreatingForbidden
	}

	return m.userRoutesClient.Create(ctx, tenantId, userId, *subscription)
}

func (m *Manager) GetUserSubscriptions(ctx context.Context, tenantId, userId string) ([]Subscription, error) {
	return m.userRoutesClient.GetUserSubscriptions(ctx, tenantId, userId)
}

func (m *Manager) UpdateSubscription(ctx context.Context, tenantId, userId, routeId string, subscription *Subscription) error {
	subscription.Id = routeId
	return m.userRoutesClient.Update(ctx, tenantId, userId, *subscription)
}
