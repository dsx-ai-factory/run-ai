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

// runaiPriorityClassLabelPrefix is the label key prefix shared by Run:AI-owned
// PriorityClasses.  Used for client-side filtering because Kubernetes label
// selectors do not support key-prefix matching.
const runaiPriorityClassLabelPrefix = "run.ai/"

type DeleteRunAIPriorityClasses struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteRunAIPriorityClasses(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteRunAIPriorityClasses {
	return &DeleteRunAIPriorityClasses{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Delete RunAI PriorityClasses", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteRunAIPriorityClasses) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	names, err := s.k8sWrapper.ListPriorityClasses(runaiPriorityClassLabelPrefix)
	if err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to list PriorityClasses: %w", err)
	}

	if len(names) == 0 {
		logging.Infof("[Uninstall] No RunAI PriorityClasses found, skipping")
		s.SetStatus(model.Installed, nil)
		return nil
	}

	logging.Infof("[Uninstall] Deleting %d RunAI PriorityClasses", len(names))
	for _, name := range names {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		if err := s.k8sWrapper.DeletePriorityClass(name); err != nil {
			s.SetStatus(model.Failed, err)
			return err
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
