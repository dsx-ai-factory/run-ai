package actions

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RegistrySecret struct {
	SecretName string
	Namespace  string
	Token      string
}

func (a RegistrySecret) Kind() string { return "k8s.registry_secret" }

func (a RegistrySecret) Args() map[string]any {
	return map[string]any{
		"name":      a.SecretName,
		"namespace": a.Namespace,
	}
}

func (a RegistrySecret) Describe() string {
	return fmt.Sprintf("apply registry credentials secret %s/%s", a.Namespace, a.SecretName)
}

func (a RegistrySecret) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	dockerConfig, err := base64.StdEncoding.DecodeString(a.Token)
	if err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("decode registry token: %w", err)
	}
	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: a.SecretName, Namespace: a.Namespace},
		Type:       corev1.SecretTypeDockerConfigJson,
		Data:       map[string][]byte{".dockerconfigjson": dockerConfig},
	}
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Apply(ctx, secret)
}
