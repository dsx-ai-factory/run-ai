package certificates

import (
	"context"
	"fmt"
	"log"

	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/dns"
	certgen "github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/generate"
	"k8s.io/client-go/kubernetes"
)

type selfSignedSigner struct {
	kbClient kubernetes.Interface
}

var _ CASigner = &selfSignedSigner{}

func (ss *selfSignedSigner) Sign(
	ctx context.Context,
	serviceName, secretName, namespace string, createIfNotExists bool) (*CertificateDetails, error) {
	cert, key, err := certgen.GenerateSelfSignedCert(
		fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		dns.GetServiceDNSNames(serviceName, namespace, nil),
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Creating tls secret %s with %s.%s certificate data", secretName, serviceName, namespace)
	if err = createOrPatchK8sTLSSecret(ctx, ss.kbClient, secretName, namespace, createIfNotExists, cert, key); err != nil {
		return nil, err
	}

	return &CertificateDetails{
		CaBundle: cert,
	}, nil
}
