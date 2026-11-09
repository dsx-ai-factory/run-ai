/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"fmt"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// InferenceEndpointNotReadyReason is the ExternalAccessReady reason while the
// InferenceEndpoint exists but its Ready condition is not yet true.
const InferenceEndpointNotReadyReason = "InferenceEndpointNotReady"

var errMultipleInferenceEndpoints = errors.New("at most one endpoint may set externalAccess.inference")

// inferenceEndpointAvailable reports whether the InferenceEndpoint API is
// installed. A missing API is a NoMatch error so callers can treat it as
// "not installed" unless a workload requests inference routing.
func (p *Plugin) inferenceEndpointAvailable() error {
	if p.mapper == nil {
		return &meta.NoKindMatchError{
			GroupKind:        endpoints.InferenceEndpointGVK.GroupKind(),
			SearchedVersions: []string{endpoints.InferenceEndpointGVK.Version},
		}
	}
	if _, err := p.mapper.RESTMapping(endpoints.InferenceEndpointGVK.GroupKind(), endpoints.InferenceEndpointGVK.Version); err != nil {
		return fmt.Errorf("InferenceEndpoint API is unavailable: %w", err)
	}
	return nil
}

func countInferenceEndpoints(cfg *endpoints.Config) int {
	count := 0
	for i := range cfg.Endpoints {
		if ea := cfg.Endpoints[i].ExternalAccess; ea != nil && ea.Inference != nil {
			count++
		}
	}
	return count
}

// inferenceWorkload is the Pod selector and container port the InferenceEndpoint targets.
type inferenceWorkload struct {
	selector map[string]string
	port     int32
}

const (
	// nimServiceKind is the kind of the NIMService manifest in the runner template.
	nimServiceKind = "NIMService"
	// nimAppLabel and nimLeaderRoleLabel are the labels the NIM operator puts on
	// the Pods of a NIMService and selects them with.
	nimAppLabel        = "app"
	nimLeaderRoleLabel = "nim-llm-role"
	nimLeaderRole      = "leader"
	// nimLWSSuffix is appended to the NIMService name to name the LeaderWorkerSet
	// of a multi-node NIMService.
	nimLWSSuffix = "-lws"
)

// resolveInferenceWorkload returns the Pod selector and container port the
// InferenceEndpoint targets. It reads nothing from the cluster: a created
// Service already carries the selector, and the selector of a referenced NIM
// operator Service follows from the NIMService.
func resolveInferenceWorkload(runner *dromosv1alpha1.WorkloadRunner, endpoint *endpoints.Endpoint) (inferenceWorkload, error) {
	if err := validateInferenceEndpointShape(endpoint); err != nil {
		return inferenceWorkload{}, err
	}
	service := firstService(endpoint)
	if service.Create != nil {
		port := service.Create.Spec.Ports[0]
		workload := inferenceWorkload{selector: service.Create.Spec.Selector, port: port.TargetPort.IntVal}
		if workload.port == 0 {
			workload.port = port.Port
		}
		return workload, nil
	}
	return nimInferenceWorkload(runner, service.Ref)
}

// nimInferenceWorkload predicts the selector the NIM operator puts on the Service
// of the runner's NIMService (GetSelectorLabels in k8s-nim-operator): app is the
// NIMService name, or the name of its LeaderWorkerSet and the leader role on a
// multi-node NIMService, where only the leader Pod serves HTTP. The operator
// serves on the Service port, so the Service port is the container port.
func nimInferenceWorkload(runner *dromosv1alpha1.WorkloadRunner, ref *endpoints.ServiceRef) (inferenceWorkload, error) {
	template := &unstructured.Unstructured{}
	if err := template.UnmarshalJSON(runner.Spec.Template.Raw); err != nil {
		return inferenceWorkload{}, fmt.Errorf("read the runner template: %w", err)
	}
	if template.GetKind() != nimServiceKind {
		return inferenceWorkload{}, fmt.Errorf(
			"externalAccess.inference with a referenced Service requires a %s, got %q", nimServiceKind, template.GetKind())
	}
	name := template.GetName()
	if ref.Name != name {
		return inferenceWorkload{}, fmt.Errorf(
			"externalAccess.inference services[0].ref must name the %s %q, got %q", nimServiceKind, name, ref.Name)
	}
	selector := map[string]string{nimAppLabel: name}
	if multiNode, found, _ := unstructured.NestedFieldNoCopy(template.Object, "spec", "multiNode"); found && multiNode != nil {
		selector = map[string]string{nimAppLabel: name + nimLWSSuffix, nimLeaderRoleLabel: nimLeaderRole}
	}
	return inferenceWorkload{selector: selector, port: ref.Port}, nil
}

