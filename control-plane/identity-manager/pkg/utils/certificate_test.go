/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateSigningCertificate_Base64DER(t *testing.T) {
	t.Parallel()
	signed := createCertificate(t, true)

	_, err := ValidateSigningCertificate(base64.StdEncoding.EncodeToString(signed))
	require.NoError(t, err)
}

func TestValidateSigningCertificate_PEM(t *testing.T) {
	t.Parallel()
	signed := createCertificate(t, true)
	encoded := pem.EncodeToMemory(&pem.Block{Type: CERTIFICATE, Bytes: signed})

	_, err := ValidateSigningCertificate(string(encoded))
	require.NoError(t, err)
}

func TestValidateSigningCertificate_CommaSeparatedBase64(t *testing.T) {
	t.Parallel()
	first := base64.StdEncoding.EncodeToString(createCertificate(t, true))
	second := base64.StdEncoding.EncodeToString(createCertificate(t, true))

	_, err := ValidateSigningCertificate(first + "," + second)
	require.NoError(t, err)
	_, err = ValidateSigningCertificate(first + ", " + second)
	require.NoError(t, err)
}

func TestValidateSigningCertificate_CommaSeparatedReturnsFirstError(t *testing.T) {
	t.Parallel()
	valid := base64.StdEncoding.EncodeToString(createCertificate(t, true))
	unsigned := base64.StdEncoding.EncodeToString(createCertificate(t, false))

	_, err := ValidateSigningCertificate("not-a-certificate," + unsigned)
	require.ErrorIs(t, err, ErrUnsignedCertificate)
	require.ErrorContains(t, err, "invalid")

	_, err = ValidateSigningCertificate(valid + "," + unsigned)
	require.NoError(t, err)
}

func TestValidateSigningCertificate_Missing(t *testing.T) {
	t.Parallel()

	_, err := ValidateSigningCertificate("  ")
	require.ErrorIs(t, err, ErrUnsignedCertificate)
	require.ErrorContains(t, err, "required")
}

func TestValidateSigningCertificate_NotACertificate(t *testing.T) {
	t.Parallel()

	_, err := ValidateSigningCertificate("not-a-certificate")
	require.ErrorIs(t, err, ErrUnsignedCertificate)
	require.ErrorContains(t, err, "invalid")
	var corrupt base64.CorruptInputError
	require.ErrorAs(t, err, &corrupt)
}

func TestValidateSigningCertificate_TamperedSignatureStillParses(t *testing.T) {
	t.Parallel()
	signed := createCertificate(t, true)
	tampered := append([]byte(nil), signed...)
	tampered[len(tampered)-2] ^= 0xff

	_, err := ValidateSigningCertificate(base64.StdEncoding.EncodeToString(tampered))
	require.NoError(t, err)
}

func TestValidateSigningCertificate_SelfSignedOutsideValidityWindow(t *testing.T) {
	t.Parallel()
	now := time.Now()
	expired := certificateTemplate("idp")
	expired.NotBefore = now.Add(-2 * time.Hour)
	expired.NotAfter = now.Add(-time.Hour)
	notYetValid := certificateTemplate("idp")
	notYetValid.NotBefore = now.Add(time.Hour)
	notYetValid.NotAfter = now.Add(2 * time.Hour)

	for _, template := range []*x509.Certificate{expired, notYetValid} {
		key := newKey(t)
		der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
		require.NoError(t, err)
		certs, err := ValidateSigningCertificate(base64.StdEncoding.EncodeToString(der))
		require.ErrorIs(t, err, ErrUnsignedCertificate)
		require.ErrorContains(t, err, "validity window")
		require.Len(t, certs, 1)
		parsed, parseErr := x509.ParseCertificate(der)
		require.NoError(t, parseErr)
		require.True(t, certs[0].ExpirationDate.Equal(parsed.NotAfter))
	}
}

func TestValidateSigningCertificate_SignedByDifferentKeyIsAccepted(t *testing.T) {
	t.Parallel()

	_, err := ValidateSigningCertificate(base64.StdEncoding.EncodeToString(createCertificate(t, false)))
	require.NoError(t, err)
}

func TestValidateSigningCertificate_LeafSignedByTrustedPublicCA(t *testing.T) {
	t.Parallel()
	leaf, _ := trustedPublicLeaf(t)

	_, err := ValidateSigningCertificate(base64.StdEncoding.EncodeToString(leaf))
	require.NoError(t, err)
}

func TestValidateSigningCertificate_IntermediateIncludedWithLeaf(t *testing.T) {
	t.Parallel()
	caCert, caKey := trustedCA(t)
	intermediateDER, intermediateKey := issueCertificate(t, caCert, caKey, "public-intermediate", true)
	intermediate, err := x509.ParseCertificate(intermediateDER)
	require.NoError(t, err)
	leafDER, _ := issueCertificate(t, intermediate, intermediateKey, "idp.example.com", false)

	chain := pem.EncodeToMemory(&pem.Block{Type: CERTIFICATE, Bytes: leafDER})
	chain = append(chain, pem.EncodeToMemory(&pem.Block{Type: CERTIFICATE, Bytes: intermediateDER})...)

	_, err = ValidateSigningCertificate(string(chain))
	require.NoError(t, err)
}

func TestValidateSigningCertificate_SignedByPrivateCA(t *testing.T) {
	t.Parallel()
	leaf, _ := trustedPublicLeaf(t)

	_, err := ValidateSigningCertificate(base64.StdEncoding.EncodeToString(leaf))
	require.NoError(t, err)
}

func trustedPublicLeaf(t *testing.T) ([]byte, *x509.CertPool) {
	caCert, caKey := trustedCA(t)
	leaf, _ := issueCertificate(t, caCert, caKey, "idp.example.com", false)
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	return leaf, roots
}

func trustedCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	der, key := issueCertificate(t, nil, nil, "public-ca", true)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return cert, key
}

func issueCertificate(t *testing.T, parent *x509.Certificate, parentKey *ecdsa.PrivateKey, commonName string, isCA bool) ([]byte, *ecdsa.PrivateKey) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	if isCA {
		template.IsCA = true
		template.BasicConstraintsValid = true
		template.KeyUsage = x509.KeyUsageCertSign
	} else {
		template.KeyUsage = x509.KeyUsageDigitalSignature
	}
	signer := parent
	signerKey := parentKey
	if signer == nil {
		signer = template
		signerKey = key
	}
	der, err := x509.CreateCertificate(rand.Reader, template, signer, &key.PublicKey, signerKey)
	require.NoError(t, err)
	return der, key
}

func createCertificate(t *testing.T, selfSigned bool) []byte {
	subjectKey := newKey(t)
	issuerKey := subjectKey
	issuer := certificateTemplate("idp")
	subject := issuer
	if !selfSigned {
		issuerKey = newKey(t)
		issuer = certificateTemplate("ca")
		subject = certificateTemplate("idp")
	}
	der, err := x509.CreateCertificate(rand.Reader, subject, issuer, &subjectKey.PublicKey, issuerKey)
	require.NoError(t, err)
	return der
}

func newKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	return key
}

func certificateTemplate(commonName string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
}
