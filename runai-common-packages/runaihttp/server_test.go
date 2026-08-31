/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaihttp

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/require"
)

func TestReadConfig_Defaults(t *testing.T) {
	defer config.FlushAll()

	cfg := readConfig()
	require.Equal(t, 8080, cfg.ServicePort)
	require.Equal(t, 15*time.Second, cfg.WriteTimeout)
	require.Equal(t, 15*time.Second, cfg.ReadTimeout)
	require.Equal(t, 60*time.Second, cfg.IdleTimeout)
	require.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
	require.Equal(t, "0.0.0.0", cfg.IPAddress)
	require.Equal(t, "localhost", cfg.HealthHost)
	require.Equal(t, "/", cfg.HealthEndpoint)
}

func TestReadConfig_CustomValues_FromEnv(t *testing.T) {
	defer config.FlushAll()

	t.Setenv("SERVICE_PORT", "8081")
	t.Setenv("SERVER_WRITE_TIMEOUT", "1m")
	t.Setenv("SERVER_READ_TIMEOUT", "1m")
	t.Setenv("SERVER_IDLE_TIMEOUT", "1m")
	t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "1m")
	t.Setenv("SERVER_IP_ADDRESS", "1.1.1.1")
	t.Setenv("SERVER_HEALTH_HOST", "host.com")
	t.Setenv("SERVER_HEALTH_ENDPOINT", "/health-check")
	cfg := readConfig()

	require.Equal(t, 8081, cfg.ServicePort)
	require.Equal(t, 1*time.Minute, cfg.WriteTimeout)
	require.Equal(t, 1*time.Minute, cfg.ReadTimeout)
	require.Equal(t, 1*time.Minute, cfg.IdleTimeout)
	require.Equal(t, 1*time.Minute, cfg.ShutdownTimeout)
	require.Equal(t, "1.1.1.1", cfg.IPAddress)
	require.Equal(t, "host.com", cfg.HealthHost)
	require.Equal(t, "/health-check", cfg.HealthEndpoint)
}
