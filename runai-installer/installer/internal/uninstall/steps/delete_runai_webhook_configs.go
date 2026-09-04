/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
)

const runaiWebhookLabelSelector = "app=runai"

// Namespaces a Run:AI webhook can route to, used as ownership evidence when
// the label is missing.
func runaiWebhookServiceNamespaces() []string {
	return []string{step.ClusterNamespace, step.CPNamespace}
}

type DeleteRunAIValidatingWebhooks struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteRunAIValidatingWebhooks(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteRunAIValidatingWebhooks {
	return &DeleteRunAIValidatingWebhooks{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Delete RunAI ValidatingWebhookConfigurations", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteRunAIValidatingWebhooks) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	names, err := s.k8sWrapper.ListValidatingWebhookConfigurations(runaiWebhookLabelSelector, runaiWebhookServiceNamespaces())
	if err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to list ValidatingWebhookConfigurations: %w", err)
	}

	if len(names) == 0 {
		logging.Infof("[Uninstall] No RunAI ValidatingWebhookConfigurations found, skipping")
		s.SetStatus(model.Installed, nil)
		return nil
	}

	logging.Infof("[Uninstall] Deleting %d RunAI ValidatingWebhookConfigurations", len(names))
	for _, name := range names {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		if err := s.k8sWrapper.DeleteValidatingWebhookConfiguration(name); err != nil {
			s.SetStatus(model.Failed, err)
			return err
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}

type DeleteRunAIMutatingWebhooks struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteRunAIMutatingWebhooks(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteRunAIMutatingWebhooks {
	return &DeleteRunAIMutatingWebhooks{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Delete RunAI MutatingWebhookConfigurations", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteRunAIMutatingWebhooks) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	names, err := s.k8sWrapper.ListMutatingWebhookConfigurations(runaiWebhookLabelSelector, runaiWebhookServiceNamespaces())
	if err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to list MutatingWebhookConfigurations: %w", err)
	}

	if len(names) == 0 {
		logging.Infof("[Uninstall] No RunAI MutatingWebhookConfigurations found, skipping")
		s.SetStatus(model.Installed, nil)
		return nil
	}

	logging.Infof("[Uninstall] Deleting %d RunAI MutatingWebhookConfigurations", len(names))
	for _, name := range names {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		if err := s.k8sWrapper.DeleteMutatingWebhookConfiguration(name); err != nil {
			s.SetStatus(model.Failed, err)
			return err
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
