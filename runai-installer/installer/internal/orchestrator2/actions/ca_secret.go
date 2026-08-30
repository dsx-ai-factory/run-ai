package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CASecret struct {
	SecretName  string
	Namespace   string
	HelmRelease string
	CACertPath  string
	Labels      map[string]string
}

func (a CASecret) Kind() string { return "k8s.ca_secret" }

func (a CASecret) Args() map[string]any {
	return map[string]any{
		"name":      a.SecretName,
		"namespace": a.Namespace,
		"caPath":    a.CACertPath,
	}
}

func (a CASecret) Describe() string {
	return fmt.Sprintf("apply ca secret %s/%s from %s", a.Namespace, a.SecretName, a.CACertPath)
}

func (a CASecret) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	caData, err := os.ReadFile(a.CACertPath)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	labels := map[string]string{"app.kubernetes.io/managed-by": "Helm"}
	for k, v := range a.Labels {
		labels[k] = v
	}
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.SecretName,
			Namespace: a.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"meta.helmclient.sh/release-name":      a.HelmRelease,
				"meta.helmclient.sh/release-namespace": a.Namespace,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{"runai-ca.pem": caData},
	}
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Apply(ctx, secret)
}
