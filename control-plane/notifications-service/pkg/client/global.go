/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"sync"
)

var (
	globalClient NotifyClienter
	once         sync.Once
)

func Init(config *NotifyClientConfig, ctx context.Context) (NotifyClienter, error) {
	return initGlobalClient(config, ctx, LoadDefaultsToConfig)
}

func RegisterNotification(origin string, notifications []*NotificationKind, removeMissing *bool) error {
	if globalClient == nil {
		return fmt.Errorf("client not initialized")
	}

	return globalClient.RegisterNotifications(origin, notifications, removeMissing)
}

func ShouldNotify(tenant string) bool {
	if globalClient == nil {
		return false
	}

	return globalClient.ShouldNotify(tenant)
}

func GetEnabledTenants() []string {
	if globalClient == nil {
		return nil
	}

	return globalClient.GetEnabledTenants()
}

func BuildNotification(category string, typ string) NotifyBuilderer {
	return NewNotifyBuilder(category, typ)
}

func RegisterErrorHandler(handler ErrorHandlerFunc) {
	if globalClient == nil {
		return
	}

	globalClient.RegisterErrorHandler(handler)
}

func SetMock(mockClient NotifyClienter) {
	globalClient = mockClient
}

func initGlobalClient(config *NotifyClientConfig, ctx context.Context, configLoader loadConfigDefaults) (NotifyClienter, error) {
	var err error
	once.Do(func() {
		config, err = configLoader(config)
		if err != nil {
			return
		}

		cc, err := NewConfigClient(config)
		if err != nil {
			return
		}

		globalClient = newNotifyClient(config, cc, DefaultErrorHandler, ctx)
	})

	return globalClient, err
}
