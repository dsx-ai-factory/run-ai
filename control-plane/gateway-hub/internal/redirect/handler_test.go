/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redirect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubClusterClient is a test double for ClusterURLProvider.
type stubClusterClient struct {
	url string
	err error
}

func (s *stubClusterClient) GetClusterURL(_ context.Context, _ string) (string, error) {
	return s.url, s.err
}

func newHandler(clusterURL string, clusterErr error) *Handler {
	return &Handler{clusterClient: &stubClusterClient{url: clusterURL, err: clusterErr}}
}

const (
	validClientID   = "runai-internal.12345678-1234-1234-1234-123456789012.workloads_auth"
	validClusterURL = "https://cluster.example.com"
	validRedirect   = "https://cluster.example.com/oauth/callback"
)

// ── ValidateRedirect ──────────────────────────────────────────────────────────

func TestValidateRedirect_ValidRequest(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+validRedirect+"&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRedirect_MissingRedirectURI(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	req := httptest.NewRequest(http.MethodGet, "/?client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing parameter: redirect_uri")
}

func TestValidateRedirect_MissingClientID(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+validRedirect, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing parameter: client_id")
}

func TestValidateRedirect_InvalidClientID(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+validRedirect+"&client_id=not-a-valid-client-id", nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid parameter: client_id")
}

func TestValidateRedirect_ClusterServiceError(t *testing.T) {
	h := newHandler("", errors.New("service unavailable"))
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+validRedirect+"&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to get cluster URL")
}

func TestValidateRedirect_InvalidRedirectURI(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	malicious := "https://evil.com/steal-token"
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+malicious+"&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid parameter: redirect_uri")
}

func TestValidateRedirect_PathTraversalRejected(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	traversal := "https://cluster.example.com/../../../etc/passwd"
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+traversal+"&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid parameter: redirect_uri")
}

func TestValidateRedirect_SubdomainAllowed(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	subdomain := "https://app.cluster.example.com/callback"
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+subdomain+"&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRedirect_SchemeMismatchRejected(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	httpRedirect := "http://cluster.example.com/callback"
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri="+httpRedirect+"&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid parameter: redirect_uri")
}

func TestValidateRedirect_ErrorPageIsHTML(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri=https://evil.com&client_id="+validClientID, nil)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.NotContains(t, w.Body.String(), "{{error_message}}", "template placeholder must be substituted")
}

func TestValidateRedirect_ForwardAuthHeader(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	// Simulate Traefik forwardAuth: no query params in the request URL,
	// original URL is in X-Forwarded-Uri.
	req := httptest.NewRequest(http.MethodGet, "/internal/plugins/workload-redirect-validation", nil)
	req.Header.Set("X-Forwarded-Uri", "/auth/realms/runai/protocol/openid-connect/auth?redirect_uri="+validRedirect+"&client_id="+validClientID)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateRedirect_ForwardAuthHeader_InvalidRedirect(t *testing.T) {
	h := newHandler(validClusterURL, nil)
	req := httptest.NewRequest(http.MethodGet, "/internal/plugins/workload-redirect-validation", nil)
	req.Header.Set("X-Forwarded-Uri", "/auth/realms/runai/protocol/openid-connect/auth?redirect_uri=https://evil.com/steal&client_id="+validClientID)
	w := httptest.NewRecorder()

	h.ValidateRedirect(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid parameter: redirect_uri")
}

// ── extractParams ─────────────────────────────────────────────────────────────

func TestExtractParams_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri=https://x.com&client_id="+validClientID, nil)
	uri, uuid, err := extractParams(req)
	require.NoError(t, err)
	assert.Equal(t, "https://x.com", uri)
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", uuid)
}

func TestExtractParams_MissingRedirectURI(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?client_id="+validClientID, nil)
	_, _, err := extractParams(req)
	assert.Error(t, err)
}

func TestExtractParams_MissingClientID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri=https://x.com", nil)
	_, _, err := extractParams(req)
	assert.Error(t, err)
}

func TestExtractParams_InvalidClientIDFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?redirect_uri=https://x.com&client_id=bad-format", nil)
	_, _, err := extractParams(req)
	assert.Error(t, err)
}

// ── isValidRedirect ───────────────────────────────────────────────────────────

func TestIsValidRedirect(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
		clusterURL  string
		want        bool
	}{
		{"exact match", "https://cluster.example.com/cb", "https://cluster.example.com", true},
		{"subdomain", "https://app.cluster.example.com/cb", "https://cluster.example.com", true},
		{"sibling subdomain", "https://app.othercluster.example.com/cb", "https://cluster.example.com", true},
		{"different domain", "https://evil.com/cb", "https://cluster.example.com", false},
		{"scheme mismatch", "http://cluster.example.com/cb", "https://cluster.example.com", false},
		{"port mismatch", "https://cluster.example.com:9000/cb", "https://cluster.example.com", false},
		{"path traversal", "https://cluster.example.com/../etc", "https://cluster.example.com", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidRedirect(tc.redirectURI, tc.clusterURL)
			assert.Equal(t, tc.want, got, "redirectURI=%q clusterURL=%q", tc.redirectURI, tc.clusterURL)
		})
	}
}

// ── serveErrorPage ────────────────────────────────────────────────────────────

func TestServeErrorPage_RendersMessage(t *testing.T) {
	w := httptest.NewRecorder()
	serveErrorPage(w, "Missing parameter: redirect_uri")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "Missing parameter: redirect_uri")
	assert.Contains(t, w.Body.String(), "/distribution-assets/cc/favicon.ico")
	assert.Contains(t, w.Body.String(), "/login/runai-theme/js/distribution-assets.js")
	assert.NotContains(t, w.Body.String(), "/login/runai-theme/img/favicon.ico")
	assert.False(t, strings.Contains(w.Body.String(), "{{"), "all template vars should be replaced")
}

func TestNewHandler_NotNil(t *testing.T) {
	h := NewHandler("http://cluster-service")
	assert.NotNil(t, h)
}
