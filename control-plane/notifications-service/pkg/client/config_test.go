/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaultsToConfig_NoConfig(t *testing.T) {
	clientConfig, _ := LoadDefaultsToConfig(&NotifyClientConfig{})
	assert.Equal(t, DEFAULT_NOTIFY_BATCH_SIZE, clientConfig.BatchSize)
	assert.Equal(t, DEFAULT_NOTIFY_BATCH_INTERVAL_SEC, clientConfig.BatchIntervalSec)
	assert.Equal(t, DEFAULT_NOTIFY_SERVICE_URL, clientConfig.NotifyUrl)
	assert.Equal(t, DEFAULT_NOTIFY_RETRY_INTERVAL_SEC, clientConfig.RetryIntervalSec)
	assert.Equal(t, DEFAULT_NOTIFY_RETRY_THRESHOLD, clientConfig.RetryThreshold)
}

func TestLoadDefaultsToConfig_PartialConfig(t *testing.T) {
	config := NotifyClientConfig{BatchSize: 50, BatchIntervalSec: 15}
	clientConfig, _ := LoadDefaultsToConfig(&config)
	assert.Equal(t, 50, clientConfig.BatchSize)
	assert.Equal(t, 15, clientConfig.BatchIntervalSec)
	assert.Equal(t, DEFAULT_NOTIFY_SERVICE_URL, clientConfig.NotifyUrl)
	assert.Equal(t, DEFAULT_NOTIFY_RETRY_INTERVAL_SEC, clientConfig.RetryIntervalSec)
	assert.Equal(t, DEFAULT_NOTIFY_RETRY_THRESHOLD, clientConfig.RetryThreshold)
}

func TestLoadDefaultsToConfig_PartialConfigWithNegative(t *testing.T) {
	config := NotifyClientConfig{BatchSize: 0, BatchIntervalSec: 0}
	clientConfig, _ := LoadDefaultsToConfig(&config)
	assert.Equal(t, DEFAULT_NOTIFY_BATCH_SIZE, clientConfig.BatchSize)
	assert.Equal(t, DEFAULT_NOTIFY_BATCH_INTERVAL_SEC, clientConfig.BatchIntervalSec)
	assert.Equal(t, DEFAULT_NOTIFY_SERVICE_URL, clientConfig.NotifyUrl)
	assert.Equal(t, DEFAULT_NOTIFY_RETRY_INTERVAL_SEC, clientConfig.RetryIntervalSec)
	assert.Equal(t, DEFAULT_NOTIFY_RETRY_THRESHOLD, clientConfig.RetryThreshold)
}

func TestGetUrlAndType_Http(t *testing.T) {
	h, scheme := getHostAndScheme("http://notifications-service:8000")
	assert.Equal(t, "notifications-service:8000", h)
	assert.Equal(t, "http", scheme)
}

func TestGetUrlAndType_Https(t *testing.T) {
	h, scheme := getHostAndScheme("https://notifications-service:8000")
	assert.Equal(t, "notifications-service:8000", h)
	assert.Equal(t, "https", scheme)
}
