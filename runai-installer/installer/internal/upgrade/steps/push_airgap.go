/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

type PushAirgapImages struct {
	*step.RunAIInstallStepImpl
	upgradeCfg *config.UpgradeConfig
}

func NewPushAirgapImages(upgradeCfg *config.UpgradeConfig) *PushAirgapImages {
	return &PushAirgapImages{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Push airgap images to registry", 60*time.Minute),
		upgradeCfg:           upgradeCfg,
	}
}

func (p *PushAirgapImages) GetStatusAsync(_ *model.RunAIInstallData) model.InstallPhase {
	return p.GetCachedStatus()
}

func (p *PushAirgapImages) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := p.CheckContextCancellation(ctx); err != nil {
		return err
	}
	p.SetStatus(model.Installing, nil)

	err := airgap.RunSetupScript(ctx,
		p.upgradeCfg.AirgapExtractedDir,
		p.upgradeCfg.RegistryURL,
		p.upgradeCfg.RegistryUsername,
		p.upgradeCfg.RegistryPassword,
		nil,
	)
	if err != nil {
		p.SetStatus(model.Failed, err)
		return err
	}

	p.SetStatus(model.Installed, nil)
	return nil
}

func (p *PushAirgapImages) GetTimeoutHint() string {
	return "Image push to registry timed out. Verify the registry is accessible and has sufficient storage."
}
