package notifiy_client

import (
	"context"
	"fmt"

	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
)

type NotifyClienterMock struct {
	// category -> type -> notification / notification kind
	registeredNotifications map[string]map[string]*notifyclient.NotificationKind
	notifiedNotifications   map[string]map[string][]*notifyclient.Notification

	errors []error
}

type NotifyBuildererMock struct {
	notifyBuilderer notifyclient.NotifyBuilderer
	notifyBuilder   *notifyclient.NotifyBuilder
	clienterMock    *NotifyClienterMock
}

func NewNotifyClienterMock() *NotifyClienterMock {
	return &NotifyClienterMock{
		registeredNotifications: map[string]map[string]*notifyclient.NotificationKind{},
		notifiedNotifications:   map[string]map[string][]*notifyclient.Notification{},
		errors:                  []error{},
	}
}

func (c *NotifyClienterMock) InitNotifyClient(ctx context.Context, config *notifyclient.NotifyClientConfig) error {
	return nil
}

func (c *NotifyClienterMock) GetEnabledTenants() []string {
	return []string{testing.TestTenant}
}

func (c *NotifyClienterMock) RegisterNotifications(origin string, notifications []*notifyclient.NotificationKind, removeMissing *bool) error {
	for _, notification := range notifications {
		if _, ok := c.registeredNotifications[notification.Category]; !ok {
			c.registeredNotifications[notification.Category] = map[string]*notifyclient.NotificationKind{}
		}

		if _, ok := c.registeredNotifications[notification.Category][notification.Type]; ok {
			return fmt.Errorf("notification already registered: %s/%s", notification.Category, notification.Type)
		}

		c.registeredNotifications[notification.Category][notification.Type] = notification
	}

	return nil
}

func (c *NotifyClienterMock) BuildNotification(category string, typ string) notifyclient.NotifyBuilderer {
	return NewNotifyClientBuildereMock(c, category, typ)
}

func (c *NotifyClienterMock) GetTenantBaseUrl(tenantId string) (string, error) {
	return "https://mock.com", nil
}

func (c *NotifyClienterMock) ResetNotifiedNotifications() {
	c.notifiedNotifications = map[string]map[string][]*notifyclient.Notification{}
	c.errors = []error{}
}

func NewNotifyClientBuildereMock(notifyClienterMock *NotifyClienterMock,
	category string, typ string,
) notifyclient.NotifyBuilderer {
	builder := notifyclient.NewNotifyBuilder(category, typ)
	return &NotifyBuildererMock{
		notifyBuilderer: builder,
		notifyBuilder:   builder,
		clienterMock:    notifyClienterMock,
	}
}

func (b *NotifyBuildererMock) Notify() {
	notification := b.notifyBuilder.Notification

	if _, ok := b.clienterMock.notifiedNotifications[notification.Category]; !ok {
		b.clienterMock.notifiedNotifications[notification.Category] = map[string][]*notifyclient.Notification{}
	}
	if _, ok := b.clienterMock.notifiedNotifications[notification.Category][notification.Type]; !ok {
		b.clienterMock.notifiedNotifications[notification.Category][notification.Type] = []*notifyclient.Notification{}
	}

	// validate this notification is registered
	if _, ok := b.clienterMock.registeredNotifications[notification.Category]; !ok {
		b.clienterMock.errors = append(b.clienterMock.errors, fmt.Errorf("category not registered: %s", notification.Category))
		return
	}
	if _, ok := b.clienterMock.registeredNotifications[notification.Category][notification.Type]; !ok {
		b.clienterMock.errors = append(b.clienterMock.errors, fmt.Errorf("type not registered: %s/%s", notification.Category, notification.Type))
		return
	}

	b.clienterMock.notifiedNotifications[notification.Category][notification.Type] = append(
		b.clienterMock.notifiedNotifications[notification.Category][notification.Type], notification)
}

func (b *NotifyBuildererMock) WithTitle(title string) notifyclient.NotifyBuilderer {
	b.notifyBuilderer = b.notifyBuilderer.WithTitle(title)
	return b
}

func (b *NotifyBuildererMock) WithDescription(description string) notifyclient.NotifyBuilderer {
	b.notifyBuilderer = b.notifyBuilderer.WithDescription(description)
	return b
}

func (b *NotifyBuildererMock) WithTenant(tenant string) notifyclient.NotifyBuilderer {
	b.notifyBuilderer = b.notifyBuilderer.WithTenant(tenant)
	return b
}

func (b *NotifyBuildererMock) WithUser(user string) notifyclient.NotifyBuilderer {
	b.notifyBuilderer = b.notifyBuilderer.WithUser(user)
	return b
}

func (b *NotifyBuildererMock) WithProperty(property string, val interface{}) notifyclient.NotifyBuilderer {
	b.notifyBuilderer = b.notifyBuilderer.WithProperty(property, val)
	return b
}
