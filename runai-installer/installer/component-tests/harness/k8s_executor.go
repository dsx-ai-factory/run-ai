/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sCall struct {
	Verb      string
	Kind      string
	Namespace string
	Name      string
	Object    client.Object
	// PatchType and PatchData are populated for "patch" calls.
	PatchType string
	PatchData []byte
}

type FakeK8sExecutor struct {
	mu    sync.Mutex
	Calls []K8sCall
	// CreateErrors maps an object name to an error returned by Create, letting
	// tests simulate a failing apply. The call is still recorded.
	CreateErrors map[string]error
	// Existing keys (namespace + name) that Get and Apply treat as already
	// present. Keying on the full ObjectKey, rather than name alone, means a
	// lookup in the wrong namespace is not falsely reported as present.
	Existing map[client.ObjectKey]bool
	// SeededObjects provides object content returned by Get (for detects that
	// inspect contents, e.g. CRD versions). A seeded key is also treated as
	// present. Only unstructured targets are populated.
	SeededObjects map[client.ObjectKey]*unstructured.Unstructured
	// AppliedManifests records the raw manifests passed to ApplyManifest, in
	// order, so tests can assert non-Helm prerequisites were applied.
	AppliedManifests [][]byte
}

func (f *FakeK8sExecutor) Get(_ context.Context, key client.ObjectKey, obj client.Object) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, K8sCall{
		Verb:      "get",
		Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace: key.Namespace,
		Name:      key.Name,
		Object:    obj,
	})
	if seeded, ok := f.SeededObjects[key]; ok {
		reqKind := obj.GetObjectKind().GroupVersionKind().Kind
		if reqKind == "" || reqKind == seeded.GetKind() {
			if u, ok := obj.(*unstructured.Unstructured); ok {
				u.Object = seeded.DeepCopy().Object
			}
			return nil
		}
	}
	if f.Existing[key] {
		return nil
	}
	return apierrors.NewNotFound(schema.GroupResource{Resource: key.Name}, key.Name)
}

func (f *FakeK8sExecutor) Create(_ context.Context, obj client.Object) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, K8sCall{
		Verb:      "create",
		Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
		Object:    obj,
	})
	if err, ok := f.CreateErrors[obj.GetName()]; ok {
		return err
	}
	f.markExisting(obj)
	return nil
}

func (f *FakeK8sExecutor) Apply(_ context.Context, obj client.Object) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, K8sCall{
		Verb:      "apply",
		Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
		Object:    obj,
	})
	if f.Existing[client.ObjectKeyFromObject(obj)] {
		return nil
	}
	if err, ok := f.CreateErrors[obj.GetName()]; ok {
		return err
	}
	f.markExisting(obj)
	return nil
}

func (f *FakeK8sExecutor) ApplyManifest(_ context.Context, manifest []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, K8sCall{Verb: "apply_manifest"})
	f.AppliedManifests = append(f.AppliedManifests, manifest)
	return nil
}

func (f *FakeK8sExecutor) Patch(_ context.Context, obj client.Object, patch client.Patch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, _ := patch.Data(obj)
	f.Calls = append(f.Calls, K8sCall{
		Verb:      "patch",
		Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
		Object:    obj,
		PatchType: string(patch.Type()),
		PatchData: data,
	})
	return nil
}

func (f *FakeK8sExecutor) Delete(_ context.Context, obj client.Object) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, K8sCall{
		Verb:      "delete",
		Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
		Object:    obj,
	})
	delete(f.Existing, client.ObjectKeyFromObject(obj))
	delete(f.SeededObjects, client.ObjectKeyFromObject(obj))
	return nil
}

func (f *FakeK8sExecutor) markExisting(obj client.Object) {
	if f.Existing == nil {
		f.Existing = map[client.ObjectKey]bool{}
	}
	f.Existing[client.ObjectKeyFromObject(obj)] = true
}

func (f *FakeK8sExecutor) CallsByVerb(verb string) []K8sCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []K8sCall
	for _, c := range f.Calls {
		if c.Verb == verb {
			out = append(out, c)
		}
	}
	return out
}
