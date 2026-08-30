package orchestrator2

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// manifestFieldOwner identifies the installer as the server-side-apply field
// manager for raw manifests it applies.
const manifestFieldOwner = client.FieldOwner("runai-installer")

type K8sExecutor interface {
	Get(ctx context.Context, key client.ObjectKey, obj client.Object) error
	Create(ctx context.Context, obj client.Object) error
	Apply(ctx context.Context, obj client.Object) error
	// ApplyManifest decodes a (possibly multi-document) YAML manifest and
	// server-side-applies each object. Used for non-Helm prerequisites whose
	// upstream distribution is raw YAML / kustomize output.
	ApplyManifest(ctx context.Context, manifest []byte) error
	// Patch applies a patch to an existing object (e.g. a JSON patch).
	Patch(ctx context.Context, obj client.Object, patch client.Patch) error
	// Delete removes an object. A not-found object is treated as success.
	Delete(ctx context.Context, obj client.Object) error
}

type realK8sExecutor struct {
	client client.Client
}

func NewK8sExecutor(c client.Client) K8sExecutor {
	return &realK8sExecutor{client: c}
}

func (e *realK8sExecutor) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return e.client.Get(ctx, key, obj)
}

func (e *realK8sExecutor) Create(ctx context.Context, obj client.Object) error {
	return e.client.Create(ctx, obj)
}

func (e *realK8sExecutor) Apply(ctx context.Context, obj client.Object) error {
	existing := obj.DeepCopyObject().(client.Object)
	err := e.client.Get(ctx, client.ObjectKeyFromObject(obj), existing)
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}
	return e.client.Create(ctx, obj)
}

func (e *realK8sExecutor) ApplyManifest(ctx context.Context, manifest []byte) error {
	objs, err := decodeManifest(manifest)
	if err != nil {
		return err
	}
	// Apply in kind-priority order (Namespaces, then CRDs, then everything
	// else), mirroring kubectl apply. The decoded documents are often not in a
	// safe order — e.g. a kustomize overlay emits namespaced resources before
	// the Namespace it adds last — so applying in document order would fail with
	// "namespace not found".
	sortByApplyPriority(objs)
	for _, obj := range objs {
		if err := e.client.Patch(ctx, obj, client.Apply, manifestFieldOwner, client.ForceOwnership); err != nil {
			return fmt.Errorf("apply %s %q: %w", obj.GetKind(), obj.GetName(), err)
		}
	}
	return nil
}

// decodeManifest splits a (possibly multi-document) YAML manifest into
// unstructured objects, skipping empty documents.
func decodeManifest(manifest []byte) ([]*unstructured.Unstructured, error) {
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifest), 4096)
	var objs []*unstructured.Unstructured
	for {
		raw := map[string]any{}
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return objs, nil
			}
			return nil, err
		}
		if len(raw) == 0 {
			continue
		}
		objs = append(objs, &unstructured.Unstructured{Object: raw})
	}
}

// sortByApplyPriority stably orders objects so cluster-scoped prerequisites
// (Namespaces, then CRDs) are applied before the resources that depend on them.
// The sort is stable, so the relative order within a tier is preserved.
func sortByApplyPriority(objs []*unstructured.Unstructured) {
	sort.SliceStable(objs, func(i, j int) bool {
		return applyPriority(objs[i].GetKind()) < applyPriority(objs[j].GetKind())
	})
}

func applyPriority(kind string) int {
	switch kind {
	case "Namespace":
		return 0
	case "CustomResourceDefinition":
		return 1
	default:
		return 2
	}
}

func (e *realK8sExecutor) Patch(ctx context.Context, obj client.Object, patch client.Patch) error {
	return e.client.Patch(ctx, obj, patch)
}

func (e *realK8sExecutor) Delete(ctx context.Context, obj client.Object) error {
	err := e.client.Delete(ctx, obj)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
