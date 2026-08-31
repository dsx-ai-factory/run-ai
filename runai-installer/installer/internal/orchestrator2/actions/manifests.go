/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyKustomize renders a kustomize overlay from a repo tarball and applies it.
type ApplyKustomize struct {
	TarballURL string
	SubPath    string
}

func (a ApplyKustomize) Kind() string { return "manifest.kustomize" }

func (a ApplyKustomize) Args() map[string]any {
	return map[string]any{"tarball": a.TarballURL, "path": a.SubPath}
}

func (a ApplyKustomize) Describe() string {
	return fmt.Sprintf("kustomize build %s (from %s) and apply", a.SubPath, a.TarballURL)
}

func (a ApplyKustomize) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	rendered, err := deps.ManifestProvider.RenderKustomize(ctx, a.TarballURL, a.SubPath)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{}, deps.K8sExecutor.ApplyManifest(ctx, rendered)
}

// ApplyRemoteManifest fetches a flat YAML manifest from a URL and applies it.
type ApplyRemoteManifest struct {
	URL string
}

func (a ApplyRemoteManifest) Kind() string { return "manifest.remote" }

func (a ApplyRemoteManifest) Args() map[string]any { return map[string]any{"url": a.URL} }

func (a ApplyRemoteManifest) Describe() string {
	return fmt.Sprintf("apply manifest from %s", a.URL)
}

func (a ApplyRemoteManifest) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	manifest, err := deps.ManifestProvider.FetchManifest(ctx, a.URL)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{}, deps.K8sExecutor.ApplyManifest(ctx, manifest)
}

// PatchDeployment applies a strategic-merge patch to an existing Deployment.
// Strategic merge is used (rather than a positional JSON patch) so the patch can
// target a container by name — Kubernetes merges the containers list by name —
// without assuming the container's index in the pod spec.
type PatchDeployment struct {
	Namespace string
	Name      string
	// StrategicMergePatch is the strategic-merge patch body (JSON).
	StrategicMergePatch []byte
}

func (a PatchDeployment) Kind() string { return "k8s.patch_deployment" }

func (a PatchDeployment) Args() map[string]any {
	return map[string]any{"namespace": a.Namespace, "name": a.Name}
}

func (a PatchDeployment) Describe() string {
	return fmt.Sprintf("patch deployment %s/%s", a.Namespace, a.Name)
}

func (a PatchDeployment) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	dep := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: a.Name, Namespace: a.Namespace},
	}
	patch := client.RawPatch(types.StrategicMergePatchType, a.StrategicMergePatch)
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Patch(ctx, dep, patch)
}

// DeleteCRD deletes a CustomResourceDefinition by name (not-found is success).
type DeleteCRD struct {
	Name string
}

func (a DeleteCRD) Kind() string { return "k8s.delete_crd" }

func (a DeleteCRD) Args() map[string]any { return map[string]any{"name": a.Name} }

func (a DeleteCRD) Describe() string {
	return fmt.Sprintf("delete CRD %s (destructive: also removes any existing custom resources of this type)", a.Name)
}

func (a DeleteCRD) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	crd := crdObject(a.Name)
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Delete(ctx, crd)
}

// crdObject builds an unstructured CustomResourceDefinition so Get/Delete work
// without registering the apiextensions types in the client scheme.
func crdObject(name string) *unstructured.Unstructured {
	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion("apiextensions.k8s.io/v1")
	crd.SetKind("CustomResourceDefinition")
	crd.SetName(name)
	return crd
}

// crdEstablishPollInterval is how often WaitForCRDEstablished re-checks the CRD.
// The overall wait is bounded by the step timeout the runner imposes.
const crdEstablishPollInterval = 3 * time.Second

// WaitForCRDEstablished blocks until a CustomResourceDefinition exists and the
// API server reports it Established (its served versions are registered and
// accept requests). It is the gate an operator-managed CRD needs before a
// custom resource of that kind can be applied: an operator installs its CRDs
// asynchronously after its Helm release is created, so applying the CR
// immediately would race the CRD registration.
type WaitForCRDEstablished struct {
	Name string
}

func (a WaitForCRDEstablished) Kind() string { return "k8s.wait_crd_established" }

func (a WaitForCRDEstablished) Args() map[string]any { return map[string]any{"name": a.Name} }

func (a WaitForCRDEstablished) Describe() string {
	return fmt.Sprintf("wait for CRD %s to be established", a.Name)
}

func (a WaitForCRDEstablished) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	for {
		established, err := a.isEstablished(ctx, deps)
		if err != nil {
			return orchestrator2.ActionResult{}, err
		}
		if established {
			return orchestrator2.ActionResult{}, nil
		}
		select {
		case <-ctx.Done():
			return orchestrator2.ActionResult{}, fmt.Errorf("waiting for CRD %s to be established: %w", a.Name, ctx.Err())
		case <-time.After(crdEstablishPollInterval):
		}
	}
}

