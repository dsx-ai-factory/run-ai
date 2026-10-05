/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	inferencev1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/inference-endpoint-controller/internal/render"
)

const (
	// fieldManager is the server-side-apply owner of every field the controller writes.
	fieldManager = "inference-endpoint-controller"
	// requeueNotReady is how long to wait for the Gateway or the EPP before looking again.
	requeueNotReady = 10 * time.Second
)

// GVKs of the two child kinds the controller handles as unstructured objects,
// so that it does not depend on the Gateway API or GIE Go modules.
var (
	poolGVK    = schema.GroupVersionKind{Group: "inference.networking.k8s.io", Version: "v1", Kind: "InferencePool"}
	routeGVK   = schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
	rewriteGVK = schema.GroupVersionKind{Group: "inference.networking.x-k8s.io", Version: "v1alpha2", Kind: "InferenceModelRewrite"}
)

// scorerPlugins maps the CRD's scorer names to GIE plugin types.
var scorerPlugins = map[inferencev1alpha1.ScorerType]string{
	inferencev1alpha1.ScorerQueue:       "queue-scorer",
	inferencev1alpha1.ScorerKVCache:     "kv-cache-utilization-scorer",
	inferencev1alpha1.ScorerPrefixCache: "prefix-cache-scorer",
}

// InferenceEndpointReconciler turns one InferenceEndpoint into its routing stack.
type InferenceEndpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// APIReader reads readiness from the API server after applying children.
	// The informer cache can still contain the previously accepted generation.
	APIReader client.Reader
}

// +kubebuilder:rbac:groups=inference.platform.run.ai,resources=inferenceendpoints,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=inference.platform.run.ai,resources=inferenceendpoints/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=inference.platform.run.ai,resources=inferenceendpoints/finalizers,verbs=update
//
// Child objects the controller creates and watches.
// +kubebuilder:rbac:groups=inference.networking.k8s.io,resources=inferencepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services;configmaps;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings,verbs=get;list;watch;create;update;patch;delete
//
// +kubebuilder:rbac:groups=inference.networking.x-k8s.io,resources=inferencemodelrewrites,verbs=get;list;watch;create;update;patch;delete
//
// Permissions the generated Roles grant to the EPP. Kubernetes only lets the
// controller create those Roles if it holds the same permissions itself.
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=inference.networking.x-k8s.io,resources=inferenceobjectives,verbs=get;list;watch

