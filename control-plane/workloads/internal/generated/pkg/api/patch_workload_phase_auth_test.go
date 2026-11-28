/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPatchWorkloadPhaseAuthorizesSync(t *testing.T) {
	auth := &mocks.IAuthorizationClient{}
	actions := map[string]bool{}
	auth.On("AuthorizeMiddleware", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if args.String(1) == "workloads" && args.String(4) == "workloadId" {
				actions[args.String(2)] = true
			}
		}).
		Return(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	controller := &WorkloadsAPIController{authorizationClient: auth}
	_ = controller.Routes()

	require.True(t, actions["sync"], "phase patch must require workloads:sync")
	require.False(t, actions["update"], "phase patch must not accept workloads:update")
	require.True(t, actions["read"])
}
