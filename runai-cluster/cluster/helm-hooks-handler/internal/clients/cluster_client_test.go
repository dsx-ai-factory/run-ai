/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigureDefaultHTTPTransport(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	caPath := filepath.Join(t.TempDir(), "ca.pem")
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	require.NoError(t, os.WriteFile(caPath, ca, 0o600))

	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	require.NoError(t, ConfigureDefaultHTTPTransport(caPath))
	response, err := http.DefaultClient.Get(server.URL)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
}
