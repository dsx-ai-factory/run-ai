/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cert

import (
	"os"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
)

const testExampleDomain = "test.example.com"

func TestGetCertHandler_LoadCertFromPath(t *testing.T) {

	// Create temp files for testing
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

	_, _ = certFile.WriteString("test cert content")
	_, _ = keyFile.WriteString("test key content")
	certFile.Close()
	keyFile.Close()

	// Create test data with SSL cert and key paths
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: false})
	runaiData.Inputs[configinputs.InstallInputSSLPath] = certFile.Name()
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = keyFile.Name()

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Verify it's a LoadCertFromPath instance
	if _, ok := handler.(*LoadCertFromPath); !ok {
		t.Errorf("expected LoadCertFromPath, got %T", handler)
	}
}

func TestGetCertHandler_GenerateSelfSignedCert_DevMode(t *testing.T) {

	// Create test data with no SSL paths and dev mode enabled
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: true}) // DevMode = true
	runaiData.Inputs[configinputs.InstallInputFQDN] = testExampleDomain
	// No SSL cert/key paths set, so they will be empty strings

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Verify it's a GenerateSelfSignedCert instance
	if _, ok := handler.(*GenerateSelfSignedCert); !ok {
		t.Errorf("expected GenerateSelfSignedCert, got %T", handler)
	}
}

func TestGetCertHandler_GenerateSelfSignedCert_MissingCertPath(t *testing.T) {

	// Create test data with only cert path but missing key path, dev mode enabled
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: true}) // DevMode = true
	runaiData.Inputs[configinputs.InstallInputFQDN] = testExampleDomain
	runaiData.Inputs[configinputs.InstallInputSSLPath] = "/some/cert/path"
	// InstallInputSSLKeyPath is not set, so it will be empty string

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Should still create GenerateSelfSignedCert because one of the paths is missing
	if _, ok := handler.(*GenerateSelfSignedCert); !ok {
		t.Errorf("expected GenerateSelfSignedCert, got %T", handler)
	}
}

func TestGetCertHandler_GenerateSelfSignedCert_MissingKeyPath(t *testing.T) {

	// Create test data with only key path but missing cert path, dev mode enabled
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: true}) // DevMode = true
	runaiData.Inputs[configinputs.InstallInputFQDN] = testExampleDomain
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = "/some/key/path"
	// InstallInputSSLPath is not set, so it will be empty string

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Should still create GenerateSelfSignedCert because one of the paths is missing
	if _, ok := handler.(*GenerateSelfSignedCert); !ok {
		t.Errorf("expected GenerateSelfSignedCert, got %T", handler)
	}
}

func TestGetCertHandler_LoadCertFromPath_NotDevMode(t *testing.T) {

	// Create test data with no SSL paths and dev mode disabled
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: false}) // DevMode = false
	// No SSL cert/key paths set, so they will be empty strings

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Should create LoadCertFromPath even with empty paths when not in dev mode
	if _, ok := handler.(*LoadCertFromPath); !ok {
		t.Errorf("expected LoadCertFromPath, got %T", handler)
	}
}

func TestGetCertHandler_LoadCertFromPath_WithBothPaths(t *testing.T) {

	// Create test data with both SSL paths set, regardless of dev mode
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: true}) // DevMode = true, but both paths are set
	runaiData.Inputs[configinputs.InstallInputSSLPath] = "/some/cert/path"
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = "/some/key/path"

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Should create LoadCertFromPath because both paths are provided
	if _, ok := handler.(*LoadCertFromPath); !ok {
		t.Errorf("expected LoadCertFromPath, got %T", handler)
	}
}

func TestGetCertHandler_DifferentHandlersForDifferentInputs(t *testing.T) {
	// First call - self-signed cert (dev mode, no paths)
	runaiData1, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: true})
	runaiData1.Inputs[configinputs.InstallInputFQDN] = "test1.example.com"
	handler1 := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData1)

	// Second call - load from path (different parameters)
	runaiData2, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: false})
	runaiData2.Inputs[configinputs.InstallInputSSLPath] = "/different/cert/path"
	runaiData2.Inputs[configinputs.InstallInputSSLKeyPath] = "/different/key/path"
	handler2 := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData2)

	// Verify different handler types are created (no more singleton)
	if _, ok := handler1.(*GenerateSelfSignedCert); !ok {
		t.Errorf("expected GenerateSelfSignedCert from first call, got %T", handler1)
	}

	if _, ok := handler2.(*LoadCertFromPath); !ok {
		t.Errorf("expected LoadCertFromPath from second call, got %T", handler2)
	}
}

func TestGetCertHandler_FQDNPassedCorrectly(t *testing.T) {

	expectedFQDN := "my-test-domain.example.com"
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: true})
	runaiData.Inputs[configinputs.InstallInputFQDN] = expectedFQDN

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Verify it's a GenerateSelfSignedCert and check the FQDN
	selfSignedHandler, ok := handler.(*GenerateSelfSignedCert)
	if !ok {
		t.Fatalf("expected GenerateSelfSignedCert, got %T", handler)
	}

	// Use reflection to check the private fqdn field
	handlerValue := reflect.ValueOf(selfSignedHandler).Elem()
	fqdnField := handlerValue.FieldByName("fqdn")
	if fqdnField.String() != expectedFQDN {
		t.Errorf("expected FQDN %s, got %s", expectedFQDN, fqdnField.String())
	}
}

func TestGetCertHandler_PathsPassedCorrectly(t *testing.T) {

	expectedCertPath := "/test/cert/path"
	expectedKeyPath := "/test/key/path"
	runaiData, _ := model.NewRunAIInstallData(&config.InstallConfig{DevMode: false})
	runaiData.Inputs[configinputs.InstallInputSSLPath] = expectedCertPath
	runaiData.Inputs[configinputs.InstallInputSSLKeyPath] = expectedKeyPath

	handler := GetCertHandler(configinputs.InstallInputSSLPath, configinputs.InstallInputSSLKeyPath, configinputs.InstallInputFQDN, runaiData)

	// Verify it's a LoadCertFromPath and check the paths
	loadCertHandler, ok := handler.(*LoadCertFromPath)
	if !ok {
		t.Fatalf("expected LoadCertFromPath, got %T", handler)
	}

	// Use reflection to check the private path fields
	handlerValue := reflect.ValueOf(loadCertHandler).Elem()
	certPathField := handlerValue.FieldByName("certPath")
	keyPathField := handlerValue.FieldByName("keyPath")

	if certPathField.String() != expectedCertPath {
		t.Errorf("expected cert path %s, got %s", expectedCertPath, certPathField.String())
	}
	if keyPathField.String() != expectedKeyPath {
		t.Errorf("expected key path %s, got %s", expectedKeyPath, keyPathField.String())
	}
}