// Reconcile renders the child objects for one InferenceEndpoint, applies them
// with server-side apply, and reports whether the Gateway accepted the route
// and the EPP is serving.
//
//  1. Fetch the endpoint. Gone means garbage collection removes the children.
//  2. Render the ten child objects from the spec.
//  3. Set the owner reference and apply each one. Any failure: Ready=False ApplyFailed.
//  4. Read the HTTPRoute parent status. Not accepted: Ready=False RouteNotAccepted.
//  5. Read the EPP Deployment. Current rollout not complete: Ready=False EppNotAvailable.
//  6. Otherwise Ready=True.
func (r *InferenceEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	endpoint := &inferencev1alpha1.InferenceEndpoint{}
	if err := r.Get(ctx, req.NamespacedName, endpoint); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	children, err := renderChildren(endpoint)
	if err != nil {
		// The CRD schema should have rejected this spec; nothing will change until the user edits it.
		return ctrl.Result{}, r.setNotReady(ctx, endpoint, inferencev1alpha1.ReasonApplyFailed, err.Error())
	}

	for _, child := range children {
		if err := controllerutil.SetControllerReference(endpoint, child, r.Scheme); err != nil {
			return ctrl.Result{}, r.setNotReady(ctx, endpoint, inferencev1alpha1.ReasonApplyFailed, err.Error())
		}
		applyConfig := client.ApplyConfigurationFromUnstructured(child)
		if err := r.Apply(ctx, applyConfig, client.FieldOwner(fieldManager), client.ForceOwnership); err != nil {
			msg := fmt.Sprintf("apply %s %s: %v", child.GetKind(), child.GetName(), err)
			if statusErr := r.setNotReady(ctx, endpoint, inferencev1alpha1.ReasonApplyFailed, msg); statusErr != nil {
				log.Error(statusErr, "Failed to record apply failure in status")
			}
			return ctrl.Result{}, fmt.Errorf("%s", msg)
		}
	}
	log.V(1).Info("Applied child objects", "count", len(children))

	// Server-side apply never deletes: when the last rewrite is removed from the
	// spec, the InferenceModelRewrite has to go explicitly.
	if len(endpoint.Spec.Rewrites) == 0 {
		if err := r.deleteRewrite(ctx, endpoint); err != nil {
			return ctrl.Result{}, err
		}
	}

	routeOK, routeMsg, err := r.routeAccepted(ctx, endpoint)
	if err != nil {
		return ctrl.Result{}, err
	}
	eppOK, eppMsg, err := r.eppAvailable(ctx, endpoint)
	if err != nil {
		return ctrl.Result{}, err
	}

	base := endpoint.DeepCopy()
	routeCondition := metav1.Condition{
		Type: inferencev1alpha1.ConditionRouteAccepted, Status: metav1.ConditionTrue,
		Reason: inferencev1alpha1.ReasonRouteAccepted, Message: routeMsg, ObservedGeneration: endpoint.Generation,
	}
	if !routeOK {
		routeCondition.Status = metav1.ConditionFalse
		routeCondition.Reason = inferencev1alpha1.ReasonRouteNotAccepted
	}
	meta.SetStatusCondition(&endpoint.Status.Conditions, routeCondition)

	ready := metav1.Condition{
		Type: inferencev1alpha1.ConditionReady, Status: metav1.ConditionTrue,
		Reason:             inferencev1alpha1.ReasonRouteAccepted,
		Message:            fmt.Sprintf("%d objects applied, route accepted, EPP available", len(children)),
		ObservedGeneration: endpoint.Generation,
	}
	switch {
	case !routeOK:
		ready.Status, ready.Reason, ready.Message = metav1.ConditionFalse, inferencev1alpha1.ReasonRouteNotAccepted, routeMsg
	case !eppOK:
		ready.Status, ready.Reason, ready.Message = metav1.ConditionFalse, inferencev1alpha1.ReasonEppNotAvailable, eppMsg
	}
	meta.SetStatusCondition(&endpoint.Status.Conditions, ready)
	endpoint.Status.ObservedGeneration = endpoint.Generation

	// A merge patch carries no resourceVersion, so a stale informer cache cannot
	// turn a routine status write into a conflict. We own the whole status.
	// NotFound means the endpoint was deleted while we were reconciling it.
	if err := r.Status().Patch(ctx, endpoint, client.MergeFrom(base)); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if ready.Status != metav1.ConditionTrue {
		return ctrl.Result{RequeueAfter: requeueNotReady}, nil
	}
	return ctrl.Result{}, nil
}

// renderChildren resolves the spec and renders the child objects.
func renderChildren(endpoint *inferencev1alpha1.InferenceEndpoint) ([]*unstructured.Unstructured, error) {
	in, err := inputFor(endpoint)
	if err != nil {
		return nil, err
	}
	return render.Render(in)
}

// inputFor projects the spec onto the flat template input. API concepts such as
// scorer types and EPP profiles are translated here so render stays API-agnostic.
func inputFor(endpoint *inferencev1alpha1.InferenceEndpoint) (render.Input, error) {
	image, err := eppImageFor(endpoint.Spec.Epp)
	if err != nil {
		return render.Input{}, err
	}
	in := render.Input{
		Name:             endpoint.Name,
		Namespace:        endpoint.Namespace,
		Selector:         endpoint.Spec.Workload.Selector.MatchLabels,
		Port:             endpoint.Spec.Workload.Port,
		GatewayName:      endpoint.Spec.Routing.GatewayRef.Name,
		GatewayNamespace: endpoint.Spec.Routing.GatewayRef.Namespace,
		Hostnames:        endpoint.Spec.Routing.Hostnames,
		EppImage:         image,
		RuntimeProfiles:  endpoint.Spec.Epp.Profile == inferencev1alpha1.EppProfileLLMDV010,
	}
	if endpoint.Spec.Epp.Scheduling != nil {
		in.ActiveProfile = string(endpoint.Spec.Epp.Scheduling.ActiveProfile)
		in.DefaultProfile = string(endpoint.Spec.Epp.Scheduling.DefaultProfile)
		for _, profile := range endpoint.Spec.Epp.Scheduling.Profiles {
			out := render.SchedulingProfile{Name: string(profile.Name)}
			for _, s := range profile.Scorers {
				out.Scorers = append(out.Scorers, render.Scorer{Plugin: scorerPlugins[s.Type], Weight: s.Weight})
			}
			in.Profiles = append(in.Profiles, out)
		}
		for _, s := range endpoint.Spec.Epp.Scheduling.Scorers {
			in.Scorers = append(in.Scorers, render.Scorer{Plugin: scorerPlugins[s.Type], Weight: s.Weight})
		}
	}
	for _, rw := range endpoint.Spec.Rewrites {
		out := render.Rewrite{Model: rw.Model}
		for _, t := range rw.Targets {
			out.Targets = append(out.Targets, render.RewriteTarget{Model: t.Model, Weight: t.Weight})
		}
		in.Rewrites = append(in.Rewrites, out)
	}
	return in, nil
}

