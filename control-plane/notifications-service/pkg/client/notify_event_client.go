/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	events_http_client "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/notify"
)

type NotificationsEventClienter interface {
	NotifyBatch(notification []events_http_client.Notification) (*events_http_client.NotifyBatch200Response, error)
}

type NotificationsEventClient struct {
	client *events_http_client.APIClient
}

func newNotificationsEventClient(config *NotifyClientConfig) NotificationsEventClienter {
	host, scheme := getHostAndScheme(config.NotifyUrl)
	cfg := events_http_client.NewConfiguration()
	cfg.Host = host
	cfg.Scheme = scheme

	return &NotificationsEventClient{
		client: events_http_client.NewAPIClient(cfg),
	}
}

func (nec *NotificationsEventClient) NotifyBatch(notifications []events_http_client.Notification) (*events_http_client.NotifyBatch200Response, error) {
	notifyResponse, res, err := nec.client.NotifyAPI.NotifyBatch(context.Background()).Notification(notifications).Execute()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return notifyResponse, getError(res)
	}

	return notifyResponse, nil
}

func getError(res *http.Response) error {
	b := make([]byte, 100)
	if res.Body != nil {
		_, _ = res.Body.Read(b)
	}

	return fmt.Errorf("failed: status: %s, error: %s", res.Status, strings.ReplaceAll(string(b), "\x00", ""))
}
