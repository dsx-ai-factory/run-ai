/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redirect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterServiceClient_APICall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"url":"https://cluster.example.com"}`))
	}))
	defer srv.Close()

	c := newClusterServiceClient(srv.URL)
	url, err := c.GetClusterURL(context.Background(), "12345678-1234-1234-1234-123456789012")
	require.NoError(t, err)
	assert.Equal(t, "https://cluster.example.com", url)
}

func TestClusterServiceClient_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newClusterServiceClient(srv.URL)
	_, err := c.GetClusterURL(context.Background(), "ffffffff-ffff-ffff-ffff-ffffffffffff")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster-service lookup")
}

// TestClusterURLEntryJSONRoundTrip guards against the bug where the url field
// was unexported, causing encoding/json to silently discard it:
//
//	json.Marshal(clusterURLEntry{URL: "https://..."}) → {}
//	json.Unmarshal("{}", &clusterURLEntry{})           → URL: ""
//
// This meant every cache.Set stored {} and every cache.Get returned URL: "".
func TestClusterURLEntryJSONRoundTrip(t *testing.T) {
	entry := clusterURLEntry{URL: "https://cluster.example.com"}

	bytes, err := json.Marshal(entry)
	require.NoError(t, err)
	assert.NotEqual(t, "{}", string(bytes), "URL field must be included in marshaled JSON")

	var result clusterURLEntry
	require.NoError(t, json.Unmarshal(bytes, &result))
	assert.Equal(t, "https://cluster.example.com", result.URL, "URL must survive json round-trip")
}

func TestClusterServiceClient_EmptyURL(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			_, _ = w.Write([]byte(`{"url":""}`))
		} else {
			_, _ = w.Write([]byte(`{"url":"https://cluster.example.com"}`))
		}
	}))
	defer srv.Close()

	c := newClusterServiceClient(srv.URL)
	uuid := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	_, err := c.GetClusterURL(context.Background(), uuid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty URL")

	// Empty URL must not be cached — second call must reach cluster-service and succeed.
	url, err := c.GetClusterURL(context.Background(), uuid)
	require.NoError(t, err)
	assert.Equal(t, "https://cluster.example.com", url)
	assert.Equal(t, 2, callCount, "second call should have hit cluster-service, not a stale cache")
}
