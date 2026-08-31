/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cert

import (
	"os"
	"testing"
)

func TestLoadCertFromPath_Success(t *testing.T) {
	certFile, err := os.CreateTemp("", "test-cert-*.crt")
	if err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	defer os.Remove(certFile.Name())
	keyFile, err := os.CreateTemp("", "test-key-*.key")
	if err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}
	defer os.Remove(keyFile.Name())

	certContent := "CERTDATA"
	keyContent := "KEYDATA"
	_, _ = certFile.WriteString(certContent)
	_, _ = keyFile.WriteString(keyContent)
	certFile.Close()
	keyFile.Close()

	h := NewLoadCertFromPath(certFile.Name(), keyFile.Name())
	crt, err := h.GetCRT()
	if err != nil || crt != certContent {
		t.Errorf("expected cert '%s', got '%s', err: %v", certContent, crt, err)
	}
	key, err := h.GetKey()
	if err != nil || key != keyContent {
		t.Errorf("expected key '%s', got '%s', err: %v", keyContent, key, err)
	}
}

func TestLoadCertFromPath_FileNotFound(t *testing.T) {
	h := NewLoadCertFromPath("/nonexistent/cert", "/nonexistent/key")
	_, err := h.GetCRT()
	if err == nil {
		t.Error("expected error for missing cert file, got nil")
	}
	_, err = h.GetKey()
	if err == nil {
		t.Error("expected error for missing key file, got nil")
	}
}

func TestGenerateSelfSignedCert_Success(t *testing.T) {
	h := NewGenerateSelfSignedCert("fqdn.example.com")
	crt, err := h.GetCRT()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if crt == "" {
		t.Error("expected non-empty certificate")
	}

	key, err := h.GetKey()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if key == "" {
		t.Error("expected non-empty key")
	}

	// Test that subsequent calls return the same cached values
	crt2, err := h.GetCRT()
	if err != nil {
		t.Errorf("expected no error on second call, got: %v", err)
	}
	if crt != crt2 {
		t.Error("expected cached certificate to be the same")
	}

	key2, err := h.GetKey()
	if err != nil {
		t.Errorf("expected no error on second call, got: %v", err)
	}
	if key != key2 {
		t.Error("expected cached key to be the same")
	}
}
