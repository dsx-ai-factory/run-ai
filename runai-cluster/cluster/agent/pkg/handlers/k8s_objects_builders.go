/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// k8sObjectBuilder is the per-kind extension point of K8sObjectsHandler.
// Every asset kind that flows through the generic k8s-objects sync endpoint
// (/api/v1/k8s-objects/updates) implements this interface and registers an
// instance in k8sObjectBuilders.
//
// To add a new kind:
//  1. Implement the three methods below for the kind's typed k8s object.
//  2. Register the builder in k8sObjectBuilders at package init.
//  3. Add the kind to getAssetKindLabel so the `run.ai/asset` label
//     identifies it on the cluster side.
//  4. Register the k8s runtime scheme for the GVK in cmd/main.go.
//  5. Grant the agent RBAC on the resource in the runai-chart.
type k8sObjectBuilder interface {
	// NewEmpty returns a zero-value typed object. The common handler uses it
	// to look up the current cluster state via the right GVK (Get/List) and
	// as the destination for inspector copy operations.
	NewEmpty() client.Object

	// Build constructs the target cluster object from a sync update. The
	// returned object must carry the k8s name/namespace (from meta/scope)
	// and the kind-specific spec decoded from the untyped objectSpec; the
	// caller stamps the Run:ai scope/kind/update-count metadata on top so
	// Build only needs to apply user-provided objectMeta labels/annotations
	// (if the kind exposes an inner ObjectMeta, as RCT does for the
	// ResourceClaim it spawns).
	Build(update K8sObjectFromBackend, namespace string) (client.Object, error)

	// CopySpec copies the kind-specific spec from the freshly-built object
	// (src, the control-plane state) onto the latest cluster read (dst),
	// preserving dst's resourceVersion so the subsequent Update doesn't
	// conflict. Called from the inspector's DeepCopySpec hook.
	CopySpec(dst, src client.Object)

	// ShouldReCreate reports whether the diff between the freshly-built
	// (expected) object and the current cluster read (existing) requires
	// delete-and-recreate instead of an in-place update. Kinds whose spec
	// the k8s API server marks immutable (RCT, as of resource.k8s.io/v1)
	// must return true as soon as the immutable fields diverge
	ShouldReCreate(expected, existing client.Object) bool
}

// k8sObjectBuilders is the kind-to-builder registry.
// anything the control plane sends with an unregistered kind is surfaced as a per-asset
// error in the next sync report rather than silently skipped,
var k8sObjectBuilders = map[assetsApi.AssetKind]k8sObjectBuilder{
	assetsApi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE: &rctBuilder{},
}

// rctBuilder reconciles resource.k8s.io/v1 ResourceClaimTemplate objects.
// The control-plane objectSpec payload is the inner ResourceClaimSpec
// (devices.requests, etc.) — that is, the spec of the ResourceClaim that
// the template spawns for each pod. The outer RCT.ObjectMeta is owned by
// us (name from asset name, namespace from scope, Run:ai labels); the
// inner RCT.Spec.ObjectMeta is where user-provided objectMeta
// labels/annotations go so they land on the spawned ResourceClaims and
// are visible to pods that select them.
type rctBuilder struct{}

func (b *rctBuilder) NewEmpty() client.Object {
	return &resourcev1.ResourceClaimTemplate{}
}

func (b *rctBuilder) Build(update K8sObjectFromBackend, namespace string) (client.Object, error) {
	if update.Spec == nil {
		return nil, fmt.Errorf("rct asset %s is missing spec", update.Meta.Name)
	}

	rct := &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      update.Meta.Name,
			Namespace: namespace,
		},
	}
	applyRunaiK8sObjectMetadata(&rct.ObjectMeta, &update.Meta, update.UpdateCount)

	// The agent needs typed k8s API objects, but the control plane carries
	// kind-specific specs as an untyped JSON map. Round-tripping through
	// json.Marshal → json.Unmarshal is the only path that honors the
	// resource.k8s.io CBOR/JSON tag conventions on nested device-request
	// structs; a direct map walk would silently drop fields the k8s types
	// name differently in Go than on the wire.
	rawSpec := update.Spec.GetObjectSpec()
	specBytes, err := json.Marshal(rawSpec)
	if err != nil {
		return nil, fmt.Errorf("marshal rct objectSpec: %w", err)
	}
	if err := json.Unmarshal(specBytes, &rct.Spec.Spec); err != nil {
		return nil, fmt.Errorf("decode rct objectSpec: %w", err)
	}

	// Inner ObjectMeta is applied to each ResourceClaim that this template
	// spawns (not to the RCT itself); it's the user-provided surface for
	// pod-selector labels. Run:ai scope/kind labels on the RCT above are
	// what a future replication controller and cleanup selectors key on.
	if update.Spec.HasObjectMeta() {
		userMeta := update.Spec.GetObjectMeta()
		if labels := userMeta.GetLabels(); len(labels) > 0 {
			rct.Spec.Labels = cloneStringMap(labels)
		}
		if annotations := userMeta.GetAnnotations(); len(annotations) > 0 {
			rct.Spec.Annotations = cloneStringMap(annotations)
		}
	}

	return rct, nil
}

func (b *rctBuilder) CopySpec(dst, src client.Object) {
	dstRct, ok := dst.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return
	}
	srcRct, ok := src.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return
	}
	srcRct.Spec.DeepCopyInto(&dstRct.Spec)
}

// ShouldReCreate for RCT returns true whenever the inner
// ResourceClaimTemplateSpec diverges.
// A metadata-only edit on the outer RCT (labels/annotations the agent
// stamps) still goes through the in-place update path because the outer
// ObjectMeta is mutable and the inner spec is unchanged.
func (b *rctBuilder) ShouldReCreate(expected, existing client.Object) bool {
	expectedRct, ok := expected.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return false
	}
	existingRct, ok := existing.(*resourcev1.ResourceClaimTemplate)
	if !ok {
		return false
	}
	return !reflect.DeepEqual(expectedRct.Spec, existingRct.Spec)
}

// applyRunaiK8sObjectMetadata stamps the Run:ai scope/kind/id/update-count
// metadata on a k8s object before create-or-update. The scope label is what
// a future replication controller will key on to decide which project
// namespaces receive a copy of this k8s object; the kind label lets any
// label-selector in the cluster (cleanup, replication, auditing) identify
// which asset family produced the object.
func applyRunaiK8sObjectMetadata(meta *metav1.ObjectMeta, assetMeta *assetsApi.AssetMetaForSync, updateCount int64) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[constants.RunaiAssetLabel] = getAssetKindLabel(assetMeta.Kind)
	meta.Labels[assetIdLabel] = assetMeta.Id
	composeScopeLabel(meta.Labels, assetMeta)

	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[assetNameAnnotation] = assetMeta.Name
	meta.Annotations[UpdateCountAnnotation] = strconv.FormatInt(updateCount, 10)
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// Package init guards against a nil entry in the builder registry, which
// would otherwise surface as a confusing nil-pointer panic on first sync.
func init() {
	for kind, builder := range k8sObjectBuilders {
		if builder == nil {
			panic(fmt.Sprintf("k8sObjectBuilders: nil builder registered for kind %q", kind))
		}
	}
}
