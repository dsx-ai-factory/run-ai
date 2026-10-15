/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package topownerlabel stamps the Run:ai workload ID on attached resources.
// Pod templates are stamped at creation and preserved on later reconciliations.
package topownerlabel

import (
	"context"
	"errors"
	"fmt"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PluginName sorts before MNNVL so identity is stamped before affinity consumes it.
const PluginName = "label-top-owner"

// Keep the legacy label while MNNVL affinity and older pod consumers use it.
//
//nolint:staticcheck // SA1019: compatibility with the existing grouping selector
const legacyTopOwnerUIDLabel = commonconstants.TopOwnerUidLabel

// podTemplatesLabeledAnnotation distinguishes resources created with ownership
// labels in their pod templates from resources that predate this plugin. It
// prevents an upgrade from changing immutable templates such as batch/v1 Jobs.
const podTemplatesLabeledAnnotation = "run.ai/top-owner-uid-template-labels"

type Plugin struct {
	reader client.Reader
}

func NewPlugin(reader client.Reader) *Plugin {
	return &Plugin{reader: reader}
}

func New(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	if deps.APIReader == nil {
		return nil, fmt.Errorf("top-owner-label plugin requires a non-nil APIReader in PluginDeps")
	}
	return NewPlugin(deps.APIReader), nil
}

func (p *Plugin) Name() string {
	return PluginName
}

// MutateResource mirrors workload-controller ownership labeling without
// changing immutable pod templates on resources that predate this plugin.
func (p *Plugin) MutateResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	state *reconcile.State,
) error {
	workloadID, stampPodTemplates, err := p.workloadIdentity(ctx, runner, resource)
	if err != nil {
		return state.AddPluginIssue(ctx, p, resource, err, true, "", issue.Internal)
	}
	stampLabels(resource, workloadID)
	if !stampPodTemplates {
		return nil
	}

	stampPodTemplateMarker(resource)
	if err := kartautils.MutatePodInstances(ctx, karta, resource, func(instance kartaresource.ExtractedInstance) error {
		stampInstance(instance, workloadID)
		return nil
	}); err != nil {
		return state.AddPluginIssue(ctx, p, resource, err, true, "", issue.Internal)
	}

	return nil
}

func stampLabels(resource *unstructured.Unstructured, workloadID string) {
	labels := resource.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	addTopOwnerLabel(labels, workloadID)
	resource.SetLabels(labels)
}

func addTopOwnerLabel(labels map[string]string, workloadID string) {
	labels[commonconstants.TopOwnerWorkloadIdLabel] = workloadID
	labels[legacyTopOwnerUIDLabel] = workloadID
}

// workloadIdentity reuses the ID stamped at creation so reconciliation cannot
// change immutable pod templates when a runner's CP annotation changes.
func (p *Plugin) workloadIdentity(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, resource *unstructured.Unstructured) (string, bool, error) {
	if p.reader == nil {
		return "", false, errors.New("top-owner-label plugin requires a reader")
	}
	workloadID := topowner.GetRunaiWorkloadId(runner.UID, runner.Annotations)

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(resource.GroupVersionKind())
	if err := p.reader.Get(ctx, client.ObjectKeyFromObject(resource), existing); err != nil {
		if apierrors.IsNotFound(err) {
			return workloadID, true, nil
		}
		return "", false, fmt.Errorf("get attached resource: %w", err)
	}

	if persistedID, err := topowner.GetTopOwnerWorkloadIdFromLabels(existing.GetLabels()); err == nil {
		workloadID = persistedID
	}
	return workloadID, existing.GetAnnotations()[podTemplatesLabeledAnnotation] == "true", nil
}

func stampPodTemplateMarker(resource *unstructured.Unstructured) {
	annotations := resource.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[podTemplatesLabeledAnnotation] = "true"
	resource.SetAnnotations(annotations)
}

func stampInstance(instance kartaresource.ExtractedInstance, workloadID string) {
	switch {
	case instance.PodTemplateSpec != nil:
		if instance.PodTemplateSpec.Labels == nil {
			instance.PodTemplateSpec.Labels = make(map[string]string)
		}
		addTopOwnerLabel(instance.PodTemplateSpec.Labels, workloadID)
	case instance.FragmentedPodSpec != nil:
		if instance.FragmentedPodSpec.Labels == nil {
			instance.FragmentedPodSpec.Labels = make(map[string]string)
		}
		addTopOwnerLabel(instance.FragmentedPodSpec.Labels, workloadID)
	}
}

var _ framework.ResourceMutatePlugin = &Plugin{}
