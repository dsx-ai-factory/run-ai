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

// runaiClusterRoleLabelPrefix is the label key prefix shared by Run:AI-owned
// ClusterRoles.  Used for client-side filtering because Kubernetes label
// selectors do not support key-prefix matching.
const runaiClusterRoleLabelPrefix = "run.ai/"

type DeleteRunAIClusterRoles struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteRunAIClusterRoles(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteRunAIClusterRoles {
	return &DeleteRunAIClusterRoles{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Delete RunAI ClusterRoles", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteRunAIClusterRoles) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	names, err := s.k8sWrapper.ListClusterRoles(runaiClusterRoleLabelPrefix)
	if err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to list ClusterRoles: %w", err)
	}

	if len(names) == 0 {
		logging.Infof("[Uninstall] No RunAI ClusterRoles found, skipping")
		s.SetStatus(model.Installed, nil)
		return nil
	}

	logging.Infof("[Uninstall] Deleting %d RunAI ClusterRoles", len(names))
	for _, name := range names {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		if err := s.k8sWrapper.DeleteClusterRole(name); err != nil {
			s.SetStatus(model.Failed, err)
			return err
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
