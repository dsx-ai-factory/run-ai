/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	nc, err := Init(&NotifyClientConfig{}, context.Background())
	assert.NotNil(t, nc)
	assert.Nil(t, err)
	globalClient = nil
}

func TestInit_Failed(t *testing.T) {
	once = sync.Once{}
	_, err := initGlobalClient(&NotifyClientConfig{}, context.Background(), func(config *NotifyClientConfig) (*NotifyClientConfig, error) {
		return nil, errors.New("failed")
	})
	assert.NotNil(t, err)
}
