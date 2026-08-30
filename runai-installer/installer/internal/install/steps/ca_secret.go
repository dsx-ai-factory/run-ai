package steps

import (
	"context"
	"fmt"
	"os"

	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InstallCAChainSecret struct {
	*step.RunAIInstallStepImpl
	secretName      string
	namespace       string
	helmReleaseName string
	labels          map[string]string
	k8sClient       k8s.K8sClients
}

func NewInstallCAChainSecret(secretName, namespace, helmReleaseName string, labels map[string]string, k8sClient k8s.K8sClients) *InstallCAChainSecret {
	name := fmt.Sprintf("Install %s secret", secretName)
	return &InstallCAChainSecret{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(name),
		secretName:           secretName,
		namespace:            namespace,
		helmReleaseName:      helmReleaseName,
		labels:               labels,
		k8sClient:            k8sClient,
	}
}

func (i *InstallCAChainSecret) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	// Skip evaluation if in terminal state or Execute() is running
	if i.ShouldSkipStatusEvaluation() {
		return i.GetCachedStatus()
	}
	secret := &corev1.Secret{}
	err := i.k8sClient.GetClient().Get(context.Background(), client.ObjectKey{
		Namespace: i.namespace,
		Name:      i.secretName,
	}, secret)

	if err != nil {
		if errors.IsNotFound(err) {
			return i.SetStatus(model.NotInstalled, nil)
		}
		return i.SetStatus(model.Failed, fmt.Errorf("failed to check if secret: %s in namespace: %s exist: [%v]", i.secretName, i.namespace, err))
	}
	return i.SetStatus(model.Installed, nil)
}

func (i *InstallCAChainSecret) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	// Check for context cancellation before starting
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	i.SetStatus(model.Installing, nil)

	secret := &corev1.Secret{}
	err := i.k8sClient.GetClient().Get(ctx, client.ObjectKey{
		Namespace: i.namespace,
		Name:      i.secretName,
	}, secret)

	if err == nil {
		i.SetStatus(model.Installed, nil)
		return nil // secret already exists, but shouldn't we check the content matches?
	}

	if !errors.IsNotFound(err) {
		error := fmt.Errorf("failed to check if secret: %s in namespace: %s exist: [%v]", i.secretName, i.namespace, err)
		i.SetStatus(model.Failed, error)
		return error
	}

	// Check for context cancellation before cert operations
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	cert, err := os.ReadFile(runaiData.GetInputValue(configinputs.InstallInputCAChainPath))
	if err != nil {
		return fmt.Errorf("failed to read cert file: %v", err)
	}

	certData := []byte(cert)

	// Check for context cancellation before creating secret
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
	}
	for k, v := range i.labels {
		labels[k] = v
	}
	annotations := map[string]string{
		"meta.helmclient.sh/release-name":      i.helmReleaseName,
		"meta.helmclient.sh/release-namespace": i.namespace,
	}
	genericSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        i.secretName,
			Namespace:   i.namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"runai-ca.pem": certData,
		},
	}

	// Create the secret in the cluster
	err = i.k8sClient.GetClient().Create(ctx, genericSecret)
	if err != nil {
		error := fmt.Errorf("failed to create CA chain secret '%s', in namespace: %s : %v", i.secretName, i.namespace, err)
		i.SetStatus(model.Failed, error)
		return error
	}
	i.SetStatus(model.Installed, nil)
	return nil
}
