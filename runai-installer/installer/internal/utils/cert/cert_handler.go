/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cert

import (
	"fmt"
	"os"
)

type CertHandler interface {
	GetCRT() (string, error)
	GetKey() (string, error)
}

type LoadCertFromPath struct {
	certPath string
	keyPath  string
}

func NewLoadCertFromPath(certPath string, keyPath string) *LoadCertFromPath {
	return &LoadCertFromPath{
		certPath: certPath,
		keyPath:  keyPath,
	}
}

func (c *LoadCertFromPath) GetCRT() (string, error) {
	cert, err := os.ReadFile(c.certPath)
	if err != nil {
		return "", fmt.Errorf("failed to read cert file: %v", err)
	}
	return string(cert), nil
}

func (c *LoadCertFromPath) GetKey() (string, error) {
	key, err := os.ReadFile(c.keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %v", err)
	}
	return string(key), nil
}

type GenerateSelfSignedCert struct {
	fqdn    string
	certPEM string
	keyPEM  string
}

func NewGenerateSelfSignedCert(fqdn string) *GenerateSelfSignedCert {
	return &GenerateSelfSignedCert{fqdn: fqdn}
}

func (c *GenerateSelfSignedCert) GetCRT() (string, error) {
	if c.certPEM != "" {
		return c.certPEM, nil
	}
	var err error
	c.certPEM, c.keyPEM, err = generateSelfSignedCert(c.fqdn)
	if err != nil {
		return "", fmt.Errorf("failed to generate self-signed cert: %v", err)
	}
	return c.certPEM, nil
}

func (c *GenerateSelfSignedCert) GetKey() (string, error) {
	if c.keyPEM != "" {
		return c.keyPEM, nil
	}
	var err error
	c.certPEM, c.keyPEM, err = generateSelfSignedCert(c.fqdn)
	if err != nil {
		return "", fmt.Errorf("failed to generate self-signed cert: %v", err)
	}
	return c.keyPEM, nil
}
