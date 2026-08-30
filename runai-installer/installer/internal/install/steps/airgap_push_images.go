package steps

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/airgap"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

type PushAirgapImages struct {
	*step.RunAIInstallStepImpl
}

func NewPushAirgapImages() *PushAirgapImages {
	return &PushAirgapImages{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Push airgap images to registry", 60*time.Minute),
	}
}

func (p *PushAirgapImages) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	return p.GetCachedStatus()
}

func (p *PushAirgapImages) Execute(ctx context.Context, runaiData *model.RunAIInstallData) error {
	if err := p.CheckContextCancellation(ctx); err != nil {
		return err
	}
	p.SetStatus(model.Installing, nil)

	registryURL := runaiData.GetInputValue(configinputs.InstallInputRegistryURL)
	extractedDir := runaiData.Config.AirgapExtractedDir
	username := runaiData.GetInputValue(configinputs.InstallInputRegistryUsername)
	password := runaiData.GetInputValue(configinputs.InstallInputRegistryPassword)

	if err := airgap.RunSetupScript(ctx, extractedDir, registryURL, username, password, nil); err != nil {
		p.SetStatus(model.Failed, err)
		return err
	}

	p.SetStatus(model.Installed, nil)
	return nil
}

func (p *PushAirgapImages) GetTimeoutHint() string {
	return "Image push to registry timed out. Verify the registry is accessible and has sufficient storage."
}
