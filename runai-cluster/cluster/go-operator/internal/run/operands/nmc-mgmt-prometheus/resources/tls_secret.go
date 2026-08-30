package resources

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	prom_common "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/shared_data"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/common/cert-utils/pkg/generate"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	tlsSecretName = "nmc-mgmt-prometheus-tls-secret"
)

type TLSSecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
	sharedData  *shared_data.SharedData
}

func NewTLSSecretResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig, sharedData *shared_data.SharedData) *TLSSecretResource {
	return &TLSSecretResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
		sharedData:  sharedData,
	}
}

func (r *TLSSecretResource) OperandName() string {
	return r.operandName
}

func (r *TLSSecretResource) Object(ctx context.Context) (client.Object, error) {
	var (
		namespace = r.runaiConfig.Namespace
	)

	secret := &corev1.Secret{}

	// Get the existing secret if it exists to consume any cluster-set values
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      tlsSecretName,
		Namespace: namespace,
	}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	secret.ObjectMeta.Name = tlsSecretName
	secret.ObjectMeta.Namespace = namespace
	secret.Type = corev1.SecretTypeTLS

	// Validate the secret has the correct fields
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	shouldRegenerate := !lo.HasKey(secret.Data, "tls.crt") || !lo.HasKey(secret.Data, "tls.key") || isCertExpiredOrExpiringSoon(secret.Data["tls.crt"])
	if shouldRegenerate {
		cert, key, err := generate.GenerateSelfSignedCert(prom_common.MgmtPrometheusDomain, []string{prom_common.MgmtPrometheusDomain})
		if err != nil {
			return nil, err
		}

		secret.Data["tls.crt"] = cert
		secret.Data["tls.key"] = key
	}

	// Store the certificate in the shared data for the remote cluster prometheus operand to use
	r.sharedData.Set(shared_data.NmcMgmtPrometheusCA, secret.Data["tls.crt"])

	return secret, nil
}

func (r *TLSSecretResource) Status(ctx context.Context) error {
	secret, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if secret.GetUID() == "" {
		return fmt.Errorf("secret [%s] is missing", secret.GetName())
	}

	return nil
}

func (r *TLSSecretResource) Enabled() bool {
	return true
}

func isCertExpiredOrExpiringSoon(certData []byte) bool {
	block, _ := pem.Decode(certData)
	if block == nil {
		return false
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	oneMonthFromNow := time.Now().AddDate(0, 1, 0)

	return oneMonthFromNow.After(cert.NotAfter)
}
