/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notifiy_client

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/utils/errors"

	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	tenants_api "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

// tenantsRequestTimeout bounds the tenants-manager calls made from the notifier
// goroutines, which run while holding the notifier lock - an unbounded request
// would block every following notifier cycle.
const tenantsRequestTimeout = 30 * time.Second

type NotifyClient interface {
	InitNotifyClient(ctx context.Context, config *notifyclient.NotifyClientConfig) error
	GetEnabledTenants(ctx context.Context) ([]string, error)
	RegisterNotifications(origin string, notifications []*notifyclient.NotificationKind, removeMissing *bool) error
	BuildNotification(category string, typ string) notifyclient.NotifyBuilderer
	GetTenantBaseUrl(ctx context.Context, tenantId string) (string, error)
}

type NotifyClienter struct {
	tenantsClient *tenants_api.APIClient
	// map of tenantId -> url
	tenantsUrlsCache      map[string]string
	tenantsUrlsCacheMutex sync.RWMutex
}

func NewNotifyClienter(conf *config.ServicesConfig) NotifyClient {
	tenantsClient := tenants_api.NewAPIClient(&tenants_api.Configuration{
		Servers: tenants_api.ServerConfigurations{
			{
				URL: conf.TenantsManagerUrl,
			},
		},
	})
	return &NotifyClienter{
		tenantsClient:         tenantsClient,
		tenantsUrlsCache:      map[string]string{},
		tenantsUrlsCacheMutex: sync.RWMutex{},
	}
}

func (c *NotifyClienter) InitNotifyClient(ctx context.Context, config *notifyclient.NotifyClientConfig) error {
	logging.Infof("Initializing notify client with config: %+v", config)

	_, err := notifyclient.Init(config, ctx)
	if err != nil {
		logging.Errorf("Failed to init notify client, error: %s", err.Error())
		return err
	}

	notifyclient.RegisterErrorHandler(func(err error) {
		logging.Errorf("Failure in notify client, error: %s", err.Error())
	})

	return nil
}

func (c *NotifyClienter) GetEnabledTenants(ctx context.Context) ([]string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, tenantsRequestTimeout)
	defer cancel()

	tenants, _, err := c.tenantsClient.InternalAPI.ManagementGetTenantsInternal(reqCtx).Execute()
	if err != nil {
		return nil, fmt.Errorf("get tenants: %w", err)
	}
	return enabledTenantIds(tenants, notifyclient.GetEnabledTenants()), nil
}

func enabledTenantIds(tenants []tenants_api.Tenant, routeIds []string) []string {
	enabledRoutes := make(map[string]bool, len(routeIds))
	for _, routeId := range routeIds {
		enabledRoutes[routeId] = true
	}

	tenantIds := make([]string, 0, len(tenants))
	for _, tenant := range tenants {
		tenantId := strconv.FormatInt(int64(tenant.Id), 10)
		if _, enabled := enabledRoutes[tenantId]; !enabled {
			continue
		}
		tenantIds = append(tenantIds, tenantId)
	}

	return tenantIds
}

func (c *NotifyClienter) RegisterNotifications(origin string, notifications []*notifyclient.NotificationKind, removeMissing *bool) error {
	return notifyclient.RegisterNotification(origin, notifications, removeMissing)
}

func (c *NotifyClienter) BuildNotification(category string, typ string) notifyclient.NotifyBuilderer {
	return notifyclient.BuildNotification(category, typ)
}

func (c *NotifyClienter) GetTenantBaseUrl(ctx context.Context, tenantId string) (string, error) {
	c.tenantsUrlsCacheMutex.RLock()
	url, ok := c.tenantsUrlsCache[tenantId]
	c.tenantsUrlsCacheMutex.RUnlock()

	if ok && url != "" {
		return url, nil
	}

	var err error
	url, err = c.getTenantBaseUrl(ctx, tenantId)
	if err != nil {
		return "", err
	}

	c.tenantsUrlsCacheMutex.Lock()
	defer c.tenantsUrlsCacheMutex.Unlock()
	c.tenantsUrlsCache[tenantId] = url
	return url, nil
}

func (c *NotifyClienter) getTenantBaseUrl(ctx context.Context, tenantId string) (string, error) {
	i, err := strconv.ParseInt(tenantId, 10, 32)
	if err != nil {
		logging.Errorf("Failed to parse tenantId %s to int, error: %s", tenantId, err.Error())
		return "", err
	}
	id := int32(i)

	reqCtx, cancel := context.WithTimeout(ctx, tenantsRequestTimeout)
	defer cancel()

	domain, _, err := c.tenantsClient.InternalAPI.GetTenantDomain(reqCtx, id).Execute()
	if err != nil {
		if !errors.IsNotFound(err) {
			logging.Errorf("Failed to get tenant domain for tenantId %s, error: %s", tenantId, err.Error())
		}
		return "", err
	}
	return domain, nil
}
