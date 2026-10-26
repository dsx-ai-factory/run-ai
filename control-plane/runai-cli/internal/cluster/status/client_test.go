/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/apis"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestClusterAPIClientDoesNotDuplicatePathPrefix(t *testing.T) {
	const expectedPath = "/cluster-api/api/v1/project/workloads/distributedinference/runai/workload/status"

	originalClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != expectedPath {
			t.Fatalf("expected request path %q, got %q", expectedPath, request.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"status":"Ready"}`)),
			Request:    request,
		}, nil
	})}
	t.Cleanup(func() { http.DefaultClient = originalClient })

	client, err := NewClusterApiClient("https://cluster.example.com/cluster-api", "token")
	if err != nil {
		t.Fatalf("create cluster API client: %v", err)
	}
	workload := apis.NewWorkload("project", api.DistributedInference, "workload", api.LWS)
	workloadStatus, err := client.GetWorkloadStatus(context.Background(), workload)
	if err != nil {
		t.Fatalf("get workload status: %v", err)
	}
	if workloadStatus.GetStatus() != "Ready" {
		t.Fatalf("expected Ready status, got %q", workloadStatus.GetStatus())
	}
}
