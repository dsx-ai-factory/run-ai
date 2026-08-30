package apiclient

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type credentialGetter struct {
	configMap *corev1.ConfigMap
	secrets   map[string]*corev1.Secret
	requests  []client.ObjectKey
}

func (g *credentialGetter) Get(_ context.Context, key client.ObjectKey, obj client.Object) error {
	g.requests = append(g.requests, key)
	switch target := obj.(type) {
	case *corev1.ConfigMap:
		g.configMap.DeepCopyInto(target)
		return nil
	case *corev1.Secret:
		secret, ok := g.secrets[key.Name]
		if !ok {
			return fmt.Errorf("secret %s not found", key.Name)
		}
		secret.DeepCopyInto(target)
		return nil
	default:
		return fmt.Errorf("unsupported object %T", obj)
	}
}

func TestUseGeneratedAdminSecret(t *testing.T) {
	tests := map[string]bool{
		"2.24.9":  false,
		"2.25":    false,
		"2.25.99": false,
		"2.26":    true,
		"2.26.0":  true,
		"3.0.0":   true,
		"0.0.0-30005734.cursorrevert-identity-f2a648d": true,
		"":               true,
		"main":           true,
		"release-2.26.x": true,
	}
	for version, expected := range tests {
		t.Run(version, func(t *testing.T) {
			assert.Equal(t, expected, useGeneratedAdminSecret(version))
		})
	}
}

func TestLoadControlPlaneCredentials(t *testing.T) {
	tests := map[string]string{
		"2.25.17": "runai-backend-tenants-manager",
		"2.26.0":  "runai-backend-gen-tenants-manager",
		"0.0.0-30005734.cursorrevert-identity-f2a648d": "runai-backend-gen-tenants-manager",
		"main": "runai-backend-gen-tenants-manager",
	}
	for version, expectedSecret := range tests {
		t.Run(version, func(t *testing.T) {
			getter := &credentialGetter{
				configMap: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "runai-backend-tenants-manager", Namespace: "runai-backend"},
					Data:       map[string]string{"ADMIN_USERNAME": "admin@example.com"},
				},
				secrets: map[string]*corev1.Secret{
					expectedSecret: {
						ObjectMeta: metav1.ObjectMeta{Name: expectedSecret, Namespace: "runai-backend"},
						Data:       map[string][]byte{"ADMIN_PASSWORD": []byte("selected-or-generated")},
					},
				},
			}

			credentials, err := LoadControlPlaneCredentials(context.Background(), getter, "runai-backend", "runai-backend", version)

			require.NoError(t, err)
			assert.Equal(t, "admin@example.com", credentials.Username)
			assert.Equal(t, "selected-or-generated", credentials.Password)
			assert.Equal(t, []client.ObjectKey{
				{Namespace: "runai-backend", Name: "runai-backend-tenants-manager"},
				{Namespace: "runai-backend", Name: expectedSecret},
			}, getter.requests)
		})
	}
}
