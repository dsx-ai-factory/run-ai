// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
	"github.com/stretchr/testify/require"
)

func TestRecommendUsesGeneratedContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/inference/right-sizing/recommendation", r.URL.Path)
		var request catalog.RightSizingRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Equal(t, "cluster", request.ClusterId)
		require.Equal(t, "NVIDIA-H100-80GB-HBM3", request.GetGpuType())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"image":"nvcr.io/nim/meta/model","tag":"1.8.0","results":[]}`))
	}))
	defer server.Close()
	request := catalog.NewRightSizingRequest("cluster", "nvcr.io/nim/meta/model", "1.8.0")
	request.SetGpuType("NVIDIA-H100-80GB-HBM3")
	response, err := New(server.URL).Recommend(context.Background(), *request)
	require.Nil(t, err)
	require.NotNil(t, response)
	require.Empty(t, response.Results)
}

func TestRecommendPreservesDependencyStatus(t *testing.T) {
	for _, status := range []int{400, 403, 502, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]any{"code": status, "message": "sizing unavailable"})
			}))
			defer server.Close()
			_, err := New(server.URL).Recommend(context.Background(), catalog.RightSizingRequest{})
			require.NotNil(t, err)
			require.Equal(t, status, int(err.Code))
		})
	}
}

func TestRecommendRejectsMalformedSuccessfulResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":`))
	}))
	defer server.Close()
	_, err := New(server.URL).Recommend(context.Background(), catalog.RightSizingRequest{})
	require.NotNil(t, err)
	require.Equal(t, http.StatusBadGateway, int(err.Code))
}