// inferenceEndpointTarget derives the InferenceEndpoint workload and routing
// from the endpoint.
func inferenceEndpointTarget(
	endpoint *endpoints.Endpoint,
	workload inferenceWorkload,
) (endpoints.InferenceEndpointTarget, error) {
	var target endpoints.InferenceEndpointTarget
	if err := validateInferenceEndpointShape(endpoint); err != nil {
		return target, err
	}
	target.Selector = workload.selector
	target.Port = workload.port
	route := endpoint.ExternalAccess.HTTPRoute
	parent := route.Spec.ParentRefs[0]
	target.GatewayName = string(parent.Name)
	if parent.Namespace != nil {
		target.GatewayNamespace = string(*parent.Namespace)
	}
	for _, hostname := range route.Spec.Hostnames {
		target.Hostnames = append(target.Hostnames, string(hostname))
	}
	return target, nil
}

// validateInferenceEndpointShape checks that every InferenceEndpoint field the
// plugin derives is representable. Admission and reconcile share it.
func validateInferenceEndpointShape(endpoint *endpoints.Endpoint) error {
	ea := endpoint.ExternalAccess
	if ea == nil || ea.Inference == nil {
		return fmt.Errorf("externalAccess.inference is required")
	}
	if err := ea.Inference.Validate(); err != nil {
		return err
	}
	if ea.HTTPRoute == nil || ea.Ingress != nil || ea.Route != nil {
		return fmt.Errorf("externalAccess.inference requires externalAccess.httpRoute without ingress or route")
	}
	service := firstService(endpoint)
	switch {
	case service != nil && service.Create != nil:
		if len(service.Create.Spec.Selector) == 0 || len(service.Create.Spec.Ports) == 0 {
			return fmt.Errorf("externalAccess.inference requires a created Service with a selector and a port")
		}
		if targetPort := service.Create.Spec.Ports[0].TargetPort; targetPort.StrVal != "" {
			return fmt.Errorf("externalAccess.inference requires a numeric Service targetPort, got %q", targetPort.StrVal)
		}
	case service != nil && service.Ref != nil:
		// The referenced Service may not exist at admission; reconcile resolves it.
		if service.Ref.Name == "" || service.Ref.Port == 0 {
			return fmt.Errorf("externalAccess.inference requires a referenced Service with a name and a port")
		}
	default:
		return fmt.Errorf("externalAccess.inference requires a created Service with a selector and a port, " +
			"or a referenced Service with a name and a port")
	}
	if service.Auth != nil {
		return fmt.Errorf("externalAccess.inference does not support services[0].auth")
	}
	route := ea.HTTPRoute
	if len(route.Spec.Hostnames) == 0 {
		return fmt.Errorf("externalAccess.inference requires httpRoute.spec.hostnames")
	}
	if len(route.Spec.ParentRefs) != 1 {
		return fmt.Errorf("externalAccess.inference requires exactly one httpRoute.spec.parentRefs entry")
	}
	parent := route.Spec.ParentRefs[0]
	if parent.Name == "" || parent.SectionName != nil || parent.Port != nil ||
		(parent.Group != nil && *parent.Group != gatewayv1.GroupName) ||
		(parent.Kind != nil && *parent.Kind != "Gateway") {
		return fmt.Errorf("externalAccess.inference requires the httpRoute parentRef to name a Gateway without sectionName or port")
	}
	return nil
}

