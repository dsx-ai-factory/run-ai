/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plan

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespaceStep struct {
	*orchestrator2.BaseStep
	namespace string
}

func newNamespaceStep(name, namespace string) *namespaceStep {
	return &namespaceStep{BaseStep: orchestrator2.NewBaseStep(name), namespace: namespace}
}

func (s *namespaceStep) BuildActions(_ *spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{
		actions.K8sCreate{Object: &corev1.Namespace{
			TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
			ObjectMeta: metav1.ObjectMeta{Name: s.namespace},
		}},
	}, nil
}

type helmStep struct {
	*orchestrator2.BaseStep
	action actions.HelmInstall
}

func newHelmStep(name string, action actions.HelmInstall) *helmStep {
	return &helmStep{BaseStep: orchestrator2.NewBaseStep(name), action: action}
}

func (s *helmStep) BuildActions(_ *spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{s.action}, nil
}
