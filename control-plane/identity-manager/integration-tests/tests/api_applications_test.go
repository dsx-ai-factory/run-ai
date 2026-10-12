/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplication_GetInstallerApp_NotFoundInNewRealm(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	_, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetInstallerServiceAccount(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}
