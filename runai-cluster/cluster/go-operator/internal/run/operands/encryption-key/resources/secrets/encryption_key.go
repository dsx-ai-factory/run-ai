package secrets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AES-256-GCM key size in bytes
	aesKeySize = 32
	// encryptionKeyOperandName is the name of the encryption key operand
	encryptionKeyOperandName = "encryption-key"
)

// NewEncryptionKeySecretResource creates a new Resource that manages the encryption key secret
func NewEncryptionKeySecretResource(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) operands.Resource {
	return &encryptionKeySecretResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

// encryptionKeySecretResource is a Resource that manages the encryption key secret for the cluster
type encryptionKeySecretResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

// OperandName returns the name of the operand
func (s *encryptionKeySecretResource) OperandName() string {
	return encryptionKeyOperandName
}

// Object returns the encryption key secret
// It creates a new secret if it doesn't exist, or returns the existing secret if it does
func (s *encryptionKeySecretResource) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = constants.RunaiAdsSecretName
		namespace = s.runaiConfig.Namespace
	)

	secret := &corev1.Secret{}

	// Get the existing secret if it exists
	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	// If secret exists and has the key, keep it
	if err == nil && len(secret.Data[constants.RunaiAdsEncryptionKeyName]) == aesKeySize {
		return secret, nil
	}

	// Generate new key
	keyHash := sha256.Sum256([]byte(s.runaiConfig.Spec.Internal.Global.ControlPlane.ClusterUUID))
	keyHex := hex.EncodeToString(keyHash[:])

	secret.ObjectMeta.Name = name
	secret.ObjectMeta.Namespace = namespace
	secret.Type = corev1.SecretTypeOpaque
	secret.StringData = map[string]string{
		constants.RunaiAdsEncryptionKeyName: keyHex[:32],
	}

	return secret, nil
}

// Status checks if the encryption key secret exists
func (s *encryptionKeySecretResource) Status(ctx context.Context) error {
	secret, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if secret.GetUID() == "" {
		return fmt.Errorf("encryption key secret [%s] is missing", secret.GetName())
	}

	return nil
}

// Enabled returns true if the encryption key secret is enabled
func (s *encryptionKeySecretResource) Enabled() bool {
	return true
}
