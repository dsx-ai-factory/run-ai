/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

type CertificateSet struct {
	CACertPath     string
	ServerCertPath string
	ServerKeyPath  string
	TempDir        string
}

func GenerateSelfSignedCerts(domain string, wildcardDomain string) (*CertificateSet, error) {
	tempDir, err := os.MkdirTemp("", "runai-e2e-certs-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	certSet := &CertificateSet{
		CACertPath:     filepath.Join(tempDir, "rootCA.pem"),
		ServerCertPath: filepath.Join(tempDir, "tls.crt"),
		ServerKeyPath:  filepath.Join(tempDir, "tls.key"),
		TempDir:        tempDir,
	}

	caKey, caCert, caCertBytes, err := generateCA()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to generate CA: %w", err)
	}

	if err := savePEMFile(certSet.CACertPath, "CERTIFICATE", caCertBytes); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to save CA cert: %w", err)
	}

	serverKey, serverCert, err := generateServerCert(caKey, caCert, domain, wildcardDomain)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to generate server cert: %w", err)
	}

	if err := savePEMFile(certSet.ServerCertPath, "CERTIFICATE", serverCert); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to save server cert: %w", err)
	}

	serverKeyBytes := x509.MarshalPKCS1PrivateKey(serverKey)
	if err := savePEMFile(certSet.ServerKeyPath, "RSA PRIVATE KEY", serverKeyBytes); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to save server key: %w", err)
	}

	return certSet, nil
}

func (cs *CertificateSet) Cleanup() error {
	if cs.TempDir != "" {
		return os.RemoveAll(cs.TempDir)
	}
	return nil
}

func generateCA() (*rsa.PrivateKey, *x509.Certificate, []byte, error) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, nil, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"RunAI E2E Test"},
			CommonName:   "RunAI E2E Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, err
	}

	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	return caKey, caCert, caCertBytes, nil
}

func generateServerCert(caKey *rsa.PrivateKey, caCert *x509.Certificate, domain string, wildcardDomain string) (*rsa.PrivateKey, []byte, error) {
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	dnsNames := []string{domain}
	if wildcardDomain != "" {
		dnsNames = append(dnsNames, wildcardDomain)
	}

	serverTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"RunAI E2E Test"},
			CommonName:   domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              dnsNames,
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	return serverKey, serverCertBytes, nil
}

func savePEMFile(path string, pemType string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{
		Type:  pemType,
		Bytes: data,
	})
}
