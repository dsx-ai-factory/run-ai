/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
)

const namespacePollingInterval = 5 * time.Second

type DeleteNamespace struct {
	*step.RunAIInstallStepImpl
	namespace  string
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteNamespace(namespace string, cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteNamespace {
	return &DeleteNamespace{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl(fmt.Sprintf("Delete namespace %s", namespace), cfg.StepTimeout),
		namespace:            namespace,
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteNamespace) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	logging.Infof("[Uninstall] Deleting namespace '%s'", s.namespace)
	if err := s.k8sWrapper.ForceDeleteNamespace(s.namespace); err != nil {
		s.SetStatus(model.Failed, err)
		return err
	}

	logging.Infof("[Uninstall] Waiting for namespace '%s' to be fully removed", s.namespace)
	for {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		exists, err := s.k8sWrapper.NamespaceExists(s.namespace)
		if err != nil {
			s.SetStatus(model.Failed, err)
			return err
		}
		if !exists {
			break
		}
		logging.Infof("[Uninstall] Namespace '%s' still terminating, retrying in 5s", s.namespace)
		select {
		case <-ctx.Done():
			err := fmt.Errorf("timed out waiting for namespace '%s' to be removed", s.namespace)
			s.SetStatus(model.Failed, err)
			return err
		case <-time.After(namespacePollingInterval):
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
