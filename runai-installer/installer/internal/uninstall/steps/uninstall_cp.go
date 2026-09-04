/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"errors"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
)

type UninstallCPChart struct {
	*step.RunAIInstallStepImpl
	helmWrapper uninstall.HelmWrapper
}

func NewUninstallCPChart(cfg *config.UninstallConfig, helmWrapper uninstall.HelmWrapper) *UninstallCPChart {
	return &UninstallCPChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Uninstall control plane chart", cfg.StepTimeout),
		helmWrapper:          helmWrapper,
	}
}

func (s *UninstallCPChart) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	// TODO: InstallPhase values (Installing/Installed) are install-centric but used as generic
	// in-progress/completed signals. Refactor to use neutral names (e.g. InProgress/Done).
	s.SetStatus(model.Installing, nil)

	waitTimeout := time.Duration(0)
	if deadline, ok := ctx.Deadline(); ok {
		waitTimeout = time.Until(deadline)
	}

	// A previously interrupted install may leave the release in pending state.
	// Fix it before uninstalling so Helm's pending-state protection does not block us.
	// We warn and continue rather than failing: if the fix fails the worst case is that
	// UninstallRelease below returns an error, which is handled there.
	if err := s.helmWrapper.FixPendingRelease(step.CPReleaseName, step.CPNamespace); err != nil {
		logging.Warnf("[Uninstall] could not fix pending CP release: %v", err)
	}

	logging.Infof("[Uninstall] Uninstalling control plane chart")
	if err := s.helmWrapper.UninstallRelease(step.CPReleaseName, step.CPNamespace, waitTimeout); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.SetStatus(model.Timeout, err)
		} else {
			s.SetStatus(model.Failed, err)
		}
		return err
	}

	logging.Infof("[Uninstall] Control plane chart uninstalled successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