// isEstablished reports whether the CRD is present with an Established=True
// condition. A not-found CRD is not an error: the operator has not created it
// yet, so the caller keeps polling.
func (a WaitForCRDEstablished) isEstablished(ctx context.Context, deps orchestrator2.Deps) (bool, error) {
	crd := crdObject(a.Name)
	err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Name: a.Name}, crd)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return hasTrueCondition(crd, "Established"), nil
}

// crdDeletePollInterval is how often WaitForCRDDeleted re-checks the CRD. The
// overall wait is bounded by the step timeout the runner imposes.
const crdDeletePollInterval = 3 * time.Second

// WaitForCRDDeleted blocks until a CustomResourceDefinition no longer exists.
// CRD deletion is asynchronous — the API server sets a deletion timestamp and
// runs the customresourcecleanup finalizer before the object disappears. A
// caller that recreates the CRD with a different set of versions must wait for
// the old one to be fully gone; otherwise the recreate races the deletion and
// the API server rejects it because the old status.storedVersions still
// references a version absent from the new spec.
type WaitForCRDDeleted struct {
	Name string
}

func (a WaitForCRDDeleted) Kind() string { return "k8s.wait_crd_deleted" }

func (a WaitForCRDDeleted) Args() map[string]any { return map[string]any{"name": a.Name} }

func (a WaitForCRDDeleted) Describe() string {
	return fmt.Sprintf("wait for CRD %s to be fully deleted", a.Name)
}

func (a WaitForCRDDeleted) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	for {
		crd := crdObject(a.Name)
		err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Name: a.Name}, crd)
		if apierrors.IsNotFound(err) {
			return orchestrator2.ActionResult{}, nil
		}
		if err != nil {
			return orchestrator2.ActionResult{}, err
		}
		select {
		case <-ctx.Done():
			return orchestrator2.ActionResult{}, fmt.Errorf("waiting for CRD %s to be deleted: %w", a.Name, ctx.Err())
		case <-time.After(crdDeletePollInterval):
		}
	}
}

// deploymentAvailablePollInterval is how often WaitForDeploymentAvailable
// re-checks the Deployment. The overall wait is bounded by the step timeout.
const deploymentAvailablePollInterval = 3 * time.Second

// WaitForDeploymentAvailable blocks until a Deployment reports an Available=True
// condition (its pods have passed their readiness probes). It gates work that
// needs an operator's admission webhook to be serving: a CRD being Established
// only means the type is registered, not that the operator pod backing its
// webhook is ready to admit custom resources, so applying a CR immediately
// after the CRD appears can be rejected by a not-yet-ready webhook.
type WaitForDeploymentAvailable struct {
	Namespace string
	Name      string
}

func (a WaitForDeploymentAvailable) Kind() string { return "k8s.wait_deployment_available" }

func (a WaitForDeploymentAvailable) Args() map[string]any {
	return map[string]any{"namespace": a.Namespace, "name": a.Name}
}

func (a WaitForDeploymentAvailable) Describe() string {
	return fmt.Sprintf("wait for deployment %s/%s to be available", a.Namespace, a.Name)
}

func (a WaitForDeploymentAvailable) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	for {
		available, err := a.isAvailable(ctx, deps)
		if err != nil {
			return orchestrator2.ActionResult{}, err
		}
		if available {
			return orchestrator2.ActionResult{}, nil
		}
		select {
		case <-ctx.Done():
			return orchestrator2.ActionResult{}, fmt.Errorf("waiting for deployment %s/%s to be available: %w", a.Namespace, a.Name, ctx.Err())
		case <-time.After(deploymentAvailablePollInterval):
		}
	}
}

// isAvailable reports whether the Deployment is present with an Available=True
// condition. A not-found Deployment is not an error: the operator's install may
// not have created it yet, so the caller keeps polling. The Deployment is read
// as unstructured so the condition check does not depend on the typed scheme.
func (a WaitForDeploymentAvailable) isAvailable(ctx context.Context, deps orchestrator2.Deps) (bool, error) {
	dep := &unstructured.Unstructured{}
	dep.SetAPIVersion("apps/v1")
	dep.SetKind("Deployment")
	dep.SetNamespace(a.Namespace)
	dep.SetName(a.Name)
	err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: a.Namespace, Name: a.Name}, dep)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return hasTrueCondition(dep, "Available"), nil
}

// hasTrueCondition reports whether obj's status.conditions contains a condition
// of the given type with status "True".
func hasTrueCondition(obj *unstructured.Unstructured, condType string) bool {
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found || err != nil {
		return false
	}
	for _, c := range conditions {
		m, ok := c.(map[string]any)
		if ok && m["type"] == condType && m["status"] == "True" {
			return true
		}
	}
	return false
}
