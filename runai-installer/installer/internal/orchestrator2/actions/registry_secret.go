/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	helmManagedByLabel             = "app.kubernetes.io/managed-by"
	helmReleaseNameAnnotation      = "meta.helm.sh/release-name"
	helmReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
	helmResourcePolicyAnnotation   = "helm.sh/resource-policy"
)

type RegistrySecret struct {
	SecretName string
	Namespace  string
	Token      string
}

func (a RegistrySecret) Kind() string { return "k8s.registry_secret" }

func (a RegistrySecret) Args() map[string]any {
	return map[string]any{
		"name":      a.SecretName,
		"namespace": a.Namespace,
	}
}

func (a RegistrySecret) Describe() string {
	return fmt.Sprintf("apply registry credentials secret %s/%s", a.Namespace, a.SecretName)
}

func (a RegistrySecret) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	dockerConfig, err := base64.StdEncoding.DecodeString(a.Token)
	if err != nil {
		return orchestrator2.ActionResult{}, fmt.Errorf("decode registry token: %w", err)
	}
	existing := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: a.SecretName, Namespace: a.Namespace},
	}
	err = deps.K8sExecutor.Get(ctx, client.ObjectKeyFromObject(existing), existing)
	if apierrors.IsNotFound(err) {
		existing.Type = corev1.SecretTypeDockerConfigJson
		existing.Data = map[string][]byte{corev1.DockerConfigJsonKey: dockerConfig}
		return orchestrator2.ActionResult{}, deps.K8sExecutor.Create(ctx, existing)
	}
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}

	original := existing.DeepCopy()
	retainHelmManagedSecret(existing)
	existing.Type = corev1.SecretTypeDockerConfigJson
	if existing.Data == nil {
		existing.Data = map[string][]byte{}
	}
	existing.Data[corev1.DockerConfigJsonKey] = dockerConfig

	return orchestrator2.ActionResult{}, deps.K8sExecutor.Patch(ctx, existing, client.MergeFrom(original))
}

func retainHelmManagedSecret(secret *corev1.Secret) {
	if secret.Labels[helmManagedByLabel] != "Helm" ||
		secret.Annotations[helmReleaseNameAnnotation] == "" ||
		secret.Annotations[helmReleaseNamespaceAnnotation] == "" {
		return
	}
	secret.Annotations[helmResourcePolicyAnnotation] = "keep"
}
