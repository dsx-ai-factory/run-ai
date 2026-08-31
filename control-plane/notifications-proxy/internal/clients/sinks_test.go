/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"encoding/json"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/stretchr/testify/require"
)

func TestDecodeToMap_AllFieldsShouldExist(t *testing.T) {
	a, b := true, false
	c := types.SinkConfigDetails{
		EmailSinkConfig: types.EmailSinkConfig{
			SmtpTlsEnabled: &a,
		},
		SlackSinkConfig: types.SlackSinkConfig{
			ApiToken: "token",
		},
		DirectNotifications: &b,
		Recipients:          []string{"user@example.com"},
	}
	m, err := DecodeToMap(c)
	require.NoError(t, err)
	bytes, err := json.Marshal(m)
	require.NoError(t, err)
	require.Equal(t, string(bytes), `{"api_token":"token","direct_notifications":false,"recipients":["user@example.com"],"smtp_tls_enabled":true}`)

	pc := types.SinkConfigDetails{}
	err = json.Unmarshal(bytes, &pc)
	require.NoError(t, err)
	require.False(t, *pc.DirectNotifications)
	require.True(t, *pc.SmtpTlsEnabled)
	require.Equal(t, pc.ApiToken, "token")
	require.Equal(t, pc.Recipients, []string{"user@example.com"})
}

func TestDecodeToMap_NoFieldsShouldExist(t *testing.T) {
	a := types.SinkConfigDetails{}
	m, err := DecodeToMap(a)
	require.NoError(t, err)
	require.Nil(t, m["smtp_tls_enabled"])
	require.Nil(t, m["direct_notifications"])
	require.Nil(t, m["api_token"])
	require.Nil(t, m["recipients"])
}
