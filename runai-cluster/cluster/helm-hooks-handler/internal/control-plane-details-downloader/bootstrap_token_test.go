/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package control_plane_details_downloader

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	schemeInit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testClusterUID     = "test-cluster-uuid"
	testBootstrapToken = "my-bootstrap-token"
	testClientSecret   = "should-not-be-used-secret"
)

// newCPServer starts an httptest server that serves a minimal cluster-config response
// and records the Authorization header it received.
func newCPServer(t *testing.T, authHeaderOut *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authHeaderOut != nil {
			*authHeaderOut = r.Header.Get("Authorization")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
}

// TestBootstrapTokenUsedWhenBothCredentialsPresent verifies that when both a bootstrap
// token and a client secret are provided, the token is used directly and the IDM
// client-credentials exchange is skipped entirely. If the secret were used instead,
// the test would fail because no IDM server is running.
func TestBootstrapTokenUsedWhenBothCredentialsPresent(t *testing.T) {
	var receivedAuth string
	server := newCPServer(t, &receivedAuth)
	defer server.Close()

	k8s := fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).Build()

	_, err := DownloadInstallationConfiguration(k8s, &hooks.ClusterHookParams{
		ControlPlaneDetails: clients.ControlPlaneDetails{
			URL:            server.URL,
			BootstrapToken: testBootstrapToken,
			ClientSecret:   testClientSecret,
		},
		ClusterDetails: hooks.ClusterDetails{
			UID: testClusterUID,
			URL: "https://cluster.example.com",
		},
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, "Bearer "+testBootstrapToken, receivedAuth,
		"bootstrap token must be used, not the client secret")
}

// TestBootstrapTokenAloneIsEnough verifies that a bootstrap token is sufficient for
// a successful CP-assisted installation even when no client secret is provided.
func TestBootstrapTokenAloneIsEnough(t *testing.T) {
	server := newCPServer(t, nil)
	defer server.Close()

	k8s := fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).Build()

	_, err := DownloadInstallationConfiguration(k8s, &hooks.ClusterHookParams{
		ControlPlaneDetails: clients.ControlPlaneDetails{
			URL:            server.URL,
			BootstrapToken: testBootstrapToken,
		},
		ClusterDetails: hooks.ClusterDetails{
			UID: testClusterUID,
			URL: "https://cluster.example.com",
		},
	}, nil)

	assert.NoError(t, err)
}

func TestBootstrapFailsFastWhenControlPlaneUnreachable(t *testing.T) {
	server := newCPServer(t, nil)
	unreachableURL := server.URL
	server.Close()

	k8s := fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).Build()

	done := make(chan error, 1)
	go func() {
		_, err := DownloadInstallationConfiguration(k8s, &hooks.ClusterHookParams{
			ControlPlaneDetails: clients.ControlPlaneDetails{
				URL:            unreachableURL,
				BootstrapToken: testBootstrapToken,
			},
			ClusterDetails: hooks.ClusterDetails{
				UID: testClusterUID,
				URL: "https://cluster.example.com",
			},
		}, nil)
		done <- err
	}()

	select {
	case err := <-done:
		require.Error(t, err, "an unreachable control-plane must produce a clear error, not a silent success")
	case <-time.After(10 * time.Second):
		t.Fatal("DownloadInstallationConfiguration hung instead of failing fast on an unreachable control-plane")
	}
}
