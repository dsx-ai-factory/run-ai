/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

const (
	CERTIFICATE       = "CERTIFICATE"
	BEGIN_CERTIFICATE = "BEGIN " + CERTIFICATE
	END_CERTIFICATE   = "END " + CERTIFICATE
)

type CertificateData struct {
	SigningAlgorithm string
	NotBefore        time.Time
	ExpirationDate   time.Time
}

func SigningCertificateExpiration(raw string) time.Time {
	certificates, err := parseSigningCertificate(raw)
	if err != nil || len(certificates) == 0 {
		return time.Time{}
	}
	return certificates[0].ExpirationDate
}

func ValidateSigningCertificate(raw string) ([]CertificateData, error) {
	certificates, err := parseSigningCertificate(raw)
	if err != nil {
		return certificates, err
	}
	for _, certificate := range certificates {
		if err := certificateCurrentlyValid(certificate); err != nil {
			return certificates, err
		}
	}
	return certificates, nil
}

func parseSigningCertificate(raw string) ([]CertificateData, error) {
	var certificatesData []CertificateData
	for _, part := range signingCertificateParts(raw) {
		certs, err := parseSigningCertificates(part)
		if len(certs) > 0 {
			certificatesData = append(certificatesData, CertificateData{
				SigningAlgorithm: certs[0].SignatureAlgorithm.String(),
				NotBefore:        certs[0].NotBefore.UTC(),
				ExpirationDate:   certs[0].NotAfter.UTC(),
			})
		}
		if err != nil {
			return certificatesData, err
		}
	}
	if len(certificatesData) == 0 {
		return nil, fmt.Errorf("signing certificate is required: %w", ErrUnsignedCertificate)
	}
	return certificatesData, nil
}


func signingCertificateParts(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	if strings.Contains(trimmed, BEGIN_CERTIFICATE) {
		return []string{trimmed}
	}
	return strings.Split(trimmed, ",")
}

func certificateCurrentlyValid(certificate CertificateData) error {
	now := time.Now()
	if now.Before(certificate.NotBefore) || now.After(certificate.ExpirationDate) {
		return fmt.Errorf("signing certificate is outside its validity window: %w", ErrUnsignedCertificate)
	}
	return nil
}

func parseSigningCertificates(raw string) ([]*x509.Certificate, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("signing certificate is required: %w", ErrUnsignedCertificate)
	}
	ders, err := decodeCertificates(trimmed)
	if err != nil {
		return nil, fmt.Errorf("signing certificate is invalid: %w: %w", ErrUnsignedCertificate, err)
	}
	if len(ders) == 0 {
		return nil, fmt.Errorf("signing certificate is invalid: %w", ErrUnsignedCertificate)
	}
	certs := make([]*x509.Certificate, 0, len(ders))
	for _, der := range ders {
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("signing certificate is invalid: %w: %w", ErrUnsignedCertificate, err)
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

func decodeCertificates(raw string) ([][]byte, error) {
	if strings.Contains(raw, BEGIN_CERTIFICATE) {
		var ders [][]byte
		rest := []byte(raw)
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}
			if block.Type != CERTIFICATE {
				return nil, fmt.Errorf("pem block is not a certificate")
			}
			ders = append(ders, block.Bytes)
		}
		if len(ders) == 0 {
			return nil, fmt.Errorf("pem block is not a certificate")
		}
		return ders, nil
	}

	compact := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, raw)
	der, err := base64.StdEncoding.DecodeString(compact)
	if err != nil {
		der, err = base64.RawStdEncoding.DecodeString(compact)
	}
	if err != nil {
		return nil, err
	}
	return [][]byte{der}, nil
}
