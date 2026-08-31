/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	events_http_client "github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/notify"
	"github.com/sethvargo/go-retry"
)

type Template int

const (
	TWorkloads Template = iota
	TJobs
)

type NotifyClienter interface {
	Notify(n *Notification)
	RegisterErrorHandler(handler ErrorHandlerFunc)
	ShouldNotify(tenant string) bool
	GetEnabledTenants() []string
	RegisterNotifications(origin string, notifications []*NotificationKind, removeMissing *bool) error
}

type NotifyClient struct {
	config           NotifyClientConfig
	eclient          NotificationsEventClienter
	cclient          ConfigClienter
	errorHandlerFunc ErrorHandlerFunc
	notifications    chan events_http_client.Notification
	batchSize        int
	batchInterval    time.Duration
	ctx              context.Context
	wg               *sync.WaitGroup
}

func NewNotifyClient(eclient NotificationsEventClienter, cclient ConfigClienter, errorHandlerFunc ErrorHandlerFunc, config NotifyClientConfig, ctx context.Context) *NotifyClient {
	nc := &NotifyClient{
		eclient:          eclient,
		cclient:          cclient,
		errorHandlerFunc: errorHandlerFunc,
		config:           config,
		batchSize:        config.BatchSize,
		notifications:    make(chan events_http_client.Notification, config.BatchSize*10),
		batchInterval:    time.Duration(config.BatchIntervalSec) * time.Second,
		ctx:              ctx,
		wg:               &sync.WaitGroup{},
	}

	go nc.batchNotifier()
	return nc
}

func newNotifyClient(config *NotifyClientConfig, cc ConfigClienter, errorHandlerFunc ErrorHandlerFunc, ctx context.Context) *NotifyClient {
	return NewNotifyClient(newNotificationsEventClient(config), cc, errorHandlerFunc, *config, ctx)
}

func (nc *NotifyClient) RegisterNotifications(origin string, notifications []*NotificationKind, removeMissing *bool) error {
	if removeMissing != nil && *removeMissing {
		err := nc.deleteUnlistedNotificationKinds(origin, notifications)
		if err != nil {
			return err
		}
	}

	return nc.registerNotificationKinds(origin, notifications)
}

func (nc *NotifyClient) RegisterErrorHandler(handler ErrorHandlerFunc) {
	nc.errorHandlerFunc = handler
}

func (nc *NotifyClient) Notify(n *Notification) {
	if !nc.ShouldNotify(n.Tenant) {
		return
	}

	select {
	case nc.notifications <- buildHttpNotification(n):
	default:
		nc.errorHandlerFunc(errors.New("failed to add message"))
	}
}

func (nc *NotifyClient) ShouldNotify(tenant string) bool {
	notify, err := nc.cclient.ShouldNotify(tenant)
	if err != nil {
		nc.errorHandlerFunc(err)
	}

	return notify
}

func (nc *NotifyClient) GetEnabledTenants() []string {
	tenants, err := nc.cclient.GetEnabledTenants()
	if err != nil {
		nc.errorHandlerFunc(err)
		return nil
	}

	return tenants
}

func (nc *NotifyClient) sendNotificationWithRetry(notificationsToSend []events_http_client.Notification) {
	if len(notificationsToSend) < 1 {
		return
	}

	retryableNotifyFunc := nc.getRetryableNotifyFunc(notificationsToSend)
	err := retry.Do(context.Background(), nc.getRetryLimits(), retryableNotifyFunc)
	if err != nil {
		nc.errorHandlerFunc(err)
	}
}

func (nc *NotifyClient) getRetryableNotifyFunc(notificationsToSend []events_http_client.Notification) retry.RetryFunc {
	return func(ctx context.Context) error {
		res, err := nc.eclient.NotifyBatch(notificationsToSend)
		if err == nil {
			return nil
		}

		if res != nil {
			notificationsToSend = notificationsToSend[res.Succeed:]
		}

		return retry.RetryableError(fmt.Errorf("failed to ingest batch: %s. Retrying\nmsgs left: %d", err.Error(), len(notificationsToSend)))
	}
}

func (nc *NotifyClient) WaitDone() {
	nc.wg.Wait()
}

