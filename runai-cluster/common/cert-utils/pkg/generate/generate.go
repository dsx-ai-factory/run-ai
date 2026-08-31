/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"
)

// Generate self signed certificate.
func GenerateSelfSignedCert(commonName string, dnsNames []string) (certBytes, keyBytes []byte, err error) {
	defaultExpiry := time.Hour * 24 * 365 * 10
	return GenerateSelfSignedCertWithDuration(commonName, dnsNames, defaultExpiry)
}

func GenerateSelfSignedCertWithDuration(commonName string, dnsNames []string, duration time.Duration) (certBytes, keyBytes []byte, err error) {
	// Generate private key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Generate certificate
	now := time.Now()
	// Generate certificate
	unsignedCert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames: dnsNames,
		IsCA:     true,

		NotBefore: now,
		NotAfter:  now.Add(duration),
	}

	// Self sign certificate
	signedCertBytes, err := x509.CreateCertificate(rand.Reader, unsignedCert, unsignedCert, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificate
	certBytes = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: signedCertBytes,
	})

	// Encode private key
	keyBytes = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return
}