func (p *Plugin) reconcileEndpointInference(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	endpoint *endpoints.Endpoint,
	backendServiceName string,
	backendServicePort int32,
	k8sClient client.Client,
) error {
	if !runner.DeletionTimestamp.IsZero() {
		return nil
	}
	if err := p.inferenceEndpointAvailable(); err != nil {
		return err
	}
	workload, err := resolveInferenceWorkload(runner, endpoint)
	if err != nil {
		return err
	}
	target, err := inferenceEndpointTarget(endpoint, workload)
	if err != nil {
		return err
	}
	composed := p.buildHTTPRoute(runner, endpoint.ExternalAccess.HTTPRoute, backendServiceName, backendServicePort)
	if p.enforceOwnership {
		// The InferenceEndpoint controller attaches its route to this Gateway, so it
		// gets the same configured-Gateway restriction as a composed HTTPRoute.
		if err := validateGatewayParents(composed, configuredRoutingConstraints()); err != nil {
			return ownershipEnforcementError("validate inference Gateway", err)
		}
	}
	desired, err := endpoints.BuildInferenceEndpoint(runner, endpoint.ExternalAccess.Inference, target)
	if err != nil {
		return err
	}
	if err := p.applyInferenceEndpoint(ctx, runner, desired, k8sClient); err != nil {
		return err
	}
	// The InferenceEndpoint's own HTTPRoute serves these hostnames. Remove the
	// composed HTTPRoute left from before inference was requested. Without the
	// HTTPRoute API there is no composed route to remove.
	if err := p.deleteIfOwned(ctx, runner, composed, k8sClient, "HTTPRoute"); err != nil && !meta.IsNoMatchError(err) {
		return err
	}
	return nil
}

func (p *Plugin) applyInferenceEndpoint(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	desired *unstructured.Unstructured,
	k8sClient client.Client,
) error {
	live := &unstructured.Unstructured{}
	live.SetGroupVersionKind(endpoints.InferenceEndpointGVK)
	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(desired), live)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("read InferenceEndpoint: %w", err)
	}
	if err == nil {
		if !metav1.IsControlledBy(live, runner) {
			return fmt.Errorf("InferenceEndpoint %s/%s is not controlled by runner %s", live.GetNamespace(), live.GetName(), runner.Name)
		}
		if !live.GetDeletionTimestamp().IsZero() {
			return fmt.Errorf("InferenceEndpoint %s/%s is terminating", live.GetNamespace(), live.GetName())
		}
		if apiequality.Semantic.DeepEqual(live.Object["spec"], desired.Object["spec"]) {
			return nil
		}
		live.Object["spec"] = desired.Object["spec"]
		// Update retains resourceVersion, so a concurrent change cannot be overwritten.
		// Strict validation rejects fields the CRD would otherwise prune and cause us to reapply forever.
		if err := k8sClient.Update(ctx, live, client.FieldValidation(metav1.FieldValidationStrict)); err != nil {
			return fmt.Errorf("update InferenceEndpoint: %w", err)
		}
		return nil
	}
	if err := controllerutil.SetControllerReference(runner, desired, k8sClient.Scheme()); err != nil {
		return fmt.Errorf("own InferenceEndpoint: %w", err)
	}
	if err := k8sClient.Create(ctx, desired, client.FieldValidation(metav1.FieldValidationStrict)); err != nil {
		// The Get above reads the informer cache, which can lag a Create that already
		// succeeded. The Owns watch delivers the object and the next reconcile converges.
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return fmt.Errorf("create InferenceEndpoint: %w", err)
	}
	return nil
}

// deleteInferenceEndpoint removes the runner's InferenceEndpoint when no endpoint
// requests inference routing. Only an endpoint controlled by this runner is deleted.
func (p *Plugin) deleteInferenceEndpoint(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	k8sClient client.Client,
) error {
	if err := p.inferenceEndpointAvailable(); err != nil {
		if meta.IsNoMatchError(err) {
			return nil
		}
		return err
	}
	name, err := endpoints.InferenceEndpointName(runner)
	if err != nil {
		// A runner without a saved identity cannot own an InferenceEndpoint.
		return nil //nolint:nilerr // nothing to delete
	}
	live := &unstructured.Unstructured{}
	live.SetGroupVersionKind(endpoints.InferenceEndpointGVK)
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: runner.Namespace, Name: name}, live); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("read InferenceEndpoint: %w", err)
	}
	if !metav1.IsControlledBy(live, runner) {
		return nil
	}
	// Preconditions protect against a delete/recreate or ownership change between
	// the read and delete. Never delete a newly substituted object on a stale read.
	uid, rv := live.GetUID(), live.GetResourceVersion()
	if err := k8sClient.Delete(ctx, live, client.Preconditions{UID: &uid, ResourceVersion: &rv}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete InferenceEndpoint: %w", err)
	}
	return nil
}

