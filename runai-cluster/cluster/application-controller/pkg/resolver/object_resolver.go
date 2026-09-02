/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resolver

import (
	"context"
	"fmt"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// ObjectResolver provides helper methods for resolving application metadata from source resources.
// This interface allows for mocking in tests and can be used across different source providers.
type ObjectResolver interface {
	GetSecretChartName(releaseName string, obj ctrlClient.Object) (string, error)
	GetProject(ctx context.Context, client ctrlClient.Client, obj ctrlClient.Object) (*kaiorg.Project, error)
	CreateObjectOwnerReference(client ctrlClient.Client, obj ctrlClient.Object) (*metav1.OwnerReference, error)
}

// DefaultObjectResolver is the default production implementation of ObjectResolver.
// It uses Helm SDK for chart name extraction and Kubernetes API for project and owner reference resolution.
// This implementation can be used by any source provider that works with Helm-based applications.
type DefaultObjectResolver struct{}

func NewDefaultObjectResolver() *DefaultObjectResolver {
	return &DefaultObjectResolver{}
}

func (d *DefaultObjectResolver) GetSecretChartName(releaseName string, obj ctrlClient.Object) (string, error) {
	settings := cli.New()

	actionConfig := new(action.Configuration)
	err := actionConfig.Init(settings.RESTClientGetter(), obj.GetNamespace(), "secret", log.Printf)
	if err != nil {
		return "", err
	}

	get := action.NewGet(actionConfig)
	rel, err := get.Run(releaseName)
	if err != nil {
		return "", err
	}

	return rel.Chart.Metadata.Name, nil
}

func (d *DefaultObjectResolver) GetProject(ctx context.Context, client ctrlClient.Client, obj ctrlClient.Object) (*kaiorg.Project, error) {
	namespace := corev1.Namespace{}
	err := client.Get(ctx, ctrlClient.ObjectKey{Name: obj.GetNamespace()}, &namespace)
	if err != nil {
		return nil, err
	}

	projectName, hasRunaiQueueLabel := namespace.Labels[constants.RunaiQueueLabel]
	if !hasRunaiQueueLabel || projectName == "" {
		return nil, fmt.Errorf("namespace does not have queue label")
	}

	kaiProject := &kaiorg.Project{}
	err = client.Get(ctx, types.NamespacedName{Name: projectName}, kaiProject)
	if err != nil {
		return nil, fmt.Errorf("failed to get kai project %s: %w", projectName, err)
	}
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:   kaiProject.Name,
			Labels: kaiProject.Labels,
		},
		Spec: kaiorg.ProjectSpec{
			Parent: kaiProject.Spec.Parent,
		},
	}, nil
}

func (d *DefaultObjectResolver) CreateObjectOwnerReference(client ctrlClient.Client, obj ctrlClient.Object) (*metav1.OwnerReference, error) {
	gvk, err := apiutil.GVKForObject(obj, client.Scheme())
	if err != nil {
		return nil, err
	}
	return &metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       obj.GetName(),
		UID:        obj.GetUID(),
	}, nil
}
