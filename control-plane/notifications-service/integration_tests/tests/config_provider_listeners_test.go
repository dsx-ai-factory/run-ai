//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/config_client"

	"github.com/stretchr/testify/require"
)

func TestGetListener_NotExist(t *testing.T) {
	_, res, err := configClient.ListenerAPI.GetListener(context.Background(), "TestGetListener_NotExist").Execute()
	require.Equal(t, http.StatusNotFound, res.StatusCode)
	require.Error(t, err)
}

func TestGetListener_Exist(t *testing.T) {
	port := 4555
	listenerConfig := config_client.ListenerData{
		ListenerId:     "TestGetListener_Exist",
		ListenerConfig: map[string]interface{}{"port": 4555},
		ListenerType:   "http_listener",
	}

	_, _, err := configClient.ListenerAPI.CreateListener(context.Background()).ListenerData(listenerConfig).Execute()
	require.NoError(t, err)
	_, err = net.DialTimeout("tcp", net.JoinHostPort("", fmt.Sprintf("%d", port)), time.Second)
	require.Error(t, err)
	res, _, err := configClient.ListenerAPI.GetListener(context.Background(), "TestGetListener_Exist").Execute()
	require.NoError(t, err)
	require.NotNil(t, res)
	_, _, err = configClient.ListenerAPI.DeleteListener(context.Background(), "TestGetListener_Exist").Execute()
	require.NoError(t, err)
}

func TestDeleteListener_Exist(t *testing.T) {
	port := 4556
	listenerConfig := config_client.ListenerData{
		ListenerId:     "TestDeleteListener_Exist",
		ListenerConfig: map[string]interface{}{"port": 4556},
		ListenerType:   "http_listener",
	}

	_, _, err := configClient.ListenerAPI.CreateListener(context.Background()).ListenerData(listenerConfig).Execute()
	require.NoError(t, err)

	_, err = net.DialTimeout("tcp", net.JoinHostPort("", fmt.Sprintf("%d", port)), time.Second)
	require.Error(t, err)

	res, r, err := configClient.ListenerAPI.DeleteListener(context.Background(), "TestDeleteListener_Exist").Execute()
	require.NoError(t, err)
	require.Equal(t, r.StatusCode, http.StatusOK)
	require.NotNil(t, res)
}

func TestDeleteListener_NotExist(t *testing.T) {
	_, res, err := configClient.ListenerAPI.DeleteListener(context.Background(), "TestDeleteAndGetListener_NotExist").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
}
