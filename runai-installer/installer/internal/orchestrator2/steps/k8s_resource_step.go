/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// k8sResourceStep is the shared implementation for steps that apply a single
// Kubernetes resource. Detect is a read-only Get of that resource: present =>
// installed. The concrete resource and apply action are supplied by the
// constructors below.
type k8sResourceStep struct {
	*orchestrator2.BaseStep
	deps    orchestrator2.Deps
	probe   client.Object
	key     client.ObjectKey
	actions []orchestrator2.PerformAction
}

func (s *k8sResourceStep) Detect(ctx context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	obj := s.probe.DeepCopyObject().(client.Object)
	err := s.deps.K8sExecutor.Get(ctx, s.key, obj)
	if err == nil {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	if apierrors.IsNotFound(err) {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}
	return orchestrator2.DetectResult{}, err
}

func (s *k8sResourceStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return s.actions, nil
}

func secretProbe() client.Object {
	return &corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"}}
}

// NewNamespace applies a namespace, detecting an existing one by Get.
func NewNamespace(name, namespace string, deps orchestrator2.Deps) orchestrator2.Step {
	return &k8sResourceStep{
		BaseStep: orchestrator2.NewBaseStep(name),
		deps:     deps,
		probe:    &corev1.Namespace{TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"}},
		key:      client.ObjectKey{Name: namespace},
		actions: []orchestrator2.PerformAction{actions.K8sApply{Object: &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}}},
	}
}

// NewRegistrySecret applies a dockerconfigjson secret, detecting an existing one by Get.
func NewRegistrySecret(name, secretName, namespace, token string, deps orchestrator2.Deps) orchestrator2.Step {
	return &k8sResourceStep{
		BaseStep: orchestrator2.NewBaseStep(name),
		deps:     deps,
		probe:    secretProbe(),
		key:      client.ObjectKey{Namespace: namespace, Name: secretName},
		actions:  []orchestrator2.PerformAction{actions.RegistrySecret{SecretName: secretName, Namespace: namespace, Token: token}},
	}
}

// NewTLSSecret applies a kubernetes.io/tls secret, detecting an existing one by Get.
func NewTLSSecret(name, secretName, namespace, certPath, keyPath string, deps orchestrator2.Deps) orchestrator2.Step {
	return &k8sResourceStep{
		BaseStep: orchestrator2.NewBaseStep(name),
		deps:     deps,
		probe:    secretProbe(),
		key:      client.ObjectKey{Namespace: namespace, Name: secretName},
		actions:  []orchestrator2.PerformAction{actions.TLSSecret{SecretName: secretName, Namespace: namespace, CertPath: certPath, KeyPath: keyPath}},
	}
}

// NewCASecret applies an Opaque CA secret, detecting an existing one by Get.
func NewCASecret(name, secretName, namespace, helmRelease, caCertPath string, labels map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	return &k8sResourceStep{
		BaseStep: orchestrator2.NewBaseStep(name),
		deps:     deps,
		probe:    secretProbe(),
		key:      client.ObjectKey{Namespace: namespace, Name: secretName},
		actions:  []orchestrator2.PerformAction{actions.CASecret{SecretName: secretName, Namespace: namespace, HelmRelease: helmRelease, CACertPath: caCertPath, Labels: labels}},
	}
}
