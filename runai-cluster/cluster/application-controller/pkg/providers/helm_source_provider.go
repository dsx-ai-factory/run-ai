/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/consts"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller/pkg/resolver"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HelmSecretType              = "helm.sh/release.v1"
	ResourceManagedByLabelName  = "app.kubernetes.io/managed-by"
	ResourceManagedByLabelValue = "Helm"
	ApplicationHelmSource       = "helm"
)

type HelmSourceProvider struct {
	resolver resolver.ObjectResolver
}

func NewHelmSourceProvider() *HelmSourceProvider {
	return &HelmSourceProvider{
		resolver: resolver.NewDefaultObjectResolver(),
	}
}

// SetObjectResolver allows injecting a custom resolver implementation (useful for testing)
func (h *HelmSourceProvider) SetObjectResolver(res resolver.ObjectResolver) {
	h.resolver = res
}

func (h *HelmSourceProvider) GetApplicationSource() string {
	return ApplicationHelmSource
}

func (h *HelmSourceProvider) GetSourceResource() ctrlClient.Object {
	return &corev1.Secret{}
}

func (h *HelmSourceProvider) FilterSourceResource(ctx context.Context, client ctrlClient.Client, sourceResource ctrlClient.Object) bool {
	secret, ok := sourceResource.(*corev1.Secret)
	if !ok || secret == nil {
		return false
	}

	// Check if it's a Helm secret
	if secret.Type != HelmSecretType {
		return false
	}

	// Check if in application namespace
	return common.IsApplicationNamespace(ctx, client, secret.Namespace)
}

func (h *HelmSourceProvider) ShouldCreateApplication(ctx context.Context, client ctrlClient.Client, sourceResource ctrlClient.Object) (bool, string, error) {
	secret, ok := sourceResource.(*corev1.Secret)
	if !ok {
		return false, "", nil
	}

	// Check if it's a Helm secret
	if secret.Type != HelmSecretType {
		return false, "", nil
	}

	// Check if in application namespace
	if !common.IsApplicationNamespace(ctx, client, secret.Namespace) {
		return false, "", nil
	}

	// Get application name from secret labels
	applicationName, ok := secret.Labels[consts.SecretNameLabel]
	if !ok || applicationName == "" {
		return false, "", nil
	}

	return true, applicationName, nil
}

func (h *HelmSourceProvider) CreateApplicationMetadata(ctx context.Context, client ctrlClient.Client, sourceResource ctrlClient.Object) (common.ApplicationMetadata, error) {
	secret, ok := sourceResource.(*corev1.Secret)
	if !ok {
		return common.ApplicationMetadata{}, fmt.Errorf("expected Secret, got %T", sourceResource)
	}

	applicationName, ok := secret.Labels[consts.SecretNameLabel]
	if !ok || applicationName == "" {
		return common.ApplicationMetadata{}, fmt.Errorf("secret missing name label")
	}

	// Get chart name using Helm SDK
	chartName, err := h.resolver.GetSecretChartName(applicationName, secret)
	if err != nil {
		return common.ApplicationMetadata{}, fmt.Errorf("failed to get chart name: %w", err)
	}

	// Get owner reference
	ownerRef, err := h.resolver.CreateObjectOwnerReference(client, secret)
	if err != nil {
		return common.ApplicationMetadata{}, fmt.Errorf("failed to get owner reference: %w", err)
	}

	// Get project labels
	projectLabels := map[string]string{}
	project, err := h.resolver.GetProject(ctx, client, secret)
	if err != nil {
		return common.ApplicationMetadata{}, fmt.Errorf("failed to get project: %w", err)
	}
	projectLabels = map[string]string{
		constants.ProjectNameLabel:    project.Name,
		constants.ProjectIdLabel:      project.Labels[v2.ProjectIdLabel],
		constants.DepartmentIdLabel:   project.Labels[v2alpha2.DepartmentIdLabel],
		constants.DepartmentNameLabel: project.Spec.Parent,
	}

	return common.ApplicationMetadata{
		Name:            applicationName,
		Namespace:       secret.Namespace,
		Type:            chartName,
		Source:          ApplicationHelmSource,
		Labels:          projectLabels,
		OwnerReferences: []metav1.OwnerReference{*ownerRef},
	}, nil
}

func (h *HelmSourceProvider) GetApplicationPhase(sourceResource ctrlClient.Object) string {
	secret, ok := sourceResource.(*corev1.Secret)
	if !ok {
		return consts.UnknownApplicationPhase
	}

	phase, ok := secret.Labels[consts.SecretPhaseLabel]
	if !ok {
		return consts.UnknownApplicationPhase
	}
	return phase
}

func (h *HelmSourceProvider) GetListOptions(namespace string) *ctrlClient.ListOptions {
	labelSelector := labels.Set{
		ResourceManagedByLabelName: ResourceManagedByLabelValue,
	}
	return &ctrlClient.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(labelSelector),
	}
}

func (h *HelmSourceProvider) BelongsTo(resource ctrlClient.Object, applicationName string) bool {
	if !h.MatchesSource(resource) {
		return false
	}

	annotations := resource.GetAnnotations()
	appName, exists := annotations[constants.ResourceReleaseAnnotationName]
	return exists && appName == applicationName
}

func (h *HelmSourceProvider) MatchesSource(resource ctrlClient.Object) bool {
	oLabels := resource.GetLabels()
	managedLabel, labelExists := oLabels[ResourceManagedByLabelName]
	if !labelExists || managedLabel != ResourceManagedByLabelValue {
		return false
	}

	annotations := resource.GetAnnotations()
	applicationAnnotation, annotationExists := annotations[constants.ResourceReleaseAnnotationName]
	return annotationExists && applicationAnnotation != ""
}

func (h *HelmSourceProvider) GetApplicationName(resource ctrlClient.Object) string {
	annotations := resource.GetAnnotations()
	return annotations[constants.ResourceReleaseAnnotationName]
}
