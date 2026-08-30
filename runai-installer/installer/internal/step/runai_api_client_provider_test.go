package step_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/apiclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type credentialGetter struct {
	configMap *corev1.ConfigMap
	secret    *corev1.Secret
}

func (g *credentialGetter) Get(_ context.Context, _ client.ObjectKey, obj client.Object) error {
	switch target := obj.(type) {
	case *corev1.ConfigMap:
		g.configMap.DeepCopyInto(target)
	case *corev1.Secret:
		g.secret.DeepCopyInto(target)
	}
	return nil
}

func TestRunAIAPIClientProviderUsesControlPlaneCredentials(t *testing.T) {
	apiclient.ResetClients()
	t.Cleanup(apiclient.ResetClients)

	var request struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_version" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"version":"2.26.0"}`))
			return
		}
		if r.URL.Path != "/api/v1/token" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode token request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"accessToken":"token"}`))
	}))
	defer server.Close()

	getter := &credentialGetter{
		configMap: &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "runai-backend-tenants-manager"},
			Data:       map[string]string{"ADMIN_USERNAME": "admin@example.com"},
		},
		secret: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "runai-backend-gen-tenants-manager"},
			Data:       map[string][]byte{"ADMIN_PASSWORD": []byte("generated-password")},
		},
	}

	provider := step.NewRunAIAPIClientProvider(getter, "2.26.0")
	result, err := provider(server.URL, "")
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if result == nil {
		t.Fatal("expected client")
	}
	if request.Username == nil || *request.Username != "admin@example.com" {
		t.Fatalf("unexpected username: %v", request.Username)
	}
	if request.Password == nil || *request.Password != "generated-password" {
		t.Fatalf("unexpected password: %v", request.Password)
	}
}
