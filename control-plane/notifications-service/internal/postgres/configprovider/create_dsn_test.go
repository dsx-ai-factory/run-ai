/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres"
	"github.com/stretchr/testify/require"
)

func TestCreateDsnHost(t *testing.T) {
	for _, tc := range []struct {
		name string
		host string
		want string
	}{
		{name: "hostname", host: "db", want: "db:5432"},
		{name: "ipv4", host: "10.0.0.5", want: "10.0.0.5:5432"},
		{name: "ipv6", host: "2600:1f18::5", want: "[2600:1f18::5]:5432"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dsn := CreateDsn(&postgres.DataBaseConfig{Host: tc.host, Port: "5432", Database: "notifications", SslMode: "disable"})
			require.Equal(t, tc.want, dsn.Host)
		})
	}
}
