/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	orgunitclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

const testIdentity = "flowmark-sa"

type fakeOrgUnits struct {
	projects []orgunitclient.Project
	deletes  atomic.Int32
}

func (f *fakeOrgUnits) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/projects"):
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orgunitclient.GetProjects200Response{Projects: f.projects})
	case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/projects/"):
		f.deletes.Add(1)
		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func managedTestProject(name string, phase orgunitclient.ProjectPhase, namespace string) orgunitclient.Project {
	project := orgunitclient.Project{
		Id: "p1", Name: name, ClusterId: "c1", Description: ownerMarker,
		CreatedBy: orgunitclient.PtrString(testIdentity),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		Status: orgunitclient.ProjectStatus{Phase: &phase},
	}
	if namespace != "" {
		project.Status.Namespace = *orgunitclient.NewNullableString(&namespace)
	}
	return project
}

func testClients(t *testing.T, orgUnits *fakeOrgUnits, objects ...k8sclient.Object) *client.Clients {
	t.Helper()
	server := httptest.NewServer(orgUnits)
	t.Cleanup(server.Close)
	return &client.Clients{
		Identity: testIdentity,
		Kube:     fake.NewClientBuilder().WithObjects(objects...).Build(),
		OrgUnits: orgunitclient.NewAPIClient(&orgunitclient.Configuration{
			Servers: orgunitclient.ServerConfigurations{{URL: server.URL}},
		}),
	}
}

func TestCleanupDeletesAProjectTheClusterNeverReported(t *testing.T) {
	orgUnits := &fakeOrgUnits{projects: []orgunitclient.Project{managedTestProject("flowmark-burst-1", orgunitclient.PROJECTPHASE_CREATING, "")}}
	clients := testClients(t, orgUnits)

	if err := CleanupProject(context.Background(), clients, "c1", "flowmark-burst-1", time.Second); err != nil {
		t.Fatalf("a project stuck in Creating is still ours to delete: %v", err)
	}
	if orgUnits.deletes.Load() != 1 {
		t.Fatalf("the control plane project must be deleted once, got %d deletes", orgUnits.deletes.Load())
	}
}

func TestCleanupDeletesTheLabelledNamespaceOfAReadyProject(t *testing.T) {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "runai-flowmark-burst-1", Labels: map[string]string{projectNamespaceLabel: "flowmark-burst-1"}}}
	orgUnits := &fakeOrgUnits{projects: []orgunitclient.Project{managedTestProject("flowmark-burst-1", orgunitclient.PROJECTPHASE_READY, namespace.Name)}}
	clients := testClients(t, orgUnits, namespace)

	if err := CleanupProject(context.Background(), clients, "c1", "flowmark-burst-1", time.Second); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	var left corev1.Namespace
	if err := clients.Kube.Get(context.Background(), k8sclient.ObjectKey{Name: namespace.Name}, &left); err == nil {
		t.Fatal("the project namespace must be deleted")
	}
	if orgUnits.deletes.Load() != 1 {
		t.Fatalf("expected one project delete, got %d", orgUnits.deletes.Load())
	}
}

func TestCleanupRefusesAProjectSomebodyElseCreated(t *testing.T) {
	project := managedTestProject("flowmark-burst-1", orgunitclient.PROJECTPHASE_READY, "")
	project.CreatedBy = orgunitclient.PtrString("a-human@example.com")
	orgUnits := &fakeOrgUnits{projects: []orgunitclient.Project{project}}
	clients := testClients(t, orgUnits)

	err := CleanupProject(context.Background(), clients, "c1", "flowmark-burst-1", time.Second)
	if err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("a foreign project must never be deleted, got %v", err)
	}
	if orgUnits.deletes.Load() != 0 {
		t.Fatal("no delete may be issued for a foreign project")
	}
}

func TestCleanupOfAMissingProjectIsANoOp(t *testing.T) {
	orgUnits := &fakeOrgUnits{}
	clients := testClients(t, orgUnits)

	if err := CleanupProject(context.Background(), clients, "c1", "flowmark-burst-gone", time.Second); err != nil {
		t.Fatalf("nothing to clean must not be an error: %v", err)
	}
	if orgUnits.deletes.Load() != 0 {
		t.Fatal("no delete may be issued when the project does not exist")
	}
}

func TestCleanupStopsBeforeTheNamespaceWhenTheClusterHasNoProjectKind(t *testing.T) {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "runai-flowmark-burst-1", Labels: map[string]string{projectNamespaceLabel: "flowmark-burst-1"}}}
	orgUnits := &fakeOrgUnits{projects: []orgunitclient.Project{managedTestProject("flowmark-burst-1", orgunitclient.PROJECTPHASE_READY, namespace.Name)}}
	clients := testClients(t, orgUnits, namespace)
	clients.Kube = fake.NewClientBuilder().WithObjects(namespace).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c k8sclient.WithWatch, key k8sclient.ObjectKey, obj k8sclient.Object, opts ...k8sclient.GetOption) error {
			if obj.GetObjectKind().GroupVersionKind() == clusterProjectKind {
				return &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: clusterProjectKind.Group, Kind: clusterProjectKind.Kind}}
			}
			return c.Get(ctx, key, obj, opts...)
		},
	}).Build()

	err := CleanupProject(context.Background(), clients, "c1", "flowmark-burst-1", time.Second)
	if err == nil || !strings.Contains(err.Error(), "kubeconfig") {
		t.Fatalf("a cluster without the Project kind must stop cleanup with a pointer to the kubeconfig, got %v", err)
	}
	if orgUnits.deletes.Load() != 1 {
		t.Fatalf("the control plane project is ours and must still be deleted, got %d deletes", orgUnits.deletes.Load())
	}
	var left corev1.Namespace
	if err := clients.Kube.Get(context.Background(), k8sclient.ObjectKey{Name: namespace.Name}, &left); err != nil {
		t.Fatal("the namespace must be left alone when the cluster cannot be verified")
	}
}
