/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"
	"os"
	"time"

	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
)

type ExtractAirgapPackage struct {
	*step.RunAIInstallStepImpl
	extractedDir string
}

func NewExtractAirgapPackage() *ExtractAirgapPackage {
	return &ExtractAirgapPackage{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Extract airgap package", 60*time.Minute),
	}
}

func (e *ExtractAirgapPackage) GetExtractedDir() string {
	return e.extractedDir
}

func (e *ExtractAirgapPackage) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	return e.GetCachedStatus()
}

func (e *ExtractAirgapPackage) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := e.CheckContextCancellation(ctx); err != nil {
		return err
	}
	e.SetStatus(model.Installing, nil)

	tarballPath := runaiData.GetInputValue(configinputs.InstallInputAirgapPackagePath)
	logging.Infof("[Airgap] Extracting package %s...", tarballPath)

	if _, err := os.Stat(tarballPath); err != nil {
		if os.IsNotExist(err) {
			logging.Errorf("[Airgap] Package not found: %s", tarballPath)
			e.SetStatus(model.Failed, fmt.Errorf("airgap package not found: %s", tarballPath))
		} else {
			logging.Errorf("[Airgap] Failed to stat package %s: %v", tarballPath, err)
			e.SetStatus(model.Failed, fmt.Errorf("failed to access airgap package %s: %v", tarballPath, err))
		}
		return e.GetError()
	}

	extractDir, err := os.MkdirTemp("", "runai-airgap-*")
	if err != nil {
		logging.Errorf("[Airgap] Failed to create temp dir: %v", err)
		e.SetStatus(model.Failed, fmt.Errorf("failed to create temp dir: %v", err))
		return e.GetError()
	}
	cleanupExtractDir := true
	defer func() {
		if cleanupExtractDir {
			_ = os.RemoveAll(extractDir)
		}
	}()

	if err := airgap.ExtractTar(ctx, tarballPath, extractDir); err != nil {
		logging.Errorf("[Airgap] Failed to extract package: %v", err)
		e.SetStatus(model.Failed, fmt.Errorf("failed to extract airgap package: %v", err))
		return e.GetError()
	}

	e.extractedDir = extractDir
	runaiData.Config.AirgapExtractedDir = extractDir
	cleanupExtractDir = false
	logging.Infof("[Airgap] Package extracted to %s", extractDir)
	e.SetStatus(model.Installed, nil)
	return nil
}
