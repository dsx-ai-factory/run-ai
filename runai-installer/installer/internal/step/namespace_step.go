/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateNamespace struct {
	*RunAIInstallStepImpl
	namespace string
	nextStep  InstallStep
	k8sClient k8s.K8sClients
}

func NewCreateNamespace(namespace string, k8sClient k8s.K8sClients) *CreateNamespace {
	name := fmt.Sprintf("Create namespace %s", namespace)
	return &CreateNamespace{
		RunAIInstallStepImpl: NewRunAIInstallStepImpl(name),
		namespace:            namespace,
		k8sClient:            k8sClient,
	}
}

func (i *CreateNamespace) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	// Skip evaluation if in terminal state or Execute() is running
	if i.ShouldSkipStatusEvaluation() {
		return i.GetCachedStatus()
	}
	logging.Infof("[RunAI] Checking if namespace '%s' exists...", i.namespace)
	namespace := &corev1.Namespace{}
	err := i.k8sClient.GetClient().Get(context.Background(), client.ObjectKey{
		Name: i.namespace,
	}, namespace)

	if err != nil {
		if errors.IsNotFound(err) {
			logging.Infof("[RunAI] Namespace '%s' does not exist", i.namespace)
			return i.SetStatus(model.NotInstalled, nil)
		}
		logging.Errorf("[RunAI] Failed to check namespace '%s': %v", i.namespace, err)
		return i.SetStatus(model.Failed, fmt.Errorf("failed to check if namespace: %s exist: [%v]", i.namespace, err))
	}
	logging.Infof("[RunAI] Namespace '%s' exists in cluster", i.namespace)
	return i.SetStatus(model.Installed, nil)
}

func (i *CreateNamespace) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	// Check for context cancellation before starting
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	logging.Infof("[RunAI] Creating namespace '%s'...", i.namespace)
	i.SetStatus(model.Installing, nil)

	// Check for context cancellation before k8s operations
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: i.namespace,
		},
	}

	existingNamespace := &corev1.Namespace{}
	err := i.k8sClient.GetClient().Get(ctx, client.ObjectKey{
		Name: i.namespace,
	}, existingNamespace)

	if err == nil {
		logging.Infof("[RunAI] Namespace '%s' exists, skipping creation", i.namespace)
		i.SetStatus(model.Installed, nil)
		return nil
	}

	if !errors.IsNotFound(err) {
		error := fmt.Errorf("failed to check if namespace: %s exist: [%v]", i.namespace, err)
		logging.Errorf("[RunAI] failed to check if namespace: %s exist: [%v]", i.namespace, err)
		i.SetStatus(model.Failed, error)
		return error
	}

	// Check for context cancellation before creating namespace
	if err := i.CheckContextCancellation(ctx); err != nil {
		return err
	}

	err = i.k8sClient.GetClient().Create(ctx, namespace)
	if err != nil {
		error := fmt.Errorf("failed to create namespace '%s': %v", i.namespace, err)
		logging.Errorf("[RunAI] failed to create namespace '%s': %v", i.namespace, err)
		i.SetStatus(model.Failed, error)
		return error
	}
	logging.Infof("[RunAI] Namespace '%s' created successfully", i.namespace)
	i.SetStatus(model.Installed, nil)
	return nil
}
