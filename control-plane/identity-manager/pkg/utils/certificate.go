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

// ValidateSigningCertificate parses a SAML IdP signing certificate and checks that it is inside its validity window.
// The certificate's signature and issuer are not checked. A private CA does not publish its certificate in the IdP metadata,
// so a leaf signed by that CA cannot be distinguished from a certificate whose signature does not verify.
// PEM values may include further certificates. Any other value may list several certificates separated by commas; each is validated on its own.
func ValidateSigningCertificate(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if strings.Contains(trimmed, BEGIN_CERTIFICATE) {
		return validateSigningCertificate(trimmed)
	}
	for _, part := range strings.Split(trimmed, ",") {
		if err := validateSigningCertificate(part); err != nil {
			return err
		}
	}
	return nil
}

func validateSigningCertificate(raw string) error {
	certs, err := parseSigningCertificates(raw)
	if err != nil {
		return err
	}
	return leafCurrentlyValid(certs[0])
}

func leafCurrentlyValid(leaf *x509.Certificate) error {
	now := time.Now()
	if now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
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
