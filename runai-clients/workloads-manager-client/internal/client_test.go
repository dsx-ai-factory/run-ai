/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
)

func TestUpdateWorkloadInternalUsesInternalEndpoint(t *testing.T) {
	requestReceived := make(chan *http.Request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestReceived <- request
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusAccepted)
		_, _ = writer.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	clientConfig := openapi.NewConfiguration()
	clientConfig.Servers = openapi.ServerConfigurations{{URL: server.URL}}
	client := &WorkloadsManagerClient{ApiClient: openapi.NewAPIClient(clientConfig)}

	_, apiErr := client.UpdateWorkloadInternal(context.Background(), "workload-id", openapi.WorkloadV2UpdateRequest{})
	require.Nil(t, apiErr)

	request := <-requestReceived
	require.Equal(t, http.MethodPut, request.Method)
	require.Equal(t, "/internal/v2/workloads/workload-id", request.URL.Path)
}
