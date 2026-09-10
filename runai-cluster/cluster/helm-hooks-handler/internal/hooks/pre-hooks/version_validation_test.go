/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetControlPlaneVersionUsesSSLCAPath(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip()
	}
	if os.Getenv("RUNAI_TEST_SSL_CERT_DIR") == "true" {
		version, err := getControlPlaneVersion(os.Getenv("RUNAI_TEST_CONTROL_PLANE_URL"), "")
		if err != nil {
			t.Fatal(err)
		}
		if version != "3.0.0" {
			t.Fatalf("unexpected version %q", version)
		}
		return
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, `{"version":"3.0.0"}`); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	caDir := t.TempDir()
	ca := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(filepath.Join(caDir, "runai-ca.pem"), ca, 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGetControlPlaneVersionUsesSSLCAPath$")
	cmd.Env = append(os.Environ(),
		"RUNAI_TEST_SSL_CERT_DIR=true",
		"RUNAI_TEST_CONTROL_PLANE_URL="+server.URL,
		"SSL_CERT_DIR="+caDir,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
}

// test function to for compareVersions
func TestCompareVersions(t *testing.T) {
	// Define test cases
	testCases := []struct {
		controlPlaneVersion string
		clusterVersion      string
		expected            bool
	}{
		{"2.21.15", "2.21.15", true},
		{"2.22.1", "2.22.2", true},
		{"2.22.0-alpha", "2.22.0-beta", true},
		{"2.22.10000932-6002de9", "2.22.10001463-run-26838-222-9ffa636", true},
		{"2.22.0", "2.23.0", false},
		{"2.22.0", "3.0.0", false},
		{"0.0.0-master-asda.3", "0.0.0-master-asda.32", true},
	}

	for _, tc := range testCases {
		result := compareVersions(tc.controlPlaneVersion, tc.clusterVersion)
		if result != nil && tc.expected {
			t.Errorf("Expected no error for control plane version %s and cluster version %s, but got error: %v", tc.controlPlaneVersion, tc.clusterVersion, result)
		} else if result == nil && !tc.expected {
			t.Errorf("Expected error for control plane version %s and cluster version %s, but got no error", tc.controlPlaneVersion, tc.clusterVersion)
		}
	}
}
