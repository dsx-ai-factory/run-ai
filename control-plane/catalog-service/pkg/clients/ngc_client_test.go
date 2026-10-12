/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ngcTagsRoundTripper func(*http.Request) (*http.Response, error)

func (f ngcTagsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSendNgcTagsRequest(t *testing.T) {
	for _, tc := range []struct {
		name    string
		image   string
		apiKey  string
		wantURL string
	}{
		{
			name:    "authenticated team image",
			image:   "nim/nvidia/nemotron-3-super-120b-a12b",
			apiKey:  "<test-api-key>",
			wantURL: "https://api.ngc.nvidia.com/v2/org/nim/team/nvidia/repos/nemotron-3-super-120b-a12b/images",
		},
		{
			name:    "authenticated org image",
			image:   "nvidia/pytorch",
			apiKey:  "<test-api-key>",
			wantURL: "https://api.ngc.nvidia.com/v2/org/nvidia/repos/pytorch/images",
		},
		{
			name:    "guest team image",
			image:   "nim/nvidia/nemotron-3-super-120b-a12b",
			wantURL: "https://api.ngc.nvidia.com/v2/repos/nim/nvidia/nemotron-3-super-120b-a12b/images",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &ngcClient{httpClient: &http.Client{Transport: ngcTagsRoundTripper(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodGet, req.Method)
				assert.Equal(t, tc.wantURL, req.URL.String())
				if tc.apiKey == "" {
					assert.Empty(t, req.Header.Get("Authorization"))
				} else {
					assert.Equal(t, "Bearer "+tc.apiKey, req.Header.Get("Authorization"))
				}
				response := httptest.NewRecorder()
				response.Header().Set("Content-Type", "application/json")
				_, err := response.WriteString(`{"images":[{"tag":"latest"},{"tag":"2.0.13"}],"requestStatus":{"statusCode":"SUCCESS"}}`)
				require.NoError(t, err)
				return response.Result(), nil
			})}}

			tags, err := client.sendNgcTagsRequest(context.Background(), tc.image, tc.apiKey)
			require.NoError(t, err)
			assert.Equal(t, []string{"latest", "2.0.13"}, tags)
		})
	}
}

func TestSendNgcImagesRequestSearchesImageIdentifierAndPaginates(t *testing.T) {
	var (
		mu      sync.Mutex
		queries []map[string]interface{}
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query map[string]interface{}
		if err := json.Unmarshal([]byte(r.URL.Query().Get("q")), &query); err != nil {
			t.Errorf("decode query: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		queries = append(queries, query)
		mu.Unlock()

		page := int(query["page"].(float64))
		resourceID := "nvidia/pytorch"
		if page == 1 {
			resourceID = "nvidia/pytorch-ltsb2"
		}
		_, err := fmt.Fprintf(w, `{
			"resultPageTotal": 2,
			"results": [{"resources": [{"resourceId": %q}, {"resourceId": %q}]}]
		}`, resourceID, resourceID)
		if err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client := &ngcClient{httpClient: server.Client()}
	images, err := client.sendNgcImagesRequest(context.Background(), server.URL, "", "nvidia/pytorch", true)

	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"nvidia/pytorch", "nvidia/pytorch-ltsb2"}, images)
	require.Len(t, queries, 2)
	assert.Equal(t, float64(ngcImagesPageSize), queries[0]["pageSize"])
	assert.NotContains(t, queries[0], "query")

	filters := queries[0]["filters"].([]interface{})
	require.Len(t, filters, 3)
	assert.Equal(t, "isPublic", filters[0].(map[string]interface{})["field"])
	assert.Equal(t, "guestAccess", filters[1].(map[string]interface{})["field"])
	assert.Equal(t, "resourceId", filters[2].(map[string]interface{})["field"])
	assert.Equal(t, "nvidia/pytorch*", filters[2].(map[string]interface{})["value"])
	assert.Equal(t, float64(0), queries[0]["page"])
	assert.Equal(t, float64(1), queries[1]["page"])
}

func TestSendNgcImagesRequestForwardsTextSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query map[string]interface{}
		if err := json.Unmarshal([]byte(r.URL.Query().Get("q")), &query); err != nil {
			t.Errorf("decode query: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		assert.Equal(t, "pytorch", query["query"])
		assert.NotContains(t, query, "filters")
		_, err := fmt.Fprint(w, `{
			"resultPageTotal": 1,
			"results": [{"resources": [{"resourceId": "nvidia/pytorch"}]}]
		}`)
		if err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client := &ngcClient{httpClient: server.Client()}
	images, err := client.sendNgcImagesRequest(context.Background(), server.URL, "api-key", " pytorch ", false)

	require.NoError(t, err)
	assert.Equal(t, []string{"nvidia/pytorch"}, images)
}

func TestBuildImagesCacheKeyIncludesRegistryUrlAndSearch(t *testing.T) {
	assert.Equal(t, "ngc_images_registry-url_", buildImagesCacheKey("registry-url", ""))
	assert.Equal(t, "ngc_images_registry-url_nvidia/pytorch", buildImagesCacheKey("registry-url", " nvidia/pytorch "))
}
