/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

// runaiResourceInterfaceLabelPrefix is the label key prefix shared by all
// Run:AI-owned ResourceInterface objects.  Used for client-side filtering
// because Kubernetes label selectors do not support key-prefix matching.
const runaiResourceInterfaceLabelPrefix = "run.ai/"

type DeleteRunAIResourceInterfaces struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteRunAIResourceInterfaces(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteRunAIResourceInterfaces {
	return &DeleteRunAIResourceInterfaces{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Delete RunAI ResourceInterfaces", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteRunAIResourceInterfaces) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	names, err := s.k8sWrapper.ListResourceInterfaces(runaiResourceInterfaceLabelPrefix)
	if err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to list ResourceInterfaces: %w", err)
	}

	if len(names) == 0 {
		logging.Infof("[Uninstall] No RunAI ResourceInterfaces found, skipping")
		s.SetStatus(model.Installed, nil)
		return nil
	}

	logging.Infof("[Uninstall] Deleting %d RunAI ResourceInterfaces", len(names))
	for _, name := range names {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		if err := s.k8sWrapper.DeleteResourceInterface(name); err != nil {
			s.SetStatus(model.Failed, err)
			return err
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
