/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestHealth is the setup sanity check: the service booted against postgres,
// nats and the mocked authz client, and serves requests.
func TestHealth(t *testing.T) {
	t.Parallel()
	resp, err := OpenapiClient.HealthAPI.Health(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
