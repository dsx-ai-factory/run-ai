package controller

import (
	"context"
	"testing"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// The project controller enqueues every kai.resources project as a cluster-scoped reconcile
// request, so mapToAllProjects is exercised here.
func TestMapToAllProjects(t *testing.T) {
	scheme := scheme_init.Scheme()
	if err := kaiorg.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add kai.resources to scheme: %v", err)
	}
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(
			&kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
			&kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-b"}},
		).
		Build()

	c := &ProjectController{client: k8sClient}
	assertMapsAllProjects(t, c.mapToAllProjects(context.Background(), nil))
}

func assertMapsAllProjects(t *testing.T, requests []ctrl.Request) {
	t.Helper()
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}
	names := map[string]bool{}
	for _, r := range requests {
		names[r.Name] = true
		if r.Namespace != "" {
			t.Errorf("expected cluster-scoped request (empty namespace), got %q", r.Namespace)
		}
	}
	if !names["proj-a"] || !names["proj-b"] {
		t.Errorf("expected requests for proj-a and proj-b, got %v", names)
	}
}
