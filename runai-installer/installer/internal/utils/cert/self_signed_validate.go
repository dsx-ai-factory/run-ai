/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cert

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"time"
)

func validateCertificateSet(set *SelfSignedCertificateSet, dnsNames []string) error {
	if err := validateCertificateSetPermissions(set); err != nil {
		return err
	}
	ca, caKey, leaf, leafKey, err := loadCertificateSet(set)
	if err != nil {
		return err
	}
	return validateCertificateSetCrypto(ca, caKey, leaf, leafKey, dnsNames)
}

func validateCertificateSetPermissions(set *SelfSignedCertificateSet) error {
	dirInfo, err := os.Stat(set.OutputDir)
	if err != nil {
		return err
	}
	if !dirInfo.IsDir() || dirInfo.Mode().Perm() != 0o700 {
		return fmt.Errorf("%s must be a directory with permissions 0700", set.OutputDir)
	}
	for _, path := range []string{set.CACertPath, set.CAKeyPath, set.CertPath, set.KeyPath} {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
			return fmt.Errorf("%s must be a regular file with permissions 0600", path)
		}
	}
	return nil
}

func loadCertificateSet(set *SelfSignedCertificateSet) (*x509.Certificate, *rsa.PrivateKey, *x509.Certificate, *rsa.PrivateKey, error) {
	ca, err := readCertificate(set.CACertPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	caKey, err := readRSAKey(set.CAKeyPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	leaf, err := readCertificate(set.CertPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	leafKey, err := readRSAKey(set.KeyPath)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return ca, caKey, leaf, leafKey, nil
}

func validateCertificateSetCrypto(ca *x509.Certificate, caKey *rsa.PrivateKey, leaf *x509.Certificate, leafKey *rsa.PrivateKey, dnsNames []string) error {
	now := time.Now()
	if now.Before(ca.NotBefore) || now.After(ca.NotAfter) || now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		return errors.New("certificate is not currently valid")
	}
	if !ca.IsCA {
		return errors.New("CA certificate is not marked as a CA")
	}
	if err := leaf.CheckSignatureFrom(ca); err != nil {
		return err
	}
	if !reflect.DeepEqual(ca.PublicKey, &caKey.PublicKey) {
		return errors.New("CA certificate and key do not match")
	}
	if !reflect.DeepEqual(leaf.PublicKey, &leafKey.PublicKey) {
		return errors.New("server certificate and key do not match")
	}
	actualNames := append([]string(nil), leaf.DNSNames...)
	sort.Strings(actualNames)
	if !reflect.DeepEqual(actualNames, dnsNames) {
		return fmt.Errorf("certificate DNS names %v do not match requested names %v", actualNames, dnsNames)
	}
	return nil
}

func readCertificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("%s does not contain a certificate", path)
	}
	return x509.ParseCertificate(block.Bytes)
}

func readRSAKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("%s does not contain an RSA private key", path)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
