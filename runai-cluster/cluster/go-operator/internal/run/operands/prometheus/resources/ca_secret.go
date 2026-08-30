package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/shared_data"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	caSecretName = "nmc-mgmt-prometheus-ca"
	caSecretKey  = "ca.crt"
)

type CaSecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
	sharedData  *shared_data.SharedData
}

func NewCaSecretResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig, sharedData *shared_data.SharedData) *CaSecretResource {
	return &CaSecretResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
		sharedData:  sharedData,
	}
}

func (r *CaSecretResource) OperandName() string {
	return r.operandName
}

func (r *CaSecretResource) Object(ctx context.Context) (client.Object, error) {
	var (
		namespace = r.runaiConfig.Namespace
	)

	secret := &corev1.Secret{}

	// Get the existing secret if it exists to consume any cluster-set values
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      caSecretName,
		Namespace: namespace,
	}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	secret.ObjectMeta.Name = caSecretName
	secret.ObjectMeta.Namespace = namespace
	secret.Type = corev1.SecretTypeOpaque

	// Validate the secret has the correct fields
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}

	if !r.Enabled() {
		return secret, nil
	}

	// Get the certificate from the shared data
	ca, ok := r.sharedData.Get(shared_data.NmcMgmtPrometheusCA)
	if !ok {
		return nil, fmt.Errorf("NMC Management Prometheus CA certificate not found in shared data")
	}
	secret.Data[caSecretKey] = ca.([]byte)

	return secret, nil
}

func (r *CaSecretResource) Status(ctx context.Context) error {
	secret, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if secret.GetUID() == "" {
		return fmt.Errorf("secret [%s] is missing", secret.GetName())
	}

	return nil
}

func (r *CaSecretResource) Enabled() bool {
	return r.runaiConfig.Spec.Internal.Global.NMC.IsEnabled()
}