// eppImageFor resolves the EPP profile to its pinned image. Image-only input
// remains supported for the approved GIE image. The CRD schema enforces the same
// rules at admission; this keeps the controller correct for objects that bypass it.
func eppImageFor(epp inferencev1alpha1.EppSpec) (string, error) {
	switch epp.Profile {
	case "":
		if epp.Image == "" {
			return "", errors.New("epp.profile or epp.image is required")
		}
		if epp.Image != inferencev1alpha1.EppImageGIEV15 {
			return "", errors.New("image-only EPP requires the approved GIE v1.5.0 image; otherwise select a supported profile")
		}
		return epp.Image, nil
	case inferencev1alpha1.EppProfileGIEV15:
		if epp.Image != "" && epp.Image != inferencev1alpha1.EppImageGIEV15 {
			return "", fmt.Errorf("epp.image must match the pinned %s profile image", epp.Profile)
		}
		return inferencev1alpha1.EppImageGIEV15, nil
	case inferencev1alpha1.EppProfileLLMDV010:
		if epp.Image != "" && epp.Image != inferencev1alpha1.EppImageLLMDV010 {
			return "", fmt.Errorf("epp.image must match the pinned %s profile image", epp.Profile)
		}
		return inferencev1alpha1.EppImageLLMDV010, nil
	default:
		return "", fmt.Errorf("unsupported epp.profile %q", epp.Profile)
	}
}

// deleteRewrite removes the endpoint's InferenceModelRewrite if it exists.
func (r *InferenceEndpointReconciler) deleteRewrite(ctx context.Context, endpoint *inferencev1alpha1.InferenceEndpoint) error {
	rewrite := &unstructured.Unstructured{}
	rewrite.SetGroupVersionKind(rewriteGVK)
	rewrite.SetName(endpoint.Name + "-rewrite")
	rewrite.SetNamespace(endpoint.Namespace)
	if err := r.Delete(ctx, rewrite); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete stale InferenceModelRewrite: %w", err)
	}
	return nil
}

// setNotReady records a terminal-looking failure on the Ready condition.
func (r *InferenceEndpointReconciler) setNotReady(ctx context.Context, endpoint *inferencev1alpha1.InferenceEndpoint, reason, message string) error {
	base := endpoint.DeepCopy()
	meta.SetStatusCondition(&endpoint.Status.Conditions, metav1.Condition{
		Type: inferencev1alpha1.ConditionReady, Status: metav1.ConditionFalse,
		Reason: reason, Message: message, ObservedGeneration: endpoint.Generation,
	})
	endpoint.Status.ObservedGeneration = endpoint.Generation
	return client.IgnoreNotFound(r.Status().Patch(ctx, endpoint, client.MergeFrom(base)))
}