func (nc *NotifyClient) getRetryLimits() retry.Backoff {
	maxRetries := nc.config.RetryThreshold
	retryInterval := retry.NewConstant(time.Duration(nc.config.RetryIntervalSec) * time.Second)

	return retry.WithMaxRetries(uint64(maxRetries), retryInterval)
}

func buildHttpNotification(n *Notification) events_http_client.Notification {
	properties := map[string]interface{}{
		"tenant_id":   n.Tenant,
		"user_id":     n.User,
		"title":       n.Title,
		"description": n.Description,
		"category":    n.Category,
		"type":        n.Type,
	}

	for name, val := range n.AdditionalProperties {
		properties[name] = val
	}

	return events_http_client.Notification{
		TenantId:   n.Tenant,
		EventType:  n.Type,
		Properties: properties,
	}
}

func (nc *NotifyClient) batchNotifier() {
	nc.listenToNotifications()
}

func (nc *NotifyClient) listenToNotifications() {
	var notificationsToSend []events_http_client.Notification

	nc.wg.Add(1)
	defer nc.wg.Done()

	for {
		select {
		case <-time.After(nc.batchInterval):
			notificationsToSend = nc.notifyBulk(notificationsToSend, 1)
		case <-nc.ctx.Done():
			nc.sendNotificationWithRetry(notificationsToSend)
			return
		case notification := <-nc.notifications:
			notificationsToSend = append(notificationsToSend, notification)
			notificationsToSend = nc.notifyBulk(notificationsToSend, nc.config.BatchSize)
		}
	}
}

func (nc *NotifyClient) notifyBulk(notificationsToSend []events_http_client.Notification, minSize int) []events_http_client.Notification {
	if len(notificationsToSend) < minSize {
		return notificationsToSend
	}

	nc.sendNotificationWithRetry(notificationsToSend)
	return []events_http_client.Notification{}
}

func (nc *NotifyClient) deleteUnlistedNotificationKinds(origin string, notifications []*NotificationKind) error {
	registeredNotifications, err := nc.cclient.GetAllOriginNotification(origin)
	if err != nil {
		nc.errorHandlerFunc(err)
		return err
	}

	return nc.deleteUnlisted(registeredNotifications, notifications)
}

func (nc *NotifyClient) registerNotificationKinds(origin string, notifications []*NotificationKind) error {
	for _, n := range notifications {
		if err := nc.createNotificationKind(origin, n); err != nil {
			return err
		}
	}

	return nil
}

func (nc *NotifyClient) deleteUnlisted(registeredNotifications []config_client.NotificationKind, notifications []*NotificationKind) error {
	notificationsMap := parseToMap(notifications)
	for _, n := range registeredNotifications {
		err := nc.deleteNotificationKind(notificationsMap, n)
		if err != nil {
			nc.errorHandlerFunc(err)
			return err
		}
	}
	return nil
}

func (nc *NotifyClient) deleteNotificationKind(notificationsMap map[string]*NotificationKind, n config_client.NotificationKind) error {
	if notificationsMap[notificationId(n.Category, n.Type)] != nil {
		return nil
	}

	return nc.cclient.DeleteNotification(n.Category, n.Type)
}

func (nc *NotifyClient) createNotificationKind(origin string, n *NotificationKind) error {
	err := nc.cclient.UpdateNotification(origin, n.Category, n.Type, n.Description, n.Template)
	if err == nil {
		return nil
	}

	if !errors.Is(err, ErrNotificationKindNotFound) {
		nc.errorHandlerFunc(err)
		return err
	}

	if err := nc.cclient.CreateNotification(origin, n.Category, n.Type, n.Description, n.Template); err != nil {
		nc.errorHandlerFunc(err)
		return err
	}
	return nil
}

func parseToMap(notifications []*NotificationKind) map[string]*NotificationKind {
	notificationsMap := map[string]*NotificationKind{}
	for _, n := range notifications {
		notificationsMap[notificationId(n.Category, n.Type)] = n
	}

	return notificationsMap
}

func notificationId(category, typ string) string {
	return fmt.Sprintf("%s:%s", category, typ)
}