// inferenceEndpointProbe is the readiness of the runner's InferenceEndpoint.
type inferenceEndpointProbe struct {
	name      string
	phase     dromosv1alpha1.PluginPhase
	condition metav1.Condition
	err       error
}

// probeInferenceEndpoint mirrors the InferenceEndpoint Ready condition. Ready
// counts only for an endpoint this runner controls, with the desired spec, at
// its current generation.
func (p *Plugin) probeInferenceEndpoint(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	endpoint *endpoints.Endpoint,
	k8sClient client.Client,
) inferenceEndpointProbe {
	result := inferenceEndpointProbe{}
	set := func(phase dromosv1alpha1.PluginPhase, status metav1.ConditionStatus, reason, message string) inferenceEndpointProbe {
		result.phase = phase
		result.condition = metav1.Condition{
			Type: endpoints.InferenceEndpointReadyCondition, Status: status,
			Reason: reason, Message: message, ObservedGeneration: runner.Generation,
		}
		return result
	}
	// Only API read failures are returned as errors; invalid input cannot be fixed by a retry.
	fail := func(err error, retry bool) inferenceEndpointProbe {
		if retry {
			result.err = err
		}
		return set(dromosv1alpha1.PluginPhaseFailed, metav1.ConditionFalse, "ReconcileFailed", err.Error())
	}
	name, err := endpoints.InferenceEndpointName(runner)
	if err != nil {
		return fail(err, false)
	}
	result.name = name
	if err := p.inferenceEndpointAvailable(); err != nil {
		return fail(err, false)
	}
	if err := validateInferenceEndpointShape(endpoint); err != nil {
		return fail(err, false)
	}
	workload, err := resolveInferenceWorkload(runner, endpoint)
	if err != nil {
		return fail(err, false)
	}
	target, err := inferenceEndpointTarget(endpoint, workload)
	if err != nil {
		return fail(err, false)
	}
	desired, err := endpoints.BuildInferenceEndpoint(runner, endpoint.ExternalAccess.Inference, target)
	if err != nil {
		return fail(err, false)
	}
	live := &unstructured.Unstructured{}
	live.SetGroupVersionKind(endpoints.InferenceEndpointGVK)
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(desired), live); err != nil {
		if apierrors.IsNotFound(err) {
			return set(dromosv1alpha1.PluginPhaseCreating, metav1.ConditionFalse, "Pending", "waiting for InferenceEndpoint creation")
		}
		return fail(fmt.Errorf("read InferenceEndpoint: %w", err), true)
	}
	if !metav1.IsControlledBy(live, runner) {
		return fail(fmt.Errorf("InferenceEndpoint is not controlled by this runner"), false)
	}
	if !live.GetDeletionTimestamp().IsZero() || !apiequality.Semantic.DeepEqual(live.Object["spec"], desired.Object["spec"]) {
		return set(dromosv1alpha1.PluginPhaseInitializing, metav1.ConditionFalse, "Pending", "waiting for desired InferenceEndpoint spec")
	}
	var observed struct {
		ObservedGeneration int64              `json:"observedGeneration"`
		Conditions         []metav1.Condition `json:"conditions"`
	}
	if status, found, err := unstructured.NestedMap(live.Object, "status"); err != nil {
		return fail(fmt.Errorf("read InferenceEndpoint status: %w", err), false)
	} else if found {
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(status, &observed); err != nil {
			return fail(fmt.Errorf("decode InferenceEndpoint status: %w", err), false)
		}
	}
	ready := meta.FindStatusCondition(observed.Conditions, "Ready")
	if ready == nil || observed.ObservedGeneration != live.GetGeneration() || ready.ObservedGeneration != live.GetGeneration() {
		return set(dromosv1alpha1.PluginPhaseInitializing, metav1.ConditionFalse, "Pending", "waiting for current InferenceEndpoint generation")
	}
	if ready.Status == metav1.ConditionTrue {
		return set(dromosv1alpha1.PluginPhaseReady, metav1.ConditionTrue, "Ready", ready.Message)
	}
	return set(dromosv1alpha1.PluginPhaseInitializing, metav1.ConditionFalse, "EndpointNotReady", ready.Message)
}