// routeAccepted reads the HTTPRoute's parent status for our Gateway and reports
// whether Accepted=True and ResolvedRefs=True refer to the current generation.
func (r *InferenceEndpointReconciler) routeAccepted(ctx context.Context, endpoint *inferencev1alpha1.InferenceEndpoint) (bool, string, error) {
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(routeGVK)
	if err := r.readinessReader().Get(ctx, types.NamespacedName{Name: endpoint.Name + "-route", Namespace: endpoint.Namespace}, route); err != nil {
		if apierrors.IsNotFound(err) {
			// The child may have been deleted since apply, including with its endpoint.
			return false, "HTTPRoute not observed yet", nil
		}
		return false, "", fmt.Errorf("get HTTPRoute: %w", err)
	}

	gatewayName := endpoint.Spec.Routing.GatewayRef.Name
	parents, _, _ := unstructured.NestedSlice(route.Object, "status", "parents")
	for _, p := range parents {
		parent, ok := p.(map[string]any)
		if !ok {
			continue
		}
		refName, _, _ := unstructured.NestedString(parent, "parentRef", "name")
		if refName != gatewayName {
			continue
		}
		accepted, acceptedMsg := conditionTrue(parent, "Accepted", route.GetGeneration())
		resolved, resolvedMsg := conditionTrue(parent, "ResolvedRefs", route.GetGeneration())
		if accepted && resolved {
			return true, fmt.Sprintf("Gateway %s accepted the route", gatewayName), nil
		}
		return false, fmt.Sprintf("Gateway %s: Accepted=%t (%s), ResolvedRefs=%t (%s)", gatewayName, accepted, acceptedMsg, resolved, resolvedMsg), nil
	}
	return false, fmt.Sprintf("Gateway %s has not reported status for the route yet", gatewayName), nil
}

// conditionTrue looks up one condition in an unstructured parent status entry.
func conditionTrue(parent map[string]any, conditionType string, generation int64) (bool, string) {
	conditions, _, _ := unstructured.NestedSlice(parent, "conditions")
	for _, c := range conditions {
		condition, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if condition["type"] != conditionType {
			continue
		}
		observed, _, _ := unstructured.NestedInt64(condition, "observedGeneration")
		if observed != generation {
			return false, fmt.Sprintf("waiting for route generation %d (observed %d)", generation, observed)
		}
		msg, _ := condition["message"].(string)
		if reason, ok := condition["reason"].(string); ok && msg == "" {
			msg = reason
		}
		return condition["status"] == string(metav1.ConditionTrue), msg
	}
	return false, "condition not reported"
}

// eppAvailable reports whether the current EPP Deployment generation has finished
// rolling out. A previous generation's available replica does not serve new config.
func (r *InferenceEndpointReconciler) eppAvailable(ctx context.Context, endpoint *inferencev1alpha1.InferenceEndpoint) (bool, string, error) {
	dep := &appsv1.Deployment{}
	if err := r.readinessReader().Get(ctx, types.NamespacedName{Name: endpoint.Name + "-epp", Namespace: endpoint.Namespace}, dep); err != nil {
		if apierrors.IsNotFound(err) {
			return false, "EPP Deployment not observed yet", nil
		}
		return false, "", fmt.Errorf("get EPP Deployment: %w", err)
	}
	replicas := int32(1)
	if dep.Spec.Replicas != nil {
		replicas = *dep.Spec.Replicas
	}
	if dep.Status.ObservedGeneration < dep.Generation || dep.Status.UpdatedReplicas < replicas ||
		dep.Status.Replicas != replicas || dep.Status.AvailableReplicas < replicas {
		return false, fmt.Sprintf("EPP Deployment %s rollout is pending: observed generation %d/%d, updated %d/%d, available %d, total %d",
			dep.Name, dep.Status.ObservedGeneration, dep.Generation, dep.Status.UpdatedReplicas,
			replicas, dep.Status.AvailableReplicas, dep.Status.Replicas), nil
	}
	return true, fmt.Sprintf("EPP Deployment %s has %d available replicas", dep.Name, dep.Status.AvailableReplicas), nil
}

// SetupWithManager registers the controller. Watching every child kind is what
// makes drift repair work: deleting or editing a child requeues its endpoint.
func (r *InferenceEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.APIReader == nil {
		r.APIReader = mgr.GetAPIReader()
	}
	pool := &unstructured.Unstructured{}
	pool.SetGroupVersionKind(poolGVK)
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(routeGVK)
	rewrite := &unstructured.Unstructured{}
	rewrite.SetGroupVersionKind(rewriteGVK)

	return ctrl.NewControllerManagedBy(mgr).
		For(&inferencev1alpha1.InferenceEndpoint{}).
		Owns(pool).
		Owns(route).
		Owns(rewrite).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Named("inferenceendpoint").
		Complete(r)
}

func (r *InferenceEndpointReconciler) readinessReader() client.Reader {
	if r.APIReader != nil {
		return r.APIReader
	}
	// Direct clients (including envtest) do not need a separate reader.
	return r.Client
}
