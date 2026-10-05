/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package inferenceendpoint

import (
	"context"
	"encoding/json"
	ep "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (p *Plugin) Status(ctx context.Context, runner *dromos.WorkloadRunner, _ client.Client, status framework.PluginStatusSetter, state *reconcile.State) error {
	status.SetAdditionalInfo(nil)
	set := func(phase dromos.PluginPhase, condition metav1.ConditionStatus, reason, message string) error {
		status.SetPhase(phase)
		status.SetCondition(metav1.Condition{Type: ReadyCondition, Status: condition, Reason: reason, Message: message, ObservedGeneration: runner.Generation})
		return nil
	}
	fail := func(message string) error {
		return set(dromos.PluginPhaseFailed, metav1.ConditionFalse, "ReconcileFailed", message)
	}
	for _, i := range state.PluginIssues(p) {
		if i.IsCritical {
			return fail(i.Error())
		}
	}
	cfg, err := ep.GetInferenceEndpoint(runner)
	if err != nil {
		return fail(err.Error())
	}
	if cfg == nil {
		return set(dromos.PluginPhaseDisabled, metav1.ConditionFalse, "Disabled", "inference routing is not configured")
	}
	if err := p.available(); err != nil {
		return fail(err.Error())
	}
	desired, err := ep.BuildInferenceEndpoint(runner, cfg)
	if err != nil {
		return fail(err.Error())
	}
	live := &unstructured.Unstructured{}
	live.SetGroupVersionKind(ep.InferenceEndpointGVK)
	if err := p.reader.Get(ctx, client.ObjectKeyFromObject(desired), live); err != nil {
		if apierrors.IsNotFound(err) {
			return set(dromos.PluginPhasePending, metav1.ConditionFalse, "Pending", "waiting for InferenceEndpoint creation")
		}
		return fail(err.Error())
	}
	if !metav1.IsControlledBy(live, runner) {
		return fail("InferenceEndpoint is not controlled by this runner")
	}
	if !live.GetDeletionTimestamp().IsZero() || !apiequality.Semantic.DeepEqual(live.Object["spec"], desired.Object["spec"]) {
		return set(dromos.PluginPhasePending, metav1.ConditionFalse, "Pending", "waiting for desired InferenceEndpoint spec")
	}
	hosts, _, err := unstructured.NestedStringSlice(live.Object, "spec", "routing", "hostnames")
	if err != nil {
		return fail(err.Error())
	}
	details, err := json.Marshal(struct {
		Name      string   `json:"name"`
		Namespace string   `json:"namespace"`
		Hostnames []string `json:"hostnames"`
	}{live.GetName(), live.GetNamespace(), hosts})
	if err != nil {
		return fail(err.Error())
	}
	status.SetAdditionalInfo(&runtime.RawExtension{Raw: details})
	var observed struct {
		ObservedGeneration int64              `json:"observedGeneration"`
		Conditions         []metav1.Condition `json:"conditions"`
	}
	observedMap, found, err := unstructured.NestedMap(live.Object, "status")
	if err != nil {
		return fail(err.Error())
	}
	if found {
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(observedMap, &observed); err != nil {
			return fail(err.Error())
		}
	}
	ready := meta.FindStatusCondition(observed.Conditions, "Ready")
	if ready == nil || observed.ObservedGeneration != live.GetGeneration() || ready.ObservedGeneration != live.GetGeneration() {
		return set(dromos.PluginPhasePending, metav1.ConditionFalse, "Pending", "waiting for current InferenceEndpoint generation")
	}
	if ready.Status == metav1.ConditionTrue {
		return set(dromos.PluginPhaseReady, metav1.ConditionTrue, "Ready", ready.Message)
	}
	return set(dromos.PluginPhasePending, metav1.ConditionFalse, "EndpointNotReady", ready.Message)
}
