/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
)

// StripRunAICRFinalizers removes finalizers from all custom resources belonging
// to the run.ai API group before those CRDs are deleted. Without this step,
// deleting a CRD while its CRs still carry finalizers leaves the CRD stuck in
// Terminating state: Kubernetes sets deletionTimestamp on every CR but cannot
// proceed because the owning controller (already removed by helm uninstall) is
// no longer present to process the finalizer and remove it.
type StripRunAICRFinalizers struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewStripRunAICRFinalizers(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *StripRunAICRFinalizers {
	return &StripRunAICRFinalizers{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Strip RunAI CR finalizers", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *StripRunAICRFinalizers) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	logging.Infof("[Uninstall] Stripping finalizers from all Run.AI CRs")
	if err := s.k8sWrapper.StripRunAICRFinalizers(ctx); err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to strip Run.AI CR finalizers: %w", err)
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
