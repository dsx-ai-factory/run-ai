/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

func TestRemoteConfigForURLCopiesBaseConfig(t *testing.T) {
	base := &rest.Config{Host: "https://management.example", BearerToken: "token"}
	remote := remoteConfigForURL(base, "https://workload.example")

	require.NotSame(t, base, remote)
	require.Equal(t, "https://management.example", base.Host)
	require.Equal(t, "https://workload.example", remote.Host)
	require.Equal(t, base.BearerToken, remote.BearerToken)
}