// inferenceEndpointURLs returns the URLs the InferenceEndpoint's route serves: the
// hostnames of the composed HTTPRoute under the endpoint's path prefix. The composed
// route itself matches "/", so its own paths would hide the prefix.
func inferenceEndpointURLs(httpRoute *gatewayv1.HTTPRoute, pathPrefix string) []string {
	if httpRoute == nil {
		return nil
	}
	urls := make([]string, 0, len(httpRoute.Spec.Hostnames))
	for _, hostname := range httpRoute.Spec.Hostnames {
		urls = append(urls, fmt.Sprintf("https://%s%s", hostname, normalizeURLPath(pathPrefix)))
	}
	return urls
}

// inferenceExternalAccessProbe reports the InferenceEndpoint in place of the
// composed HTTPRoute. Hostnames come from the composed httpRoute, which the
// InferenceEndpoint's own route serves under pathPrefix.
func inferenceExternalAccessProbe(probe inferenceEndpointProbe, httpRoute *gatewayv1.HTTPRoute, pathPrefix string) externalAccessProbeResult {
	reason := InferenceEndpointNotReadyReason
	switch probe.phase {
	case dromosv1alpha1.PluginPhaseReady:
		reason = ExternalAccessReadyConditionName
	case dromosv1alpha1.PluginPhaseCreating:
		reason = ExternalAccessReasonNotFound
	}
	message := fmt.Sprintf("inferenceendpoint %s: %s", probe.name, probe.condition.Message)
	conditionMessage := fmt.Sprintf("InferenceEndpoint %s: %s", probe.name, probe.condition.Message)
	var urls []endpoints.ServiceURLStatus
	if probe.phase == dromosv1alpha1.PluginPhaseReady {
		message = fmt.Sprintf("inferenceendpoint %s is ready", probe.name)
		conditionMessage = fmt.Sprintf("InferenceEndpoint %s is ready", probe.name)
		routeURLs := inferenceEndpointURLs(httpRoute, pathPrefix)
		if len(routeURLs) > 0 {
			message = fmt.Sprintf("%s, url: %s", message, routeURLs[0])
			conditionMessage = fmt.Sprintf("%s, url: %s", conditionMessage, routeURLs[0])
		}
		urls = serviceURLs(endpoints.ExternalAccessKindInferenceEndpoint, probe.name, routeURLs)
	}
	result := endpointPhaseResult{phase: probe.phase, reason: reason, message: message}
	resource := endpoints.ExternalAccessResourceStatus{
		Kind:    endpoints.ExternalAccessKindInferenceEndpoint,
		Name:    probe.name,
		Phase:   detailsPhaseFromPluginPhase(probe.phase),
		Reason:  reason,
		Message: message,
		Hosts:   httpRouteHosts(httpRoute),
	}
	return externalAccessProbeResult{
		result:           result,
		conditionStatus:  conditionStatusFromResult(result, metav1.ConditionFalse),
		conditionMessage: conditionMessage,
		status: &endpoints.ExternalAccessStatus{
			Phase:     resource.Phase,
			Reason:    resource.Reason,
			Message:   resource.Message,
			Resources: []endpoints.ExternalAccessResourceStatus{resource},
		},
		serviceURLs: urls,
		err:         probe.err,
	}
}

func hasInferenceEndpointCondition(runner *dromosv1alpha1.WorkloadRunner) bool {
	if runner == nil || runner.Status.Plugins[PluginName] == nil {
		return false
	}
	return meta.FindStatusCondition(
		runner.Status.Plugins[PluginName].Conditions, endpoints.InferenceEndpointReadyCondition,
	) != nil
}

// setInferenceEndpointDisabled marks a previously reported InferenceEndpointReady
// condition as disabled once no endpoint requests inference routing.
func setInferenceEndpointDisabled(runner *dromosv1alpha1.WorkloadRunner, status framework.PluginStatusSetter) {
	if !hasInferenceEndpointCondition(runner) {
		return
	}
	status.SetCondition(metav1.Condition{
		Type: endpoints.InferenceEndpointReadyCondition, Status: metav1.ConditionFalse,
		Reason: "Disabled", Message: "inference routing is not configured", ObservedGeneration: runner.Generation,
	})
}
