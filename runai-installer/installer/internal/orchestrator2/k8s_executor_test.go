package orchestrator2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestSortByApplyPriority verifies the manifest apply order puts the Namespace
// first and the CRD second, even when the source document order lists a
// namespaced resource first (as a kustomize overlay does). Without this, the
// namespaced ServiceAccount would be applied before its namespace exists.
func TestSortByApplyPriority(t *testing.T) {
	manifest := []byte(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: training-operator
  namespace: kubeflow
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tfjobs.kubeflow.org
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubeflow
`)

	objs, err := decodeManifest(manifest)
	require.NoError(t, err)
	require.Len(t, objs, 3)

	sortByApplyPriority(objs)

	assert.Equal(t, "Namespace", objs[0].GetKind(), "namespace applied first")
	assert.Equal(t, "CustomResourceDefinition", objs[1].GetKind(), "CRD applied before namespaced resources")
	assert.Equal(t, "ServiceAccount", objs[2].GetKind())
}

// TestSortByApplyPriority_StableWithinTier verifies the sort preserves the
// document order of resources that share a priority tier.
func TestSortByApplyPriority_StableWithinTier(t *testing.T) {
	manifest := []byte(`
apiVersion: v1
kind: Service
metadata:
  name: a
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: b
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: c
`)
	objs, err := decodeManifest(manifest)
	require.NoError(t, err)
	sortByApplyPriority(objs)

	var names []string
	for _, o := range objs {
		names = append(names, o.GetName())
	}
	assert.Equal(t, []string{"a", "b", "c"}, names, "same-tier resources keep document order")
}

func getUnstructured(t *testing.T, c client.Client, gvk schema.GroupVersionKind, key client.ObjectKey) error {
	t.Helper()
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	return c.Get(context.Background(), key, obj)
}

// TestApplyManifest_AppliesAllDocuments verifies the real executor decodes a
// multi-document manifest and server-side-applies every object, so each ends up
// present in the cluster regardless of its order in the source document.
func TestApplyManifest_AppliesAllDocuments(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	exec := NewK8sExecutor(c)

	manifest := []byte(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: training-operator
  namespace: kubeflow
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubeflow
`)
	require.NoError(t, exec.ApplyManifest(context.Background(), manifest))

	require.NoError(t, getUnstructured(t, c,
		schema.GroupVersionKind{Version: "v1", Kind: "Namespace"},
		client.ObjectKey{Name: "kubeflow"}), "namespace applied")
	require.NoError(t, getUnstructured(t, c,
		schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"},
		client.ObjectKey{Namespace: "kubeflow", Name: "training-operator"}), "namespaced resource applied")
}

// TestApplyManifest_Idempotent verifies re-applying the same manifest succeeds
// (server-side apply is an upsert), so a re-run does not error.
func TestApplyManifest_Idempotent(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	exec := NewK8sExecutor(c)

	manifest := []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: mpi-operator
`)
	require.NoError(t, exec.ApplyManifest(context.Background(), manifest))
	require.NoError(t, exec.ApplyManifest(context.Background(), manifest), "re-apply must not error")

	require.NoError(t, getUnstructured(t, c,
		schema.GroupVersionKind{Version: "v1", Kind: "Namespace"},
		client.ObjectKey{Name: "mpi-operator"}))
}

// TestApplyManifest_SkipsEmptyDocuments verifies blank YAML documents (common
// from kustomize output with leading/trailing separators) are ignored rather
// than producing an apply error.
func TestApplyManifest_SkipsEmptyDocuments(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	exec := NewK8sExecutor(c)

	manifest := []byte(`
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubeflow
---
`)
	require.NoError(t, exec.ApplyManifest(context.Background(), manifest))
	require.NoError(t, getUnstructured(t, c,
		schema.GroupVersionKind{Version: "v1", Kind: "Namespace"},
		client.ObjectKey{Name: "kubeflow"}))
}
