package certificates

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type CertificateDetails struct {
	CaBundle []byte
}

type CASigner interface {
	Sign(ctx context.Context, serviceName, secretName, namespace string, createIfNotExists bool) (*CertificateDetails, error)
}

func NewSigner(conf *rest.Config) CASigner {
	return &selfSignedSigner{kubernetes.NewForConfigOrDie(conf)}
}

func NewCertificateDetails(caBundle []byte) *CertificateDetails {
	return &CertificateDetails{
		CaBundle: caBundle,
	}
}
