/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

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

	client := registryTestClient(t, server)
	images, err := client.sendNgcImagesRequest(context.Background(), "http://registry.example", "", "nvidia/pytorch", true)

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

	client := registryTestClient(t, server)
	images, err := client.sendNgcImagesRequest(context.Background(), "http://registry.example", "api-key", " pytorch ", false)

	require.NoError(t, err)
	assert.Equal(t, []string{"nvidia/pytorch"}, images)
}

func TestBuildImagesCacheKeyIncludesRegistryUrlAndSearch(t *testing.T) {
	assert.Equal(t, "ngc_images_registry-url_", buildImagesCacheKey("registry-url", ""))
	assert.Equal(t, "ngc_images_registry-url_nvidia/pytorch", buildImagesCacheKey("registry-url", " nvidia/pytorch "))
}

func registryTestClient(t *testing.T, server *httptest.Server) *ngcClient {
	t.Helper()
	target, err := url.Parse(server.URL)
	require.NoError(t, err)
	base := server.Client().Transport
	return &ngcClient{
		lookupHost: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		},
		httpClient: &http.Client{Transport: ngcTagsRoundTripper(func(req *http.Request) (*http.Response, error) {
			cloned := req.Clone(req.Context())
			cloned.URL.Scheme = target.Scheme
			cloned.URL.Host = target.Host
			return base.RoundTrip(cloned)
		})},
	}
}

func TestRegistryURLPolicy(t *testing.T) {
	publicLookup := func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("8.8.8.8")}, nil
	}
	for _, tc := range []struct {
		name    string
		rawURL  string
		lookup  func(context.Context, string) ([]net.IP, error)
		wantErr bool
	}{
		{name: "public name", rawURL: "https://registry.example/v2/search", lookup: publicLookup},
		{name: "rfc1918 registry", rawURL: "https://10.1.2.3/v2/search"},
		{name: "ula registry", rawURL: "https://[fd12::1]/v2/search"},
		{name: "loopback", rawURL: "http://127.0.0.1", wantErr: true},
		{name: "mapped loopback", rawURL: "http://[::ffff:127.0.0.1]", wantErr: true},
		{name: "link local", rawURL: "http://169.254.169.254/latest", wantErr: true},
		{name: "mapped link local", rawURL: "http://[::ffff:169.254.169.254]", wantErr: true},
		{name: "unspecified v4", rawURL: "http://0.0.0.0", wantErr: true},
		{name: "unspecified v4 subnet", rawURL: "http://0.1.2.3", wantErr: true},
		{name: "unspecified v6", rawURL: "http://[::]", wantErr: true},
		{name: "localhost name", rawURL: "http://localhost.", wantErr: true},
		{name: "metadata name", rawURL: "http://metadata.google.internal", wantErr: true},
		{name: "aws ipv6 metadata", rawURL: "http://[fd00:ec2::254]", wantErr: true},
		{name: "file scheme", rawURL: "file:///etc/passwd", wantErr: true},
		{name: "name resolving to loopback", rawURL: "http://rebind.example", lookup: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("8.8.8.8"), net.ParseIP("127.0.0.1")}, nil
		}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &ngcClient{lookupHost: tc.lookup}
			err := client.validateRegistryBaseURL(context.Background(), tc.rawURL)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestDialRegistryUsesCheckedAddress(t *testing.T) {
	var dialed string
	client := &ngcClient{
		lookupHost: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		},
		dialContext: func(_ context.Context, _ string, addr string) (net.Conn, error) {
			dialed = addr
			return nil, fmt.Errorf("stopped after dial")
		},
	}
	_, err := client.dialRegistry(context.Background(), "tcp", "rebind.example:443")
	require.Error(t, err)
	assert.Equal(t, "8.8.8.8:443", dialed)

	blocked := &ngcClient{
		lookupHost: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
		dialContext: func(context.Context, string, string) (net.Conn, error) {
			t.Fatal("dialed a blocked address")
			return nil, nil
		},
	}
	_, err = blocked.dialRegistry(context.Background(), "tcp", "rebind.example:443")
	require.Error(t, err)
}

func TestDialRegistryFallsBackToNextCheckedAddress(t *testing.T) {
	lookups := 0
	var dialed []string
	clientConn, serverConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	client := &ngcClient{
		lookupHost: func(context.Context, string) ([]net.IP, error) {
			lookups++
			return []net.IP{net.ParseIP("2001:db8::1"), net.ParseIP("8.8.8.8")}, nil
		},
		dialContext: func(_ context.Context, _ string, addr string) (net.Conn, error) {
			dialed = append(dialed, addr)
			if len(dialed) == 1 {
				return nil, fmt.Errorf("unreachable")
			}
			return clientConn, nil
		},
	}
	conn, err := client.dialRegistry(context.Background(), "tcp", "registry.example:443")
	require.NoError(t, err)
	require.Equal(t, clientConn, conn)
	assert.Equal(t, []string{"[2001:db8::1]:443", "8.8.8.8:443"}, dialed)
	assert.Equal(t, 1, lookups)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	stopped := &ngcClient{
		lookupHost: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("8.8.8.8"), net.ParseIP("1.1.1.1")}, nil
		},
		dialContext: func(context.Context, string, string) (net.Conn, error) {
			t.Fatal("dialed after the deadline")
			return nil, nil
		},
	}
	_, err = stopped.dialRegistry(canceled, "tcp", "registry.example:443")
	require.Error(t, err)
}

func TestRegistryClientBlocksRedirectToLoopback(t *testing.T) {
	client := newRegistryHTTPClient(&ngcClient{}, time.Second)
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/next", nil)
	require.NoError(t, err)
	require.Error(t, client.CheckRedirect(req, nil))

	allowed, err := http.NewRequest(http.MethodGet, "https://10.1.2.3/v2/search", nil)
	require.NoError(t, err)
	require.NoError(t, client.CheckRedirect(allowed, nil))
}

func TestSendNgcImagesRequestRejectsLoopbackBeforeDial(t *testing.T) {
	client := &ngcClient{httpClient: &http.Client{Transport: ngcTagsRoundTripper(func(*http.Request) (*http.Response, error) {
		t.Fatal("request was sent")
		return nil, nil
	})}}
	_, err := client.sendNgcImagesRequest(context.Background(), "http://127.0.0.1/v2/search", "secret", "", false)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret")
}

func TestNgcGetRequestOmitsUpstreamBody(t *testing.T) {
	const upstream = "upstream-secret-body"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(upstream))
	}))
	t.Cleanup(server.Close)

	client := &ngcClient{httpClient: server.Client()}
	_, status, err := client.ngcGetRequest(context.Background(), server.URL, "")
	require.Error(t, err)
	assert.Equal(t, http.StatusBadGateway, status)
	assert.NotContains(t, err.Error(), upstream)
}
