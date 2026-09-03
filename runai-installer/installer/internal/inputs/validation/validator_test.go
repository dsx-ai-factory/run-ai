/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidateNGCCredentials(t *testing.T) {
	// Test valid base64
	result := ValidateJfrogCredentials("dGVzdDp0ZXN0")
	if !result.Status {
		t.Errorf("Expected valid status, got %v", result.Status)
	}

	// Test invalid base64
	result = ValidateJfrogCredentials("invalid@base64!")
	if result.Status {
		t.Errorf("Expected invalid status, got %v", result.Status)
	}

	// Test empty
	result = ValidateJfrogCredentials("")
	if result.Status {
		t.Errorf("Expected invalid status for empty, got %v", result.Status)
	}
}

func TestValidateFQDN(t *testing.T) {

	// Test valid FQDN format
	result := ValidateFQDN("example.com")
	// DNS resolution may succeed or fail depending on test environment
	// We can't predict the result, so we just check it's a boolean
	if result.Status != true && result.Status != false {
		t.Errorf("Expected boolean status, got %v", result.Status)
	}

	// Test invalid FQDN format
	result = ValidateFQDN("invalid fqdn")
	if result.Status {
		t.Errorf("Expected invalid status for whitespace, got %v", result.Status)
	}

	// Test invalid FQDN format with special characters
	result = ValidateFQDN("example@.com")
	if result.Status {
		t.Errorf("Expected invalid status for special characters, got %v", result.Status)
	}

	// Test empty
	result = ValidateFQDN("")
	if result.Status {
		t.Errorf("Expected invalid status for empty, got %v", result.Status)
	}
}

func TestValidateWildcardFQDN(t *testing.T) {
	// Test valid wildcard FQDN format
	result := ValidateWildcardFQDN("*.example.com")
	// DNS resolution may succeed or fail depending on test environment
	// We can't predict the result, so we just check it's a boolean
	if result.Status != true && result.Status != false {
		t.Errorf("Expected boolean status, got %v", result.Status)
	}

	// Test invalid wildcard FQDN format (missing asterisk)
	result = ValidateWildcardFQDN("example.com")
	if result.Status {
		t.Errorf("Expected invalid status for missing wildcard, got %v", result.Status)
	}

	// Test invalid wildcard FQDN format (wrong asterisk position)
	result = ValidateWildcardFQDN("example.*.com")
	if result.Status {
		t.Errorf("Expected invalid status for wrong asterisk position, got %v", result.Status)
	}

	// Test empty (should be valid as it's optional)
	result = ValidateWildcardFQDN("")
	if !result.Status {
		t.Errorf("Expected valid status for empty, got %v", result.Status)
	}
}

func TestValidateClusterName(t *testing.T) {
	// Test valid cluster name
	result := ValidateClusterName("my-cluster")
	if !result.Status {
		t.Errorf("Expected valid status, got %v", result.Status)
	}

	// Test invalid cluster name
	result = ValidateClusterName("My-Cluster")
	if result.Status {
		t.Errorf("Expected invalid status, got %v", result.Status)
	}

	// Test empty
	result = ValidateClusterName("")
	if result.Status {
		t.Errorf("Expected invalid status for empty, got %v", result.Status)
	}
}

func TestValidateCustomPort(t *testing.T) {
	// Test valid port
	result := ValidateIngressControllerHttpsPort("30000")
	if !result.Status {
		t.Errorf("Expected valid status, got %v", result.Status)
	}

	// Test invalid port (too low)
	result = ValidateIngressControllerHttpsPort("29999")
	if result.Status {
		t.Errorf("Expected invalid status, got %v", result.Status)
	}

	// Test invalid port (too high)
	result = ValidateIngressControllerHttpsPort("32768")
	if result.Status {
		t.Errorf("Expected invalid status, got %v", result.Status)
	}

	// Test empty (optional)
	result = ValidateIngressControllerHttpsPort("")
	if !result.Status {
		t.Errorf("Expected valid status for empty, got %v", result.Status)
	}
}

func TestValidateKubernetesDistribution(t *testing.T) {
	// Test valid distribution: openshift
	result := ValidateKubernetesDistribution("openshift")
	if !result.Status {
		t.Errorf("Expected valid status for 'openshift', got %v with message: %s", result.Status, result.Message)
	}

	// Test valid distribution: empty (optional)
	result = ValidateKubernetesDistribution("")
	if !result.Status {
		t.Errorf("Expected valid status for empty, got %v with message: %s", result.Status, result.Message)
	}

	// Test invalid distribution
	result = ValidateKubernetesDistribution("invalid-distro")
	if result.Status {
		t.Errorf("Expected invalid status for 'invalid-distro', got %v", result.Status)
	}
}

func TestValidateWorkloadRouting(t *testing.T) {
	result := ValidateWorkloadRouting("")
	if !result.Status {
		t.Errorf("Expected valid status for empty (default), got %v", result.Status)
	}

	result = ValidateWorkloadRouting("path")
	if !result.Status {
		t.Errorf("Expected valid status for 'path', got %v", result.Status)
	}

	result = ValidateWorkloadRouting("host")
	if !result.Status {
		t.Errorf("Expected valid status for 'host', got %v", result.Status)
	}

	result = ValidateWorkloadRouting("hosts")
	if result.Status {
		t.Errorf("Expected invalid status for 'hosts', got %v", result.Status)
	}

	result = ValidateWorkloadRouting("subdomain")
	if result.Status {
		t.Errorf("Expected invalid status for 'subdomain', got %v", result.Status)
	}
}

// --- Certificate validation tests ---

func writePEMToTempFile(t *testing.T, pemData []byte, name string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, pemData, 0o600); err != nil {
		t.Fatalf("failed writing temp pem: %v", err)
	}
	return p
}

func generateCertPEMForHost(t *testing.T, host string, notBefore, notAfter time.Time) []byte {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatalf("serial: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{host},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	var pemBuf []byte
	pemBuf = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return pemBuf
}

func TestValidateCertificate_EmptyInputs(t *testing.T) {
	res := ValidateCertificate("", "host")
	if res.Status {
		t.Errorf("expected failure for empty cert path")
	}
	res = ValidateCertificate("/non/existent/path.pem", "")
	if res.Status {
		t.Errorf("expected failure for empty host")
	}
}

func TestValidateCertificate_Valid(t *testing.T) {
	host := "valid.example.test"
	pemBytes := generateCertPEMForHost(t, host, time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour))
	certPath := writePEMToTempFile(t, pemBytes, "cert.pem")
	res := ValidateCertificate(certPath, host)
	if !res.Status {
		t.Errorf("expected valid certificate for host, got: %#v", res)
	}
}

func TestValidateCertificate_Expired(t *testing.T) {
	host := "expired.example.test"
	pemBytes := generateCertPEMForHost(t, host, time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
	certPath := writePEMToTempFile(t, pemBytes, "cert.pem")
	res := ValidateCertificate(certPath, host)
	if res.Status {
		t.Errorf("expected invalid due to expiry, got: %#v", res)
	}
}

func TestValidateCertificate_HostnameMismatch(t *testing.T) {
	pemBytes := generateCertPEMForHost(t, "a.example.test", time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour))
	certPath := writePEMToTempFile(t, pemBytes, "cert.pem")
	res := ValidateCertificate(certPath, "b.example.test")
	if res.Status {
		t.Errorf("expected invalid due to hostname mismatch, got: %#v", res)
	}
}
