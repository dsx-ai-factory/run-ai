package steps

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/airgap"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

type ExtractAirgapPackage struct {
	*step.RunAIInstallStepImpl
	upgradeCfg *config.UpgradeConfig
}

func NewExtractAirgapPackage(upgradeCfg *config.UpgradeConfig) *ExtractAirgapPackage {
	return &ExtractAirgapPackage{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Extract airgap package", 60*time.Minute),
		upgradeCfg:           upgradeCfg,
	}
}

func (e *ExtractAirgapPackage) GetStatusAsync(_ *model.RunAIInstallData) model.InstallPhase {
	return e.GetCachedStatus()
}

func (e *ExtractAirgapPackage) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := e.CheckContextCancellation(ctx); err != nil {
		return err
	}
	e.SetStatus(model.Installing, nil)

	tarballPath := e.upgradeCfg.AirgapPackagePath
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

	e.upgradeCfg.AirgapExtractedDir = extractDir
	cleanupExtractDir = false
	logging.Infof("[Airgap] Package extracted to %s", extractDir)
	e.SetStatus(model.Installed, nil)
	return nil
}
