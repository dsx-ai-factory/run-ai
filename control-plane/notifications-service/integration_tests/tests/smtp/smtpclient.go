/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/integration_tests/mailhog/mailhogclient"
	"github.com/sethvargo/go-retry"
)

type Client struct {
	client *mailhogclient.APIClient
}

func NewSmtpClient(host string) *Client {
	config := mailhogclient.NewConfiguration()
	config.Host = host
	config.Scheme = "http"
	config.DefaultHeader["Content-Type"] = "application/json"

	return &Client{
		client: mailhogclient.NewAPIClient(config),
	}
}

func (s *Client) GetMessages() (*mailhogclient.Messages, *http.Response, error) {
	request := s.client.DefaultAPI.ApiV2MessagesGet(context.Background())
	return s.client.DefaultAPI.ApiV2MessagesGetExecute(request)
}

func (s *Client) GetMessagesWithRetry(maxRetries uint64, untilTrue func(msgs *mailhogclient.Messages, resp *http.Response, err error) bool) (*mailhogclient.Messages, *http.Response, error) {
	var (
		err  error
		resp *http.Response
		msgs *mailhogclient.Messages
	)

	retryableGetMessagesFunc := func(ctx context.Context) error {
		msgs, resp, err = s.GetMessages()
		if untilTrue(msgs, resp, err) {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("failed retry condition. Retrying... (Error:  %w)", err))
	}

	retryErr := retry.Do(
		context.Background(),
		retry.WithMaxRetries(maxRetries, retry.NewConstant(1*time.Second)),
		retryableGetMessagesFunc)

	if retryErr != nil {
		err = fmt.Errorf("RetryError: %w", retryErr)
	}
	return msgs, resp, err
}
