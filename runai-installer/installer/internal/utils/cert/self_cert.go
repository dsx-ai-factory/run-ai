package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// generateSelfSignedCert returns the PEM-encoded cert and key as strings
func generateSelfSignedCert(fqdn string) (certPEM string, keyPEM string, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate private key: %v", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: fqdn,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{fqdn},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", fmt.Errorf("create certificate: %v", err)
	}

	// Encode certificate to PEM string
	var certPEMBuilder strings.Builder
	err = pem.Encode(&certPEMBuilder, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return "", "", fmt.Errorf("encode certificate PEM: %v", err)
	}

	// Encode private key to PEM string
	var keyPEMBuilder strings.Builder
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	err = pem.Encode(&keyPEMBuilder, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return "", "", fmt.Errorf("encode key PEM: %v", err)
	}

	return certPEMBuilder.String(), keyPEMBuilder.String(), nil
}
