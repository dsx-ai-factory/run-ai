/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/maypok86/otter/v2"
	"github.com/stretchr/testify/assert"
)

func TestShouldNotify_InCacheReturnTrue(t *testing.T) {
	cache := otter.Must[string, bool](&otter.Options[string, bool]{
		ExpiryCalculator: otter.ExpiryWriting[string, bool](5 * time.Second),
	})
	sc := ConfigClient{
		client: config_client.NewAPIClient(config_client.NewConfiguration()),
		cache:  cache,
	}

	cache.Set("tenantId", true)
	res, err := sc.ShouldNotify("tenantId")
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestShouldNotify_InCacheReturnFalse(t *testing.T) {
	cache := otter.Must[string, bool](&otter.Options[string, bool]{
		ExpiryCalculator: otter.ExpiryWriting[string, bool](time.Duration(5 * time.Second)),
	})
	sc := ConfigClient{
		client: config_client.NewAPIClient(config_client.NewConfiguration()),
		cache:  cache,
	}

	cache.Set("tenantId", false)
	res, err := sc.ShouldNotify("tenantId")
	assert.Nil(t, err)
	assert.False(t, res)
}
