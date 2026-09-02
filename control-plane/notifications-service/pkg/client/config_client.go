/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/maypok86/otter/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/config_client"
)

type ConfigClienter interface {
	ShouldNotify(tenant string) (bool, error)
	GetEnabledTenants() ([]string, error)
	CreateNotification(origin, category, typ, description, template string) error
	UpdateNotification(origin, category, typ, description, template string) error
	DeleteNotification(category, typ string) error
	GetAllOriginNotification(origin string) ([]config_client.NotificationKind, error)
}

var ErrNotificationKindNotFound = errors.New("notification kind not found")

type ConfigClient struct {
	client *config_client.APIClient
	cache  *otter.Cache[string, bool]
}

func NewConfigClient(config *NotifyClientConfig) (ConfigClienter, error) {
	cache := otter.Must[string, bool](&otter.Options[string, bool]{
		ExpiryCalculator: otter.ExpiryWriting[string, bool](time.Duration(config.CacheTTLSec * int(time.Second))),
	})

	host, scheme := getHostAndScheme(config.ConfigUrl)
	cfgServer := config_client.NewConfiguration()
	cfgServer.Host = host
	cfgServer.Scheme = scheme

	return &ConfigClient{
		client: config_client.NewAPIClient(cfgServer),
		cache:  cache,
	}, nil
}

func (cc *ConfigClient) ShouldNotify(tenant string) (bool, error) {
	notify, exist := cc.cache.GetIfPresent(tenant)
	if exist {
		return notify, nil
	}

	tenantRoutes, res, err := cc.client.TenantRouteAPI.GetRoutes(context.Background(), tenant).Execute()
	if err != nil {
		return false, err
	}
	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("cant get routes for tenant %s, status code %d, response: %s", tenant, res.StatusCode, res.Status)
	}

	shouldNotify := len(tenantRoutes) > 0
	cc.cache.Set(tenant, shouldNotify)
	return shouldNotify, nil
}

func (cc *ConfigClient) GetEnabledTenants() ([]string, error) {
	tenantIds, res, err := cc.client.TenantRouteAPI.GetAllTenantsWithRoutes(context.Background()).Execute()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cant get tenants with enabled notifications %s", res.Status)
	}

	return tenantIds, nil
}

func (cc *ConfigClient) DeleteNotification(category, typ string) error {
	_, res, err := cc.client.RegisterNotificationAPI.DeleteNotificationKind(context.Background()).Category(category).Type_(typ).Execute()
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", res.Status)
	}

	return err
}

func (cc *ConfigClient) GetAllOriginNotification(origin string) ([]config_client.NotificationKind, error) {
	notifications, res, err := cc.client.RegisterNotificationAPI.GetNotificationKinds(context.Background()).Origin(origin).Execute()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", res.Status)
	}

	return notifications, err
}

func (cc *ConfigClient) CreateNotification(origin, category, typ, description, template string) error {
	_, res, err := cc.client.RegisterNotificationAPI.CreateNotificationKind(context.Background()).NotificationKind(config_client.NotificationKind{
		Origin:      origin,
		Category:    category,
		Type:        typ,
		Description: description,
		Template:    template,
	}).Execute()
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return ErrNotificationKindNotFound
		}
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s", res.Status)
	}

	return nil
}

func (cc *ConfigClient) UpdateNotification(origin, category, typ, description, template string) error {
	_, res, err := cc.client.RegisterNotificationAPI.UpdateNotificationKind(context.Background()).NotificationKind(config_client.NotificationKind{
		Origin:      origin,
		Category:    category,
		Type:        typ,
		Description: description,
		Template:    template,
	}).Execute()
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return ErrNotificationKindNotFound
		}
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	if res.StatusCode == http.StatusNotFound {
		return ErrNotificationKindNotFound
	}
	return fmt.Errorf("request failed to update notification kind: %s", res.Status)
}
